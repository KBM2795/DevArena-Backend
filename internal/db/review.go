package db

import (
	"context"
	"encoding/json"
	"time"
)

// FullReview represents the complete AI review with submission details
type FullReview struct {
	// Submission info
	SubmissionID string `json:"submission_id"`
	ChallengeID  string `json:"challenge_id"`
	RepoURL      string `json:"repo_url"`
	SubmittedAt  string `json:"submitted_at"`

	// Challenge info
	ChallengeTitle string   `json:"challenge_title"`
	MaxScore       int      `json:"max_score"`
	TechStack      []string `json:"tech_stack"`

	// Review info
	OverallScore int              `json:"overall_score"`
	Categories   []ReviewCategory `json:"categories"`
	Feedback     string           `json:"feedback"`
	Suggestions  []string         `json:"suggestions"`
	ReviewedAt   string           `json:"reviewed_at"`
}

// ReviewCategory represents a scored category
type ReviewCategory struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Notes string `json:"notes"`
}

// GetReviewForChallenge returns the AI review for a user's submission to a specific challenge
func (db *Database) GetReviewForChallenge(clerkUserID string, challengeID string) (*FullReview, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	// Get submission and review details
	query := `
		SELECT 
			s.id as submission_id,
			s.challenge_id,
			s.repo_url,
			TO_CHAR(s.created_at, 'YYYY-MM-DD HH24:MI') as submitted_at,
			c.title as challenge_title,
			c.max_score,
			c.tech_stack,
			COALESCE(r.overall_score, s.score) as overall_score,
			COALESCE(r.categories, '[]'::jsonb) as categories,
			COALESCE(r.feedback, '') as feedback,
			COALESCE(r.suggestions, '[]'::jsonb) as suggestions,
			COALESCE(TO_CHAR(r.reviewed_at, 'YYYY-MM-DD HH24:MI'), TO_CHAR(s.updated_at, 'YYYY-MM-DD HH24:MI')) as reviewed_at
		FROM submissions s
		JOIN challenges c ON s.challenge_id = c.id
		LEFT JOIN ai_reviews r ON s.id = r.submission_id
		WHERE s.user_id = $1 AND s.challenge_id = $2
		ORDER BY s.created_at DESC
		LIMIT 1
	`

	var review FullReview
	var categoriesJSON []byte
	var suggestionsJSON []byte

	err = db.Pool.QueryRow(ctx, query, internalUserID, challengeID).Scan(
		&review.SubmissionID,
		&review.ChallengeID,
		&review.RepoURL,
		&review.SubmittedAt,
		&review.ChallengeTitle,
		&review.MaxScore,
		&review.TechStack,
		&review.OverallScore,
		&categoriesJSON,
		&review.Feedback,
		&suggestionsJSON,
		&review.ReviewedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(categoriesJSON) > 0 {
		if err := json.Unmarshal(categoriesJSON, &review.Categories); err != nil {
			review.Categories = []ReviewCategory{}
		}
	}
	if len(suggestionsJSON) > 0 {
		if err := json.Unmarshal(suggestionsJSON, &review.Suggestions); err != nil {
			review.Suggestions = []string{}
		}
	}

	return &review, nil
}
