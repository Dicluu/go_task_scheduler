package register

import (
	"auth/internal/application/usecase"
	"auth/internal/application/usecase/registeruc"
	"auth/internal/lib/logger/sl"
	resp "auth/pkg/api/resp"
	"auth/pkg/kafka"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
}

type Usecase interface {
	Register(ctx context.Context, email string, password string) (int64, error)
}

func New(log *slog.Logger, u Usecase) http.HandlerFunc {
	const op = "http-server.handlers.user.register.New"
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

		log.Info("request body decoded")

		err = validator.New().Struct(req)
		if err != nil {
			log.Info("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(err.(validator.ValidationErrors)))

			return
		}

		userID, err := u.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, usecase.ErrUserAlreadyExists) {
				render.JSON(w, r, resp.Error("user already exists"))

				return
			}

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("user registered", slog.Int64("user_id", userID))

		render.JSON(w, r, &Response{
			Response: resp.OK(),
		})

		return
	}
}

func NewUsecase(s registeruc.UserSaver, l *slog.Logger, p *kafka.Producer) Usecase {
	return &registeruc.Usecase{
		Saver:    s,
		Log:      l,
		Producer: p,
	}
}
