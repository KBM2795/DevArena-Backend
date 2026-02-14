package db

import (
	"context"
	"fmt"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// SaveTestResults persists the raw and parsed test results
func (db *Database) SaveTestResults(submissionID string, result *models.TestResult, rawOutput string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := fmt.Sprintf("tr-%s", submissionID)

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO submission_test_results (
			id, submission_id, tests_total, tests_passed, tests_failed, raw_output, build_success
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (submission_id) DO UPDATE SET
			tests_total = EXCLUDED.tests_total,
			tests_passed = EXCLUDED.tests_passed,
			tests_failed = EXCLUDED.tests_failed,
			raw_output = EXCLUDED.raw_output,
			build_success = EXCLUDED.build_success
	`,
		id,
		submissionID,
		result.Total,
		result.Passed,
		result.Failed,
		rawOutput,
		result.Total > 0, // Assume build success if tests ran
	)

	return err
}

// SaveScore persists the final calculated score
func (db *Database) SaveScore(submissionID string, result *models.SubmissionScore) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := fmt.Sprintf("score-%s", submissionID)

	_, err := db.Pool.Exec(ctx, `
		INSERT INTO submission_scores (
			id, submission_id, functionality_score, code_quality_score, constraint_score, architecture_score, final_score, max_score
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (submission_id) DO UPDATE SET
			functionality_score = EXCLUDED.functionality_score,
			code_quality_score = EXCLUDED.code_quality_score,
			constraint_score = EXCLUDED.constraint_score,
			architecture_score = EXCLUDED.architecture_score,
			final_score = EXCLUDED.final_score,
			max_score = EXCLUDED.max_score,
			updated_at = NOW()
	`,
		id,
		submissionID,
		result.FunctionalityScore,
		result.CodeQualityScore,
		result.ConstraintScore,
		result.ArchitectureScore,
		result.FinalScore,
		result.MaxScore,
	)

	return err
}

// GetSubmissionResults retrieves detailed test results for a submission
func (db *Database) GetSubmissionResults(submissionID string) (*models.TestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tr models.TestResult
	var rawOutput string

	err := db.Pool.QueryRow(ctx, `
		SELECT tests_total, tests_passed, tests_failed, raw_output
		FROM submission_test_results
		WHERE submission_id = $1
	`, submissionID).Scan(&tr.Total, &tr.Passed, &tr.Failed, &rawOutput)

	if err != nil {
		return nil, err
	}

	// We don't store individual test details in DB to save space
	// We could parse rawOutput if needed, but for now just counts
	return &tr, nil
}
