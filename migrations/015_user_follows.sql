-- Migration 015: User Follows
-- Adds follow/unfollow functionality between users

-- Core follows junction table
CREATE TABLE IF NOT EXISTS public.user_follows (
    follower_id  character varying NOT NULL,
    following_id character varying NOT NULL,
    created_at   timestamp with time zone DEFAULT now(),
    CONSTRAINT user_follows_pkey PRIMARY KEY (follower_id, following_id),
    CONSTRAINT user_follows_no_self CHECK (follower_id != following_id),
    CONSTRAINT user_follows_follower_fkey FOREIGN KEY (follower_id)
        REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT user_follows_following_fkey FOREIGN KEY (following_id)
        REFERENCES public.users(id) ON DELETE CASCADE
);

-- Indexes for fast lookups in both directions
CREATE INDEX IF NOT EXISTS idx_user_follows_follower  ON user_follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_following ON user_follows(following_id);

-- Denormalized counts on users table for fast reads
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS followers_count integer DEFAULT 0;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS following_count integer DEFAULT 0;
