package index

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"task/internal/domain/models/task"
	"task/internal/lib/logger/sl"
	resp "task/pkg/api/resp"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const (
	DEFAULT_PAGE  = 0
	DEFAULT_LIMIT = 20
)

type Response struct {
	Data []task.Task   `json:"data"`
	Resp resp.Response `json:"resp"`
}

type Usecase interface {
	Tasks(ctx context.Context, limit, offset int, userId int64) ([]task.Task, error)
}

func New(log *slog.Logger, usecase Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.task.index.New"

		log := log.With(
			slog.String("op", op),
			slog.String("req-id", middleware.GetReqID(r.Context())),
		)

		page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
		if err != nil || page == 0 {
			log.Info("failed to parse int for page; default value applied")
			page = DEFAULT_PAGE
		}

		limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
		if err != nil || limit == 0 {
			log.Info("failed to parse int for limit; default value applied")
			limit = DEFAULT_LIMIT
		}

		userId := r.Context().Value("user").(int64)

		tasks, err := usecase.Tasks(r.Context(), int(limit), int(page), userId)
		if err != nil {
			log.Error("failed to get index page", sl.Err(err))

			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		render.JSON(w, r, Response{
			Data: tasks,
			Resp: resp.OK(),
		})
	}
}
