package db

import (
	"context"
	"time"
)

// UpdateUserStats recalculates and updates stats for a user:
//   - Total Score: Sum of (Challenge Submissions * 100) + (Open Showcases * 80) + (Likes received * 1)
//   - Challenges Completed: Count of unique challenges submitted
//   - Streak: Current and Longest daily streak
//   - Rank: Position based on total score
func (db *Database) UpdateUserStats(ctx context.Context, userID string) error {
	// 1. Calculate Total Score & Challenges Completed
	var stats struct {
		TotalScore          int
		ChallengesCompleted int
	}

	err := db.Pool.QueryRow(ctx, `
		WITH SubmissionPoints AS (
			SELECT 
				COALESCE(SUM(CASE WHEN challenge_id IS NOT NULL THEN 100 ELSE 80 END), 0) as sub_score,
				COUNT(DISTINCT challenge_id) as challenges_done
			FROM submissions
			WHERE user_id = $1
		),
		LikePoints AS (
			SELECT COALESCE(COUNT(*), 0) as like_score
			FROM submission_likes l
			JOIN submissions s ON l.submission_id = s.id
			WHERE s.user_id = $1
		)
		SELECT 
			(SELECT sub_score FROM SubmissionPoints) + (SELECT like_score FROM LikePoints) as total_score,
			(SELECT challenges_done FROM SubmissionPoints) as challenges_completed
	`, userID).Scan(&stats.TotalScore, &stats.ChallengesCompleted)
	if err != nil {
		return err
	}

	// 2. Calculate Streaks (consecutive calendar days user made submissions)
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
	now := time.Now().Truncate(24 * time.Hour)
	latest := dates[0].Truncate(24 * time.Hour)

	daysSinceLatest := int(now.Sub(latest).Hours() / 24)

	if daysSinceLatest > 1 {
		currentStreak = 0 // Streak broken
	} else {
		currentStreak = 1
	}

	for i := 0; i < len(dates)-1; i++ {
		curr := dates[i].Truncate(24 * time.Hour)
		prev := dates[i+1].Truncate(24 * time.Hour)

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
	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

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
