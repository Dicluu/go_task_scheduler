package loginuc

import (
	"auth/internal/application/usecase"
	"auth/internal/domain/models"
	"auth/internal/lib/jwt/jwtgen"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginUseCase struct {
	Log      *slog.Logger
	Ugetter  UserGetter
	RtSaver  RefreshTokenSaver
	Secret   string
	TokenTTL time.Duration
	RtTTL    time.Duration
}

type UserGetter interface {
	User(ctx context.Context, email string) (*models.User, error)
}

type RefreshTokenSaver interface {
	SaveRefreshToken(ctx context.Context, rt *models.RefreshToken) error
}

func (u *LoginUseCase) Login(ctx context.Context, email, password string) (string, string, error) {
	const op = "application.usecase.loginuc.login_usecase.Login"
	log := u.Log.With("op", op)

	user, err := u.Ugetter.User(ctx, email)
	if err != nil {
		log.Error("error when getting user", sl.Err(err))

		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Warn("failed to login", slog.Int64("user_id", user.ID))

		return "", "", usecase.ErrWrongLoginOrPassword
	}

	token, err := jwtgen.NewToken(user, u.Secret, u.TokenTTL)
	if err != nil {
		log.Error("failed to generate JWT token", sl.Err(err))

		return "", "", err
	}

	refreshToken, err := jwtgen.NewRefreshToken(user, u.Secret, u.RtTTL)
	if err != nil {
		log.Error("failed to generate JWT token", sl.Err(err))

		return "", "", err
	}

	err = u.RtSaver.SaveRefreshToken(ctx, &models.RefreshToken{Token: refreshToken, UserID: user.ID})
	if err != nil {
		if errors.Is(err, storage.ErrRefreshTokenExists) {
			log.Error("failed to generate unique JWT", sl.Err(err))

			return "", "", err
		}

		log.Error("failed to save refresh token", sl.Err(err))

		return "", "", err
	}

	return token, refreshToken, nil
}
