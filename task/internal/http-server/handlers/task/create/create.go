package create

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"task/internal/domain/models"
	"task/internal/lib/logger/sl"
	"task/internal/middlewares/auth"
	resp "task/pkg/api/resp"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type Request struct {
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at" validate:"required"`
}

type Response struct {
	TaskId int64         `json:"task_id"`
	Resp   resp.Response `json:"resp"`
}

type Usecase interface {
	Save(ctx context.Context, task *models.Task) (int64, error)
}

func New(log *slog.Logger, usecase Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.task.create.New"

		log := log.With(
			slog.String("op", op),
			slog.String("req-id", middleware.GetReqID(r.Context())),
		)

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

		user, err := auth.GetUser(r)
		if err != nil {
			if errors.Is(err, models.ErrUnauthorized) {
				log.Warn("failed to get user from context; maybe middleware is not provided properly")

				render.JSON(w, r, resp.Error("unauthorized"))

				return
			}

			log.Error("failed to get user from context", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		taskId, err := usecase.Save(r.Context(), &models.Task{
			Name:        req.Name,
			Description: req.Description,
			StartsAt:    req.StartsAt,
			UserId:      user.Id,
		})
		if err != nil {
			log.Error("failed to save task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		render.JSON(w, r, Response{TaskId: taskId, Resp: resp.OK()})
	}
}
