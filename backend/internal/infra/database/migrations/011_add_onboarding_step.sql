-- ==========================================
-- Migration 011: Add onboarding_step to users
-- ==========================================

-- Adicionar coluna onboarding_step para rastrear o passo atual do onboarding
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_step INTEGER NOT NULL DEFAULT 0;

-- Criar índice para consultas por onboarding_step
CREATE INDEX IF NOT EXISTS idx_users_onboarding_step ON users(onboarding_step);

-- Atualizar usuários existentes baseado no estado atual
-- Se já completou onboarding, step = 5
UPDATE users SET onboarding_step = 5 WHERE has_completed_onboarding = true;

-- Se tem pagamento pago mas não completou onboarding, step = 1 (precisa configurar negócio)
UPDATE users u SET onboarding_step = 1 
WHERE u.has_completed_onboarding = false 
AND u.onboarding_step = 0
AND EXISTS (
    SELECT 1 FROM payments p 
    WHERE p.user_id = u.id AND p.status = 'paid'
);

-- Se tem business_profile, atualizar para step 2
UPDATE users u SET onboarding_step = 2
WHERE u.has_completed_onboarding = false
AND u.onboarding_step = 1
AND EXISTS (
    SELECT 1 FROM business_profiles bp WHERE bp.user_id = u.id
);

-- Se tem concorrentes, atualizar para step 3
UPDATE users u SET onboarding_step = 3
WHERE u.has_completed_onboarding = false
AND u.onboarding_step = 2
AND EXISTS (
    SELECT 1 FROM competitors c WHERE c.user_id = u.id
);

-- Se tem integração WordPress, atualizar para step 4
UPDATE users u SET onboarding_step = 4
WHERE u.has_completed_onboarding = false
AND u.onboarding_step = 3
AND EXISTS (
    SELECT 1 FROM integrations i WHERE i.user_id = u.id AND i.type = 'wordpress'
);
