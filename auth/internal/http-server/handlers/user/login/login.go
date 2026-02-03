package login

import (
	"auth/internal/domain/models"
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
	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
)

type Request struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserGetter interface {
	User(ctx context.Context, email string) (*models.User, error)
}

type RefreshTokenSaver interface {
	SaveRefreshToken(ctx context.Context, rt models.RefreshToken) error
}

func New(log *slog.Logger, getter UserGetter, rts RefreshTokenSaver, secret string, tokenTTL time.Duration, rtTTL time.Duration) http.HandlerFunc {
	const op = "http-server.handlers.user.login.New"
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

		err = validator.New().Struct(req)
		if err != nil {
			log.Info("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(err.(validator.ValidationErrors)))

			return
		}

		user, err := getter.User(r.Context(), req.Email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				log.Info("user not found", sl.Err(err))

				render.JSON(w, r, resp.Error("user not found"))

				return
			}

			log.Error("error when getting user", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(req.Password))
		if err != nil {
			log.Info("failed to login", slog.Int64("user_id", user.ID))

			render.JSON(w, r, resp.Error("failed to login. Incorrect login or password"))

			return
		}

		token, err := jwtgen.NewToken(user, secret, tokenTTL)
		if err != nil {
			log.Error("failed to generate JWT token", sl.Err(err))

			resp.Error("internal error")

			return
		}

		refreshToken, err := jwtgen.NewRefreshToken(user, secret, rtTTL)
		if err != nil {
			log.Error("failed to generate JWT token", sl.Err(err))

			resp.Error("internal error")

			return
		}

		err = rts.SaveRefreshToken(r.Context(), models.RefreshToken{Token: refreshToken, UserID: user.ID})
		if err != nil {
			if errors.Is(err, storage.ErrRefreshTokenExists) {
				log.Error("failed to generate unique JWT", sl.Err(err))

				render.JSON(w, r, resp.Error("internal error"))

				return
			}

			log.Error("failed to save refresh token", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		render.JSON(w, r, &Response{
			Response:     resp.OK(),
			Token:        token,
			RefreshToken: refreshToken,
		})

		return
	}
}
