package models

import "time"

type RefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiredAt time.Duration
	Used      bool
}
