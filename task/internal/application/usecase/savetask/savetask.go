package savetask

import (
	"context"
	"task/internal/domain/models/task"
	"time"
)

type Usecase struct {
	saver TaskSaver
}

type TaskSaver interface {
	SaveTask(ctx context.Context, name, description string, startsAt time.Time, userId int64) (int64, error)
}

func New(saver TaskSaver) *Usecase {
	return &Usecase{saver: saver}
}

func (u *Usecase) Save(ctx context.Context, task *task.Task) (int64, error) {
	recordId, err := u.saver.SaveTask(ctx, task.Name, task.Description, task.StartsAt, task.UserId)

	return recordId, err
}
