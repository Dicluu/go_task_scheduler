package index

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"task/internal/domain/models"
	"task/internal/lib/logger/sl"
	"task/internal/middlewares/auth"
	resp "task/pkg/api/resp"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const (
	DEFAULT_PAGE  = 0
	DEFAULT_LIMIT = 20
)

type Response struct {
	Data []models.Task `json:"data"`
	Resp resp.Response `json:"resp"`
}

type Usecase interface {
	Tasks(ctx context.Context, limit, offset int, userId int64) ([]models.Task, error)
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

		tasks, err := usecase.Tasks(r.Context(), int(limit), int(page), user.Id)
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
