package register

import (
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	resp "auth/pkg/api/resp"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, pasHash []byte) (int64, error)
}

func New(log *slog.Logger, saver UserSaver) http.HandlerFunc {
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

		pass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Error("failed to generate password")

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		userID, err := saver.SaveUser(r.Context(), req.Email, pass)
		if err != nil {
			if errors.Is(err, storage.ErrUserExists) {
				log.Info("user already exists", sl.Err(err))

				render.JSON(w, r, resp.Error("user already exists"))
				return
			}

			log.Error("errors occurred", sl.Err(err))

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
