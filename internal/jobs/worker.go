package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/KBM2795/DevArena-Backend/internal/config"
	"github.com/KBM2795/DevArena-Backend/internal/db"
	"github.com/KBM2795/DevArena-Backend/internal/evaluator"
	"github.com/KBM2795/DevArena-Backend/internal/models"
)

// Worker handles evaluation jobs one at a time
type Worker struct {
	db           *db.Database
	config       *config.Config
	stopCh       chan struct{} // signal to stop the worker
	doneCh       chan struct{} // signals that the worker has stopped
	pollInterval time.Duration
}

// NewWorker creates a new evaluation worker
func NewWorker(database *db.Database, cfg *config.Config) *Worker {
	return &Worker{
		db:           database,
		config:       cfg,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
		pollInterval: 5 * time.Second,
	}
}

// Start begins the worker loop in a goroutine
// 1. Recovers any stuck jobs from a previous crash
// 2. Polls the DB every 5 seconds for pending submissions
// 3. Processes ONE job at a time (no parallel processing)
func (w *Worker) Start() {
	// Step 1: Recover stuck jobs from previous crash/restart
	recovered, err := w.db.RecoverStuckSubmissions()
	if err != nil {
		log.Printf("[Worker] WARNING: Failed to recover stuck submissions: %v", err)
	} else if recovered > 0 {
		log.Printf("[Worker] Recovered %d stuck submissions back to pending", recovered)
	}

	// Step 2: Start the worker loop in a goroutine
	go w.loop()
	log.Println("[Worker] Evaluation worker started — polling every 5 seconds")
}

// Stop gracefully stops the worker (waits for current job to finish)
func (w *Worker) Stop() {
	log.Println("[Worker] Stopping evaluation worker...")
	close(w.stopCh) // signal the loop to stop
	<-w.doneCh      // wait for the loop to finish
	log.Println("[Worker] Evaluation worker stopped")
}

// loop is the main worker loop — runs forever until Stop() is called
func (w *Worker) loop() {
	defer close(w.doneCh)

	for {
		select {
		case <-w.stopCh:
			return
		default:
			// Try to claim and process the next pending job
			processed := w.processNext()

			if !processed {
				// No work to do — sleep before checking again
				select {
				case <-w.stopCh:
					return
				case <-time.After(w.pollInterval):
					// Continue to next iteration
				}
			}
			// If we processed a job, immediately check for the next one
			// (no sleep — keeps the queue moving)
		}
	}
}

// processNext claims one pending submission and processes it
// Returns true if a job was found and processed
func (w *Worker) processNext() bool {
	// Atomically claim the next pending submission
	sub, err := w.db.ClaimNextPending()
	if err != nil {
		// No pending submissions — this is normal
		return false
	}
	if sub == nil {
		return false
	}

	log.Printf("[Worker] Processing submission %s (challenge: %s, repo: %s)", sub.ID, sub.ChallengeID, sub.RepoURL)

	// Process the submission through the evaluation pipeline
	w.evaluate(sub)

	return true
}

// evaluate runs the full evaluation pipeline for a submission
func (w *Worker) evaluate(sub *db.SubmissionDetail) {
	// ─── Step 0: Get Challenge Template ──────────────────
	template, err := w.db.GetTemplateByChallenge(sub.ChallengeID)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get challenge template: %v", err)
		log.Printf("[Worker] [%s] ❌ %s", sub.ID, errMsg)
		w.db.UpdateSubmissionStatus(sub.ID, "failed", &errMsg)
		return
	}

	// ─── Step 1: Run Pipeline ────────────────────────────
	// Pipeline handles: Clone -> Validate -> Fetch Tests -> Docker -> Parse -> Score
	log.Printf("[Worker] [%s] Starting evaluation pipeline...", sub.ID)
	w.db.UpdateSubmissionStatus(sub.ID, "testing", nil)

	// Create pipeline config
	pipelineCfg := evaluator.PipelineConfig{
		GitHubToken: w.config.GitHub.PAT,
		Docker:      evaluator.DefaultDockerConfig(),
	}

	result, repoDir, cleanup, err := evaluator.RunPipeline(
		sub.RepoURL,
		sub.Branch,
		sub.ChallengeID,
		template.TestRepoURL,
		template.ForbiddenPackages,
		sub.MaxScore,
		pipelineCfg,
	)

	// Ensure cleanup happens after we are done with everything (including AI)
	if cleanup != nil {
		defer cleanup()
	}

	// ─── Step 2: Handle Pipeline Errors ──────────────────
	if err != nil {
		errMsg := err.Error()
		log.Printf("[Worker] [%s] ❌ Pipeline failed: %v", sub.ID, err)
		w.db.UpdateSubmissionStatus(sub.ID, "failed", &errMsg)
		return
	}

	// ─── Step 3: Save Results ────────────────────────────
	// Save detailed test results (counts + raw output)
	if err := w.db.SaveTestResults(sub.ID, result.TestResult, ""); err != nil {
		log.Printf("[Worker] [%s] ⚠️ Failed to save test results: %v", sub.ID, err)
	}

	var aiReview *models.AIReview

	// ─── Step 3.5: Run AI Review ─────────────────────────
	// Only run AI review if build succeeded (tests ran) AND we have a valid repoDir
	if result.TestResult.Passed >= 0 && repoDir != "" {
		// Update status to reviewing
		w.db.UpdateSubmissionStatus(sub.ID, "reviewing", nil)

		// Run AI Review (Step 2 of pipeline)
		// We pass the context, submission, repoDir, test results, config, and DB pool
		// We create a background context for the AI call to ensure it has its own timeout if needed
		aiCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		var err error
		aiReview, err = evaluator.RunAIReview(aiCtx, sub, repoDir, result.TestResult, w.config, w.db)
		if err != nil {
			log.Printf("[Worker] [%s] ⚠️ AI Review failed: %v", sub.ID, err)
		} else {
			if err := w.db.SaveAIReview(context.Background(), aiReview); err != nil {
				log.Printf("[Worker] [%s] ⚠️ Failed to save AI Review: %v", sub.ID, err)
			} else {
				log.Printf("[Worker] [%s] ✨ AI Review saved (Quality: %d, Arch: %d)", sub.ID, aiReview.CodeQualityScore, aiReview.ArchitectureScore)
			}
		}
	}

	// ─── Step 4: Calculate & Save Final Score ────────────────
	// Initialize score structure
	finalScore := &models.SubmissionScore{
		SubmissionID: sub.ID,
		MaxScore:     result.MaxScore,
	}

	// Fetch challenge to get Rubric
	challenge, err := w.db.GetChallengeByID(sub.ChallengeID)
	var rubric *models.Rubric
	if err == nil && challenge.Rubric != nil {
		var r models.Rubric
		if err := json.Unmarshal(challenge.Rubric, &r); err == nil {
			rubric = &r
		}
	}

	// Calculate scores
	if aiReview != nil {
		finalScore.CalculateFromParts(result.Score, aiReview, rubric)
	} else {
		// Fallback if AI failed: score is based purely on tests
		// We treat test result as "Functionality" partition
		// If we have a rubric, we should ideally scale it, but for now we just map it directly
		// or maybe we should still use CalculateFromParts with empty/zero AI review?
		// For safety, let's just use the old logic for fallback but maybe respect Functionality weight?
		// actually, if AI fails, we probably just want to give points for Functionality portion.

		// If AI failed, we can't give points for other categories.
		// So Final Score = Functionality Score * Functionality Weight
		if rubric != nil {
			finalScore.FunctionalityScore = result.Score
			finalScore.FinalScore = int(float64(result.Score) * (float64(rubric.Functionality) / 100.0))
		} else {
			finalScore.FunctionalityScore = result.Score
			// Default fallback: 50% weight for functionality if we assume default rubric
			finalScore.FinalScore = int(float64(result.Score) * 0.5)
		}
	}

	// Save calculated score to DB
	if err := w.db.SaveScore(sub.ID, finalScore); err != nil {
		log.Printf("[Worker] [%s] ⚠️ Failed to save score: %v", sub.ID, err)
	}

	// ─── Step 5: Update User Stats ──────────────────────
	// Recalculate total score, rank, streaks, etc.
	if err := w.db.UpdateUserStats(context.Background(), sub.UserID); err != nil {
		log.Printf("[Worker] [%s] ⚠️ Failed to update user stats: %v", sub.ID, err)
	}

	// ─── Step 6: Create Notification ────────────────────
	notifTitle := fmt.Sprintf("Challenge evaluated: %s", sub.ChallengeTitle)
	notifMsg := fmt.Sprintf("Your submission for '%s' has been evaluated. Score: %d/%d", sub.ChallengeTitle, finalScore.FinalScore, finalScore.MaxScore)
	notifLink := fmt.Sprintf("/challenges?id=%s", sub.ChallengeID) // Or direct to submission view if available

	if err := w.db.CreateNotification(sub.UserID, notifTitle, notifMsg, "review", notifLink); err != nil {
		log.Printf("[Worker] [%s] ⚠️ Failed to create notification: %v", sub.ID, err)
	}

	w.db.UpdateSubmissionStatus(sub.ID, "completed", &sub.Branch)
	log.Printf("[Worker] [%s] ✅ Evaluation complete (Final Score: %d/%d)", sub.ID, finalScore.FinalScore, finalScore.MaxScore)
}
