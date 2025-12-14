-- ==========================================
-- Migration 008: Create Article Jobs Table
-- ==========================================

CREATE TABLE IF NOT EXISTS article_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    progress INTEGER NOT NULL DEFAULT 0,
    payload JSONB NOT NULL,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_article_jobs_user_id ON article_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_article_jobs_status ON article_jobs(status);
CREATE INDEX IF NOT EXISTS idx_article_jobs_type ON article_jobs(type);
CREATE INDEX IF NOT EXISTS idx_article_jobs_created_at ON article_jobs(created_at DESC);