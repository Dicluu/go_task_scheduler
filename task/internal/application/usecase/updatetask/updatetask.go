package updatetask

import (
	"context"
	"task/internal/domain/models/task"
	"time"
)

type Usecase struct {
	updater TaskUpdater
}

type TaskUpdater interface {
	UpdateTask(ctx context.Context, name, description string, startsAt time.Time, taskId int64) error
}

func New(taskUpdater TaskUpdater) *Usecase {
	return &Usecase{
		updater: taskUpdater,
	}
}

func (u *Usecase) Update(ctx context.Context, task *task.Task) error {
	return u.updater.UpdateTask(ctx, task.Name, task.Description, task.StartsAt, task.Id)
}
