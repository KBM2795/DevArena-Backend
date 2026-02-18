package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// OnboardingData represents the data submitted during onboarding
type OnboardingData struct {
	Experience   string   `json:"experience"`
	Paths        []string `json:"paths"`
	Technologies []string `json:"technologies"`
}

// SaveOnboardingData saves user onboarding data to the starter_packs table
// and auto-assigns challenges based on the user's selected paths and technologies
func (db *Database) SaveOnboardingData(clerkUserID string, onboardingData struct {
	Experience   string   `json:"experience"`
	Paths        []string `json:"paths"`
	Technologies []string `json:"technologies"`
}) error {
	log.Printf("[Onboarding] Starting for user: %s, experience: %s", clerkUserID, onboardingData.Experience)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the internal user ID from clerk_user_id
	var internalUserID string
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Convert slices to JSON for JSONB columns
	pathsJSON, err := json.Marshal(onboardingData.Paths)
	if err != nil {
		return err
	}

	techJSON, err := json.Marshal(onboardingData.Technologies)
	if err != nil {
		return err
	}

	// Map experience level to difficulty and find matching challenges
	difficultyFilter := mapExperienceToDifficulty(onboardingData.Experience)
	challengeIDs, err := db.findMatchingChallenges(ctx, onboardingData.Technologies, difficultyFilter)
	if err != nil {
		// Continue with empty challenge list - don't fail the whole onboarding
		challengeIDs = []string{}
	}

	challengeIDsJSON, err := json.Marshal(challengeIDs)
	if err != nil {
		return err
	}

	// Generate a new UUID for the starter pack
	starterPackID := uuid.New().String()

	// Insert or update the starter pack with challenge IDs
	query := `
		INSERT INTO starter_packs (id, user_id, experience, paths, technologies, challenge_ids, total_challenges, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			experience = EXCLUDED.experience,
			paths = EXCLUDED.paths,
			technologies = EXCLUDED.technologies,
			challenge_ids = EXCLUDED.challenge_ids,
			total_challenges = EXCLUDED.total_challenges,
			updated_at = NOW()
	`

	// Convert JSON byte slices to strings for SimpleProtocol compatibility
	_, err = db.Pool.Exec(ctx, query, starterPackID, internalUserID, onboardingData.Experience, string(pathsJSON), string(techJSON), string(challengeIDsJSON), len(challengeIDs))
	if err != nil {
		return fmt.Errorf("failed to insert starter pack: %w", err)
	}

	// Mark user's onboarding as completed
	updateUserQuery := `
		UPDATE users SET onboarding_completed = TRUE, updated_at = NOW()
		WHERE id = $1
	`
	_, err = db.Pool.Exec(ctx, updateUserQuery, internalUserID)
	if err != nil {
		return fmt.Errorf("failed to update onboarding status: %w", err)
	}

	log.Printf("[Onboarding] Completed for user %s with %d challenges assigned", clerkUserID, len(challengeIDs))
	return nil
}

// mapExperienceToDifficulty maps user experience level to challenge difficulties
func mapExperienceToDifficulty(experience string) []string {
	switch experience {
	case "beginner":
		return []string{"Easy"}
	case "intermediate":
		return []string{"Easy", "Medium"}
	case "advanced":
		return []string{"Medium", "Hard"}
	case "expert":
		return []string{"Hard"}
	default:
		return []string{"Easy", "Medium", "Hard"}
	}
}

// findMatchingChallenges finds challenges that match the user's technologies and difficulty
func (db *Database) findMatchingChallenges(ctx context.Context, technologies []string, difficulties []string) ([]string, error) {
	if len(technologies) == 0 {
		return []string{}, nil
	}

	// Check if there are published challenges
	var publishedCount int
	db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM challenges WHERE is_published = TRUE").Scan(&publishedCount)
	if publishedCount == 0 {
		return []string{}, nil
	}

	// Query challenges matching difficulty
	query := `
		SELECT id FROM challenges 
		WHERE is_published = TRUE 
		AND difficulty = ANY($1)
		ORDER BY 
			CASE difficulty 
				WHEN 'Easy' THEN 1 
				WHEN 'Medium' THEN 2 
				WHEN 'Hard' THEN 3 
			END,
			created_at DESC
		LIMIT 10
	`

	rows, err := db.Pool.Query(ctx, query, difficulties)
	if err != nil {
		return nil, fmt.Errorf("failed to query challenges: %w", err)
	}
	defer rows.Close()

	var challengeIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		challengeIDs = append(challengeIDs, id)
	}

	return challengeIDs, rows.Err()
}
