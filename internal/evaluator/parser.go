package evaluator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// ParseVitestOutput parses the JSON output from Vitest
// The output might contain non-JSON text before/after the actual JSON, so we extract it
func ParseVitestOutput(rawOutput string) (*models.TestResult, error) {
	// Vitest JSON reporter outputs a JSON object
	// But there might be npm warnings or other text around it
	jsonStr := extractJSON(rawOutput)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in vitest output")
	}

	var vitest models.VitestOutput
	if err := json.Unmarshal([]byte(jsonStr), &vitest); err != nil {
		// Log the raw output for debugging
		fmt.Printf("[Parser] Failed to parse JSON. Raw output:\n%s\n", rawOutput)
		return nil, fmt.Errorf("failed to parse vitest JSON: %w", err)
	}

	result := &models.TestResult{
		Passed: vitest.NumPassedTests,
		Failed: vitest.NumFailedTests,
		Total:  vitest.NumTotalTests,
	}

	// Extract individual test details
	for _, suite := range vitest.TestResults {
		for _, assertion := range suite.AssertionResults {
			detail := models.TestDetail{
				Name:   assertion.Title,
				Status: assertion.Status,
			}
			if len(assertion.FailureMessages) > 0 {
				detail.Error = assertion.FailureMessages[0]
				// Truncate error messages to prevent huge DB entries
				if len(detail.Error) > 500 {
					detail.Error = detail.Error[:500] + "..."
				}
			}
			result.Details = append(result.Details, detail)
		}
	}

	// Fallback: if no assertions found but we have counts
	if result.Total == 0 && len(result.Details) > 0 {
		result.Total = len(result.Details)
	}

	return result, nil
}

// extractJSON finds the first valid JSON object in a string
// Vitest might print warnings before/after the JSON
func extractJSON(s string) string {
	// Find the first '{' that starts a JSON object
	start := strings.Index(s, `{"`)
	if start == -1 {
		return ""
	}

	// Walk forward to find the matching closing brace
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	return ""
}

// CalculateScore computes the final score based on test results
func CalculateScore(result *models.TestResult, maxScore int) int {
	if result.Total == 0 {
		return 0
	}
	score := (result.Passed * maxScore) / result.Total
	return score
}
