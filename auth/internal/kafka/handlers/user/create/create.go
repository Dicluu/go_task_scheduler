package create

import (
	"auth/internal/domain/models"
	k "auth/pkg/kafka"
	"encoding/json"
	"fmt"
	"log/slog"
)

type User struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
}

type Handler struct {
	Log *slog.Logger
}

func (h *Handler) PrepareMessage(message any) ([]byte, error) {
	const op = "internal.kafka.handlers.user.create"
	log := h.Log.With(slog.String("op", op), slog.Any("message", message))

	m, ok := message.(models.User)
	if !ok {
		log.Error("wrong type provided to handler")

		return []byte{}, k.ErrWrongTypeProvided
	}

	var user User
	user.Id = m.ID
	user.Email = m.Email

	res, err := json.Marshal(user)
	if err != nil {
		log.Error("failed to marshal message")

		return []byte{}, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func New(log *slog.Logger) *Handler {
	return &Handler{
		Log: log,
	}
}
