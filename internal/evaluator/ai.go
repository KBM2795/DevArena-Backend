package evaluator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/KBM2795/DevArena-Backend/internal/config"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/KBM2795/DevArena-Backend/internal/models"
	"google.golang.org/genai"
)

type AIClient struct {
	client *genai.Client
	model  string
	db     *db.Database
}

func NewAIClient(cfg *config.Config, database *db.Database) (*AIClient, error) {
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
		db:     database,
	}, nil
}

// ── Prompt Version fetched from DB ───────────────────────────────────
// prompt_versions schema:
//   id          varchar  PK   (e.g. "prompt-code-review-v1")
//   name        varchar       (e.g. "code_review_v1")
//   step        varchar       (e.g. "code_review")  ← matched via EvaluationType
//   model       varchar       (e.g. "gpt-4o")
//   prompt_text text          ← the SYSTEM prompt stored in DB
//   is_active   boolean
//   created_at  timestamp
//   "Version"   varchar       (e.g. "1", "2")       ← pick the highest

// fetchSystemPrompt queries prompt_versions for the latest active prompt
// matching the given step (e.g. "code_review"). It picks the row with the
// highest Version number. Returns (promptText, promptVersionID, error).
func (ai *AIClient) fetchSystemPrompt(ctx context.Context, step string) (string, string, error) {
	query := `
		SELECT id, prompt_text
		FROM prompt_versions
		WHERE step = $1 AND is_active = true
		ORDER BY CAST("Version" AS INTEGER) DESC
		LIMIT 1
	`

	var promptVersionID string
	var promptText string

	err := ai.db.Pool.QueryRow(ctx, query, step).Scan(&promptVersionID, &promptText)
	if err != nil {
		return "", "", fmt.Errorf("no active prompt for step %q: %w", step, err)
	}

	log.Printf("[AI] Using prompt version %s for step %s", promptVersionID, step)
	return promptText, promptVersionID, nil
}

// ── Default fallback prompt ──────────────────────────────────────────
const defaultSystemPrompt = `You are an expert code reviewer for DevArena. Analyze the submitted code and return a structured JSON evaluation.

You will receive:
- challenge context (title, description, requirements, constraints, rubric)
- submission files (source code)
- test summary (tests passed/failed)

Return ONLY valid JSON in this exact format:
{
  "scores": {
    "code_quality": <int>,
    "constraints": <int>,
    "architecture": <int>
  },
  "strengths": ["..."],
  "issues": ["..."],
  "improvements": ["..."]
}

RULES:
- Scores must be integers
- Scores must NOT exceed the rubric max for each category
- No negative scores
- All fields are required
- Do NOT calculate final score
- Do NOT modify functionality score`

// ── GenerateReview ───────────────────────────────────────────────────
func (ai *AIClient) GenerateReview(ctx context.Context, input models.AIReviewInput) (*models.AIReviewOutput, string, error) {
	// 1. Fetch the system prompt from DB (highest active version)
	systemPrompt, promptVersionID, err := ai.fetchSystemPrompt(ctx, input.EvaluationType)
	if err != nil {
		log.Printf("[AI] Warning: %v — falling back to default prompt", err)
		systemPrompt = defaultSystemPrompt
		promptVersionID = "fallback"
	}

	// 2. Build the user prompt with dynamic challenge + submission data
	userPrompt := buildUserPrompt(input)

	// 3. Combine: system prompt first, then user prompt
	fullPrompt := systemPrompt + "\n\n" + userPrompt

	// 4. Configure generation for JSON output
	genConfig := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		Temperature:      PtrFloat32(0.2), // Low temp for deterministic output
	}

	// 5. Call Gemini
	resp, err := ai.client.Models.GenerateContent(ctx, ai.model, genai.Text(fullPrompt), genConfig)
	if err != nil {
		return nil, promptVersionID, fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, promptVersionID, fmt.Errorf("empty response from gemini")
	}

	// 6. Extract JSON from response
	var jsonStr string
	for _, part := range resp.Candidates[0].Content.Parts {
		jsonStr += part.Text
	}

	// Clean up markdown code blocks if present
	jsonStr = cleanJSON(jsonStr)

	// 7. Parse into structured output
	var output models.AIReviewOutput
	if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
		log.Printf("[AI] Failed to parse JSON: %s", jsonStr)
		return nil, promptVersionID, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &output, promptVersionID, nil
}

// ── buildUserPrompt creates the dynamic data portion ─────────────────
// This is the "user" part — challenge details + submission files.
// The system prompt (from DB) already defines the AI's role and output format.
func buildUserPrompt(input models.AIReviewInput) string {
	var b strings.Builder

	// Challenge context
	b.WriteString(fmt.Sprintf("CHALLENGE CONTEXT:\nTitle: %s\n", input.ChallengeCtx.Title))
	b.WriteString(fmt.Sprintf("Description: %s\n", input.ChallengeCtx.DescriptionMD))
	b.WriteString(fmt.Sprintf("Requirements: %s\n", input.ChallengeCtx.RequirementsMD))
	b.WriteString(fmt.Sprintf("Constraints: %s\n", input.ChallengeCtx.ConstraintsMD))
	b.WriteString(fmt.Sprintf("Rubric: %v\n", input.ChallengeCtx.Rubric))

	// Test summary
	b.WriteString(fmt.Sprintf("\nSUBMISSION CONTEXT:\nTest Results: Passed %d / Failed %d\n",
		input.SubmissionCtx.TestSummary.TestsPassed,
		input.SubmissionCtx.TestSummary.TestsFailed,
	))

	// Source files
	b.WriteString("\nFiles:\n")
	for name, content := range input.SubmissionCtx.Files {
		if len(content) > 50000 {
			b.WriteString(fmt.Sprintf("\n--- File: %s (TRUNCATED) ---\n%s\n", name, content[:50000]))
		} else {
			b.WriteString(fmt.Sprintf("\n--- File: %s ---\n%s\n", name, content))
		}
	}

	return b.String()
}

// ── Helpers ──────────────────────────────────────────────────────────

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
