package deletetask

import (
	"context"
	"task/internal/domain/models"
	"task/internal/domain/models/task"
)

type Usecase struct {
	deleter TaskDeleter
}

type TaskDeleter interface {
	Task(ctx context.Context, taskId int64) (task.Task, error)
	DeleteTask(ctx context.Context, taskId int64) error
}

func New(deleter TaskDeleter) *Usecase {
	return &Usecase{
		deleter: deleter,
	}
}

func (u *Usecase) Delete(ctx context.Context, task *task.Task, userId int64) error {
	t, err := u.deleter.Task(ctx, task.Id)
	if err != nil {
		return err
	}

	if !t.CanBeDeletedBy(userId) {
		return models.ErrCannotDeleteRecord
	}

	err = u.deleter.DeleteTask(ctx, task.Id)
	if err != nil {
		return err
	}

	return nil
}
