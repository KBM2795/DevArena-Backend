package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AIReview represents an AI code review for a submission
type AIReview struct {
	ID           string           `json:"id" gorm:"primaryKey;type:varchar(255)"`
	SubmissionID string           `json:"submission_id" gorm:"type:varchar(255);not null;index"`
	OverallScore int              `json:"overall_score" gorm:"not null;comment:Score 0-100"`
	Categories   ReviewCategories `json:"categories" gorm:"type:jsonb"`
	Feedback     string           `json:"feedback" gorm:"type:text"`
	Suggestions  Suggestions      `json:"suggestions" gorm:"type:jsonb"`
	ReviewedAt   time.Time        `json:"reviewed_at" gorm:"autoCreateTime"`

	// Relationships
	Submission Submission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
}

// ReviewCategory represents a single category score in the AI review
type ReviewCategory struct {
	Name  string `json:"name"`  // e.g., "Code Quality", "Best Practices", "Architecture"
	Score int    `json:"score"` // 0-100
	Notes string `json:"notes"` // Specific feedback for this category
}

// ReviewCategories is a slice of ReviewCategory that implements GORM's scanner/valuer
type ReviewCategories []ReviewCategory

func (r ReviewCategories) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *ReviewCategories) Scan(value interface{}) error {
	if value == nil {
		*r = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

// Suggestions is a string slice that implements GORM's scanner/valuer
type Suggestions []string

func (s Suggestions) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Suggestions) Scan(value interface{}) error {
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

// Standard review categories
const (
	CategoryCodeQuality     = "Code Quality"
	CategoryBestPractices   = "Best Practices"
	CategoryArchitecture    = "Architecture"
	CategoryDocumentation   = "Documentation"
	CategoryTestCoverage    = "Test Coverage"
	CategorySecurity        = "Security"
	CategoryPerformance     = "Performance"
	CategoryMaintainability = "Maintainability"
)
