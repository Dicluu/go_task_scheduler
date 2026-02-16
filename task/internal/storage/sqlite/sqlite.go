package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"task/internal/domain/models"
	"task/internal/storage"
	"time"

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

func (s *Storage) SaveTask(ctx context.Context, name, description string, startsAt time.Time, userId int64) (int64, error) {
	const op = "storage.sqlite.SaveTask"

	res, err := s.db.ExecContext(ctx, "INSERT INTO tasks(name, description, starts_at, user_id, status) VALUES(?,?,?,?,?)", name, description, startsAt, userId, models.TASK_STATUS_TODO)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	taskId, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return taskId, nil
}

func (s *Storage) Task(ctx context.Context, taskId int64) (models.Task, error) {
	const op = "storage.sqlite.Task"

	var t models.Task

	err := s.db.QueryRowContext(ctx, "SELECT name, description, starts_at, user_id, status FROM tasks WHERE id = ?", taskId).
		Scan(&t.Name, &t.Description, &t.StartsAt, &t.UserId, &t.Status)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) || errors.Is(err, sql.ErrNoRows) {
			return models.Task{}, fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
		}

		return models.Task{}, fmt.Errorf("%s: %w", op, err)
	}

	return t, nil
}

func (s *Storage) UpdateTask(ctx context.Context, name, description, status string, startsAt time.Time, taskId int64) error {
	const op = "storage.sqlite.UpdateTask"

	_, err := s.db.ExecContext(ctx, "UPDATE tasks SET name = ?, description = ?, starts_at = ?, status = ? WHERE id = ?", name, description, startsAt, status, taskId)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteTask(ctx context.Context, taskId int64) error {
	const op = "storage.sqlite.DeleteTask"

	_, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", taskId)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Tasks(ctx context.Context, limit, offset int, userId int64) ([]models.Task, error) {
	const op = "storage.sqlite.Tasks"

	tasks := make([]models.Task, 0)

	res, err := s.db.QueryContext(ctx, "SELECT id, name, description, starts_at, status FROM tasks WHERE user_id = ? LIMIT ? OFFSET ?", userId, limit, offset)
	if err != nil {
		return []models.Task{}, fmt.Errorf("%s: %w", op, err)
	}

	var r models.Task

	for res.Next() {
		err := res.Scan(&r.Id, &r.Name, &r.Description, &r.StartsAt, &r.Status)
		if err != nil {
			return []models.Task{}, fmt.Errorf("%s: %w", op, err)
		}

		tasks = append(tasks, r)
	}

	return tasks, nil
}
