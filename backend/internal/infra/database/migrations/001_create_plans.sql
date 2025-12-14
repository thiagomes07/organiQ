-- ==========================================
-- Migration 001: Create Plans Table
-- ==========================================

CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    max_articles INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    features JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plans_name ON plans(name);
CREATE INDEX IF NOT EXISTS idx_plans_active ON plans(active);

-- Seed initial plans
INSERT INTO plans (name, max_articles, price, features) VALUES
    ('Free', 0, 0.00, '["Teste grátis", "Suporte limitado"]'),
    ('Starter', 5, 49.90, '["5 matérias/mês", "SEO básico", "Suporte email"]'),
    ('Pro', 15, 99.90, '["15 matérias/mês", "SEO avançado", "Suporte prioritário"]'),
    ('Enterprise', 50, 249.90, '["50 matérias/mês", "SEO premium", "Suporte dedicado"]')
ON CONFLICT (name) DO NOTHING;