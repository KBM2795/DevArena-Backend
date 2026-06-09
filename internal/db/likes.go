package db

import (
	"context"
	"fmt"
	"time"
)

// ToggleLike toggles the like state on a project showcase for a user
// Returns (liked bool, likeCount int, err error)
func (db *Database) ToggleLike(clerkUserID string, submissionID string) (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Get internal user ID for the user performing the like
	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return false, 0, fmt.Errorf("user not found: %w", err)
	}

	// 2. Get owner user ID of the submission to update their stats later
	var ownerUserID string
	err = db.Pool.QueryRow(ctx, "SELECT user_id FROM submissions WHERE id = $1", submissionID).Scan(&ownerUserID)
	if err != nil {
		return false, 0, fmt.Errorf("submission not found: %w", err)
	}

	// 3. Check if like exists
	var exists bool
	err = db.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM submission_likes WHERE user_id = $1 AND submission_id = $2)
	`, internalUserID, submissionID).Scan(&exists)
	if err != nil {
		return false, 0, err
	}

	var liked bool
	if exists {
		// Unlike
		_, err = db.Pool.Exec(ctx, `
			DELETE FROM submission_likes WHERE user_id = $1 AND submission_id = $2
		`, internalUserID, submissionID)
		if err != nil {
			return false, 0, fmt.Errorf("failed to unlike: %w", err)
		}
		liked = false
	} else {
		// Like
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO submission_likes (user_id, submission_id, created_at)
			VALUES ($1, $2, NOW())
		`, internalUserID, submissionID)
		if err != nil {
			return false, 0, fmt.Errorf("failed to like: %w", err)
		}
		liked = true
	}

	// 4. Get updated like count
	var likeCount int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM submission_likes WHERE submission_id = $1
	`, submissionID).Scan(&likeCount)
	if err != nil {
		return liked, 0, err
	}

	// 5. Update submission owner's global score and stats since like count changed
	_ = db.UpdateUserStats(ctx, ownerUserID)

	return liked, likeCount, nil
}

// GetLikeStatus returns the total like count for a submission and whether a specific user liked it
func (db *Database) GetLikeStatus(clerkUserID string, submissionID string) (int, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var likeCount int
	err := db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM submission_likes WHERE submission_id = $1
	`, submissionID).Scan(&likeCount)
	if err != nil {
		return 0, false, err
	}

	var hasLiked bool
	if clerkUserID != "" {
		var internalUserID string
		err = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&internalUserID)
		if err == nil {
			db.Pool.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM submission_likes WHERE user_id = $1 AND submission_id = $2)
			`, internalUserID, submissionID).Scan(&hasLiked)
		}
	}

	return likeCount, hasLiked, nil
}
