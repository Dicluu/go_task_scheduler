package fetchtask

import (
	"context"
	"task/internal/domain/models"
)

type Usecase struct {
	fetcher TaskFetcher
}

type TaskFetcher interface {
	Task(ctx context.Context, taskId int64) (models.Task, error)
}

func (u *Usecase) Task(ctx context.Context, taskId, userId int64) (models.Task, error) {
	t, err := u.fetcher.Task(ctx, taskId)

	if !t.CanBeViewedBy(userId) {
		return models.Task{}, models.ErrCannotViewRecord
	}

	return t, err
}

func New(fetcher TaskFetcher) *Usecase {
	return &Usecase{
		fetcher: fetcher,
	}
}
