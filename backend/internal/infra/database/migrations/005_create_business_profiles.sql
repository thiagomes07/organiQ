-- ==========================================
-- Migration 005: Create Business Profiles Table
-- ==========================================

CREATE TABLE IF NOT EXISTS business_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    primary_objective VARCHAR(20) NOT NULL,
    secondary_objective VARCHAR(20),
    location JSONB NOT NULL,
    site_url TEXT,
    has_blog BOOLEAN NOT NULL DEFAULT FALSE,
    blog_urls JSONB,
    brand_file_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_business_profiles_user_id ON business_profiles(user_id);