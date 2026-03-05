package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"notification/internal/domain/models"
	"notification/internal/storage"
	"strconv"
	"strings"

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

func (s *Storage) SaveUser(ctx context.Context, user models.User) (int64, error) {
	const op = "storage.sqlite.SaveUser"

	res, err := s.db.ExecContext(ctx, "INSERT INTO users(user_id, email) VALUES(?, ?)", user.UserId, user.Email)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userId, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}

// TODO: use domain model as param
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

// TODO: use domain model as param
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

func (s *Storage) Users(ctx context.Context, userIds []int64) (map[int64]models.User, error) {
	const op = "storage.sqlite.Users"

	users := make(map[int64]models.User)
	ids := make([]string, 0)

	for _, id := range userIds {
		ids = append(ids, strconv.Itoa(int(id)))
	}

	prepare := strings.Join(ids, ",")

	res, err := s.db.QueryContext(ctx, fmt.Sprintf("SELECT id, email, user_id FROM users WHERE user_id IN (%s)", prepare))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var r models.User

	for res.Next() {
		err := res.Scan(&r.Id, &r.Email, &r.UserId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		users[r.UserId] = r
	}

	return users, nil
}
