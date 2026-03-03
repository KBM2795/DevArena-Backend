package evaluator

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/KBM2795/DevArena-Backend/internal/config"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// Returns the AIReview model to be saved by the caller
func RunAIReview(ctx context.Context, sub *db.SubmissionDetail, repoDir string, testResult *models.TestResult, cfg *config.Config, database *db.Database) (*models.AIReview, error) {
	log.Printf("[AI Pipeline] Starting AI Review for submission %s...", sub.ID)

	// 1. Initialize AI Client
	ai, err := NewAIClient(cfg, database)
	if err != nil {
		return nil, fmt.Errorf("failed to init AI client: %v", err)
	}

	// 2. Gather Source Files
	files, err := gatherSourceFiles(repoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to gather source files: %w", err)
	}

	// 3. Prepare Input Context
	input := models.AIReviewInput{
		EvaluationType: "code_review",
		ChallengeCtx: models.AIReviewChallengeCtx{
			ChallengeID:   sub.ChallengeID,
			Title:         sub.ChallengeTitle,
			Difficulty:    "Medium", // Hardcoded for now, should come from sub/challenge
			DescriptionMD: "See challenge description",
			Rubric:        nil,
		},
		SubmissionCtx: models.AIReviewSubmissionCtx{
			Files: files,
			TestSummary: models.TestSummary{
				TestsPassed: testResult.Passed,
				TestsFailed: testResult.Failed,
			},
		},
	}

	// 4. Call AI Generation
	reviewOutput, promptVersionID, err := ai.GenerateReview(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// 5. Construct Result
	aiReview := &models.AIReview{
		SubmissionID:      sub.ID,
		PromptVersionID:   promptVersionID,
		CodeQualityScore:  reviewOutput.Scores.CodeQuality,
		ConstraintScore:   reviewOutput.Scores.Constraints,
		ArchitectureScore: reviewOutput.Scores.Architecture,
		StrengthsJSON:     models.StringArray(reviewOutput.Strengths),
		IssuesJSON:        models.StringArray(reviewOutput.Issues),
		ImprovementsJSON:  models.StringArray(reviewOutput.Improvements),
	}

	return aiReview, nil
}

// gatherSourceFiles walks the repo and returns a map of filename -> content
// It ignores .git, node_modules, and non-text files
func gatherSourceFiles(repoDir string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(repoDir, path)
		// Skip directories and ignored paths
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "dist" || d.Name() == ".next" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only read text files of interest
		ext := filepath.Ext(path)
		if !isInterestingExtension(ext) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		// Limit file size to 50KB to prevent context bloat
		if len(content) > 50000 {
			files[relPath] = string(content[:50000]) + "\n...[TRUNCATED]"
		} else {
			files[relPath] = string(content)
		}

		return nil
	})

	return files, err
}

func isInterestingExtension(ext string) bool {
	switch strings.ToLower(ext) {
	case ".js", ".ts", ".jsx", ".tsx", ".css", ".html", ".go", ".py", ".sql", ".md", ".json":
		return true
	default:
		return false
	}
}
