package fetchtasks

import (
	"context"
	"task/internal/domain/models"
)

type Usecase struct {
	Fetcher TaskFetcher
}

type TaskFetcher interface {
	Tasks(ctx context.Context, limit, offset int, userId int64) ([]models.Task, error)
}

func New(fetcher TaskFetcher) *Usecase {
	return &Usecase{
		Fetcher: fetcher,
	}
}

func (u *Usecase) Tasks(ctx context.Context, limit, offset int, userId int64) ([]models.Task, error) {
	return u.Fetcher.Tasks(ctx, limit, offset, userId)
}
