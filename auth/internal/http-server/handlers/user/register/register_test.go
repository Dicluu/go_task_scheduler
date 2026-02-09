package register

import (
	register "auth/internal/http-server/handlers/user/register/mocks"
	"auth/internal/lib/logger/handlers/slogdiscard"
	"auth/internal/storage"
	"auth/pkg/testhlpr"
	"context"
	"encoding/json"
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
		respError string
		mockErr   error
	}{
		{
			name:     "success",
			email:    "test@gmail.com",
			password: "12345",
		},
		{
			name:      "empty email",
			password:  "12345",
			respError: "field Email is a required field",
		},
		{
			name:      "empty password",
			email:     "test@gmail.com",
			respError: "field Password is a required field",
		},
		{
			name:      "empty email and password",
			respError: "field Email is a required field, field Password is a required field",
		},
		{
			name:      "wrong email",
			email:     "test",
			password:  "12345",
			respError: "field Email is not valid",
		},
		{
			name:      "user already exists",
			email:     "test@test.ru",
			password:  "12345",
			respError: "user already exists",
			mockErr:   storage.ErrUserExists,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			userSaverMock := register.NewMockUserSaver(t)

			if tc.respError == "" || tc.mockErr != nil {
				userSaverMock.EXPECT().
					SaveUser(context.Background(), tc.email, mock.AnythingOfType("[]uint8")).
					Return(int64(1), tc.mockErr).
					Once()
			}

			handler := New(slogdiscard.NewDiscardLogger(), userSaverMock)
			input := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, tc.email, tc.password)
			body := testhlpr.SendRequest(handler, t, "/register", "POST", input, http.StatusOK)

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
