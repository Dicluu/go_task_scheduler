package models

import "time"

type Notification struct {
	To        int64
	TaskName  string
	StartTime time.Time
}
