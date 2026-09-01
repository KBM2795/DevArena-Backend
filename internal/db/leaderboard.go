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

// LeaderboardResult contains top users and surrounding users around the requesting user
type LeaderboardResult struct {
	Top          []LeaderboardEntry `json:"top"`
	Surroundings []LeaderboardEntry `json:"surroundings"`
	UserRank     int                `json:"user_rank"`
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

// GetLeaderboardWithSurroundings returns top users (e.g. 20) and, if the user is outside the top 20,
// 3 users above them, the user themselves, and 3 users below them.
func (db *Database) GetLeaderboardWithSurroundings(clerkUserID string, limit int) (*LeaderboardResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if limit < 1 || limit > 50 {
		limit = 20
	}

	result := &LeaderboardResult{
		Top:          []LeaderboardEntry{},
		Surroundings: []LeaderboardEntry{},
	}

	// 1. Fetch top N users
	topEntries, err := db.GetLeaderboard(limit)
	if err != nil {
		return nil, err
	}
	if topEntries != nil {
		result.Top = topEntries
	}

	// 2. If requesting user is authenticated, find their rank and surroundings
	if clerkUserID != "" {
		rankQuery := `
			WITH ranked AS (
				SELECT clerk_user_id, ROW_NUMBER() OVER (ORDER BY total_score DESC) as rank
				FROM users
				WHERE total_score > 0
			)
			SELECT rank FROM ranked WHERE clerk_user_id = $1
		`
		var userRank int
		err := db.Pool.QueryRow(ctx, rankQuery, clerkUserID).Scan(&userRank)
		if err == nil && userRank > 0 {
			result.UserRank = userRank

			// If the user is outside the top list, fetch 3 above + user + 3 below
			if userRank > limit {
				surroundingQuery := `
					WITH ranked AS (
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
					)
					SELECT 
						rank, id, display_name, username, avatar_url,
						total_score, challenges_completed, current_streak
					FROM ranked
					WHERE rank BETWEEN ($1 - 3) AND ($1 + 3)
					ORDER BY rank ASC
				`
				sRows, err := db.Pool.Query(ctx, surroundingQuery, userRank)
				if err == nil {
					defer sRows.Close()
					for sRows.Next() {
						var sEntry LeaderboardEntry
						if err := sRows.Scan(
							&sEntry.Rank, &sEntry.UserID, &sEntry.DisplayName,
							&sEntry.Username, &sEntry.AvatarURL, &sEntry.TotalScore,
							&sEntry.ChallengesCompleted, &sEntry.CurrentStreak,
						); err == nil {
							result.Surroundings = append(result.Surroundings, sEntry)
						}
					}
				}
			}
		}
	}

	return result, nil
}
