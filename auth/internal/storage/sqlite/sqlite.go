package sqlite

import (
	"auth/internal/domain/models"
	"auth/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.sqlite.User"

	var user models.User

	err := s.db.QueryRowContext(ctx, "SELECT id, email, pass_hash FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return &models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) UserById(ctx context.Context, id int64) (*models.User, error) {
	const op = "storage.sqlite.UserById"

	var user models.User

	err := s.db.QueryRowContext(ctx, "SELECT id, email, pass_hash FROM users WHERE id = ?", id).Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return &models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	var isAdmin bool

	err := s.db.QueryRowContext(ctx, "SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.sqlite.App"

	var app models.App

	err := s.db.QueryRowContext(ctx, "SELECT id, name FROM apps WHERE id = ?", appID).Scan(&app.ID, &app.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}

func (s *Storage) RefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	const op = "storage.sqlite.RefreshToken"

	var rt *models.RefreshToken

	err := s.db.QueryRowContext(ctx, "SELECT id, user_id, token, used FROM refresh_tokens WHERE token = ?", token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.Used)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &models.RefreshToken{}, fmt.Errorf("%s: %w", op, storage.ErrRefreshTokenNotFound)
		}

		return &models.RefreshToken{}, fmt.Errorf("%s: %w", op, err)
	}

	return rt, nil
}

// TODO: refactor, do not use domain model in storage layer
func (s *Storage) SaveRefreshToken(ctx context.Context, rt *models.RefreshToken) error {
	const op = "storage.sqlite.SaveRefreshToken"

	_, err := s.db.ExecContext(ctx, "INSERT INTO refresh_tokens(user_id, token, used) VALUES(?,?,?)", rt.UserID, rt.Token, false)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return fmt.Errorf("%s: %w", op, storage.ErrRefreshTokenExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// TODO: refactor, do not use domain model in storage layer
func (s *Storage) UseRefreshToken(ctx context.Context, rt *models.RefreshToken) error {
	const op = "storage.sqlite.UseRefreshToken"

	_, err := s.db.ExecContext(ctx, "UPDATE refresh_tokens SET used = ? WHERE id = ?", true, rt.ID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, storage.ErrRefreshTokenNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
