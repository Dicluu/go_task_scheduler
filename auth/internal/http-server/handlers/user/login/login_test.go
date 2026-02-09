package login

import (
	"auth/internal/application/usecase"
	login "auth/internal/http-server/handlers/user/login/mocks"
	"auth/internal/lib/logger/handlers/slogdiscard"
	"auth/internal/storage"
	"auth/pkg/testhlpr"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveUserHandler(t *testing.T) {
	testTable := []struct {
		name      string
		email     string
		password  string
		emptyBody bool
		respError string
		mockErr   error
	}{
		{
			name:     "success",
			email:    "test@test.ru",
			password: "password",
		},
		{
			name:      "empty email",
			email:     "",
			password:  "password",
			respError: "field Email is a required field",
		},
		{
			name:      "empty password",
			email:     "test@test.ru",
			password:  "",
			respError: "field Password is a required field",
		},
		{
			name:      "user not found",
			email:     "test@test.ru",
			password:  "password",
			mockErr:   storage.ErrUserNotFound,
			respError: "user not found",
		},
		{
			name:      "wrong login or password",
			email:     "test@test.ru",
			password:  "password",
			mockErr:   usecase.ErrWrongLoginOrPassword,
			respError: "failed to login. Incorrect login or password",
		},
		{
			name:      "other error",
			email:     "test@test.ru",
			password:  "password",
			mockErr:   errors.New("test error"),
			respError: "internal error",
		},
		{
			name:      "empty body",
			email:     "test@test.ru",
			password:  "password",
			emptyBody: true,
			respError: "invalid request",
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			usecaseMock := login.NewMockUsecase(t)

			if tc.respError == "" || tc.mockErr != nil {
				usecaseMock.EXPECT().
					Login(context.Background(), tc.email, tc.password).
					Return(string(mock.AnythingOfType("String")), string(mock.AnythingOfType("String")), tc.mockErr).
					Once()
			}

			handler := New(slogdiscard.NewDiscardLogger(), usecaseMock)

			var input string
			if tc.emptyBody == false {
				input = fmt.Sprintf(`{"email": "%s", "password": "%s"}`, tc.email, tc.password)
			}

			body := testhlpr.SendRequest(handler, t, "/refresh", "POST", input, http.StatusOK)

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
