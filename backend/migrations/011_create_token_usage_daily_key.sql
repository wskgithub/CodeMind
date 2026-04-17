-- ============================================================
-- Migration 011: 创建 Key 级每日用量汇总表
-- ============================================================
-- 为按 API Key 维度统计提供高性能查询支持
-- ============================================================

CREATE TABLE IF NOT EXISTS token_usage_daily_key (
    id                  BIGSERIAL       PRIMARY KEY,
    api_key_id          BIGINT          NOT NULL REFERENCES api_keys(id),
    user_id             BIGINT          NOT NULL REFERENCES users(id),
    usage_date          DATE            NOT NULL,
    prompt_tokens       BIGINT          NOT NULL DEFAULT 0,
    completion_tokens   BIGINT          NOT NULL DEFAULT 0,
    total_tokens        BIGINT          NOT NULL DEFAULT 0,
    request_count       INTEGER         NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    UNIQUE (api_key_id, usage_date)
);

CREATE INDEX IF NOT EXISTS idx_token_usage_daily_key_user_date ON token_usage_daily_key(user_id, usage_date);

COMMENT ON TABLE token_usage_daily_key IS 'API Key 级每日用量汇总表';

-- 回溯迁移历史数据（幂等，可重复执行）
-- 使用 Asia/Shanghai 时区提取日期，与现有每日汇总逻辑保持一致
INSERT INTO token_usage_daily_key (
    api_key_id, user_id, usage_date,
    prompt_tokens, completion_tokens, total_tokens, request_count,
    created_at, updated_at
)
SELECT
    api_key_id,
    user_id,
    (created_at AT TIME ZONE 'Asia/Shanghai')::date AS usage_date,
    SUM(prompt_tokens),
    SUM(completion_tokens),
    SUM(total_tokens),
    COUNT(*)::int,
    NOW(),
    NOW()
FROM token_usage
GROUP BY api_key_id, user_id, (created_at AT TIME ZONE 'Asia/Shanghai')::date
ON CONFLICT (api_key_id, usage_date) DO UPDATE SET
    user_id             = EXCLUDED.user_id,
    prompt_tokens       = EXCLUDED.prompt_tokens,
    completion_tokens   = EXCLUDED.completion_tokens,
    total_tokens        = EXCLUDED.total_tokens,
    request_count       = EXCLUDED.request_count,
    updated_at          = NOW();
