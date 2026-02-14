package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AIReport represents Step 3 — AI-generated professional feedback report
// Must NOT rescore or reinterpret raw code
type AIReport struct {
	ID                 string      `json:"id" gorm:"primaryKey;type:varchar(255)"`
	SubmissionID       string      `json:"submission_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	SummaryMD          string      `json:"summary_md" gorm:"type:text"`
	DetailedFeedbackMD string      `json:"detailed_feedback_md" gorm:"type:text"` // JSON string with per-category feedback
	DosJSON            StringArray `json:"dos" gorm:"type:jsonb;column:dos_json"`
	DontsJSON          StringArray `json:"donts" gorm:"type:jsonb;column:donts_json"`
	NextStepsJSON      StringArray `json:"next_steps" gorm:"type:jsonb;column:next_steps_json"`
	PromptVersionID    string      `json:"prompt_version_id,omitempty" gorm:"type:varchar(255)"`
	RawResponse        string      `json:"raw_response,omitempty" gorm:"type:text"`
	CreatedAt          time.Time   `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	Submission Submission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
}

// AIReportInput is the structured payload sent to the AI for report generation (Step 3)
type AIReportInput struct {
	EvaluationType string                `json:"evaluation_type"` // "final_report"
	ChallengeInfo  AIReportChallengeInfo `json:"challenge_info"`
	ScoreBreakdown AIReportScores        `json:"score_breakdown"`
	Strengths      []string              `json:"strengths"`
	Issues         []string              `json:"issues"`
	Improvements   []string              `json:"improvements"`
}

type AIReportChallengeInfo struct {
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
}

type AIReportScores struct {
	FinalScore    int `json:"final_score"`
	MaxScore      int `json:"max_score"`
	Functionality int `json:"functionality"`
	CodeQuality   int `json:"code_quality"`
	Constraints   int `json:"constraints"`
	Architecture  int `json:"architecture"`
}

// AIReportOutput is the expected AI response (strict JSON)
type AIReportOutput struct {
	Summary          string            `json:"summary"`
	DetailedFeedback map[string]string `json:"detailed_feedback"`
	Dos              []string          `json:"dos"`
	Donts            []string          `json:"donts"`
	NextSteps        []string          `json:"next_steps"`
}

// DetailedFeedback is used to parse/store the detailed_feedback_md JSON
type DetailedFeedback struct {
	Functionality string `json:"functionality"`
	CodeQuality   string `json:"code_quality"`
	Constraints   string `json:"constraints"`
	Architecture  string `json:"architecture"`
}

func (d DetailedFeedback) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DetailedFeedback) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}
