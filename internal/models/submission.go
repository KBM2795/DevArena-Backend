package models

import "time"

// Submission represents a user's showcase project submission
type Submission struct {
	ID                 string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID             string     `json:"user_id" gorm:"type:varchar(255);not null;index"`
	ChallengeID        *string    `json:"challenge_id,omitempty" gorm:"type:varchar(255);index"` // Nullable for open showcase
	Title              string     `json:"title" gorm:"type:text"`
	RepoURL            string     `json:"repo_url" gorm:"type:text;not null"`
	VideoURL           string     `json:"video_url" gorm:"type:text"`
	ProjectDescription string     `json:"project_description" gorm:"type:text"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Challenge *Challenge `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
}

// SubmissionRequest represents the API request for creating a submission
type SubmissionRequest struct {
	ChallengeID *string `json:"challenge_id"` // Optional for open showcase
	Title       string  `json:"title" binding:"required"`
	RepoURL     string  `json:"repo_url" binding:"required"`
	VideoURL    string  `json:"video_url"`
	Description string  `json:"description"`
}
