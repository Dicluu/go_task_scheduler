package updatetask

import (
	"context"
	"task/internal/domain/models"
	"time"
)

type Usecase struct {
	updater TaskUpdater
}

type TaskUpdater interface {
	Task(ctx context.Context, taskId int64) (models.Task, error)
	UpdateTask(ctx context.Context, name, description, status string, startsAt time.Time, taskId int64) error
}

func New(taskUpdater TaskUpdater) *Usecase {
	return &Usecase{
		updater: taskUpdater,
	}
}

func (u *Usecase) Update(ctx context.Context, inTask *models.Task, userId int64) error {

	t, err := u.updater.Task(ctx, inTask.Id)
	if err != nil {
		return err
	}

	if !t.CanBeUpdatedBy(userId) {
		return models.ErrCannotUpdateRecord
	}

	err = u.updater.UpdateTask(ctx, inTask.Name, inTask.Description, inTask.Status, inTask.StartsAt, inTask.Id)

	return nil
}
