package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// FullReview represents the complete evaluation result for a submission
// Combines data from: submission, submission_scores, ai_reviews, ai_reports, submission_test_results
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

	// Evaluation status
	EvaluationStatus string `json:"evaluation_status"`

	// Final Score (from submission_scores — computed in Go)
	FinalScore         int `json:"final_score"`
	FunctionalityScore int `json:"functionality_score"`
	CodeQualityScore   int `json:"code_quality_score"`
	ConstraintScore    int `json:"constraint_score"`
	ArchitectureScore  int `json:"architecture_score"`

	// Test Results (from submission_test_results — Step 1)
	TestResults *TestResultSummary `json:"test_results,omitempty"`

	// AI Review (from ai_reviews — Step 2)
	Strengths    []string `json:"strengths"`
	Issues       []string `json:"issues"`
	Improvements []string `json:"improvements"`

	// AI Report (from ai_reports — Step 3)
	Report *ReportSummary `json:"report,omitempty"`

	ReviewedAt string `json:"reviewed_at"`
}

// TestResultSummary is a subset of test result data for the review response
type TestResultSummary struct {
	BuildSuccess bool   `json:"build_success"`
	TestsTotal   int    `json:"tests_total"`
	TestsPassed  int    `json:"tests_passed"`
	TestsFailed  int    `json:"tests_failed"`
	ExecutionMs  int    `json:"execution_time_ms"`
	RawOutput    string `json:"raw_output,omitempty"`
}

// ReportSummary is a subset of AI report data for the review response
type ReportSummary struct {
	SummaryMD          string   `json:"summary_md"`
	DetailedFeedbackMD string   `json:"detailed_feedback_md"`
	Dos                []string `json:"dos"`
	Donts              []string `json:"donts"`
	NextSteps          []string `json:"next_steps"`
}

// GetReviewForChallenge returns the full evaluation result for a user's submission to a challenge
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

	// Get submission + score + challenge info
	query := `
		SELECT 
			s.id as submission_id,
			s.challenge_id,
			s.repo_url,
			TO_CHAR(s.created_at, 'YYYY-MM-DD HH24:MI') as submitted_at,
			c.title as challenge_title,
			c.max_score,
			c.tech_stack,
			s.evaluation_status,
			COALESCE(sc.final_score, 0) as final_score,
			COALESCE(sc.functionality_score, 0) as functionality_score,
			COALESCE(sc.code_quality_score, 0) as code_quality_score,
			COALESCE(sc.constraint_score, 0) as constraint_score,
			COALESCE(sc.architecture_score, 0) as architecture_score,
			COALESCE(TO_CHAR(
				GREATEST(
					COALESCE(r.reviewed_at, s.created_at),
					COALESCE(s.evaluation_completed_at, s.created_at)
				), 'YYYY-MM-DD HH24:MI'
			), '') as reviewed_at
		FROM submissions s
		JOIN challenges c ON s.challenge_id = c.id
		LEFT JOIN submission_scores sc ON s.id = sc.submission_id
		LEFT JOIN ai_reviews r ON s.id = r.submission_id
		WHERE s.user_id = $1 AND s.challenge_id = $2
		ORDER BY s.created_at DESC
		LIMIT 1
	`

	var review FullReview
	err = db.Pool.QueryRow(ctx, query, internalUserID, challengeID).Scan(
		&review.SubmissionID,
		&review.ChallengeID,
		&review.RepoURL,
		&review.SubmittedAt,
		&review.ChallengeTitle,
		&review.MaxScore,
		&review.TechStack,
		&review.EvaluationStatus,
		&review.FinalScore,
		&review.FunctionalityScore,
		&review.CodeQualityScore,
		&review.ConstraintScore,
		&review.ArchitectureScore,
		&review.ReviewedAt,
	)
	if err != nil {
		return nil, err
	}

	// Fetch AI review details (strengths, issues, improvements)
	var strengthsJSON, issuesJSON, improvementsJSON []byte
	aiReviewQuery := `
		SELECT 
			COALESCE(strengths_json, '[]'::jsonb),
			COALESCE(issues_json, '[]'::jsonb),
			COALESCE(improvements_json, '[]'::jsonb)
		FROM ai_reviews
		WHERE submission_id = $1
		ORDER BY reviewed_at DESC
		LIMIT 1
	`
	err = db.Pool.QueryRow(ctx, aiReviewQuery, review.SubmissionID).Scan(
		&strengthsJSON, &issuesJSON, &improvementsJSON,
	)
	if err == nil {
		json.Unmarshal(strengthsJSON, &review.Strengths)
		json.Unmarshal(issuesJSON, &review.Issues)
		json.Unmarshal(improvementsJSON, &review.Improvements)
	}
	if review.Strengths == nil {
		review.Strengths = []string{}
	}
	if review.Issues == nil {
		review.Issues = []string{}
	}
	if review.Improvements == nil {
		review.Improvements = []string{}
	}

	// Fetch test results (Step 1)
	testQuery := `
		SELECT build_success, tests_total, tests_passed, tests_failed, execution_time_ms, COALESCE(raw_output, '')
		FROM submission_test_results
		WHERE submission_id = $1
	`
	var tr TestResultSummary
	err = db.Pool.QueryRow(ctx, testQuery, review.SubmissionID).Scan(
		&tr.BuildSuccess, &tr.TestsTotal, &tr.TestsPassed, &tr.TestsFailed, &tr.ExecutionMs, &tr.RawOutput,
	)
	if err == nil {
		review.TestResults = &tr
	}

	// Fetch AI report (Step 3)
	reportQuery := `
		SELECT summary_md, COALESCE(detailed_feedback_md, ''), 
			COALESCE(dos_json, '[]'::jsonb), 
			COALESCE(donts_json, '[]'::jsonb), 
			COALESCE(next_steps_json, '[]'::jsonb)
		FROM ai_reports
		WHERE submission_id = $1
	`
	var rpt ReportSummary
	var dosJSON, dontsJSON, nextStepsJSON []byte
	err = db.Pool.QueryRow(ctx, reportQuery, review.SubmissionID).Scan(
		&rpt.SummaryMD, &rpt.DetailedFeedbackMD, &dosJSON, &dontsJSON, &nextStepsJSON,
	)
	if err == nil {
		json.Unmarshal(dosJSON, &rpt.Dos)
		json.Unmarshal(dontsJSON, &rpt.Donts)
		json.Unmarshal(nextStepsJSON, &rpt.NextSteps)
		if rpt.Dos == nil {
			rpt.Dos = []string{}
		}
		if rpt.Donts == nil {
			rpt.Donts = []string{}
		}
		if rpt.NextSteps == nil {
			rpt.NextSteps = []string{}
		}
		review.Report = &rpt
	}

	return &review, nil
}

// SaveAIReview inserts a new AI review record
func (db *Database) SaveAIReview(ctx context.Context, r *models.AIReview) error {
	query := `
		INSERT INTO ai_reviews (
			id, submission_id, 
			code_quality_score, constraint_score, architecture_score,
			strengths_json, issues_json, improvements_json,
			prompt_version_id, raw_response
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	// Handle optional foreign key
	var promptVersionID interface{} = r.PromptVersionID
	if r.PromptVersionID == "" {
		promptVersionID = nil
	}

	_, err := db.Pool.Exec(ctx, query,
		r.SubmissionID,
		r.CodeQualityScore, r.ConstraintScore, r.ArchitectureScore,
		r.StrengthsJSON, r.IssuesJSON, r.ImprovementsJSON,
		promptVersionID, r.RawResponse,
	)
	return err
}
