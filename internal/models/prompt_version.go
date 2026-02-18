package models

import "time"

// PromptVersion tracks which AI prompt/model produced each review or report
type PromptVersion struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Name       string    `json:"name" gorm:"type:varchar(100);not null"` // e.g. "code_review_v1"
	Step       string    `json:"step" gorm:"type:varchar(20);not null"`  // "code_review" or "report"
	Model      string    `json:"model" gorm:"type:varchar(50);not null"` // "gpt-4o", "gpt-4"
	PromptText string    `json:"prompt_text" gorm:"type:text;not null"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}
