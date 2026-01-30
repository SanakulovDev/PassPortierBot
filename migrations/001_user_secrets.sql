-- ============================================================================
-- PassPortierBot: user_secrets Table Schema
-- ============================================================================
-- This table stores encrypted user secrets with a unique constraint on
-- (user_id, key_name) to enable efficient UPSERT operations.
--
-- NOTE: If using GORM, this migration happens automatically via AutoMigrate.
--       This file serves as documentation and for manual/raw SQL setups.
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_secrets (
    -- Primary key with auto-increment
    id BIGSERIAL PRIMARY KEY,
    
    -- Telegram user ID (indexed for fast lookups)
    user_id BIGINT NOT NULL,
    
    -- Secret key name (e.g., "instagram", "gmail", "wifi")
    key_name VARCHAR(255) NOT NULL,
    
    -- Base64 encoded encrypted value: Salt[16] + Nonce[12] + Ciphertext + AuthTag[16]
    encrypted_value TEXT NOT NULL,
    
    -- Timestamps for auditing
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE  -- Soft delete support
);

-- ============================================================================
-- CRITICAL: Unique constraint for UPSERT support
-- This enables ON CONFLICT (user_id, key_name) DO UPDATE
-- ============================================================================
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_secrets_user_key 
    ON user_secrets (user_id, key_name) 
    WHERE deleted_at IS NULL;

-- Index for fast user lookups
CREATE INDEX IF NOT EXISTS idx_user_secrets_user_id 
    ON user_secrets (user_id);

-- ============================================================================
-- Example UPSERT Query (for reference):
-- 
-- INSERT INTO user_secrets (user_id, key_name, encrypted_value, updated_at)
-- VALUES ($1, $2, $3, NOW())
-- ON CONFLICT (user_id, key_name) 
-- DO UPDATE SET 
--     encrypted_value = EXCLUDED.encrypted_value,
--     updated_at = NOW();
-- ============================================================================
