package refresh

import (
	"auth/internal/domain/models"
	"auth/internal/lib/jwt"
	"auth/internal/lib/jwt/jwtgen"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	resp "auth/pkg/api/resp"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

type Request struct {
	RefreshToken string `json:"token" validate:"required"`
}

type Response struct {
	resp.Response
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenStorage interface {
	RefreshToken(ctx context.Context, token string) (models.RefreshToken, error)
	SaveRefreshToken(ctx context.Context, rt models.RefreshToken) error
}

type UserStorage interface {
	UserById(ctx context.Context, id int64) (*models.User, error)
}

func New(log *slog.Logger, rtStorage RefreshTokenStorage, uStorage UserStorage, secret string, tokenTTL, rtTTL time.Duration) http.HandlerFunc {
	const op = "http-server.handlers.user.refresh.New"
	return func(w http.ResponseWriter, r *http.Request) {
		log := log.With(slog.String("op", op), slog.String("req-id", middleware.GetReqID(r.Context())))

		if r.ContentLength == 0 {
			log.Warn("missing request body")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		token, err := rtStorage.RefreshToken(r.Context(), req.RefreshToken)
		if err != nil {
			if errors.Is(err, storage.ErrRefreshTokenNotFound) {
				log.Warn("refresh token not found", sl.Err(err))

				render.JSON(w, r, resp.Error("internal error"))

				return
			}

			log.Error("failed to refresh token", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		if token.Used {
			log.Info("token is used", slog.Int64("token_id", token.ID))

			render.JSON(w, r, resp.Error("token is unavailable"))

			return
		}

		_, err = jwt.Parse(token.Token, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithExpirationRequired())
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				log.Info("token expired", slog.Int64("token_id", token.ID))

				render.JSON(w, r, resp.Error("token expired"))

				return
			}
		}

		// TODO: deactivate if used

		user, err := uStorage.UserById(r.Context(), token.UserID)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Error("user not found", slog.Int64("token_id", token.ID))

				render.JSON(w, r, resp.Error("user not found"))

				return
			}

			log.Error("errors occurred")

			render.JSON(w, r, "internal error")

			return
		}

		jwtToken, err := jwtgen.NewToken(user, secret, tokenTTL)
		if err != nil {
			log.Error("failed to generate access token", sl.Err(err))

			render.JSON(w, r, "internal error")

			return
		}

		refreshToken, err := jwtgen.NewRefreshToken(user, secret, rtTTL)
		if err != nil {
			log.Error("failed to generate refresh token", sl.Err(err))

			render.JSON(w, r, "internal error")

			return
		}

		render.JSON(w, r, &Response{
			Response:     resp.OK(),
			Token:        jwtToken,
			RefreshToken: refreshToken,
		})

	}
}
