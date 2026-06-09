-- ============================================================================
-- Migration: Clean up unwanted AI/Grader tables and setup showcase comments & likes
-- ============================================================================

BEGIN;

-- 1. Drop AI/Grader-related tables that are no longer needed
DROP TABLE IF EXISTS public.ai_reviews CASCADE;
DROP TABLE IF EXISTS public.ai_reports CASCADE;
DROP TABLE IF EXISTS public.submission_test_results CASCADE;
DROP TABLE IF EXISTS public.submission_scores CASCADE;
DROP TABLE IF EXISTS public.challenge_templates CASCADE;
DROP TABLE IF EXISTS public.prompt_versions CASCADE;

-- 2. Alter submissions table
-- - Make challenge_id nullable for open showcases
-- - Add video_url and project_description
-- - Remove obsolete evaluator pipeline columns
ALTER TABLE public.submissions 
    ALTER COLUMN challenge_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS title TEXT,
    ADD COLUMN IF NOT EXISTS video_url TEXT,
    ADD COLUMN IF NOT EXISTS project_description TEXT,
    DROP COLUMN IF EXISTS evaluation_status,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS evaluation_started_at,
    DROP COLUMN IF EXISTS evaluation_completed_at,
    DROP COLUMN IF EXISTS commit_hash;

-- 3. Create comments table for global (submission_id is null) and project-specific chat
CREATE TABLE IF NOT EXISTS public.comments (
    id            UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id       CHARACTER VARYING NOT NULL,
    challenge_id  CHARACTER VARYING, -- nullable, for challenge-level general chat
    submission_id CHARACTER VARYING, -- nullable, for project-level chat
    message       TEXT NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT comments_pkey PRIMARY KEY (id),
    CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT comments_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id) ON DELETE CASCADE,
    CONSTRAINT comments_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES public.submissions(id) ON DELETE CASCADE
);

-- 4. Create likes table for project showcase likes
CREATE TABLE IF NOT EXISTS public.submission_likes (
    user_id       CHARACTER VARYING NOT NULL,
    submission_id CHARACTER VARYING NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT submission_likes_pkey PRIMARY KEY (user_id, submission_id),
    CONSTRAINT submission_likes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT submission_likes_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES public.submissions(id) ON DELETE CASCADE
);

-- 5. Create indices for faster queries
CREATE INDEX IF NOT EXISTS idx_comments_challenge_id ON public.comments(challenge_id);
CREATE INDEX IF NOT EXISTS idx_comments_submission_id ON public.comments(submission_id);
CREATE INDEX IF NOT EXISTS idx_comments_created_at ON public.comments(created_at ASC);
CREATE INDEX IF NOT EXISTS idx_submission_likes_submission_id ON public.submission_likes(submission_id);

COMMIT;
