package models

import "time"

// SubmissionTestResult represents Step 1 — Docker test execution output
// This is deterministic — no AI involved
type SubmissionTestResult struct {
	ID                    string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	SubmissionID          string    `json:"submission_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	BuildSuccess          bool      `json:"build_success" gorm:"not null;default:false"`
	TestsTotal            int       `json:"tests_total" gorm:"not null;default:0"`
	TestsPassed           int       `json:"tests_passed" gorm:"not null;default:0"`
	TestsFailed           int       `json:"tests_failed" gorm:"not null;default:0"`
	FunctionalityScore    int       `json:"functionality_score" gorm:"not null;default:0"`
	MaxFunctionalityScore int       `json:"max_functionality_score" gorm:"not null;default:0"`
	ExecutionTimeMs       int       `json:"execution_time_ms" gorm:"default:0"`
	MemoryUsageMb         int       `json:"memory_usage_mb" gorm:"default:0"`
	RawOutput             string    `json:"raw_output,omitempty" gorm:"type:text"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Submission Submission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
}

// CalculateFunctionalityScore computes the deterministic functionality score
// Formula: (TestsPassed / TestsTotal) * MaxFunctionalityScore
func (r *SubmissionTestResult) CalculateFunctionalityScore() int {
	if r.TestsTotal == 0 {
		return 0
	}
	return (r.TestsPassed * r.MaxFunctionalityScore) / r.TestsTotal
}
