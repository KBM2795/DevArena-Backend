package models

import (
	"time"
)

// AIReview represents Step 2 — AI code review for a submission
// AI returns category scores + structured feedback; never computes final score
type AIReview struct {
	ID           string `json:"id" gorm:"primaryKey;type:varchar(255)"`
	SubmissionID string `json:"submission_id" gorm:"type:varchar(255);not null;index"`

	// Structured category scores (must not exceed rubric max)
	CodeQualityScore  int `json:"code_quality_score" gorm:"not null;default:0"`
	ConstraintScore   int `json:"constraint_score" gorm:"not null;default:0"`
	ArchitectureScore int `json:"architecture_score" gorm:"not null;default:0"`

	// Structured feedback arrays
	StrengthsJSON    StringArray `json:"strengths" gorm:"type:jsonb;column:strengths_json"`
	IssuesJSON       StringArray `json:"issues" gorm:"type:jsonb;column:issues_json"`
	ImprovementsJSON StringArray `json:"improvements" gorm:"type:jsonb;column:improvements_json"`

	// Auditing
	PromptVersionID string `json:"prompt_version_id,omitempty" gorm:"type:varchar(255)"`
	RawResponse     string `json:"raw_response,omitempty" gorm:"type:text"`

	ReviewedAt time.Time `json:"reviewed_at" gorm:"autoCreateTime"`

	// Relationships
	Submission Submission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
}

// AIReviewInput is the structured payload sent to the AI for code review (Step 2)
type AIReviewInput struct {
	EvaluationType string                `json:"evaluation_type"` // "code_review"
	ChallengeCtx   AIReviewChallengeCtx  `json:"challenge_context"`
	SubmissionCtx  AIReviewSubmissionCtx `json:"submission_context"`
}

type AIReviewChallengeCtx struct {
	ChallengeID    string         `json:"challenge_id"`
	Title          string         `json:"title"`
	Difficulty     string         `json:"difficulty"`
	DescriptionMD  string         `json:"description_md"`
	RequirementsMD string         `json:"requirements_md"`
	ConstraintsMD  string         `json:"constraints_md"`
	Rubric         map[string]int `json:"rubric"`
}

type AIReviewSubmissionCtx struct {
	Files       map[string]string `json:"files"`
	TestSummary TestSummary       `json:"test_summary"`
}

type TestSummary struct {
	TestsPassed int `json:"tests_passed"`
	TestsFailed int `json:"tests_failed"`
}

// AIReviewOutput is the expected AI response (strict JSON)
type AIReviewOutput struct {
	Scores       AIReviewScores `json:"scores"`
	Strengths    []string       `json:"strengths"`
	Issues       []string       `json:"issues"`
	Improvements []string       `json:"improvements"`
}

type AIReviewScores struct {
	CodeQuality  int `json:"code_quality"`
	Constraints  int `json:"constraints"`
	Architecture int `json:"architecture"`
}
