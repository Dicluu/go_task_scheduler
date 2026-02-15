package update

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"task/internal/domain/models"
	"task/internal/domain/models/task"
	"task/internal/lib/logger/sl"
	"task/internal/storage"
	resp "task/pkg/api/resp"
	"time"

	"github.com/go-chi/chi/v5"
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
	Resp resp.Response `json:"resp"`
}

type Usecase interface {
	Update(ctx context.Context, task *task.Task, userId int64) error
}

func New(log *slog.Logger, usecase Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.task.update.New"

		log := log.With(
			slog.String("op", op),
			log.With("req-id", middleware.GetReqID(r.Context())),
		)

		taskId, err := strconv.ParseInt(chi.URLParam(r, "task"), 10, 64)
		if err != nil {
			log.Warn("failed to parse int", slog.String("task-id", chi.URLParam(r, "task")), sl.Err(err))

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		log = log.With(slog.Int64("task-id", taskId))

		if r.ContentLength == 0 {
			log.Warn("missing request body")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		var req Request
		err = render.DecodeJSON(r.Body, &req)
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

		userId := r.Context().Value("user").(int64)

		err = usecase.Update(r.Context(), &task.Task{
			Id:          taskId,
			Name:        req.Name,
			Description: req.Description,
			StartsAt:    req.StartsAt,
		}, userId)
		if err != nil {
			if errors.Is(err, models.ErrCannotUpdateRecord) {
				log.Warn("attempt to update record of another person")

				render.JSON(w, r, resp.Error("you can't update this task"))

				return
			}

			if errors.Is(err, storage.ErrTaskNotFound) {
				log.Info("task not found")

				render.JSON(w, r, resp.Error("task not found"))

				return
			}

			log.Error("failed to update task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		render.JSON(w, r, Response{Resp: resp.OK()})
	}
}
