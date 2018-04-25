package structs

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int32 `gorm:"primary_key"`
	FirstName string
	LastName  string
	Username  sql.NullString
	TgID      int64
	CreatedAt time.Time
}

type InvalidMessage struct {
	ID        int32 `gorm:"primary_key"`
	UserID    int32
	Message   string
	CreatedAt time.Time
}

type Task struct {
	ID    int32 `gorm:"primary_key"`
	Task  string
	Image string
}

type UserTask struct {
	ID     int32 `gorm:"primary_key"`
	UserID int32
	TaskID int32
}
