package models

import "time"

// LeaderboardEntry represents a user's position on the leaderboard
type LeaderboardEntry struct {
	Rank           int       `json:"rank"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"display_name"`
	AvatarURL      string    `json:"avatar_url"`
	TotalScore     int       `json:"total_score"`
	ProblemsSolved int       `json:"problems_solved"`
	CurrentStreak  int       `json:"current_streak"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// LeaderboardFilter defines filtering options for leaderboard queries
type LeaderboardFilter struct {
	Period string `json:"period"` // "all_time", "weekly", "monthly"
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
