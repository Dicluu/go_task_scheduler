package show

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

type Usecase interface {
	Task(ctx context.Context, taskId int64) (task.Task, error)
}

type Response struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	StartsAt    string        `json:"starts_at"`
	Resp        resp.Response `json:"resp"`
	UserId      int64         `json:"-"`
}

func New(log *slog.Logger, usecase Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.task.show.New"

		log := log.With(
			slog.String("op", op),
			slog.String("uri-param", chi.URLParam(r, "task")),
			slog.String("req-id", middleware.GetReqID(r.Context())),
		)

		taskId, err := strconv.ParseInt(chi.URLParam(r, "task"), 10, 64)
		if err != nil {
			log.Warn("failed to convert URI param to int", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to resolve URI param"))

			return
		}

		task, err := usecase.Task(r.Context(), taskId)
		if err != nil {
			if errors.Is(err, storage.ErrTaskNotFound) {
				log.Warn("task not found", sl.Err(err))

				render.JSON(w, r, resp.Error("task not found"))

				return
			}

			log.Error("failed to fetch task", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		render.JSON(w, r, Response{
			Name:        task.Name,
			Description: task.Description,
			StartsAt:    task.StartsAt.Format("2006.01.02 15:04:05"),
			UserId:      task.UserId,
			Resp:        resp.OK(),
		})

	}
}
