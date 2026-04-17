-- ============================================================
-- Migration 010: 创建 LLM 训练数据表
-- ============================================================
-- 记录 LLM 代理请求的完整输入/输出，用于后续模型训练
-- ============================================================

CREATE TABLE IF NOT EXISTS llm_training_data (
    id                  BIGSERIAL       PRIMARY KEY,
    user_id             BIGINT          NOT NULL REFERENCES users(id),
    api_key_id          BIGINT          NOT NULL REFERENCES api_keys(id),
    request_type        VARCHAR(30)     NOT NULL,
    model               VARCHAR(100)    NOT NULL,
    is_stream           BOOLEAN         NOT NULL DEFAULT FALSE,

    request_body        JSONB           NOT NULL,
    response_body       JSONB,

    prompt_tokens       INTEGER         NOT NULL DEFAULT 0,
    completion_tokens   INTEGER         NOT NULL DEFAULT 0,
    total_tokens        INTEGER         NOT NULL DEFAULT 0,

    duration_ms         INTEGER,
    status_code         INTEGER         NOT NULL DEFAULT 200,
    client_ip           VARCHAR(45),

    is_excluded         BOOLEAN         NOT NULL DEFAULT FALSE,

    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_training_data_user_created ON llm_training_data(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_training_data_model        ON llm_training_data(model);
CREATE INDEX IF NOT EXISTS idx_training_data_type         ON llm_training_data(request_type);
CREATE INDEX IF NOT EXISTS idx_training_data_created      ON llm_training_data(created_at);
CREATE INDEX IF NOT EXISTS idx_training_data_excluded     ON llm_training_data(is_excluded) WHERE is_excluded = FALSE;

COMMENT ON TABLE  llm_training_data             IS 'LLM 请求/响应记录，用于模型训练数据采集';
COMMENT ON COLUMN llm_training_data.request_body  IS '完整请求体 JSON';
COMMENT ON COLUMN llm_training_data.response_body IS '完整响应体 JSON（流式请求为组装后的等效非流式格式；Embedding 请求为 NULL）';
COMMENT ON COLUMN llm_training_data.is_excluded   IS '是否已从训练集中排除';

-- 在 system_configs 中插入训练数据采集开关（默认开启）
INSERT INTO system_configs (config_key, config_value, description)
VALUES ('system.training_data_collection', 'true', '是否开启 LLM 请求/响应数据采集（用于模型训练）')
ON CONFLICT (config_key) DO NOTHING;
