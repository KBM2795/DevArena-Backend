package db

import (
	"context"
	"fmt"
	"time"
)

// CreateSubmission inserts a new submission with "pending" status
func (db *Database) CreateSubmission(userID, challengeID, repoURL, branch string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID from clerk user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		userID,
	).Scan(&internalUserID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Generate submission ID
	submissionID := fmt.Sprintf("sub-%s-%d", challengeID, time.Now().UnixMilli())

	// Insert submission
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO submissions (id, user_id, challenge_id, repo_url, branch, evaluation_status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`, submissionID, internalUserID, challengeID, repoURL, branch)
	if err != nil {
		return "", fmt.Errorf("failed to create submission: %w", err)
	}

	return submissionID, nil
}

// UpdateSubmissionStatus updates the evaluation_status and optional error message
func (db *Database) UpdateSubmissionStatus(submissionID, status string, errorMsg *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE submissions SET evaluation_status = $1`
	args := []interface{}{status}
	argIdx := 2

	// Set timestamps based on status
	switch status {
	case "testing", "reviewing":
		if status == "testing" {
			query += fmt.Sprintf(", evaluation_started_at = NOW()")
		}
	case "completed", "failed":
		query += fmt.Sprintf(", evaluation_completed_at = NOW()")
	}

	// Optional error message
	if errorMsg != nil {
		query += fmt.Sprintf(", error_message = $%d", argIdx)
		args = append(args, *errorMsg)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, submissionID)

	_, err := db.Pool.Exec(ctx, query, args...)
	return err
}

// GetSubmissionByID returns a submission by its ID
func (db *Database) GetSubmissionByID(submissionID string) (*SubmissionDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sub SubmissionDetail
	err := db.Pool.QueryRow(ctx, `
		SELECT 
			s.id, s.user_id, s.challenge_id, s.repo_url, s.branch,
			s.evaluation_status, COALESCE(s.error_message, ''),
			c.title, c.max_score,
			COALESCE(ct.test_repo_url, '') as test_repo_url
		FROM submissions s
		JOIN challenges c ON s.challenge_id = c.id
		LEFT JOIN challenge_templates ct ON c.id = ct.challenge_id
		WHERE s.id = $1
	`, submissionID).Scan(
		&sub.ID, &sub.UserID, &sub.ChallengeID, &sub.RepoURL, &sub.Branch,
		&sub.EvaluationStatus, &sub.ErrorMessage,
		&sub.ChallengeTitle, &sub.MaxScore,
		&sub.TestRepoURL,
	)
	if err != nil {
		return nil, err
	}

	return &sub, nil
}

// SubmissionDetail holds all info the evaluator needs
type SubmissionDetail struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	ChallengeID      string `json:"challenge_id"`
	RepoURL          string `json:"repo_url"`
	Branch           string `json:"branch"`
	EvaluationStatus string `json:"evaluation_status"`
	ErrorMessage     string `json:"error_message,omitempty"`
	ChallengeTitle   string `json:"challenge_title"`
	MaxScore         int    `json:"max_score"`
	TestRepoURL      string `json:"test_repo_url"`
}

// GetSubmissionStatus returns just the status for polling
func (db *Database) GetSubmissionStatus(submissionID, clerkUserID string) (*SubmissionStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var resp SubmissionStatusResponse
	err := db.Pool.QueryRow(ctx, `
		SELECT 
			s.id, s.evaluation_status, COALESCE(s.error_message, ''),
			COALESCE(sc.final_score, 0), c.max_score
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		JOIN challenges c ON s.challenge_id = c.id
		LEFT JOIN submission_scores sc ON s.id = sc.submission_id
		WHERE s.id = $1 AND u.clerk_user_id = $2
	`, submissionID, clerkUserID).Scan(
		&resp.ID, &resp.EvaluationStatus, &resp.ErrorMessage,
		&resp.FinalScore, &resp.MaxScore,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// SubmissionStatusResponse is returned when polling status
type SubmissionStatusResponse struct {
	ID               string `json:"id"`
	EvaluationStatus string `json:"evaluation_status"`
	ErrorMessage     string `json:"error_message,omitempty"`
	FinalScore       int    `json:"final_score"`
	MaxScore         int    `json:"max_score"`
}

// RecoverStuckSubmissions resets any submissions stuck in "testing" or "reviewing"
// back to "pending" — called on server startup to recover from crashes
func (db *Database) RecoverStuckSubmissions() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tag, err := db.Pool.Exec(ctx, `
		UPDATE submissions 
		SET evaluation_status = 'pending',
		    error_message = 'Retrying after server restart'
		WHERE evaluation_status IN ('testing', 'reviewing', 'generating')
	`)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

// ClaimNextPending atomically claims the next pending submission for processing
// Uses UPDATE ... RETURNING to prevent two workers from grabbing the same job
func (db *Database) ClaimNextPending() (*SubmissionDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var sub SubmissionDetail
	err := db.Pool.QueryRow(ctx, `
		UPDATE submissions 
		SET evaluation_status = 'testing',
		    evaluation_started_at = NOW()
		WHERE id = (
			SELECT s.id FROM submissions s
			WHERE s.evaluation_status = 'pending'
			ORDER BY s.created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id
	`).Scan(&sub.ID)
	if err != nil {
		return nil, err // no rows = no pending submissions
	}

	// Now fetch the full submission detail
	fullSub, err := db.GetSubmissionByID(sub.ID)
	if err != nil {
		return nil, err
	}

	return fullSub, nil
}

// CountSubmissionsForChallenge returns how many submissions a user has for a challenge
func (db *Database) CountSubmissionsForChallenge(clerkUserID, challengeID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM submissions s
		JOIN users u ON s.user_id = u.id
		WHERE u.clerk_user_id = $1 AND s.challenge_id = $2
		AND s.evaluation_status != 'failed'
	`, clerkUserID, challengeID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SubmissionSummary is a lightweight submission entry for listing
type SubmissionSummary struct {
	ID               string `json:"id"`
	EvaluationStatus string `json:"evaluation_status"`
	FinalScore       int    `json:"final_score"`
	ErrorMessage     string `json:"error_message,omitempty"`
	CreatedAt        string `json:"created_at"`
	Attempt          int    `json:"attempt"`
}

// GetSubmissionsForChallenge returns all submissions for a user+challenge pair
func (db *Database) GetSubmissionsForChallenge(clerkUserID, challengeID string) ([]SubmissionSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.Pool.Query(ctx, `
		SELECT s.id, s.evaluation_status, 
		       COALESCE(sc.final_score, 0),
		       COALESCE(s.error_message, ''),
		       s.created_at
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN submission_scores sc ON s.id = sc.submission_id
		WHERE u.clerk_user_id = $1 AND s.challenge_id = $2
		AND s.evaluation_status != 'failed'
		ORDER BY s.created_at ASC
	`, clerkUserID, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []SubmissionSummary
	attempt := 1
	for rows.Next() {
		var s SubmissionSummary
		var createdAt time.Time
		err := rows.Scan(&s.ID, &s.EvaluationStatus, &s.FinalScore, &s.ErrorMessage, &createdAt)
		if err != nil {
			return nil, err
		}
		s.CreatedAt = createdAt.Format(time.RFC3339)
		s.Attempt = attempt
		attempt++
		subs = append(subs, s)
	}

	return subs, nil
}
