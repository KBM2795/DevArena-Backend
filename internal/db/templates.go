package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TemplateResponse represents a challenge template in API responses
type TemplateResponse struct {
	ID                string          `json:"id"`
	ChallengeID       string          `json:"challenge_id"`
	RepoTemplateURL   string          `json:"repo_template_url"`
	TestRepoURL       string          `json:"test_repo_url,omitempty"`
	EntryFile         string          `json:"entry_file"`
	AllowedEditPaths  []string        `json:"allowed_edit_paths"`
	ReadonlyPaths     []string        `json:"readonly_paths"`
	ForbiddenPackages []string        `json:"forbidden_packages"`
	TemplateTree      json.RawMessage `json:"template_tree"`
}

// GetTemplateByChallenge retrieves the template for a specific challenge
func (db *Database) GetTemplateByChallenge(challengeID string) (*TemplateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			id,
			challenge_id,
			repo_template_url,
			COALESCE(test_repo_url, ''),
			COALESCE(entry_file, 'src/index.tsx'),
			COALESCE(allowed_edit_paths, '[]'),
			COALESCE(readonly_paths, '[]'),
			COALESCE(forbidden_packages, '[]'),
			COALESCE(template_tree, '[]')
		FROM challenge_templates
		WHERE challenge_id = $1
	`

	var t TemplateResponse
	var allowedEditPathsJSON, readonlyPathsJSON, forbiddenPackagesJSON string
	var templateTreeBytes []byte

	err := db.Pool.QueryRow(ctx, query, challengeID).Scan(
		&t.ID,
		&t.ChallengeID,
		&t.RepoTemplateURL,
		&t.TestRepoURL,
		&t.EntryFile,
		&allowedEditPathsJSON,
		&readonlyPathsJSON,
		&forbiddenPackagesJSON,
		&templateTreeBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Parse JSON arrays
	json.Unmarshal([]byte(allowedEditPathsJSON), &t.AllowedEditPaths)
	json.Unmarshal([]byte(readonlyPathsJSON), &t.ReadonlyPaths)
	json.Unmarshal([]byte(forbiddenPackagesJSON), &t.ForbiddenPackages)
	t.TemplateTree = templateTreeBytes

	// Ensure arrays are not nil
	if t.AllowedEditPaths == nil {
		t.AllowedEditPaths = []string{}
	}
	if t.ReadonlyPaths == nil {
		t.ReadonlyPaths = []string{}
	}
	if t.ForbiddenPackages == nil {
		t.ForbiddenPackages = []string{}
	}

	return &t, nil
}
