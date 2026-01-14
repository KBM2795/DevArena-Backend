package models

import "time"

// SubmissionStatus represents the status of a submission
type SubmissionStatus string

const (
	StatusPending           SubmissionStatus = "pending"
	StatusRunning           SubmissionStatus = "running"
	StatusAccepted          SubmissionStatus = "accepted"
	StatusWrongAnswer       SubmissionStatus = "wrong_answer"
	StatusTimeLimitExceed   SubmissionStatus = "time_limit_exceeded"
	StatusMemoryLimitExceed SubmissionStatus = "memory_limit_exceeded"
	StatusRuntimeError      SubmissionStatus = "runtime_error"
	StatusCompileError      SubmissionStatus = "compile_error"
)

// Submission represents a user's code submission
type Submission struct {
	ID            string           `json:"id" gorm:"primaryKey;type:varchar(255)"`
	UserID        string           `json:"user_id" gorm:"type:varchar(255);not null;index"`
	ChallengeID   string           `json:"challenge_id" gorm:"type:varchar(255);not null;index"`
	Code          string           `json:"code" gorm:"type:text;not null"`
	Language      string           `json:"language" gorm:"type:varchar(50);not null"`
	Status        SubmissionStatus `json:"status" gorm:"type:varchar(50);not null;default:pending"`
	Score         int              `json:"score" gorm:"default:0"`
	ExecutionTime int              `json:"execution_time" gorm:"comment:Execution time in ms"`
	MemoryUsed    int              `json:"memory_used" gorm:"comment:Memory used in KB"`
	ErrorMessage  string           `json:"error_message,omitempty" gorm:"type:text"`
	TestsPassed   int              `json:"tests_passed" gorm:"default:0"`
	TestsTotal    int              `json:"tests_total" gorm:"default:0"`
	CreatedAt     time.Time        `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Challenge Challenge `json:"challenge,omitempty" gorm:"foreignKey:ChallengeID"`
}

// SubmissionResult represents the result of running a submission
type SubmissionResult struct {
	SubmissionID  string           `json:"submission_id"`
	Status        SubmissionStatus `json:"status"`
	Score         int              `json:"score"`
	ExecutionTime int              `json:"execution_time"`
	MemoryUsed    int              `json:"memory_used"`
	TestResults   []TestResult     `json:"test_results"`
}

// TestResult represents the result of a single test case
type TestResult struct {
	TestCaseID string `json:"test_case_id"`
	Passed     bool   `json:"passed"`
	Output     string `json:"output,omitempty"`
	Expected   string `json:"expected,omitempty"`
	Error      string `json:"error,omitempty"`
}
