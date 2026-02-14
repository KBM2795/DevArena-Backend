package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/KBM2795/DevArena-Backend/internal/config"
	"github.com/KBM2795/DevArena-Backend/internal/models"
	"google.golang.org/genai"
)

type AIClient struct {
	client *genai.Client
	model  string
}

func NewAIClient(cfg *config.Config) (*AIClient, error) {
	if cfg.Gemini.APIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is missing")
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: cfg.Gemini.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &AIClient{
		client: client,
		model:  "gemini-2.5-flash",
	}, nil
}

func (ai *AIClient) GenerateReview(ctx context.Context, input models.AIReviewInput) (*models.AIReviewOutput, error) {
	prompt, err := constructReviewPrompt(input)
	if err != nil {
		return nil, err
	}

	// Configure generation for JSON output
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		Temperature:      PtrFloat32(0.2), // Low temp for deterministic output
	}

	resp, err := ai.client.Models.GenerateContent(ctx, ai.model, genai.Text(prompt), config)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	// Extract JSON from response
	var jsonStr string
	for _, part := range resp.Candidates[0].Content.Parts {
		jsonStr += part.Text
	}

	// Clean up markdown code blocks if present
	jsonStr = cleanJSON(jsonStr)

	var output models.AIReviewOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		log.Printf("[AI] Failed to parse JSON: %s", jsonStr)
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &output, nil
}

func constructReviewPrompt(input models.AIReviewInput) (string, error) {
	// Build prompt with context
	var fileContext strings.Builder
	for name, content := range input.SubmissionCtx.Files {
		// Skip large files or binaries if necessary
		if len(content) > 50000 {
			fileContext.WriteString(fmt.Sprintf("\n--- File: %s (TRUNCATED) ---\n%s\n", name, content[:50000]))
		} else {
			fileContext.WriteString(fmt.Sprintf("\n--- File: %s ---\n%s\n", name, content))
		}
	}

	prompt := fmt.Sprintf(`You are an expert Senior Software Engineer acting as a code reviewer.
Your task is to review a student submission for a coding challenge.

CHALLENGE CONTEXT:
Title: %s
Description: %s
Requirements: %s
Constraints: %s
Rubric: %v

SUBMISSION CONTEXT:
Test Results: Passed %d / Failed %d
Files:
%s

INSTRUCTIONS:
1. Analyze the code for Quality, Constraints, and Architecture.
2. Cross-reference with the Rubric to assign scores.
3. Identify Strengths, Issues, and specific Improvements.
4. Output STRICT JSON in the following format:

{
  "scores": {
    "code_quality": <0-100>,
    "constraints": <0-100>,
    "architecture": <0-100>
  },
  "strengths": ["point 1", "point 2"],
  "issues": ["point 1", "point 2"],
  "improvements": ["actionable advice 1", "actionable advice 2"]
}

Ensure the JSON is valid and contains no other text.`,
		input.ChallengeCtx.Title,
		input.ChallengeCtx.DescriptionMD,
		input.ChallengeCtx.RequirementsMD,
		input.ChallengeCtx.ConstraintsMD,
		input.ChallengeCtx.Rubric,
		input.SubmissionCtx.TestSummary.TestsPassed,
		input.SubmissionCtx.TestSummary.TestsFailed,
		fileContext.String(),
	)

	return prompt, nil
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func PtrFloat32(v float32) *float32 {
	return &v
}
