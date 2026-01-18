-- ==========================================
-- Migration 009: Create Article Ideas Table
-- ==========================================

CREATE TABLE IF NOT EXISTS article_ideas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES article_jobs(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    summary TEXT NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT FALSE,
    feedback TEXT,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_article_ideas_user_id ON article_ideas(user_id);
CREATE INDEX IF NOT EXISTS idx_article_ideas_job_id ON article_ideas(job_id);
CREATE INDEX IF NOT EXISTS idx_article_ideas_approved ON article_ideas(approved);
CREATE INDEX IF NOT EXISTS idx_article_ideas_generated_at ON article_ideas(generated_at);
