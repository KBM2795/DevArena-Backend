package models

import (
	"time"
)

// Notification represents a user notification
type Notification struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:varchar(255);not null"`
	Title     string    `json:"title" gorm:"type:varchar(255);not null"`
	Message   string    `json:"message" gorm:"type:text"`
	Type      string    `json:"type" gorm:"type:varchar(50);default:'info'"`
	Link      string    `json:"link" gorm:"type:varchar(255)"`
	IsRead    bool      `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
