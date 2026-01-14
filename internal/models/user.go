package models

import "time"

// User represents a DevArena user
type User struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	ClerkUserID   string    `json:"clerk_user_id" gorm:"uniqueIndex;type:varchar(255);not null"`
	Email         string    `json:"email" gorm:"uniqueIndex;type:varchar(255);not null"`
	Username      string    `json:"username" gorm:"uniqueIndex;type:varchar(100)"`
	DisplayName   string    `json:"display_name" gorm:"type:varchar(255)"`
	AvatarURL     string    `json:"avatar_url" gorm:"type:text"`
	Bio           string    `json:"bio" gorm:"type:text"`
	CurrentStreak int       `json:"current_streak" gorm:"default:0"`
	LongestStreak int       `json:"longest_streak" gorm:"default:0"`
	TotalScore    int       `json:"total_score" gorm:"default:0"`
	Rank          int       `json:"rank" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Submissions []Submission `json:"submissions,omitempty" gorm:"foreignKey:UserID"`
}

// UserStats represents aggregated user statistics
type UserStats struct {
	UserID         string  `json:"user_id"`
	ProblemsSolved int     `json:"problems_solved"`
	AcceptanceRate float64 `json:"acceptance_rate"`
	CurrentStreak  int     `json:"current_streak"`
	HoursCoded     float64 `json:"hours_coded"`
	TopPercentile  float64 `json:"top_percentile"`
}
