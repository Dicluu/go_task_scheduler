package notifytask

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"task/internal/domain/models"
	"task/internal/http-server/adapters"
	"task/pkg/helpers"
)

type TaskStorage interface {
	TasksReadyNotify(ctx context.Context) ([]models.Task, error)
	MarkAsNotified(ctx context.Context, tasks []models.Task) error
}

type NotifyAdapter interface {
	Notify(tasks []models.Task) error
}

type Usecase struct {
	taskStorage TaskStorage
	adapter     NotifyAdapter
	batch       int
	log         *slog.Logger
}

func New(taskStorage TaskStorage, adapter NotifyAdapter, batch int, log *slog.Logger) *Usecase {
	return &Usecase{
		taskStorage: taskStorage,
		adapter:     adapter,
		batch:       batch,
		log:         log,
	}
}

func (u *Usecase) NotifyTasks(ctx context.Context) error {
	const op = "application.usecase.notifytask.NotifyTask"
	log := u.log.With(slog.String("op", op))

	toNotify, err := u.taskStorage.TasksReadyNotify(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if len(toNotify) == 0 {
		log.Info("no notifications are ready to send")

		return nil
	}

	log.Info("notification are ready to send", slog.Int("count", len(toNotify)))

	chunks := helpers.ChunkSlice(toNotify, u.batch)

	for i, chunk := range chunks {
		err = u.adapter.Notify(chunk)
		if err != nil {
			if errors.Is(err, adapters.ErrFailedConn) {
				return fmt.Errorf("%s: %w", op, err)
			}

			log.Error("failed to send notifications", err)
		}

		err := u.taskStorage.MarkAsNotified(ctx, chunk)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		log.Info("chunk sent", slog.Int("chunk-num", i))
	}

	log.Info("notifications are sent")

	return nil
}
