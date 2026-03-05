package sendemail

import (
	"context"
	"fmt"
	"log/slog"
	"notification/internal/domain/models"
	"notification/internal/lib/logger/sl"
)

type UserFetcher interface {
	Users(ctx context.Context, userIds []int64) (map[int64]models.User, error)
}

type SendEmailAdapter interface {
	Send(to, subject, body string) error
}

type Usecase struct {
	fetcher UserFetcher
	sender  SendEmailAdapter
	log     *slog.Logger
}

func (u *Usecase) Send(ctx context.Context, notifications []models.Notification) error {
	const op = "internal.application.usecase.sendemail.Send"
	log := u.log.With(slog.String("op", op))
	ids := make([]int64, 0)

	for _, n := range notifications {
		ids = append(ids, n.To)
	}

	users, err := u.fetcher.Users(ctx, ids)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	emails := make(map[int64]string)

	for _, u := range users {
		emails[u.UserId] = u.Email
	}

	for _, n := range notifications {
		s := fmt.Sprintf("Time to make task: %s", n.TaskName)
		b := fmt.Sprintf("Task '%s' is scheduled for %s", n.TaskName, n.StartTime.Format("02.01.2006 15:04:05"))
		err := u.sender.Send(emails[n.To], s, b)
		if err != nil {
			log.Error("failed to send message", slog.Int64("user_id", n.To), sl.Err(err))

			continue
		}

		log.Info("success sent message", slog.Int64("user_id", n.To))
	}

	return nil
}

func New(f UserFetcher, s SendEmailAdapter, l *slog.Logger) *Usecase {
	return &Usecase{
		fetcher: f,
		sender:  s,
		log:     l,
	}
}
