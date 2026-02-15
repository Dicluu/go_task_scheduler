package remove

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"task/internal/domain/models/task"
	"task/internal/lib/logger/sl"
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
	Delete(ctx context.Context, task *task.Task) error
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

		err = usecase.Delete(r.Context(), &task.Task{
			Id: taskId,
		})
		if err != nil {
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
