-- ──────────────────────────────────
-- 014: 拆分 base_url 为按协议区分的 URL
-- 将单一 base_url 拆分为 openai_base_url 和 anthropic_base_url，
-- 支持同一服务商不同协议使用不同的 Base URL
-- ──────────────────────────────────

-- 模板表：添加新列（如果不存在）
ALTER TABLE third_party_provider_templates ADD COLUMN IF NOT EXISTS openai_base_url VARCHAR(500);
ALTER TABLE third_party_provider_templates ADD COLUMN IF NOT EXISTS anthropic_base_url VARCHAR(500);

-- 用户服务表：添加新列（如果不存在）
ALTER TABLE user_third_party_providers ADD COLUMN IF NOT EXISTS openai_base_url VARCHAR(500);
ALTER TABLE user_third_party_providers ADD COLUMN IF NOT EXISTS anthropic_base_url VARCHAR(500);

-- 迁移已有数据：根据 format 将旧 base_url 值写入对应的新列
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'third_party_provider_templates' AND column_name = 'base_url') THEN
        UPDATE third_party_provider_templates SET
            openai_base_url    = CASE WHEN format IN ('openai', 'all') THEN base_url ELSE openai_base_url END,
            anthropic_base_url = CASE WHEN format IN ('anthropic', 'all') THEN base_url ELSE anthropic_base_url END
        WHERE base_url IS NOT NULL AND base_url != '';

        ALTER TABLE third_party_provider_templates DROP COLUMN base_url;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_third_party_providers' AND column_name = 'base_url') THEN
        UPDATE user_third_party_providers SET
            openai_base_url    = CASE WHEN format IN ('openai', 'all') THEN base_url ELSE openai_base_url END,
            anthropic_base_url = CASE WHEN format IN ('anthropic', 'all') THEN base_url ELSE anthropic_base_url END
        WHERE base_url IS NOT NULL AND base_url != '';

        ALTER TABLE user_third_party_providers DROP COLUMN base_url;
    END IF;
END $$;

COMMENT ON COLUMN third_party_provider_templates.openai_base_url IS 'OpenAI 协议 Base URL，format 含 openai 时必填';
COMMENT ON COLUMN third_party_provider_templates.anthropic_base_url IS 'Anthropic 协议 Base URL，format 含 anthropic 时必填';
COMMENT ON COLUMN user_third_party_providers.openai_base_url IS 'OpenAI 协议 Base URL，format 含 openai 时必填';
COMMENT ON COLUMN user_third_party_providers.anthropic_base_url IS 'Anthropic 协议 Base URL，format 含 anthropic 时必填';
