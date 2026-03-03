-- ============================================================================
-- DevArena — Full Database Schema (Production Setup)
-- ============================================================================
--
-- Consolidated migration to create all tables from scratch on a fresh database.
-- Matches the live production schema exactly. No seed data included.
--
-- Tables (13):
--   users, starter_packs, challenges, tags, challenge_tags,
--   challenge_templates, prompt_versions, submissions, ai_reviews,
--   submission_test_results, submission_scores, ai_reports, notifications
--
-- Prerequisites: PostgreSQL 14+
-- Usage: psql -U <user> -d <database> -f 000_full_schema.sql
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. USERS
-- ============================================================================

CREATE TABLE public.users (
    id                    CHARACTER VARYING NOT NULL,
    clerk_user_id         CHARACTER VARYING NOT NULL UNIQUE,
    email                 CHARACTER VARYING NOT NULL UNIQUE,
    username              CHARACTER VARYING UNIQUE,
    display_name          CHARACTER VARYING,
    avatar_url            TEXT,
    bio                   TEXT,
    github_username       CHARACTER VARYING,
    github_connected      BOOLEAN DEFAULT FALSE,
    onboarding_completed  BOOLEAN DEFAULT FALSE,
    current_streak        INTEGER DEFAULT 0,
    longest_streak        INTEGER DEFAULT 0,
    total_score           INTEGER DEFAULT 0,
    rank                  INTEGER DEFAULT 0,
    challenges_completed  INTEGER DEFAULT 0,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT users_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_users_clerk_user_id ON public.users(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON public.users(email);
CREATE INDEX IF NOT EXISTS idx_users_github_username ON public.users(github_username);
CREATE INDEX IF NOT EXISTS idx_users_total_score ON public.users(total_score DESC);

-- ============================================================================
-- 2. STARTER PACKS
-- ============================================================================

CREATE TABLE public.starter_packs (
    id                CHARACTER VARYING NOT NULL,
    user_id           CHARACTER VARYING NOT NULL UNIQUE,
    experience        CHARACTER VARYING,
    paths             JSONB DEFAULT '[]'::jsonb,
    technologies      JSONB DEFAULT '[]'::jsonb,
    challenge_ids     JSONB DEFAULT '[]'::jsonb,
    current_progress  INTEGER DEFAULT 0,
    total_challenges  INTEGER DEFAULT 0,
    is_active         BOOLEAN DEFAULT TRUE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT starter_packs_pkey PRIMARY KEY (id),
    CONSTRAINT starter_packs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE INDEX IF NOT EXISTS idx_starter_packs_user_id ON public.starter_packs(user_id);

-- ============================================================================
-- 3. CHALLENGES
-- ============================================================================

CREATE TABLE public.challenges (
    id                CHARACTER VARYING NOT NULL,
    title             CHARACTER VARYING NOT NULL,
    description       TEXT NOT NULL,
    difficulty        CHARACTER VARYING NOT NULL,
    type              CHARACTER VARYING DEFAULT 'project'::character varying,
    max_score         INTEGER DEFAULT 100,
    repo_template_url TEXT,
    requirements      JSONB DEFAULT '[]'::jsonb,
    tech_stack        JSONB DEFAULT '[]'::jsonb,
    estimated_hours   INTEGER DEFAULT 4,
    is_published      BOOLEAN DEFAULT FALSE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    description_md    TEXT,
    rubric            JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT challenges_pkey PRIMARY KEY (id)
);

-- ============================================================================
-- 4. TAGS
-- ============================================================================

CREATE TABLE public.tags (
    id          CHARACTER VARYING NOT NULL,
    name        CHARACTER VARYING NOT NULL UNIQUE,
    slug        CHARACTER VARYING NOT NULL UNIQUE,
    category    CHARACTER VARYING,
    color       CHARACTER VARYING,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT tags_pkey PRIMARY KEY (id)
);

-- ============================================================================
-- 5. CHALLENGE_TAGS (Junction Table)
-- ============================================================================

CREATE TABLE public.challenge_tags (
    challenge_id  CHARACTER VARYING NOT NULL,
    tag_id        CHARACTER VARYING NOT NULL,
    CONSTRAINT challenge_tags_pkey PRIMARY KEY (challenge_id, tag_id),
    CONSTRAINT challenge_tags_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id),
    CONSTRAINT challenge_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id)
);

-- ============================================================================
-- 6. CHALLENGE TEMPLATES
-- ============================================================================

CREATE TABLE public.challenge_templates (
    id                  CHARACTER VARYING NOT NULL,
    challenge_id        CHARACTER VARYING NOT NULL UNIQUE,
    repo_template_url   TEXT NOT NULL,
    entry_file          CHARACTER VARYING DEFAULT 'src/index.tsx'::character varying,
    allowed_edit_paths  JSONB DEFAULT '[]'::jsonb,
    readonly_paths      JSONB DEFAULT '[]'::jsonb,
    forbidden_packages  JSONB DEFAULT '[]'::jsonb,
    template_tree       JSONB DEFAULT '[]'::jsonb,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    test_repo_url       TEXT,
    CONSTRAINT challenge_templates_pkey PRIMARY KEY (id),
    CONSTRAINT challenge_templates_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id)
);

CREATE INDEX IF NOT EXISTS idx_challenge_templates_challenge_id ON public.challenge_templates(challenge_id);

-- ============================================================================
-- 7. PROMPT VERSIONS (Auditing & A/B Testing)
-- ============================================================================

CREATE TABLE public.prompt_versions (
    id          CHARACTER VARYING NOT NULL,
    name        CHARACTER VARYING NOT NULL,
    step        CHARACTER VARYING NOT NULL CHECK (step::text = ANY (ARRAY['code_review'::character varying, 'report'::character varying]::text[])),
    model       CHARACTER VARYING NOT NULL,
    prompt_text TEXT NOT NULL,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "Version"   BIGINT,
    CONSTRAINT prompt_versions_pkey PRIMARY KEY (id)
);

-- ============================================================================
-- 8. SUBMISSIONS
-- ============================================================================

CREATE TABLE public.submissions (
    id                       CHARACTER VARYING NOT NULL,
    user_id                  CHARACTER VARYING NOT NULL,
    challenge_id             CHARACTER VARYING NOT NULL,
    repo_url                 TEXT NOT NULL,
    branch                   CHARACTER VARYING DEFAULT 'main'::character varying,
    commit_hash              CHARACTER VARYING,
    created_at               TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at               TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    evaluation_status        CHARACTER VARYING NOT NULL DEFAULT 'pending'::character varying,
    error_message            TEXT,
    evaluation_started_at    TIMESTAMP WITH TIME ZONE,
    evaluation_completed_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT submissions_pkey PRIMARY KEY (id),
    CONSTRAINT submissions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id),
    CONSTRAINT submissions_challenge_id_fkey FOREIGN KEY (challenge_id) REFERENCES public.challenges(id)
);

CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON public.submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_submissions_challenge_id ON public.submissions(challenge_id);

-- ============================================================================
-- 9. AI REVIEWS (Step 2 — AI Code Review)
-- ============================================================================

CREATE TABLE public.ai_reviews (
    id                  CHARACTER VARYING NOT NULL,
    submission_id       CHARACTER VARYING NOT NULL,
    reviewed_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    code_quality_score  INTEGER NOT NULL DEFAULT 0,
    constraint_score    INTEGER NOT NULL DEFAULT 0,
    architecture_score  INTEGER NOT NULL DEFAULT 0,
    strengths_json      JSONB DEFAULT '[]'::jsonb,
    issues_json         JSONB DEFAULT '[]'::jsonb,
    improvements_json   JSONB DEFAULT '[]'::jsonb,
    prompt_version_id   CHARACTER VARYING,
    raw_response        TEXT,
    CONSTRAINT ai_reviews_pkey PRIMARY KEY (id),
    CONSTRAINT ai_reviews_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES public.submissions(id),
    CONSTRAINT ai_reviews_prompt_version_id_fkey FOREIGN KEY (prompt_version_id) REFERENCES public.prompt_versions(id)
);

CREATE INDEX IF NOT EXISTS idx_ai_reviews_submission_id ON public.ai_reviews(submission_id);

-- ============================================================================
-- 10. SUBMISSION TEST RESULTS (Step 1 — Docker Test Execution)
-- ============================================================================

CREATE TABLE public.submission_test_results (
    id                      CHARACTER VARYING NOT NULL,
    submission_id           CHARACTER VARYING NOT NULL UNIQUE,
    build_success           BOOLEAN NOT NULL DEFAULT FALSE,
    tests_total             INTEGER NOT NULL DEFAULT 0,
    tests_passed            INTEGER NOT NULL DEFAULT 0,
    tests_failed            INTEGER NOT NULL DEFAULT 0,
    functionality_score     INTEGER NOT NULL DEFAULT 0,
    max_functionality_score INTEGER NOT NULL DEFAULT 0,
    execution_time_ms       INTEGER DEFAULT 0,
    memory_usage_mb         INTEGER DEFAULT 0,
    raw_output              TEXT,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT submission_test_results_pkey PRIMARY KEY (id),
    CONSTRAINT submission_test_results_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES public.submissions(id)
);

CREATE INDEX IF NOT EXISTS idx_submission_test_results_submission_id 
    ON public.submission_test_results(submission_id);

-- ============================================================================
-- 11. SUBMISSION SCORES (Final Computed Score — Calculated in Go)
-- ============================================================================

CREATE TABLE public.submission_scores (
    id                  CHARACTER VARYING NOT NULL,
    submission_id       CHARACTER VARYING NOT NULL UNIQUE,
    functionality_score INTEGER NOT NULL DEFAULT 0,
    code_quality_score  INTEGER NOT NULL DEFAULT 0,
    constraint_score    INTEGER NOT NULL DEFAULT 0,
    architecture_score  INTEGER NOT NULL DEFAULT 0,
    final_score         INTEGER NOT NULL DEFAULT 0,
    max_score           INTEGER NOT NULL DEFAULT 100,
    created_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT submission_scores_pkey PRIMARY KEY (id),
    CONSTRAINT submission_scores_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES public.submissions(id)
);

CREATE INDEX IF NOT EXISTS idx_submission_scores_submission_id 
    ON public.submission_scores(submission_id);

-- ============================================================================
-- 12. AI REPORTS (Step 3 — Professional Feedback Report)
-- ============================================================================

CREATE TABLE public.ai_reports (
    id                   CHARACTER VARYING NOT NULL,
    submission_id        CHARACTER VARYING NOT NULL UNIQUE,
    summary_md           TEXT,
    detailed_feedback_md TEXT,
    dos_json             JSONB DEFAULT '[]'::jsonb,
    donts_json           JSONB DEFAULT '[]'::jsonb,
    next_steps_json      JSONB DEFAULT '[]'::jsonb,
    prompt_version_id    CHARACTER VARYING,
    raw_response         TEXT,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT ai_reports_pkey PRIMARY KEY (id),
    CONSTRAINT ai_reports_prompt_version_id_fkey FOREIGN KEY (prompt_version_id) REFERENCES public.prompt_versions(id),
    CONSTRAINT ai_reports_submission_id_fkey FOREIGN KEY (submission_id) REFERENCES public.submissions(id)
);

CREATE INDEX IF NOT EXISTS idx_ai_reports_submission_id 
    ON public.ai_reports(submission_id);

-- ============================================================================
-- 13. NOTIFICATIONS
-- ============================================================================

CREATE TABLE public.notifications (
    id          UUID NOT NULL DEFAULT gen_random_uuid(),
    user_id     CHARACTER VARYING NOT NULL,
    title       CHARACTER VARYING NOT NULL,
    message     TEXT,
    type        CHARACTER VARYING DEFAULT 'info'::character varying,
    link        CHARACTER VARYING,
    is_read     BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT notifications_pkey PRIMARY KEY (id),
    CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON public.notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON public.notifications(is_read);

-- ============================================================================
-- DONE! All 13 tables, indexes, and constraints created.
-- ============================================================================

COMMIT;
