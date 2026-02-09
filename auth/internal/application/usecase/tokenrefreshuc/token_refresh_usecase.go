package tokenrefreshuc

import (
	"auth/internal/application/usecase"
	"auth/internal/domain/models"
	"auth/internal/lib/jwt/jwtgen"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RefreshTokenStorage interface {
	RefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	SaveRefreshToken(ctx context.Context, rt *models.RefreshToken) error
	UseRefreshToken(ctx context.Context, rt *models.RefreshToken) error
}

type UserStorage interface {
	UserById(ctx context.Context, id int64) (*models.User, error)
}

type TokenRefreshUseCase struct {
	Log       *slog.Logger
	RtStorage RefreshTokenStorage
	UStorage  UserStorage
	TokenTTL  time.Duration
	RtTTL     time.Duration
}

func (u *TokenRefreshUseCase) RefreshSession(ctx context.Context, secret string, refreshToken string) (string, string, error) {
	const op = "application.usecase.tokenrefreshuc.token_refresh_usecase.RefreshSession"
	u.Log = u.Log.With("op", op)

	token, err := u.RtStorage.RefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, storage.ErrRefreshTokenNotFound) {
			u.Log.Warn("refresh token not found", slog.String("token", refreshToken))

			return "", "", err
		}

		return "", "", err
	}

	if token.Used {
		u.Log.Info("refresh token is used", slog.String("token", refreshToken))

		return "", "", usecase.ErrTokenUsed
	}

	_, err = jwt.Parse(token.Token, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}, jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			u.Log.Info("token expired", sl.Err(err))

			return "", "", err
		}

		u.Log.Error("failed to parse jwt token", sl.Err(err))

		return "", "", err
	}

	err = u.RtStorage.UseRefreshToken(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrRefreshTokenNotFound) {
			u.Log.Error("token not found, seems sql query works not properly", slog.String("token", token.Token), sl.Err(err))

			return "", "", err
		}

		u.Log.Error("failed to use refresh token", sl.Err(err))

		return "", "", err
	}

	user, err := u.UStorage.UserById(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			u.Log.Error("user not found", slog.Int64("token_id", token.ID))

			return "", "", err
		}

		return "", "", err
	}

	jwtToken, err := jwtgen.NewToken(user, secret, u.TokenTTL)
	if err != nil {
		u.Log.Error("failed to generate access token", sl.Err(err))

		return "", "", err
	}

	refreshToken, err = jwtgen.NewRefreshToken(user, secret, u.RtTTL)
	if err != nil {
		u.Log.Error("failed to generate refresh token", sl.Err(err))

		return "", "", err
	}

	err = u.RtStorage.SaveRefreshToken(ctx, &models.RefreshToken{UserID: user.ID, Token: refreshToken})
	if err != nil {
		if errors.Is(err, storage.ErrRefreshTokenExists) {
			u.Log.Error(fmt.Sprintf("token generated duplicate %s", refreshToken), sl.Err(err))

			return "", "", err
		}

		u.Log.Error("save refresh token failed", sl.Err(err))

		return "", "", err
	}

	return jwtToken, refreshToken, nil
}
