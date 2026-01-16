-- Create users table for DevArena
-- Run this in your PostgreSQL database

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    clerk_user_id VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) UNIQUE,
    display_name VARCHAR(255),
    avatar_url TEXT,
    bio TEXT,
    github_username VARCHAR(100),
    github_connected BOOLEAN DEFAULT FALSE,
    onboarding_completed BOOLEAN DEFAULT FALSE,
    current_streak INTEGER DEFAULT 0,
    longest_streak INTEGER DEFAULT 0,
    total_score INTEGER DEFAULT 0,
    rank INTEGER DEFAULT 0,
    challenges_completed INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_clerk_user_id ON users(clerk_user_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_github_username ON users(github_username);

-- Create starter_packs table
CREATE TABLE IF NOT EXISTS starter_packs (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    experience VARCHAR(50),
    paths JSONB DEFAULT '[]',
    technologies JSONB DEFAULT '[]',
    challenge_ids JSONB DEFAULT '[]',
    current_progress INTEGER DEFAULT 0,
    total_challenges INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_starter_packs_user_id ON starter_packs(user_id);

-- Create challenges table
CREATE TABLE IF NOT EXISTS challenges (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    difficulty VARCHAR(20) NOT NULL,
    type VARCHAR(50) DEFAULT 'project',
    max_score INTEGER DEFAULT 100,
    repo_template_url TEXT,
    requirements JSONB DEFAULT '[]',
    tech_stack JSONB DEFAULT '[]',
    estimated_hours INTEGER DEFAULT 4,
    is_published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(100),
    color VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create challenge_tags junction table
CREATE TABLE IF NOT EXISTS challenge_tags (
    challenge_id VARCHAR(255) REFERENCES challenges(id) ON DELETE CASCADE,
    tag_id VARCHAR(255) REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (challenge_id, tag_id)
);

-- Create submissions table
CREATE TABLE IF NOT EXISTS submissions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id VARCHAR(255) NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    repo_url TEXT NOT NULL,
    branch VARCHAR(100) DEFAULT 'main',
    commit_hash VARCHAR(64),
    status VARCHAR(50) DEFAULT 'pending',
    score INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submissions_user_id ON submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_submissions_challenge_id ON submissions(challenge_id);

-- Create ai_reviews table
CREATE TABLE IF NOT EXISTS ai_reviews (
    id VARCHAR(255) PRIMARY KEY,
    submission_id VARCHAR(255) NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    overall_score INTEGER NOT NULL,
    categories JSONB DEFAULT '[]',
    feedback TEXT,
    suggestions JSONB DEFAULT '[]',
    reviewed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_reviews_submission_id ON ai_reviews(submission_id);
