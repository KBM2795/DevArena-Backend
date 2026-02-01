package db

import (
	"context"
	"time"
)

// ActivityEntry represents daily activity data for the contribution graph
type ActivityEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetUserActivity returns daily submission counts for a user in a given year
func (db *Database) GetUserActivity(clerkUserID string, year int) ([]ActivityEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First, get the internal user ID
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		return nil, err
	}

	// Query submissions grouped by date
	query := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			COUNT(*) as count
		FROM submissions
		WHERE user_id = $1 
		AND EXTRACT(YEAR FROM created_at) = $2
		GROUP BY TO_CHAR(created_at, 'YYYY-MM-DD')
		ORDER BY date ASC
	`

	rows, err := db.Pool.Query(ctx, query, internalUserID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ActivityEntry
	for rows.Next() {
		var entry ActivityEntry
		if err := rows.Scan(&entry.Date, &entry.Count); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}
