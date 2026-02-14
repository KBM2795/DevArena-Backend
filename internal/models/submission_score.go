package models

import "time"

// SubmissionScore represents the final computed score
// Calculated in Go only — never by AI
// FinalScore = FunctionalityScore + CodeQualityScore + ConstraintScore + ArchitectureScore
type SubmissionScore struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(255)"`
	SubmissionID       string    `json:"submission_id" gorm:"type:varchar(255);not null;uniqueIndex"`
	FunctionalityScore int       `json:"functionality_score" gorm:"not null;default:0"`
	CodeQualityScore   int       `json:"code_quality_score" gorm:"not null;default:0"`
	ConstraintScore    int       `json:"constraint_score" gorm:"not null;default:0"`
	ArchitectureScore  int       `json:"architecture_score" gorm:"not null;default:0"`
	FinalScore         int       `json:"final_score" gorm:"not null;default:0"`
	MaxScore           int       `json:"max_score" gorm:"not null;default:100"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Submission Submission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
}

// Calculate computes the final score from all components
func (s *SubmissionScore) Calculate() {
	s.FinalScore = s.FunctionalityScore + s.CodeQualityScore + s.ConstraintScore + s.ArchitectureScore
	if s.FinalScore > s.MaxScore {
		s.FinalScore = s.MaxScore
	}
}

// CalculateFromParts builds the score from test results and AI review using Rubric weights
func (s *SubmissionScore) CalculateFromParts(functionalityScore int, review *AIReview, rubric *Rubric) {
	s.FunctionalityScore = functionalityScore
	s.CodeQualityScore = review.CodeQualityScore
	s.ConstraintScore = review.ConstraintScore
	s.ArchitectureScore = review.ArchitectureScore

	// Calculate weighted score
	// Formula: (Score * Weight) / 100
	// We assume input scores are 0-100 and weights sum to 100

	if rubric == nil {
		// Fallback to default weights if no rubric provided
		// Default: Func 50%, Quality 25%, Constraints 15%, Arch 10%
		rubric = &Rubric{
			Functionality: 50,
			CodeQuality:   25,
			Constraints:   15,
			Architecture:  10,
		}
	}

	weightedScore := 0.0
	weightedScore += float64(s.FunctionalityScore) * (float64(rubric.Functionality) / 100.0)
	weightedScore += float64(s.CodeQualityScore) * (float64(rubric.CodeQuality) / 100.0)
	weightedScore += float64(s.ConstraintScore) * (float64(rubric.Constraints) / 100.0)
	weightedScore += float64(s.ArchitectureScore) * (float64(rubric.Architecture) / 100.0)

	s.FinalScore = int(weightedScore)
	if s.FinalScore > s.MaxScore {
		s.FinalScore = s.MaxScore
	}
}
