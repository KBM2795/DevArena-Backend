package models

import (
	"encoding/json"
	"time"
)

// ChallengeTemplate represents the template configuration for a challenge
type ChallengeTemplate struct {
	ID                string          `json:"id" gorm:"primaryKey;type:varchar(255)"`
	ChallengeID       string          `json:"challenge_id" gorm:"type:varchar(255);not null;unique"`
	RepoTemplateURL   string          `json:"repo_template_url" gorm:"type:text;not null"`
	TestRepoURL       string          `json:"test_repo_url" gorm:"type:text"` // Private test repo URL
	EntryFile         string          `json:"entry_file" gorm:"type:varchar(255);default:src/index.tsx"`
	AllowedEditPaths  StringArray     `json:"allowed_edit_paths" gorm:"type:jsonb"`
	ReadonlyPaths     StringArray     `json:"readonly_paths" gorm:"type:jsonb"`
	ForbiddenPackages StringArray     `json:"forbidden_packages" gorm:"type:jsonb"`
	TemplateTree      json.RawMessage `json:"template_tree" gorm:"type:jsonb"`
	CreatedAt         time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// TemplateResponse is the API response for challenge templates
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
