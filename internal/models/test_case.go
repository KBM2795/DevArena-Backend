package models

import "time"

// TestCase represents a test case for a challenge
type TestCase struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	ChallengeID    string    `json:"challenge_id" gorm:"type:varchar(255);not null;index"`
	Input          string    `json:"input" gorm:"type:text;not null"`
	ExpectedOutput string    `json:"expected_output" gorm:"type:text;not null"`
	IsHidden       bool      `json:"is_hidden" gorm:"default:false"` // Hidden test cases not shown to users
	OrderIndex     int       `json:"order_index" gorm:"default:0"`
	Weight         int       `json:"weight" gorm:"default:1"` // Score weight for this test case
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationship
	Challenge Challenge `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
}
