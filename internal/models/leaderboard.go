package models

import "time"

// LeaderboardEntry represents a user's position on the leaderboard
type LeaderboardEntry struct {
	Rank                int       `json:"rank"`
	UserID              string    `json:"user_id"`
	Username            string    `json:"username"`
	DisplayName         string    `json:"display_name"`
	AvatarURL           string    `json:"avatar_url"`
	GitHubUsername      string    `json:"github_username"`
	TotalScore          int       `json:"total_score"`
	ChallengesCompleted int       `json:"challenges_completed"`
	AverageReviewScore  float64   `json:"average_review_score"`
	CurrentStreak       int       `json:"current_streak"`
	LastActivityAt      time.Time `json:"last_activity_at"`
}

// LeaderboardFilter defines filtering options for leaderboard queries
type LeaderboardFilter struct {
	Period     string `json:"period"`     // "all_time", "weekly", "monthly"
	TechStack  string `json:"tech_stack"` // Filter by technology
	Difficulty string `json:"difficulty"` // Filter by challenge difficulty
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}
