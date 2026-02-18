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
	Technologies        []string `json:"technologies"`
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
