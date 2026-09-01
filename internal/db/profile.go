package db

import (
	"context"
	"encoding/json"
	"time"
)

// UserProfile contains safe user data for API responses
// NOTE: Does NOT include email, clerk_user_id, or internal id for security
type UserProfile struct {
	Username            string   `json:"username"`
	DisplayName         string   `json:"display_name"`
	AvatarURL           string   `json:"avatar_url"`
	Bio                 string   `json:"bio"`
	GithubUsername      string   `json:"github_username,omitempty"`
	GithubConnected     bool     `json:"github_connected"`
	CurrentStreak       int      `json:"current_streak"`
	LongestStreak       int      `json:"longest_streak"`
	TotalScore          int      `json:"total_score"`
	Rank                int      `json:"rank"`
	ChallengesCompleted int      `json:"challenges_completed"`
	FollowersCount      int      `json:"followers_count"`
	FollowingCount      int      `json:"following_count"`
	Technologies        []string `json:"technologies"`
}

// PublicUserProfile extends UserProfile with follow status and content for the public profile page
type PublicUserProfile struct {
	ID                  string             `json:"id"`
	Username            string             `json:"username"`
	DisplayName         string             `json:"display_name"`
	AvatarURL           string             `json:"avatar_url"`
	Bio                 string             `json:"bio"`
	GithubUsername      string             `json:"github_username,omitempty"`
	GithubConnected     bool               `json:"github_connected"`
	CurrentStreak       int                `json:"current_streak"`
	LongestStreak       int                `json:"longest_streak"`
	TotalScore          int                `json:"total_score"`
	Rank                int                `json:"rank"`
	ChallengesCompleted int                `json:"challenges_completed"`
	FollowersCount      int                `json:"followers_count"`
	FollowingCount      int                `json:"following_count"`
	IsFollowing         bool               `json:"is_following"`
	Technologies        []string           `json:"technologies"`
	Submissions         []SubmissionDetail `json:"submissions"`
}

// GetUserProfile retrieves the user's public profile by clerk_user_id
func (db *Database) GetUserProfile(clerkUserID string) (*UserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profile := &UserProfile{}
	var technologiesJSON []byte

	// Query user data and join with starter_packs for technologies
	query := `
		SELECT 
			COALESCE(u.username, '') as username,
			COALESCE(u.display_name, '') as display_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio,
			COALESCE(u.github_username, '') as github_username,
			u.github_connected,
			u.current_streak,
			u.longest_streak,
			u.total_score,
			u.rank,
			u.challenges_completed,
			COALESCE(u.followers_count, 0) as followers_count,
			COALESCE(u.following_count, 0) as following_count,
			COALESCE(sp.technologies, '[]'::jsonb) as technologies
		FROM users u
		LEFT JOIN starter_packs sp ON sp.user_id = u.id
		WHERE u.clerk_user_id = $1
	`

	err := db.Pool.QueryRow(ctx, query, clerkUserID).Scan(
		&profile.Username,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.Bio,
		&profile.GithubUsername,
		&profile.GithubConnected,
		&profile.CurrentStreak,
		&profile.LongestStreak,
		&profile.TotalScore,
		&profile.Rank,
		&profile.ChallengesCompleted,
		&profile.FollowersCount,
		&profile.FollowingCount,
		&technologiesJSON,
	)
	if err != nil {
		return nil, err
	}

	// Parse technologies JSON array
	if len(technologiesJSON) > 0 {
		if err := json.Unmarshal(technologiesJSON, &profile.Technologies); err != nil {
			profile.Technologies = []string{}
		}
	}

	return profile, nil
}

// GetPublicUserProfile retrieves a user's public profile by username, with follow status relative to the requesting user
func (db *Database) GetPublicUserProfile(username string, requestingClerkUserID string) (*PublicUserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	profile := &PublicUserProfile{}
	var technologiesJSON []byte

	query := `
		SELECT 
			u.id,
			COALESCE(u.username, '') as username,
			COALESCE(u.display_name, '') as display_name,
			COALESCE(u.avatar_url, '') as avatar_url,
			COALESCE(u.bio, '') as bio,
			COALESCE(u.github_username, '') as github_username,
			u.github_connected,
			u.current_streak,
			u.longest_streak,
			u.total_score,
			u.rank,
			u.challenges_completed,
			COALESCE(u.followers_count, 0) as followers_count,
			COALESCE(u.following_count, 0) as following_count,
			COALESCE(sp.technologies, '[]'::jsonb) as technologies
		FROM users u
		LEFT JOIN starter_packs sp ON sp.user_id = u.id
		WHERE u.username = $1 OR u.id = $1
	`

	err := db.Pool.QueryRow(ctx, query, username).Scan(
		&profile.ID,
		&profile.Username,
		&profile.DisplayName,
		&profile.AvatarURL,
		&profile.Bio,
		&profile.GithubUsername,
		&profile.GithubConnected,
		&profile.CurrentStreak,
		&profile.LongestStreak,
		&profile.TotalScore,
		&profile.Rank,
		&profile.ChallengesCompleted,
		&profile.FollowersCount,
		&profile.FollowingCount,
		&technologiesJSON,
	)
	if err != nil {
		return nil, err
	}

	// Parse technologies JSON array
	if len(technologiesJSON) > 0 {
		if err := json.Unmarshal(technologiesJSON, &profile.Technologies); err != nil {
			profile.Technologies = []string{}
		}
	}

	// Check if the requesting user follows this profile
	if requestingClerkUserID != "" {
		var requestingID string
		err = db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE clerk_user_id = $1 OR id = $1", requestingClerkUserID).Scan(&requestingID)
		if err == nil && requestingID != "" {
			db.Pool.QueryRow(ctx, `
				SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)
			`, requestingID, profile.ID).Scan(&profile.IsFollowing)
		}
	}

	// Fetch user's showcase submissions (latest 10)
	subQuery := `
		SELECT 
			s.id, s.user_id, u2.clerk_user_id,
			COALESCE(u2.username, '') as username,
			COALESCE(u2.display_name, '') as display_name,
			COALESCE(u2.avatar_url, '') as avatar_url,
			s.challenge_id,
			COALESCE(c.title, '') as challenge_title,
			COALESCE(s.title, '') as title,
			s.repo_url,
			COALESCE(s.video_url, '') as video_url,
			COALESCE(s.project_description, '') as project_description,
			s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM submission_likes sl WHERE sl.submission_id = s.id) as like_count,
			false as user_has_liked
		FROM submissions s
		JOIN users u2 ON u2.id = s.user_id
		LEFT JOIN challenges c ON c.id = s.challenge_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT 50
	`

	rows, err := db.Pool.Query(ctx, subQuery, profile.ID)
	if err == nil {
		defer rows.Close()
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
				break
			}
			profile.Submissions = append(profile.Submissions, sub)
		}
	}

	if profile.Submissions == nil {
		profile.Submissions = []SubmissionDetail{}
	}

	return profile, nil
}
