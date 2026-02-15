package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"task/internal/lib/logger/sl"
	resp "task/pkg/api/resp"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

type authClaims struct {
	jwt.RegisteredClaims
	UserId int64  `json:"uid"`
	Type   string `json:"type"`
}

func Middleware(log *slog.Logger, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				log.Warn("authorization header was not provided")

				render.JSON(w, r, resp.Error("user is unauthorized"))

				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				log.Warn("bearer prefix not provided to token", slog.String("token", authHeader))

				render.JSON(w, r, resp.Error("invalid request"))

				return
			}

			stringToken := strings.TrimPrefix(authHeader, "Bearer ")

			var claims authClaims
			_, err := jwt.ParseWithClaims(stringToken, &claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			}, jwt.WithExpirationRequired())
			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					log.Info("token is expired", slog.String("token", stringToken))

					render.JSON(w, r, resp.Error("token is expired"))

					return
				}

				if errors.Is(err, jwt.ErrSignatureInvalid) {
					log.Warn("token signature is invalid", slog.String("token", stringToken))

					render.JSON(w, r, resp.Error("internal error"))

					return
				}

				log.Error("failed to parse token", slog.String("token", stringToken), sl.Err(err))

				render.JSON(w, r, resp.Error("invalid token provided"))

				return
			}

			if claims.Type != "access" {
				log.Warn("not access token was provided", slog.String("token", stringToken))

				render.JSON(w, r, resp.Error("wrong token's type"))

				return
			}

			// TODO: make user as domain, not int value
			// TODO: move key to const
			ctx := context.WithValue(r.Context(), "user", claims.UserId)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
