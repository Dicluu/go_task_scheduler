package update

import (
	"context"
	"notification/internal/domain/models"
)

type Usecase struct {
	saver UserSaver
}

type UserSaver interface {
	SaveUser(ctx context.Context, user models.User) (int64, error)
}

func (u *Usecase) SaveUser(ctx context.Context, email string, userId int64) (int64, error) {
	var user models.User
	user.Email = email
	user.UserId = userId

	userId, err := u.saver.SaveUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func New(s UserSaver) *Usecase {
	return &Usecase{
		saver: s,
	}
}
