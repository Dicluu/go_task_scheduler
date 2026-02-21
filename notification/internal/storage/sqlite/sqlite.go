package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"notification/internal/storage"

	"github.com/mattn/go-sqlite3"
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

func (s *Storage) SaveUser(ctx context.Context, email string) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	res, err := s.db.ExecContext(ctx, "INSERT INTO users(email) VALUES(?)", email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}

func (s *Storage) UpdateUser(ctx context.Context, email string, userId int64) error {
	const op = "storage.sqlite.UpdateUser"

	_, err := s.db.ExecContext(ctx, "UPDATE users SET email = ? WHERE id = ?", email, userId)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, userId int64) error {
	const op = "storage.sqlite.DeleteUser"

	_, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userId)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
