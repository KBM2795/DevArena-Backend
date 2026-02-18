-- Add index on total_score for optimized leaderboard performance
-- This is critical for scaling to 200+ concurrent users

CREATE INDEX IF NOT EXISTS idx_users_total_score ON users(total_score DESC);
