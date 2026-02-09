package refresh

import (
	"auth/internal/application/usecase"
	"auth/internal/application/usecase/tokenrefreshuc"
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

type Usecase interface {
	RefreshSession(ctx context.Context, secret string, refreshToken string) (string, string, error)
}

type Request struct {
	RefreshToken string `json:"token" validate:"required"`
}

type Response struct {
	resp.Response
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(log *slog.Logger, usecase Usecase, secret string) http.HandlerFunc {
	const op = "http-server.handlers.user.refresh.New"
	return func(w http.ResponseWriter, r *http.Request) {
		hlog := log.With(slog.String("op", op), slog.String("req-id", middleware.GetReqID(r.Context())))

		if r.ContentLength == 0 {
			hlog.Warn("missing request body")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			hlog.Error("failed to decode request", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		jwtToken, refreshToken, err := usecase.RefreshSession(r.Context(), secret, req.RefreshToken)
		if err != nil {
			handleError(w, r, err)

			return
		}

		render.JSON(w, r, &Response{
			Response:     resp.OK(),
			Token:        jwtToken,
			RefreshToken: refreshToken,
		})
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, storage.ErrRefreshTokenNotFound):
		render.JSON(w, r, resp.Error("refresh token not found"))
	case errors.Is(err, usecase.ErrTokenUsed):
		render.JSON(w, r, resp.Error("refresh token is unavailable"))
	case errors.Is(err, jwt.ErrTokenExpired):
		render.JSON(w, r, resp.Error("token expired"))
	default:
		render.JSON(w, r, resp.Error("internal error"))
	}
}

func NewUsecase(log *slog.Logger, rtStorage tokenrefreshuc.RefreshTokenStorage, uStorage tokenrefreshuc.UserStorage, tokenTTL, rtTTL time.Duration) *tokenrefreshuc.TokenRefreshUseCase {
	return &tokenrefreshuc.TokenRefreshUseCase{
		Log:       log,
		RtStorage: rtStorage,
		UStorage:  uStorage,
		TokenTTL:  tokenTTL,
		RtTTL:     rtTTL,
	}
}
