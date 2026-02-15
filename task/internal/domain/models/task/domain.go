package task

import "time"

// TODO: add status

type Task struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	UserId      int64     `json:"-"`
}
