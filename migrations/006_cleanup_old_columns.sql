-- Migration 006: Cleanup Old Columns
-- Removes deprecated columns that are replaced by the Evaluation Pipeline (005)
-- 
-- ⚠️  WARNING: This is DESTRUCTIVE. Only run after confirming 005 applied successfully
--     and no code references the old columns anymore.

-- ============================================================================
-- 1. DROP OLD COLUMNS FROM ai_reviews
-- 
-- Removed:
--   overall_score  → replaced by submission_scores.final_score (computed in Go)
--   categories     → replaced by code_quality_score, constraint_score, architecture_score
--   feedback       → replaced by strengths_json, issues_json, improvements_json
--   suggestions    → replaced by improvements_json
-- ============================================================================

ALTER TABLE ai_reviews DROP COLUMN IF EXISTS overall_score;
ALTER TABLE ai_reviews DROP COLUMN IF EXISTS categories;
ALTER TABLE ai_reviews DROP COLUMN IF EXISTS feedback;
ALTER TABLE ai_reviews DROP COLUMN IF EXISTS suggestions;

-- ============================================================================
-- 2. DROP OLD COLUMNS FROM submissions
-- 
-- Removed:
--   status  → replaced by evaluation_status (more granular pipeline tracking)
--   score   → replaced by submission_scores table (proper score breakdown)
-- ============================================================================

ALTER TABLE submissions DROP COLUMN IF EXISTS status;
ALTER TABLE submissions DROP COLUMN IF EXISTS score;

-- ============================================================================
-- 3. ADD NOT NULL CONSTRAINT ON evaluation_status
-- Now that 'status' is gone, evaluation_status becomes the primary status field
-- ============================================================================

ALTER TABLE submissions ALTER COLUMN evaluation_status SET NOT NULL;
ALTER TABLE submissions ALTER COLUMN evaluation_status SET DEFAULT 'pending';

-- ============================================================================
-- 4. ADD ON DELETE CASCADE to ai_reports and submission tables
-- Ensures cleanup when submissions are deleted
-- ============================================================================

-- ai_reports: update FK to cascade on delete
ALTER TABLE ai_reports DROP CONSTRAINT IF EXISTS ai_reports_submission_id_fkey;
ALTER TABLE ai_reports ADD CONSTRAINT ai_reports_submission_id_fkey 
    FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE;

-- submission_test_results: update FK to cascade on delete
ALTER TABLE submission_test_results DROP CONSTRAINT IF EXISTS submission_test_results_submission_id_fkey;
ALTER TABLE submission_test_results ADD CONSTRAINT submission_test_results_submission_id_fkey 
    FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE;

-- submission_scores: update FK to cascade on delete
ALTER TABLE submission_scores DROP CONSTRAINT IF EXISTS submission_scores_submission_id_fkey;
ALTER TABLE submission_scores ADD CONSTRAINT submission_scores_submission_id_fkey 
    FOREIGN KEY (submission_id) REFERENCES submissions(id) ON DELETE CASCADE;

-- ============================================================================
-- 5. VERIFICATION — Confirm final clean schema
-- ============================================================================

-- ai_reviews should now have ONLY these columns:
SELECT column_name, data_type, is_nullable, column_default 
FROM information_schema.columns 
WHERE table_name = 'ai_reviews' 
ORDER BY ordinal_position;

-- submissions should now have ONLY these columns:
SELECT column_name, data_type, is_nullable, column_default 
FROM information_schema.columns 
WHERE table_name = 'submissions' 
ORDER BY ordinal_position;
