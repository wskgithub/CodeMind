-- ──────────────────────────────────
-- 013: 第三方模型服务相关表
-- ──────────────────────────────────

-- 第三方模型服务模板（管理员配置）
-- 管理员预置常见 LLM 服务商模板，用户创建时仅需填入 API Key
CREATE TABLE third_party_provider_templates (
    id                 BIGSERIAL PRIMARY KEY,
    name               VARCHAR(100) NOT NULL,
    openai_base_url    VARCHAR(500),
    anthropic_base_url VARCHAR(500),
    models             JSONB NOT NULL DEFAULT '[]',
    format             VARCHAR(20) NOT NULL DEFAULT 'openai',
    description        VARCHAR(500),
    icon               VARCHAR(100),
    status             SMALLINT NOT NULL DEFAULT 1,
    sort_order         INT NOT NULL DEFAULT 0,
    created_by         BIGINT NOT NULL REFERENCES users(id),
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE third_party_provider_templates IS '第三方模型服务模板（管理员配置）';
COMMENT ON COLUMN third_party_provider_templates.openai_base_url IS 'OpenAI 协议 Base URL，format 含 openai 时必填';
COMMENT ON COLUMN third_party_provider_templates.anthropic_base_url IS 'Anthropic 协议 Base URL，format 含 anthropic 时必填';
COMMENT ON COLUMN third_party_provider_templates.models IS '模型名称列表，JSON 字符串数组';
COMMENT ON COLUMN third_party_provider_templates.format IS '支持的协议格式: openai, anthropic, all';
COMMENT ON COLUMN third_party_provider_templates.icon IS '图标标识，用于前端展示';
COMMENT ON COLUMN third_party_provider_templates.sort_order IS '排序权重，越小越靠前';

CREATE INDEX idx_tppt_status ON third_party_provider_templates(status) WHERE deleted_at IS NULL;

-- 用户第三方模型服务（用户独立绑定）
CREATE TABLE user_third_party_providers (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users(id),
    name               VARCHAR(100) NOT NULL,
    openai_base_url    VARCHAR(500),
    anthropic_base_url VARCHAR(500),
    api_key_encrypted  TEXT NOT NULL,
    models             JSONB NOT NULL DEFAULT '[]',
    format             VARCHAR(20) NOT NULL DEFAULT 'openai',
    template_id        BIGINT REFERENCES third_party_provider_templates(id),
    status             SMALLINT NOT NULL DEFAULT 1,
    created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE user_third_party_providers IS '用户第三方模型服务（每个用户独立绑定）';
COMMENT ON COLUMN user_third_party_providers.openai_base_url IS 'OpenAI 协议 Base URL，format 含 openai 时必填';
COMMENT ON COLUMN user_third_party_providers.anthropic_base_url IS 'Anthropic 协议 Base URL，format 含 anthropic 时必填';
COMMENT ON COLUMN user_third_party_providers.api_key_encrypted IS 'AES-256-GCM 加密的 API Key';
COMMENT ON COLUMN user_third_party_providers.models IS '模型名称列表，JSON 字符串数组';
COMMENT ON COLUMN user_third_party_providers.format IS '支持的协议格式: openai, anthropic, all';
COMMENT ON COLUMN user_third_party_providers.template_id IS '关联的模板 ID，NULL 表示自定义创建';

CREATE INDEX idx_utpp_user_id ON user_third_party_providers(user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_utpp_user_name ON user_third_party_providers(user_id, name) WHERE deleted_at IS NULL;

-- 第三方模型服务用量记录（与平台用量分开统计）
CREATE TABLE third_party_token_usage (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL,
    provider_id       BIGINT NOT NULL,
    api_key_id        BIGINT NOT NULL,
    model             VARCHAR(100) NOT NULL,
    prompt_tokens     INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens      INT NOT NULL DEFAULT 0,
    request_type      VARCHAR(30) NOT NULL,
    duration_ms       INT,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE third_party_token_usage IS '第三方模型服务用量记录（仅供参考，以服务商为准）';

CREATE INDEX idx_tptu_user_created ON third_party_token_usage(user_id, created_at);
CREATE INDEX idx_tptu_provider ON third_party_token_usage(provider_id);

-- 为训练数据表添加来源标识
ALTER TABLE llm_training_data ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'platform';
ALTER TABLE llm_training_data ADD COLUMN IF NOT EXISTS third_party_provider_id BIGINT;

COMMENT ON COLUMN llm_training_data.source IS '数据来源: platform=平台内置, third_party=第三方服务';
COMMENT ON COLUMN llm_training_data.third_party_provider_id IS '第三方服务 ID，仅 source=third_party 时有值';

CREATE INDEX IF NOT EXISTS idx_training_data_source ON llm_training_data(source);
