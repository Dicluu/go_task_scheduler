package remove

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"task/internal/domain/models"
	"task/internal/lib/logger/sl"
	"task/internal/middlewares/auth"
	"task/internal/storage"
	resp "task/pkg/api/resp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	Resp resp.Response `json:"resp"`
}

type Usecase interface {
	Delete(ctx context.Context, task *models.Task, userId int64) error
}

func New(log *slog.Logger, usecase Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.task.delete.New"

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

		err = usecase.Delete(r.Context(), &models.Task{
			Id: taskId,
		}, user.Id)
		if err != nil {
			if errors.Is(err, models.ErrCannotDeleteRecord) {
				log.Warn("attempt to delete record of another person")

				render.JSON(w, r, resp.Error("you can't delete this task"))

				return
			}

			if errors.Is(err, storage.ErrTaskNotFound) {
				log.Info("task not found")

				render.JSON(w, r, resp.Error("task not found"))

				return
			}

			log.Error("failed to delete task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		render.JSON(w, r, Response{Resp: resp.OK()})
	}
}
