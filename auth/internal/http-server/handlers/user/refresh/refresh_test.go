package refresh

import (
	"auth/internal/application/usecase"
	refresh "auth/internal/http-server/handlers/user/refresh/mocks"
	"auth/internal/lib/logger/handlers/slogdiscard"
	"auth/internal/storage"
	"auth/pkg/testhlpr"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSaveUserHandler(t *testing.T) {
	testTable := []struct {
		name      string
		token     string
		emptyBody bool
		respError string
		mockErr   error
	}{
		{
			name:  "success",
			token: "test_token",
		},
		{
			name:      "token not found",
			token:     "test_token",
			respError: "refresh token not found",
			mockErr:   storage.ErrRefreshTokenNotFound,
		},
		{
			name:      "used token",
			token:     "test_token",
			respError: "refresh token is unavailable",
			mockErr:   usecase.ErrTokenUsed,
		},
		{
			name:      "token expired",
			token:     "test_token",
			respError: "token expired",
			mockErr:   jwt.ErrTokenExpired,
		},
		{
			name:      "other errors",
			token:     "test_token",
			respError: "internal error",
			mockErr:   errors.New("test error"),
		},
		{
			name:      "empty body",
			token:     "",
			emptyBody: true,
			respError: "invalid request",
		},
	}

	for _, tc := range testTable {
		t.Run(tc.name, func(t *testing.T) {

			usecaseMock := refresh.NewMockUsecase(t)

			if tc.respError == "" || tc.mockErr != nil {
				usecaseMock.EXPECT().
					RefreshSession(context.Background(), string(mock.AnythingOfType("String")), tc.token).
					Return(string(mock.AnythingOfType("String")), string(mock.AnythingOfType("String")), tc.mockErr).
					Once()
			}

			handler := New(slogdiscard.NewDiscardLogger(), usecaseMock, string(mock.AnythingOfType("String")))

			var input string
			if tc.emptyBody == false {
				input = fmt.Sprintf(`{"token": "%s"}`, tc.token)
			}

			body := testhlpr.SendRequest(handler, t, "/refresh", "POST", input, http.StatusOK)

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
