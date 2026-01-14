package models

import "time"

// Difficulty represents challenge difficulty level
type Difficulty string

const (
	DifficultyEasy   Difficulty = "Easy"
	DifficultyMedium Difficulty = "Medium"
	DifficultyHard   Difficulty = "Hard"
)

// Challenge represents a coding challenge
type Challenge struct {
	ID           string     `json:"id" gorm:"primaryKey;type:varchar(255)"`
	Title        string     `json:"title" gorm:"type:varchar(255);not null"`
	Description  string     `json:"description" gorm:"type:text;not null"`
	Difficulty   Difficulty `json:"difficulty" gorm:"type:varchar(20);not null"`
	MaxScore     int        `json:"max_score" gorm:"not null"`
	TimeLimit    int        `json:"time_limit" gorm:"comment:Time limit in seconds"`
	MemoryLimit  int        `json:"memory_limit" gorm:"comment:Memory limit in MB"`
	StarterCode  string     `json:"starter_code" gorm:"type:text"`
	SolutionCode string     `json:"solution_code,omitempty" gorm:"type:text"`
	IsPublished  bool       `json:"is_published" gorm:"default:false"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`

	// Computed fields (not stored, calculated at query time)
	SuccessRate float64 `json:"success_rate" gorm:"-"`

	// Relationships
	Tags        []Tag        `json:"tags" gorm:"many2many:challenge_tags"`
	TestCases   []TestCase   `json:"test_cases,omitempty" gorm:"foreignKey:ChallengeID"`
	Submissions []Submission `json:"submissions,omitempty" gorm:"foreignKey:ChallengeID"`
}

// ChallengeResponse is the API response for challenges (with computed fields)
type ChallengeResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Difficulty  Difficulty `json:"difficulty"`
	MaxScore    int        `json:"max_score"`
	SuccessRate float64    `json:"success_rate"`
	Tags        []string   `json:"tags"`
	IsSolved    bool       `json:"is_solved"`
}
