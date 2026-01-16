package models

import "time"

// SubmissionStatus represents the status of a submission
type SubmissionStatus string

const (
	StatusPending   SubmissionStatus = "pending"   // Waiting to be reviewed
	StatusReviewing SubmissionStatus = "reviewing" // AI is reviewing
	StatusReviewed  SubmissionStatus = "reviewed"  // Review complete
	StatusFailed    SubmissionStatus = "failed"    // Review failed (invalid repo, etc.)
)

// Submission represents a user's GitHub repository submission for a challenge
type Submission struct {
	ID          string           `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID      string           `json:"user_id" gorm:"type:varchar(255);not null;index"`
	ChallengeID string           `json:"challenge_id" gorm:"type:varchar(255);not null;index"`
	RepoURL     string           `json:"repo_url" gorm:"type:text;not null"`
	Branch      string           `json:"branch" gorm:"type:varchar(100);default:main"`
	CommitHash  string           `json:"commit_hash" gorm:"type:varchar(64)"`
	Status      SubmissionStatus `json:"status" gorm:"type:varchar(50);not null;default:pending"`
	Score       int              `json:"score" gorm:"default:0;comment:Overall score 0-100"`
	CreatedAt   time.Time        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time        `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Challenge Challenge  `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
	AIReviews []AIReview `json:"ai_reviews,omitempty" gorm:"foreignKey:SubmissionID"`
}

// SubmissionRequest represents the API request for creating a submission
type SubmissionRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	RepoURL     string `json:"repo_url" binding:"required,url"`
	Branch      string `json:"branch"`
}
