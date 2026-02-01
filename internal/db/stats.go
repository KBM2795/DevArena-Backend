package db

import (
	"context"
	"time"
)

// UserStats represents comprehensive user statistics
type UserStats struct {
	CurrentStreak       int     `json:"current_streak"`
	LongestStreak       int     `json:"longest_streak"`
	TotalScore          int     `json:"total_score"`
	Rank                int     `json:"rank"`
	ChallengesCompleted int     `json:"challenges_completed"`
	TotalSubmissions    int     `json:"total_submissions"`
	AcceptanceRate      float64 `json:"acceptance_rate"`
}

// RecentSubmission represents a submission with challenge info
type RecentSubmission struct {
	ID          string   `json:"id"`
	ChallengeID string   `json:"challenge_id"`
	Title       string   `json:"title"`
	Difficulty  string   `json:"difficulty"`
	Score       int      `json:"score"`
	MaxScore    int      `json:"max_score"`
	Status      string   `json:"status"`
	TechStack   []string `json:"tech_stack"`
	SubmittedAt string   `json:"submitted_at"`
}

// GetUserStats returns comprehensive stats for a user
func (db *Database) GetUserStats(clerkUserID string) (*UserStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID and basic stats
	var stats UserStats
	var internalUserID string

	err := db.Pool.QueryRow(ctx, `
		SELECT id, current_streak, longest_streak, total_score, rank, challenges_completed
		FROM users WHERE clerk_user_id = $1
	`, clerkUserID).Scan(
		&internalUserID,
		&stats.CurrentStreak,
		&stats.LongestStreak,
		&stats.TotalScore,
		&stats.Rank,
		&stats.ChallengesCompleted,
	)
	if err != nil {
		return nil, err
	}

	// Get total submissions count
	db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM submissions WHERE user_id = $1
	`, internalUserID).Scan(&stats.TotalSubmissions)

	// Calculate acceptance rate (reviewed submissions with score > 70%)
	var reviewed, passed int
	db.Pool.QueryRow(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'reviewed'),
			COUNT(*) FILTER (WHERE status = 'reviewed' AND score >= 70)
		FROM submissions WHERE user_id = $1
	`, internalUserID).Scan(&reviewed, &passed)

	if reviewed > 0 {
		stats.AcceptanceRate = float64(passed) / float64(reviewed) * 100
	}

	return &stats, nil
}

// GetRecentSubmissions returns the user's recent submissions with challenge details
func (db *Database) GetRecentSubmissions(clerkUserID string, limit int) ([]RecentSubmission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			s.id,
			s.challenge_id,
			c.title,
			c.difficulty,
			s.score,
			c.max_score,
			s.status,
			c.tech_stack,
			TO_CHAR(s.created_at, 'YYYY-MM-DD') as submitted_at
		FROM submissions s
		JOIN challenges c ON s.challenge_id = c.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2
	`

	rows, err := db.Pool.Query(ctx, query, internalUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []RecentSubmission
	for rows.Next() {
		var sub RecentSubmission
		if err := rows.Scan(
			&sub.ID,
			&sub.ChallengeID,
			&sub.Title,
			&sub.Difficulty,
			&sub.Score,
			&sub.MaxScore,
			&sub.Status,
			&sub.TechStack,
			&sub.SubmittedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, sub)
	}

	return submissions, rows.Err()
}

// TechFocus represents a technology category and its usage percentage
type TechFocus struct {
	Name       string `json:"name"`
	Count      int    `json:"count"`
	Percentage int    `json:"percentage"`
}

// GetUserTechFocus returns tech category breakdown from user's solved challenges
func (db *Database) GetUserTechFocus(clerkUserID string) ([]TechFocus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	// Get all tech_stack from submitted challenges and count occurrences
	// Use jsonb_array_elements_text since tech_stack is stored as JSON array
	query := `
		SELECT tech, COUNT(*) as count
		FROM (
			SELECT DISTINCT s.challenge_id, jsonb_array_elements_text(c.tech_stack::jsonb) as tech
			FROM submissions s
			JOIN challenges c ON s.challenge_id = c.id
			WHERE s.user_id = $1 AND s.status = 'reviewed'
		) sub
		GROUP BY tech
		ORDER BY count DESC
		LIMIT 6
	`

	rows, err := db.Pool.Query(ctx, query, internalUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var techCounts []struct {
		Tech  string
		Count int
	}
	var total int

	for rows.Next() {
		var t struct {
			Tech  string
			Count int
		}
		if err := rows.Scan(&t.Tech, &t.Count); err != nil {
			return nil, err
		}
		techCounts = append(techCounts, t)
		total += t.Count
	}

	// Convert to percentages
	var result []TechFocus
	for _, tc := range techCounts {
		percentage := 0
		if total > 0 {
			percentage = (tc.Count * 100) / total
		}
		result = append(result, TechFocus{
			Name:       tc.Tech,
			Count:      tc.Count,
			Percentage: percentage,
		})
	}

	return result, rows.Err()
}
