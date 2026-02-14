package models

import "time"

// EvaluationStatus represents the pipeline status of a submission
type EvaluationStatus string

const (
	EvalPending   EvaluationStatus = "pending"   // Waiting to be evaluated
	EvalQueued    EvaluationStatus = "queued"    // In evaluation queue
	EvalTesting   EvaluationStatus = "testing"   // Step 1: Docker tests running
	EvalReviewing EvaluationStatus = "reviewing" // Step 2: AI code review
	EvalReporting EvaluationStatus = "reporting" // Step 3: AI report generation
	EvalCompleted EvaluationStatus = "completed" // All steps done
	EvalFailed    EvaluationStatus = "failed"    // Pipeline failed
)

// Submission represents a user's GitHub repository submission for a challenge
type Submission struct {
	ID          string `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID      string `json:"user_id" gorm:"type:varchar(255);not null;index"`
	ChallengeID string `json:"challenge_id" gorm:"type:varchar(255);not null;index"`
	RepoURL     string `json:"repo_url" gorm:"type:text;not null"`
	Branch      string `json:"branch" gorm:"type:varchar(100);default:main"`
	CommitHash  string `json:"commit_hash" gorm:"type:varchar(64)"`

	// Evaluation pipeline status
	EvaluationStatus      EvaluationStatus `json:"evaluation_status" gorm:"type:varchar(30);not null;default:pending"`
	ErrorMessage          string           `json:"error_message,omitempty" gorm:"type:text"`
	EvaluationStartedAt   *time.Time       `json:"evaluation_started_at,omitempty"`
	EvaluationCompletedAt *time.Time       `json:"evaluation_completed_at,omitempty"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User       User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Challenge  Challenge             `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
	AIReviews  []AIReview            `json:"ai_reviews,omitempty" gorm:"foreignKey:SubmissionID"`
	TestResult *SubmissionTestResult `json:"test_result,omitempty" gorm:"foreignKey:SubmissionID"`
	Score      *SubmissionScore      `json:"score_breakdown,omitempty" gorm:"foreignKey:SubmissionID"`
	Report     *AIReport             `json:"report,omitempty" gorm:"foreignKey:SubmissionID"`
}

// SubmissionRequest represents the API request for creating a submission
type SubmissionRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	RepoURL     string `json:"repo_url" binding:"required,url"`
	Branch      string `json:"branch"`
}
