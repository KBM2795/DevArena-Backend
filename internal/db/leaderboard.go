package db

import (
	"context"
	"time"
)

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank                int    `json:"rank"`
	UserID              string `json:"user_id"`
	DisplayName         string `json:"display_name"`
	Username            string `json:"username"`
	AvatarURL           string `json:"avatar_url"`
	TotalScore          int    `json:"total_score"`
	ChallengesCompleted int    `json:"challenges_completed"`
	CurrentStreak       int    `json:"current_streak"`
}

// GetLeaderboard returns the top N users ordered by total_score
func (db *Database) GetLeaderboard(limit int) ([]LeaderboardEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			ROW_NUMBER() OVER (ORDER BY total_score DESC) as rank,
			id,
			COALESCE(display_name, username, 'Anonymous') as display_name,
			COALESCE(username, '') as username,
			COALESCE(avatar_url, '') as avatar_url,
			total_score,
			challenges_completed,
			current_streak
		FROM users 
		WHERE total_score > 0
		ORDER BY total_score DESC
		LIMIT $1
	`

	rows, err := db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(
			&entry.Rank,
			&entry.UserID,
			&entry.DisplayName,
			&entry.Username,
			&entry.AvatarURL,
			&entry.TotalScore,
			&entry.ChallengesCompleted,
			&entry.CurrentStreak,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
