package login

import (
	"auth/internal/application/usecase"
	"auth/internal/application/usecase/loginuc"
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
)

type Usecase interface {
	Login(ctx context.Context, email, password string) (string, string, error)
}

type Request struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(log *slog.Logger, usecase Usecase) http.HandlerFunc {
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

		token, refreshToken, err := usecase.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			handleError(w, r, err)

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

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, storage.ErrUserNotFound):
		render.JSON(w, r, resp.Error("user not found"))
	case errors.Is(err, usecase.ErrWrongLoginOrPassword):
		render.JSON(w, r, resp.Error("failed to login. Incorrect login or password"))
	default:
		render.JSON(w, r, resp.Error("internal error"))
	}
}

func NewUsecase(log *slog.Logger, ugetter loginuc.UserGetter, rtsaver loginuc.RefreshTokenSaver, secret string, tokenTTL, rtTTL time.Duration) *loginuc.LoginUseCase {
	return &loginuc.LoginUseCase{
		Log:      log,
		Ugetter:  ugetter,
		RtSaver:  rtsaver,
		Secret:   secret,
		TokenTTL: tokenTTL,
		RtTTL:    rtTTL,
	}
}
