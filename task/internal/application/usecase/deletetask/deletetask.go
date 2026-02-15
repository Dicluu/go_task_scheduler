package deletetask

import (
	"context"
	"task/internal/domain/models/task"
)

type Usecase struct {
	deleter TaskDeleter
}

type TaskDeleter interface {
	DeleteTask(ctx context.Context, taskId int64) error
}

func New(deleter TaskDeleter) *Usecase {
	return &Usecase{
		deleter: deleter,
	}
}

func (u *Usecase) Delete(ctx context.Context, task *task.Task) error {
	return u.deleter.DeleteTask(ctx, task.Id)
}
