package db

import (
	"context"
	"time"
)

// UpdateUserStats recalculates and updates stats for a user
// - Total Score: Sum of best scores per challenge
// - Challenges Completed: Count of unique challenges with completed submissions
// - Streak: Current and Longest daily streak
// - Rank: Position based on total score
func (db *Database) UpdateUserStats(ctx context.Context, userID string) error {
	// 1. Calculate Total Score & Challenges Completed
	var stats struct {
		TotalScore          int
		ChallengesCompleted int
	}

	// We sum the MAXIMUM score the user has achieved for each challenge
	err := db.Pool.QueryRow(ctx, `
		WITH BestScores AS (
			SELECT s.challenge_id, MAX(sc.final_score) as max_score
			FROM submissions s
			JOIN submission_scores sc ON s.id = sc.submission_id
			WHERE s.user_id = $1
			GROUP BY s.challenge_id
		)
		SELECT 
			COALESCE(SUM(max_score), 0),
			COUNT(*)
		FROM BestScores
	`, userID).Scan(&stats.TotalScore, &stats.ChallengesCompleted)
	if err != nil {
		return err
	}

	// 2. Calculate Streaks (handled in Go for easier logic)
	// Fetch all unique submission dates
	rows, err := db.Pool.Query(ctx, `
		SELECT DISTINCT DATE(created_at) as sub_date
		FROM submissions
		WHERE user_id = $1
		ORDER BY sub_date DESC
	`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var dates []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return err
		}
		dates = append(dates, d)
	}

	currentStreak, longestStreak := calculateStreaks(dates)

	// 3. Update User Table with Score, Challenges, Streaks
	_, err = db.Pool.Exec(ctx, `
		UPDATE users 
		SET total_score = $2, 
			challenges_completed = $3,
			current_streak = $4,
			longest_streak = $5,
			updated_at = NOW()
		WHERE id = $1
	`, userID, stats.TotalScore, stats.ChallengesCompleted, currentStreak, longestStreak)
	if err != nil {
		return err
	}

	// 4. Update Rank (Global)
	// This is slightly expensive but ok for now.
	// Rank = Count of users with higher score + 1
	_, err = db.Pool.Exec(ctx, `
		UPDATE users u
		SET rank = (
			SELECT COUNT(*) + 1
			FROM users u2
			WHERE u2.total_score > u.total_score
		)
		WHERE u.id = $1
	`, userID)

	return err
}

// calculateStreaks computes current and longest daily streaks
func calculateStreaks(dates []time.Time) (int, int) {
	if len(dates) == 0 {
		return 0, 0
	}

	currentStreak := 0
	longestStreak := 0
	tempStreak := 1

	// Dates are ordered DESC (newest first)
	// Check today vs newest submission
	now := time.Now().Truncate(24 * time.Hour)
	latest := dates[0].Truncate(24 * time.Hour)

	// If latest submission was today or yesterday, streak is active
	daysSinceLatest := int(now.Sub(latest).Hours() / 24)

	if daysSinceLatest > 1 {
		currentStreak = 0 // Streak broken
	} else {
		currentStreak = 1 // At least 1 (the latest day)
	}

	// Iterate backwards through dates (newest to oldest) to find consecutive days
	for i := 0; i < len(dates)-1; i++ {
		curr := dates[i].Truncate(24 * time.Hour)
		prev := dates[i+1].Truncate(24 * time.Hour) // previous date in time (next in list)

		diff := int(curr.Sub(prev).Hours() / 24)

		if diff == 1 {
			tempStreak++
		} else {
			if tempStreak > longestStreak {
				longestStreak = tempStreak
			}
			tempStreak = 1
		}
	}
	// Check last streak
	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	// If current streak is active (meaning top of list is part of a sequence connected to today)
	// We need to calculate how far back the connected sequence goes from the top
	// We already did a full pass for longest, let's just re-evaluate current from top

	// Reset current run
	if daysSinceLatest <= 1 {
		currentRun := 1
		for i := 0; i < len(dates)-1; i++ {
			curr := dates[i].Truncate(24 * time.Hour)
			prev := dates[i+1].Truncate(24 * time.Hour)
			if int(curr.Sub(prev).Hours()/24) == 1 {
				currentRun++
			} else {
				break
			}
		}
		currentStreak = currentRun
	} else {
		currentStreak = 0
	}

	return currentStreak, longestStreak
}
