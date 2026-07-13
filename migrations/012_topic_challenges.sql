-- ============================================================================
-- Migration 012: Add support for topic-only challenges
-- These are challenges with only a topic and description — no pre-built
-- GitHub repo template. Users build from scratch based on the topic info.
-- ============================================================================

BEGIN;

-- 1. Add a CHECK constraint to validate allowed challenge types
--    (existing rows with 'project', 'feature', 'refactor', 'bugfix' are safe)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'challenges_type_check'
    ) THEN
        ALTER TABLE public.challenges
            ADD CONSTRAINT challenges_type_check
            CHECK (type IN ('project', 'feature', 'refactor', 'bugfix', 'topic'));
    END IF;
END $$;

-- 2. Ensure repo_template_url explicitly allows NULL (it already does,
--    but this documents the intent for topic challenges)
COMMENT ON COLUMN public.challenges.repo_template_url IS
    'Starter template repo URL. NULL for topic-only challenges.';

-- 3. Add an index on type for filtering by challenge type
CREATE INDEX IF NOT EXISTS idx_challenges_type ON public.challenges(type);

COMMIT;
