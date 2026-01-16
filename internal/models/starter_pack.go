package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StarterPack represents a personalized learning pack generated from onboarding
type StarterPack struct {
	ID              string      `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID          string      `json:"user_id" gorm:"uniqueIndex;type:varchar(255);not null"`
	Experience      string      `json:"experience" gorm:"type:varchar(50)"` // beginner, intermediate, advanced, expert
	Paths           StringArray `json:"paths" gorm:"type:jsonb"`            // frontend, backend, fullstack, ai, mobile
	Technologies    StringArray `json:"technologies" gorm:"type:jsonb"`     // Selected technologies
	Challenges      StringArray `json:"challenge_ids" gorm:"type:jsonb"`    // Recommended challenge IDs
	CurrentProgress int         `json:"current_progress" gorm:"default:0"`  // Completed challenges count
	TotalChallenges int         `json:"total_challenges" gorm:"default:0"`  // Total challenges in pack
	IsActive        bool        `json:"is_active" gorm:"default:true"`
	CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// StarterPackChallenge represents a challenge in a starter pack with ordering
type StarterPackChallenge struct {
	ID            string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	StarterPackID string     `json:"starter_pack_id" gorm:"type:varchar(255);not null;index"`
	ChallengeID   string     `json:"challenge_id" gorm:"type:varchar(255);not null;index"`
	OrderIndex    int        `json:"order_index" gorm:"default:0"`
	IsCompleted   bool       `json:"is_completed" gorm:"default:false"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Score         int        `json:"score" gorm:"default:0"`

	// Relationships
	StarterPack StarterPack `json:"starter_pack,omitempty" gorm:"foreignKey:StarterPackID"`
	Challenge   Challenge   `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
}

// StarterPackResponse is the API response for a starter pack
type StarterPackResponse struct {
	ID              string                  `json:"id"`
	Experience      string                  `json:"experience"`
	Paths           []string                `json:"paths"`
	Technologies    []string                `json:"technologies"`
	Challenges      []ChallengeWithProgress `json:"challenges"`
	CurrentProgress int                     `json:"current_progress"`
	TotalChallenges int                     `json:"total_challenges"`
	ProgressPercent float64                 `json:"progress_percent"`
}

// ChallengeWithProgress represents a challenge with user progress
type ChallengeWithProgress struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Difficulty  Difficulty `json:"difficulty"`
	TechStack   []string   `json:"tech_stack"`
	OrderIndex  int        `json:"order_index"`
	IsCompleted bool       `json:"is_completed"`
	Score       int        `json:"score,omitempty"`
}

// StringArray is a helper type for JSON arrays in PostgreSQL
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}
