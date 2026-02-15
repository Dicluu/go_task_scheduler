package fetchtask

import (
	"context"
	"task/internal/domain/models/task"
)

type Usecase struct {
	fetcher TaskFetcher
}

type TaskFetcher interface {
	Task(ctx context.Context, taskId int64) (task.Task, error)
}

func (u *Usecase) Task(ctx context.Context, taskId int64) (task.Task, error) {
	task, err := u.fetcher.Task(ctx, taskId)

	return task, err
}

func New(fetcher TaskFetcher) *Usecase {
	return &Usecase{
		fetcher: fetcher,
	}
}
