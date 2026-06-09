package db

import (
	"context"
	"fmt"
	"time"
)

// CommentDetail represents a structured chat comment
type CommentDetail struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ClerkUserID  string    `json:"clerk_user_id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url"`
	ChallengeID  *string   `json:"challenge_id,omitempty"`
	SubmissionID *string   `json:"submission_id,omitempty"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetComments retrieves comments. If submissionID is provided, gets project chat. If challengeID is provided and submissionID is nil, gets challenge chat. If both are nil, gets global chat.
func (db *Database) GetComments(challengeID, submissionID *string) ([]CommentDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rows interface{}
	var query string
	var err error

	baseQuery := `
		SELECT 
			c.id, c.user_id, u.clerk_user_id, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
			c.challenge_id, c.submission_id, c.message, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
	`

	if submissionID != nil && *submissionID != "" {
		query = baseQuery + " WHERE c.submission_id = $1 ORDER BY c.created_at ASC"
		rows, err = db.Pool.Query(ctx, query, *submissionID)
	} else if challengeID != nil && *challengeID != "" {
		query = baseQuery + " WHERE c.challenge_id = $1 AND c.submission_id IS NULL ORDER BY c.created_at ASC"
		rows, err = db.Pool.Query(ctx, query, *challengeID)
	} else {
		query = baseQuery + " WHERE c.challenge_id IS NULL AND c.submission_id IS NULL ORDER BY c.created_at ASC LIMIT 100"
		rows, err = db.Pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}

	pgxRows, ok := rows.(interface {
		Next() bool
		Scan(dest ...any) error
		Close()
		Err() error
	})
	if !ok {
		return nil, fmt.Errorf("unexpected rows type")
	}
	defer pgxRows.Close()

	var comments []CommentDetail
	for pgxRows.Next() {
		var comm CommentDetail
		err := pgxRows.Scan(
			&comm.ID, &comm.UserID, &comm.ClerkUserID, &comm.Username, &comm.DisplayName, &comm.AvatarURL,
			&comm.ChallengeID, &comm.SubmissionID, &comm.Message, &comm.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comm)
	}

	return comments, pgxRows.Err()
}

// SaveComment inserts a new comment into the database
func (db *Database) SaveComment(clerkUserID string, challengeID, submissionID *string, message string) (*CommentDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var commentID string
	query := `
		INSERT INTO comments (user_id, challenge_id, submission_id, message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	err = db.Pool.QueryRow(ctx, query, internalUserID, challengeID, submissionID, message).Scan(&commentID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert comment: %w", err)
	}

	// Fetch detail of inserted comment
	var comm CommentDetail
	fetchQuery := `
		SELECT 
			c.id, c.user_id, u.clerk_user_id, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
			c.challenge_id, c.submission_id, c.message, c.created_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`
	err = db.Pool.QueryRow(ctx, fetchQuery, commentID).Scan(
		&comm.ID, &comm.UserID, &comm.ClerkUserID, &comm.Username, &comm.DisplayName, &comm.AvatarURL,
		&comm.ChallengeID, &comm.SubmissionID, &comm.Message, &comm.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load saved comment: %w", err)
	}

	return &comm, nil
}
