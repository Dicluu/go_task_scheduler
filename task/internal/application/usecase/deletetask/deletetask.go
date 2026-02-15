package deletetask

import (
	"context"
	"task/internal/domain/models"
)

type Usecase struct {
	deleter TaskDeleter
}

type TaskDeleter interface {
	Task(ctx context.Context, taskId int64) (models.Task, error)
	DeleteTask(ctx context.Context, taskId int64) error
}

func New(deleter TaskDeleter) *Usecase {
	return &Usecase{
		deleter: deleter,
	}
}

func (u *Usecase) Delete(ctx context.Context, task *models.Task, userId int64) error {
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
