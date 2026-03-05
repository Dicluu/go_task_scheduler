package dto

import (
	"notification/internal/domain/models"
	"time"
)

type NotifyRequestDTO struct {
	UserId   int64
	TaskName string
	Date     time.Time
}

func (r *NotifyRequestDTO) ToDomain() *models.Notification {
	return &models.Notification{
		To:        r.UserId,
		TaskName:  r.TaskName,
		StartTime: r.Date,
	}
}
