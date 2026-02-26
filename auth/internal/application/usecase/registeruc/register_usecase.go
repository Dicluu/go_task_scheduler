package registeruc

import (
	"auth/internal/application/usecase"
	"auth/internal/domain/models"
	kafkaIn "auth/internal/kafka"
	"auth/internal/kafka/handlers/user/create"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	"auth/pkg/kafka"
	"context"
	"errors"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type Usecase struct {
	Saver    UserSaver
	Producer *kafka.Producer
	Log      *slog.Logger
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, pasHash []byte) (int64, error)
}

func (u *Usecase) Register(ctx context.Context, email string, pass string) (int64, error) {
	const op = "application.usecase.registeruc.register_usecase.Register"
	log := u.Log.With(slog.String("op", op))

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password")

		return 0, err
	}

	userID, err := u.Saver.SaveUser(ctx, email, hash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Info("user already exists", sl.Err(err))

			return 0, usecase.ErrUserAlreadyExists
		}

		log.Error("error appears when save user: ", sl.Err(err))

		return 0, err
	}

	var user models.User
	user.ID = userID
	user.Email = email

	err = u.Producer.Produce(newKafkaHandler(u.Log), user, kafkaIn.UserCreateTopic)
	if err != nil {
		log.Error("failed to send kafka message", sl.Err(err))

		return 0, err
	}

	log.Info("kafka event created")

	return userID, nil
}

func newKafkaHandler(log *slog.Logger) kafka.Handler {
	return &create.Handler{
		Log: log,
	}
}
