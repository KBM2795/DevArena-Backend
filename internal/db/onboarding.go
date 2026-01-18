package db

import (
	"context"
	"encoding/json"
	"fmt"
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
func (db *Database) SaveOnboardingData(clerkUserID string, onboardingData struct {
	Experience   string   `json:"experience"`
	Paths        []string `json:"paths"`
	Technologies []string `json:"technologies"`
}) error {
	fmt.Printf("[DEBUG] SaveOnboardingData called with clerkUserID: %s\n", clerkUserID)
	fmt.Printf("[DEBUG] Onboarding data: Experience=%s, Paths=%v, Technologies=%v\n",
		onboardingData.Experience, onboardingData.Paths, onboardingData.Technologies)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First, get the internal user ID from clerk_user_id
	var internalUserID string
	fmt.Printf("[DEBUG] Looking up user with clerk_user_id: %s\n", clerkUserID)
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM users WHERE clerk_user_id = $1",
		clerkUserID,
	).Scan(&internalUserID)
	if err != nil {
		fmt.Printf("[DEBUG] ERROR: Failed to find user with clerk_user_id %s: %v\n", clerkUserID, err)
		return fmt.Errorf("failed to find user: %w", err)
	}
	fmt.Printf("[DEBUG] Found internal user ID: %s\n", internalUserID)

	// Convert slices to JSON for JSONB columns
	pathsJSON, err := json.Marshal(onboardingData.Paths)
	if err != nil {
		fmt.Printf("[DEBUG] ERROR: Failed to marshal paths: %v\n", err)
		return err
	}

	techJSON, err := json.Marshal(onboardingData.Technologies)
	if err != nil {
		fmt.Printf("[DEBUG] ERROR: Failed to marshal technologies: %v\n", err)
		return err
	}
	fmt.Printf("[DEBUG] Marshaled JSON - paths: %s, tech: %s\n", string(pathsJSON), string(techJSON))

	// Generate a new UUID for the starter pack
	starterPackID := uuid.New().String()
	fmt.Printf("[DEBUG] Generated starter pack ID: %s\n", starterPackID)

	// Insert or update the starter pack
	query := `
		INSERT INTO starter_packs (id, user_id, experience, paths, technologies, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			experience = EXCLUDED.experience,
			paths = EXCLUDED.paths,
			technologies = EXCLUDED.technologies,
			updated_at = NOW()
	`

	fmt.Printf("[DEBUG] Executing INSERT into starter_packs...\n")
	_, err = db.Pool.Exec(ctx, query, starterPackID, internalUserID, onboardingData.Experience, pathsJSON, techJSON)
	if err != nil {
		fmt.Printf("[DEBUG] ERROR: Failed to insert starter pack: %v\n", err)
		return err
	}
	fmt.Printf("[DEBUG] Successfully inserted/updated starter pack\n")

	// Mark user's onboarding as completed
	updateUserQuery := `
		UPDATE users SET onboarding_completed = TRUE, updated_at = NOW()
		WHERE id = $1
	`
	fmt.Printf("[DEBUG] Updating onboarding_completed for user ID: %s\n", internalUserID)
	_, err = db.Pool.Exec(ctx, updateUserQuery, internalUserID)
	if err != nil {
		fmt.Printf("[DEBUG] ERROR: Failed to update onboarding_completed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Successfully marked onboarding as completed\n")
	}

	return err
}
