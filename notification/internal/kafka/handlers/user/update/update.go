package update

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

type UserRequest struct {
	UserId int64  `json:"id"`
	Email  string `json:"email"`
}

type Usecase interface {
	SaveUser(ctx context.Context, email string, userId int64) (int64, error)
}

type Handler struct {
	Usecase Usecase
	Log     *slog.Logger
}

func New(u Usecase, l *slog.Logger) *Handler {
	return &Handler{
		Usecase: u,
		Log:     l,
	}
}

func (h *Handler) HandleMessage(ctx context.Context, m []byte) error {
	const op = "internal.kafka.handlers.user.update.HandleMessage"
	log := h.Log.With(slog.String("op", op))

	var req UserRequest
	err := json.Unmarshal(m, &req)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = h.Usecase.SaveUser(ctx, req.Email, req.UserId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user saved", slog.Any("user", req))

	return nil
}
