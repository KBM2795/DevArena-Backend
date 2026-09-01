package db

import (
	"context"
	"fmt"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// ToggleFollow follows or unfollows a user. Returns (isNowFollowing, newFollowersCount, error).
func (db *Database) ToggleFollow(clerkUserID string, targetInternalID string) (bool, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Get internal user ID for the follower
	var followerID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1", clerkUserID).Scan(&followerID)
	if err != nil {
		return false, 0, fmt.Errorf("user not found: %w", err)
	}

	// 2. Resolve target user ID by id or username
	var resolvedTargetID string
	err = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE id = $1 OR username = $1", targetInternalID).Scan(&resolvedTargetID)
	if err != nil {
		return false, 0, fmt.Errorf("target user not found")
	}
	targetInternalID = resolvedTargetID

	// 3. Prevent self-follow
	if followerID == targetInternalID {
		return false, 0, fmt.Errorf("cannot follow yourself")
	}

	// 4. Check if follow already exists
	var exists bool
	err = db.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)
	`, followerID, targetInternalID).Scan(&exists)
	if err != nil {
		return false, 0, err
	}

	var isFollowing bool
	if exists {
		// Unfollow
		_, err = db.Pool.Exec(ctx, `
			DELETE FROM user_follows WHERE follower_id = $1 AND following_id = $2
		`, followerID, targetInternalID)
		if err != nil {
			return false, 0, fmt.Errorf("failed to unfollow: %w", err)
		}

		// Decrement counts
		_, _ = db.Pool.Exec(ctx, `UPDATE users SET followers_count = GREATEST(followers_count - 1, 0) WHERE id = $1`, targetInternalID)
		_, _ = db.Pool.Exec(ctx, `UPDATE users SET following_count = GREATEST(following_count - 1, 0) WHERE id = $1`, followerID)

		isFollowing = false
	} else {
		// Follow
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO user_follows (follower_id, following_id, created_at)
			VALUES ($1, $2, NOW())
		`, followerID, targetInternalID)
		if err != nil {
			return false, 0, fmt.Errorf("failed to follow: %w", err)
		}

		// Increment counts
		_, _ = db.Pool.Exec(ctx, `UPDATE users SET followers_count = followers_count + 1 WHERE id = $1`, targetInternalID)
		_, _ = db.Pool.Exec(ctx, `UPDATE users SET following_count = following_count + 1 WHERE id = $1`, followerID)

		isFollowing = true

		// Create notification for the target user
		var followerName string
		_ = db.Pool.QueryRow(ctx, "SELECT COALESCE(display_name, username, 'Someone') FROM users WHERE id = $1", followerID).Scan(&followerName)

		var followerUsername string
		_ = db.Pool.QueryRow(ctx, "SELECT COALESCE(username, '') FROM users WHERE id = $1", followerID).Scan(&followerUsername)

		link := ""
		if followerUsername != "" {
			link = "/users/" + followerUsername
		}

		_ = db.CreateNotification(
			targetInternalID,
			"New Follower",
			followerName+" started following you",
			"follow",
			link,
		)
	}

	// 5. Get updated followers count
	var followersCount int
	err = db.Pool.QueryRow(ctx, `SELECT followers_count FROM users WHERE id = $1`, targetInternalID).Scan(&followersCount)
	if err != nil {
		return isFollowing, 0, nil
	}

	return isFollowing, followersCount, nil
}

// GetFollowStatus checks if the requesting user follows the target user
func (db *Database) GetFollowStatus(clerkUserID string, targetInternalID string) (bool, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Resolve target user ID and get follow counts
	var resolvedTargetID string
	var followersCount, followingCount int
	err := db.Pool.QueryRow(ctx, `
		SELECT id, COALESCE(followers_count, 0), COALESCE(following_count, 0) FROM users WHERE id = $1 OR username = $1
	`, targetInternalID).Scan(&resolvedTargetID, &followersCount, &followingCount)
	if err != nil {
		return false, 0, 0, fmt.Errorf("target user not found: %w", err)
	}
	targetInternalID = resolvedTargetID

	// Check if the requesting user follows the target
	var isFollowing bool
	if clerkUserID != "" {
		var followerID string
		err = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1 OR id = $1", clerkUserID).Scan(&followerID)
		if err == nil {
			db.Pool.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)
			`, followerID, targetInternalID).Scan(&isFollowing)
		}
	}

	return isFollowing, followersCount, followingCount, nil
}

// GetFollowers returns a paginated list of a user's followers
func (db *Database) GetFollowers(targetInternalID string, requestingClerkUserID string, page, limit int) ([]models.FollowUser, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Resolve target user ID
	var resolvedTargetID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE id = $1 OR username = $1", targetInternalID).Scan(&resolvedTargetID)
	if err != nil {
		return nil, 0, fmt.Errorf("target user not found: %w", err)
	}
	targetInternalID = resolvedTargetID

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get the requesting user's internal ID for is_following check
	var requestingID string
	if requestingClerkUserID != "" {
		_ = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1 OR id = $1", requestingClerkUserID).Scan(&requestingID)
	}

	// Total count
	var total int
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_follows WHERE following_id = $1`, targetInternalID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch followers with is_following enrichment
	query := `
		SELECT 
			u.id,
			COALESCE(u.username, '') as username,
			COALESCE(u.display_name, '') as display_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio,
			u.total_score,
			u.rank,
			u.challenges_completed,
			(SELECT COUNT(*) FROM submissions s WHERE s.user_id = u.id) as project_count,
			COALESCE(u.followers_count, 0) as followers_count,
			CASE WHEN $3 != '' THEN
				EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $3 AND following_id = u.id)
			ELSE false END as is_following
		FROM user_follows uf
		JOIN users u ON u.id = uf.follower_id
		WHERE uf.following_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $2 OFFSET $4
	`

	rows, err := db.Pool.Query(ctx, query, targetInternalID, limit, requestingID, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.FollowUser
	for rows.Next() {
		var u models.FollowUser
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.Bio,
			&u.TotalScore, &u.Rank, &u.ChallengesCompleted, &u.ProjectCount,
			&u.FollowersCount, &u.IsFollowing,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

// GetFollowing returns a paginated list of users that the target user follows
func (db *Database) GetFollowing(targetInternalID string, requestingClerkUserID string, page, limit int) ([]models.FollowUser, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Resolve target user ID
	var resolvedTargetID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE id = $1 OR username = $1", targetInternalID).Scan(&resolvedTargetID)
	if err != nil {
		return nil, 0, fmt.Errorf("target user not found: %w", err)
	}
	targetInternalID = resolvedTargetID

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	offset := (page - 1) * limit

	var requestingID string
	if requestingClerkUserID != "" {
		_ = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1 OR id = $1", requestingClerkUserID).Scan(&requestingID)
	}

	var total int
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM user_follows WHERE follower_id = $1`, targetInternalID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			u.id,
			COALESCE(u.username, '') as username,
			COALESCE(u.display_name, '') as display_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio,
			u.total_score,
			u.rank,
			u.challenges_completed,
			(SELECT COUNT(*) FROM submissions s WHERE s.user_id = u.id) as project_count,
			COALESCE(u.followers_count, 0) as followers_count,
			CASE WHEN $3 != '' THEN
				EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $3 AND following_id = u.id)
			ELSE false END as is_following
		FROM user_follows uf
		JOIN users u ON u.id = uf.following_id
		WHERE uf.follower_id = $1
		ORDER BY uf.created_at DESC
		LIMIT $2 OFFSET $4
	`

	rows, err := db.Pool.Query(ctx, query, targetInternalID, limit, requestingID, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.FollowUser
	for rows.Next() {
		var u models.FollowUser
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.Bio,
			&u.TotalScore, &u.Rank, &u.ChallengesCompleted, &u.ProjectCount,
			&u.FollowersCount, &u.IsFollowing,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

// GetFollowingFeedSubmissions returns showcase submissions from users that the requesting user follows
func (db *Database) GetFollowingFeedSubmissions(clerkUserID string, page, limit int) ([]SubmissionDetail, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1 OR id = $1", clerkUserID).Scan(&internalUserID)
	if err != nil {
		return nil, 0, fmt.Errorf("user not found: %w", err)
	}

	// Count total submissions from followed users
	var total int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM submissions s
		WHERE s.user_id IN (
			SELECT following_id FROM user_follows WHERE follower_id = $1
		)
	`, internalUserID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Fetch submissions from followed users
	query := `
		SELECT 
			s.id, s.user_id, u.clerk_user_id,
			COALESCE(u.username, '') as username,
			COALESCE(u.display_name, '') as display_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			s.challenge_id,
			COALESCE(c.title, '') as challenge_title,
			COALESCE(s.title, '') as title,
			s.repo_url,
			COALESCE(s.video_url, '') as video_url,
			COALESCE(s.project_description, '') as project_description,
			s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM submission_likes sl WHERE sl.submission_id = s.id) as like_count,
			EXISTS(SELECT 1 FROM submission_likes sl WHERE sl.submission_id = s.id AND sl.user_id = $1) as user_has_liked
		FROM submissions s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.user_id IN (
			SELECT following_id FROM user_follows WHERE follower_id = $1
		)
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Pool.Query(ctx, query, internalUserID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var submissions []SubmissionDetail
	for rows.Next() {
		var sub SubmissionDetail
		if err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ClerkUserID,
			&sub.Username, &sub.DisplayName, &sub.AvatarURL,
			&sub.ChallengeID, &sub.ChallengeTitle,
			&sub.Title, &sub.RepoURL, &sub.VideoURL,
			&sub.ProjectDescription,
			&sub.CreatedAt, &sub.UpdatedAt,
			&sub.LikeCount, &sub.UserHasLiked,
		); err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, sub)
	}

	return submissions, total, nil
}

// SearchUsers searches users by username or display_name using ILIKE
func (db *Database) SearchUsers(queryStr string, requestingClerkUserID string, limit int) ([]models.FollowUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if limit < 1 || limit > 20 {
		limit = 5
	}

	var requestingID string
	if requestingClerkUserID != "" {
		_ = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1 OR id = $1", requestingClerkUserID).Scan(&requestingID)
	}

	searchPattern := "%" + queryStr + "%"

	query := `
		SELECT 
			u.id,
			COALESCE(u.username, '') as username,
			COALESCE(u.display_name, '') as display_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio,
			u.total_score,
			u.rank,
			u.challenges_completed,
			(SELECT COUNT(*) FROM submissions s WHERE s.user_id = u.id) as project_count,
			COALESCE(u.followers_count, 0) as followers_count,
			CASE WHEN $3 != '' THEN
				EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $3 AND following_id = u.id)
			ELSE false END as is_following
		FROM users u
		WHERE u.onboarding_completed = true
			AND (u.username ILIKE $1 OR u.display_name ILIKE $1)
		ORDER BY u.followers_count DESC, u.total_score DESC
		LIMIT $2
	`

	rows, err := db.Pool.Query(ctx, query, searchPattern, limit, requestingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.FollowUser
	for rows.Next() {
		var u models.FollowUser
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.AvatarURL, &u.Bio,
			&u.TotalScore, &u.Rank, &u.ChallengesCompleted, &u.ProjectCount,
			&u.FollowersCount, &u.IsFollowing,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
