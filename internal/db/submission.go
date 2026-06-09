package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SubmissionDetail holds all info needed for project showcases
type SubmissionDetail struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	ClerkUserID        string     `json:"clerk_user_id"`
	Username           string     `json:"username"`
	DisplayName        string     `json:"display_name"`
	AvatarURL          string     `json:"avatar_url"`
	ChallengeID        *string    `json:"challenge_id,omitempty"`
	ChallengeTitle     string     `json:"challenge_title,omitempty"`
	Title              string     `json:"title"`
	RepoURL            string     `json:"repo_url"`
	VideoURL           string     `json:"video_url"`
	ProjectDescription string     `json:"project_description"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LikeCount          int        `json:"like_count"`
	UserHasLiked       bool       `json:"user_has_liked"`
}

// CreateOrUpdateSubmission creates a new project showcase or updates it if it's challenge-based and already exists
func (db *Database) CreateOrUpdateSubmission(clerkUserID string, challengeID *string, title, repoURL, videoURL, description string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID from clerk user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		return "", false, fmt.Errorf("user not found: %w", err)
	}

	// For challenge-based submissions, check if the user has already submitted a project
	if challengeID != nil && *challengeID != "" {
		var existingID string
		err := db.Pool.QueryRow(ctx, `
			SELECT id FROM submissions 
			WHERE user_id = $1 AND challenge_id = $2
		`, internalUserID, *challengeID).Scan(&existingID)

		if err == nil {
			// Submission exists -> Update it (Edit Mode)
			_, err = db.Pool.Exec(ctx, `
				UPDATE submissions 
				SET title = $1, repo_url = $2, video_url = $3, project_description = $4, updated_at = NOW()
				WHERE id = $5
			`, title, repoURL, videoURL, description, existingID)
			if err != nil {
				return "", false, fmt.Errorf("failed to update submission: %w", err)
			}
			return existingID, true, nil
		}
	}

	// Create new submission (either open showcase or first-time challenge submission)
	submissionID := fmt.Sprintf("sub-%s", uuid.New().String()[:18])

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO submissions (id, user_id, challenge_id, title, repo_url, video_url, project_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`, submissionID, internalUserID, challengeID, title, repoURL, videoURL, description)
	if err != nil {
		return "", false, fmt.Errorf("failed to create submission: %w", err)
	}

	return submissionID, false, nil
}

// GetSubmissionByID returns a single submission by its ID, with like counts and user like status
func (db *Database) GetSubmissionByID(submissionID string, currentClerkUserID string) (*SubmissionDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID for like checking
	var currentUserID string
	if currentClerkUserID != "" {
		_ = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", currentClerkUserID).Scan(&currentUserID)
	}

	var sub SubmissionDetail
	var challengeTitle sql.NullString

	query := `
		SELECT 
			s.id, s.user_id, u.clerk_user_id, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
			s.challenge_id, c.title as challenge_title, COALESCE(s.title, ''), s.repo_url, COALESCE(s.video_url, ''), COALESCE(s.project_description, ''),
			s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM submission_likes WHERE submission_id = s.id) as like_count,
			EXISTS(SELECT 1 FROM submission_likes WHERE submission_id = s.id AND user_id = $2) as user_has_liked
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN challenges c ON s.challenge_id = c.id
		WHERE s.id = $1
	`
	err := db.Pool.QueryRow(ctx, query, submissionID, currentUserID).Scan(
		&sub.ID, &sub.UserID, &sub.ClerkUserID, &sub.Username, &sub.DisplayName, &sub.AvatarURL,
		&sub.ChallengeID, &challengeTitle, &sub.Title, &sub.RepoURL, &sub.VideoURL, &sub.ProjectDescription,
		&sub.CreatedAt, &sub.UpdatedAt,
		&sub.LikeCount, &sub.UserHasLiked,
	)
	if err != nil {
		return nil, err
	}

	if challengeTitle.Valid {
		sub.ChallengeTitle = challengeTitle.String
	}

	return &sub, nil
}

// GetSubmissionsForChallenge returns all user submissions for a specific challenge (public)
func (db *Database) GetSubmissionsForChallenge(challengeID string, currentClerkUserID string) ([]SubmissionDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID for like checking
	var currentUserID string
	if currentClerkUserID != "" {
		_ = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", currentClerkUserID).Scan(&currentUserID)
	}

	query := `
		SELECT 
			s.id, s.user_id, u.clerk_user_id, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
			s.challenge_id, c.title as challenge_title, COALESCE(s.title, ''), s.repo_url, COALESCE(s.video_url, ''), COALESCE(s.project_description, ''),
			s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM submission_likes WHERE submission_id = s.id) as like_count,
			EXISTS(SELECT 1 FROM submission_likes WHERE submission_id = s.id AND user_id = $2) as user_has_liked
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN challenges c ON s.challenge_id = c.id
		WHERE s.challenge_id = $1
		ORDER BY like_count DESC, s.created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query, challengeID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []SubmissionDetail
	for rows.Next() {
		var sub SubmissionDetail
		var challengeTitle sql.NullString
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ClerkUserID, &sub.Username, &sub.DisplayName, &sub.AvatarURL,
			&sub.ChallengeID, &challengeTitle, &sub.Title, &sub.RepoURL, &sub.VideoURL, &sub.ProjectDescription,
			&sub.CreatedAt, &sub.UpdatedAt,
			&sub.LikeCount, &sub.UserHasLiked,
		)
		if err != nil {
			return nil, err
		}
		if challengeTitle.Valid {
			sub.ChallengeTitle = challengeTitle.String
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

// GetAllShowcases returns ALL paginated submissions (open projects + challenge submissions).
// search is case-insensitive and filters title, description, username, display_name, challenge title.
// Pass limit=0 to return all results (no limit).
func (db *Database) GetOpenShowcases(currentClerkUserID string, search string, offset, limit int) ([]SubmissionDetail, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID for like checking
	var currentUserID string
	if currentClerkUserID != "" {
		_ = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", currentClerkUserID).Scan(&currentUserID)
	}

	// Count query — $1 = search (empty string means "no filter")
	countQuery := `
		SELECT COUNT(*)
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN challenges c ON s.challenge_id = c.id
		WHERE (
			$1 = ''
			OR s.title ILIKE '%' || $1 || '%'
			OR s.project_description ILIKE '%' || $1 || '%'
			OR u.username ILIKE '%' || $1 || '%'
			OR u.display_name ILIKE '%' || $1 || '%'
			OR c.title ILIKE '%' || $1 || '%'
		)
	`

	var total int
	if err := db.Pool.QueryRow(ctx, countQuery, search).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data query — params: $1=currentUserID, $2=search, $3=limit, $4=offset
	dataQuery := `
		SELECT
			s.id, s.user_id, u.clerk_user_id, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
			s.challenge_id, c.title as challenge_title, COALESCE(s.title, ''), s.repo_url, COALESCE(s.video_url, ''), COALESCE(s.project_description, ''),
			s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM submission_likes WHERE submission_id = s.id) as like_count,
			EXISTS(SELECT 1 FROM submission_likes WHERE submission_id = s.id AND user_id = $1) as user_has_liked
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN challenges c ON s.challenge_id = c.id
		WHERE (
			$2 = ''
			OR s.title ILIKE '%' || $2 || '%'
			OR s.project_description ILIKE '%' || $2 || '%'
			OR u.username ILIKE '%' || $2 || '%'
			OR u.display_name ILIKE '%' || $2 || '%'
			OR c.title ILIKE '%' || $2 || '%'
		)
		ORDER BY like_count DESC, s.created_at DESC
		LIMIT NULLIF($3::bigint, 0)
		OFFSET $4
	`

	rows, err := db.Pool.Query(ctx, dataQuery, currentUserID, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var subs []SubmissionDetail
	for rows.Next() {
		var sub SubmissionDetail
		var challengeTitle sql.NullString
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ClerkUserID, &sub.Username, &sub.DisplayName, &sub.AvatarURL,
			&sub.ChallengeID, &challengeTitle, &sub.Title, &sub.RepoURL, &sub.VideoURL, &sub.ProjectDescription,
			&sub.CreatedAt, &sub.UpdatedAt,
			&sub.LikeCount, &sub.UserHasLiked,
		)
		if err != nil {
			return nil, 0, err
		}
		if challengeTitle.Valid {
			sub.ChallengeTitle = challengeTitle.String
		}
		subs = append(subs, sub)
	}

	return subs, total, nil
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
	`, clerkUserID, challengeID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetUserSubmissions returns all showcases submitted by a specific user
func (db *Database) GetUserSubmissions(clerkUserID string) ([]SubmissionDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			s.id, s.user_id, u.clerk_user_id, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
			s.challenge_id, c.title as challenge_title, COALESCE(s.title, ''), s.repo_url, COALESCE(s.video_url, ''), COALESCE(s.project_description, ''),
			s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM submission_likes WHERE submission_id = s.id) as like_count,
			EXISTS(SELECT 1 FROM submission_likes WHERE submission_id = s.id AND user_id = $2) as user_has_liked
		FROM submissions s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN challenges c ON s.challenge_id = c.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query, internalUserID, internalUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []SubmissionDetail
	for rows.Next() {
		var sub SubmissionDetail
		var challengeTitle sql.NullString
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ClerkUserID, &sub.Username, &sub.DisplayName, &sub.AvatarURL,
			&sub.ChallengeID, &challengeTitle, &sub.Title, &sub.RepoURL, &sub.VideoURL, &sub.ProjectDescription,
			&sub.CreatedAt, &sub.UpdatedAt,
			&sub.LikeCount, &sub.UserHasLiked,
		)
		if err != nil {
			return nil, err
		}
		if challengeTitle.Valid {
			sub.ChallengeTitle = challengeTitle.String
		}
		subs = append(subs, sub)
	}

	return subs, nil
}
