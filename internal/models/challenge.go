package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Difficulty represents challenge difficulty level
type Difficulty string

const (
	DifficultyEasy   Difficulty = "Easy"
	DifficultyMedium Difficulty = "Medium"
	DifficultyHard   Difficulty = "Hard"
)

// ChallengeType represents the type of challenge
type ChallengeType string

const (
	ChallengeTypeProject  ChallengeType = "project"  // Build a full project
	ChallengeTypeFeature  ChallengeType = "feature"  // Add feature to existing project
	ChallengeTypeRefactor ChallengeType = "refactor" // Refactor/improve code
	ChallengeTypeBugfix   ChallengeType = "bugfix"   // Fix bugs in codebase
)

// Challenge represents a DevArena coding challenge
type Challenge struct {
	ID              string        `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Title           string        `json:"title" gorm:"type:varchar(255);not null"`
	Description     string        `json:"description" gorm:"type:text;not null"`
	Difficulty      Difficulty    `json:"difficulty" gorm:"type:varchar(20);not null"`
	Type            ChallengeType `json:"type" gorm:"type:varchar(50);default:project"`
	MaxScore        int           `json:"max_score" gorm:"not null;default:100"`
	RepoTemplateURL string        `json:"repo_template_url" gorm:"type:text"` // Starter template repo
	Requirements    Requirements  `json:"requirements" gorm:"type:jsonb"`     // Review criteria
	TechStack       TechStack     `json:"tech_stack" gorm:"type:jsonb"`       // Expected technologies
	EstimatedHours  int           `json:"estimated_hours" gorm:"default:4"`   // Estimated completion time
	IsPublished     bool          `json:"is_published" gorm:"default:false"`
	CreatedAt       time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time     `json:"updated_at" gorm:"autoUpdateTime"`

	// Computed fields (not stored, calculated at query time)
	SuccessRate     float64 `json:"success_rate" gorm:"-"`
	SubmissionCount int     `json:"submission_count" gorm:"-"`

	// Relationships
	Tags        []Tag        `json:"tags" gorm:"many2many:challenge_tags"`
	Submissions []Submission `json:"submissions,omitempty" gorm:"foreignKey:ChallengeID"`
}

// Requirements represents the review criteria for a challenge
type Requirements []string

func (r Requirements) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *Requirements) Scan(value interface{}) error {
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

// TechStack represents expected technologies for a challenge
type TechStack []string

func (t TechStack) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TechStack) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// ChallengeResponse is the API response for challenges (with computed fields)
type ChallengeResponse struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Difficulty      Difficulty    `json:"difficulty"`
	Type            ChallengeType `json:"type"`
	MaxScore        int           `json:"max_score"`
	SuccessRate     float64       `json:"success_rate"`
	SubmissionCount int           `json:"submission_count"`
	TechStack       []string      `json:"tech_stack"`
	EstimatedHours  int           `json:"estimated_hours"`
	Tags            []string      `json:"tags"`
	IsSolved        bool          `json:"is_solved"`
}
