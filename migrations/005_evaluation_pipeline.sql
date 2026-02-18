-- Migration 005: Evaluation Pipeline
-- Implements DevArena Evaluation Spec v1
-- Adds: submission_test_results, submission_scores, ai_reports, prompt_versions
-- Alters: submissions, ai_reviews, challenges, challenge_templates

-- ============================================================================
-- 1. PROMPT VERSIONS TABLE
-- Track which prompt produced which AI review/report for auditing & A/B testing
-- ============================================================================

CREATE TABLE IF NOT EXISTS prompt_versions (
    id          VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,                          -- e.g. "code_review_v1", "report_v2"
    step        VARCHAR(20) NOT NULL CHECK (step IN ('code_review', 'report')),
    model       VARCHAR(50) NOT NULL,                           -- "gpt-4o", "gpt-4", "gpt-4o-mini"
    prompt_text TEXT NOT NULL,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- 2. SUBMISSION TEST RESULTS TABLE (Step 1 Output)
-- Deterministic Docker test execution results — no AI involved
-- ============================================================================

CREATE TABLE IF NOT EXISTS submission_test_results (
    id                      VARCHAR(255) PRIMARY KEY,
    submission_id           VARCHAR(255) NOT NULL UNIQUE REFERENCES submissions(id) ON DELETE CASCADE,
    build_success           BOOLEAN NOT NULL DEFAULT FALSE,
    tests_total             INTEGER NOT NULL DEFAULT 0,
    tests_passed            INTEGER NOT NULL DEFAULT 0,
    tests_failed            INTEGER NOT NULL DEFAULT 0,
    functionality_score     INTEGER NOT NULL DEFAULT 0,
    max_functionality_score INTEGER NOT NULL DEFAULT 0,
    execution_time_ms       INTEGER DEFAULT 0,
    memory_usage_mb         INTEGER DEFAULT 0,
    raw_output              TEXT,                               -- Raw Docker stdout/stderr for debugging
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submission_test_results_submission_id 
    ON submission_test_results(submission_id);

-- ============================================================================
-- 3. SUBMISSION SCORES TABLE (Final Computed Score)
-- Calculated in Go only — never by AI
-- final_score = functionality + code_quality + constraints + architecture
-- ============================================================================

CREATE TABLE IF NOT EXISTS submission_scores (
    id                  VARCHAR(255) PRIMARY KEY,
    submission_id       VARCHAR(255) NOT NULL UNIQUE REFERENCES submissions(id) ON DELETE CASCADE,
    functionality_score INTEGER NOT NULL DEFAULT 0,
    code_quality_score  INTEGER NOT NULL DEFAULT 0,
    constraint_score    INTEGER NOT NULL DEFAULT 0,
    architecture_score  INTEGER NOT NULL DEFAULT 0,
    final_score         INTEGER NOT NULL DEFAULT 0,
    max_score           INTEGER NOT NULL DEFAULT 100,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submission_scores_submission_id 
    ON submission_scores(submission_id);

-- ============================================================================
-- 4. AI REPORTS TABLE (Step 3 Output)
-- Professional feedback report — AI must NOT rescore or reinterpret code
-- ============================================================================

CREATE TABLE IF NOT EXISTS ai_reports (
    id                   VARCHAR(255) PRIMARY KEY,
    submission_id        VARCHAR(255) NOT NULL UNIQUE REFERENCES submissions(id) ON DELETE CASCADE,
    summary_md           TEXT,
    detailed_feedback_md TEXT,                                  -- JSON with per-category feedback
    dos_json             JSONB DEFAULT '[]',
    donts_json           JSONB DEFAULT '[]',
    next_steps_json      JSONB DEFAULT '[]',
    prompt_version_id    VARCHAR(255) REFERENCES prompt_versions(id),
    raw_response         TEXT,                                  -- Raw AI output for auditing
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_reports_submission_id 
    ON ai_reports(submission_id);

-- ============================================================================
-- 5. ALTER SUBMISSIONS TABLE
-- Add granular evaluation pipeline status tracking
-- ============================================================================

-- Add evaluation pipeline status (more granular than current 'status')
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS evaluation_status VARCHAR(30) DEFAULT 'pending';

-- Add error tracking
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS error_message TEXT;

-- Add pipeline timing
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS evaluation_started_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS evaluation_completed_at TIMESTAMP WITH TIME ZONE;

-- Add constraint for valid evaluation statuses
-- Valid: pending, queued, testing, reviewing, reporting, completed, failed
COMMENT ON COLUMN submissions.evaluation_status IS 
    'Pipeline status: pending → queued → testing → reviewing → reporting → completed | failed';

-- ============================================================================
-- 6. ALTER AI_REVIEWS TABLE
-- Restructure: remove AI-controlled overall_score, add structured category scores
-- ============================================================================

-- Add new structured score columns
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS code_quality_score INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS constraint_score INTEGER NOT NULL DEFAULT 0;
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS architecture_score INTEGER NOT NULL DEFAULT 0;

-- Add structured feedback arrays
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS strengths_json JSONB DEFAULT '[]';
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS issues_json JSONB DEFAULT '[]';
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS improvements_json JSONB DEFAULT '[]';

-- Add prompt versioning for auditing
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS prompt_version_id VARCHAR(255) REFERENCES prompt_versions(id);

-- Add raw response storage for auditing
ALTER TABLE ai_reviews ADD COLUMN IF NOT EXISTS raw_response TEXT;

-- NOTE: We keep the old columns (overall_score, categories, feedback, suggestions) 
-- for backward compatibility. They can be dropped in a future migration once
-- all code is updated to use the new columns.
-- 
-- To drop later:
-- ALTER TABLE ai_reviews DROP COLUMN IF EXISTS overall_score;
-- ALTER TABLE ai_reviews DROP COLUMN IF EXISTS categories;
-- ALTER TABLE ai_reviews DROP COLUMN IF EXISTS feedback;
-- ALTER TABLE ai_reviews DROP COLUMN IF EXISTS suggestions;

-- ============================================================================
-- 7. ALTER CHALLENGES TABLE
-- Add rubric (dynamic per challenge) and normalize max_score to 100
-- ============================================================================

-- Add rubric column — defines scoring categories & weights per challenge
-- Example: {"functionality": 50, "code_quality": 25, "constraints": 15, "architecture": 10}
ALTER TABLE challenges ADD COLUMN IF NOT EXISTS rubric JSONB DEFAULT '{}';

-- Normalize all challenges to max_score = 100
UPDATE challenges SET max_score = 100 WHERE max_score != 100;

-- Backfill rubric for existing challenges
-- Default rubric: functionality 50, code_quality 25, constraints 15, architecture 10
UPDATE challenges SET rubric = '{
    "functionality": 50,
    "code_quality": 25,
    "constraints": 15,
    "architecture": 10
}'::jsonb
WHERE rubric = '{}'::jsonb OR rubric IS NULL;

-- ============================================================================
-- 8. ALTER CHALLENGE TEMPLATES TABLE
-- Add private test repo URL for Docker eval
-- ============================================================================

ALTER TABLE challenge_templates ADD COLUMN IF NOT EXISTS test_repo_url TEXT;

-- Backfill test repo URLs (private repos for grading)
-- Format: https://github.com/autonerveai27/<challenge-slug>-tests
UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-1-css-flexbox-tests' 
WHERE challenge_id = 'challenge-1' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-2-react-counter-tests' 
WHERE challenge_id = 'challenge-2' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-3-infinite-scroll-tests' 
WHERE challenge_id = 'challenge-3' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-4-custom-dropdown-tests' 
WHERE challenge_id = 'challenge-4' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-5-nodejs-file-upload-tests' 
WHERE challenge_id = 'challenge-5' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-6-rate-limiter-tests' 
WHERE challenge_id = 'challenge-6' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-7-sql-complex-join-tests' 
WHERE challenge_id = 'challenge-7' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-8-sentiment-analysis-tests' 
WHERE challenge_id = 'challenge-8' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-9-image-classification-tests' 
WHERE challenge_id = 'challenge-9' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-10-debug-login-tests' 
WHERE challenge_id = 'challenge-10' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-11-go-concurrency-tests' 
WHERE challenge_id = 'challenge-11' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-12-task-manager-tests' 
WHERE challenge_id = 'challenge-12' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-13-threejs-cube-tests' 
WHERE challenge_id = 'challenge-13' AND test_repo_url IS NULL;

UPDATE challenge_templates SET test_repo_url = 'https://github.com/autonerveai27/challenge-14-nextjs-blog-tests' 
WHERE challenge_id = 'challenge-14' AND test_repo_url IS NULL;

-- ============================================================================
-- 9. VERIFICATION QUERIES
-- Run these to confirm the migration applied correctly
-- ============================================================================

-- Verify new tables exist
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name IN ('submission_test_results', 'submission_scores', 'ai_reports', 'prompt_versions')
ORDER BY table_name;
    
-- Verify submissions has new columns
SELECT column_name, data_type FROM information_schema.columns 
WHERE table_name = 'submissions' 
AND column_name IN ('evaluation_status', 'error_message', 'evaluation_started_at', 'evaluation_completed_at');

-- Verify ai_reviews has new columns
SELECT column_name, data_type FROM information_schema.columns 
WHERE table_name = 'ai_reviews' 
AND column_name IN ('code_quality_score', 'constraint_score', 'architecture_score', 'strengths_json', 'issues_json', 'improvements_json', 'prompt_version_id', 'raw_response');

-- Verify challenges rubric + normalized max_score
SELECT id, title, max_score, rubric FROM challenges ORDER BY id;

-- Verify challenge_templates has test_repo_url
SELECT challenge_id, test_repo_url FROM challenge_templates ORDER BY challenge_id;
