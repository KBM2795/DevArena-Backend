package models

import "time"

// UserFollow represents a follow relationship between two users
type UserFollow struct {
	FollowerID  string    `json:"follower_id"`
	FollowingID string    `json:"following_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// FollowUser is a user card for search and follower/following lists (YouTube style)
type FollowUser struct {
	ID                  string `json:"id"`
	Username            string `json:"username"`
	DisplayName         string `json:"display_name"`
	AvatarURL           string `json:"avatar_url"`
	Bio                 string `json:"bio"`
	TotalScore          int    `json:"total_score"`
	Rank                int    `json:"rank"`
	ChallengesCompleted int    `json:"challenges_completed"`
	ProjectCount        int    `json:"project_count"`
	FollowersCount      int    `json:"followers_count"`
	IsFollowing         bool   `json:"is_following"`
}
