-- ============================================================
-- Migration 015: 添加缓存 Token 字段
-- ============================================================
-- 为 LLM 缓存命中统计功能提供数据支持
-- ============================================================

-- 1. 为 token_usage 表添加缓存字段
ALTER TABLE token_usage
    ADD COLUMN IF NOT EXISTS cache_creation_input_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cache_read_input_tokens INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN token_usage.cache_creation_input_tokens IS '缓存创建 Token 数（首次写入缓存）';
COMMENT ON COLUMN token_usage.cache_read_input_tokens IS '缓存命中 Token 数（从缓存读取）';

-- 2. 为 token_usage_daily 表添加缓存字段
ALTER TABLE token_usage_daily
    ADD COLUMN IF NOT EXISTS cache_creation_input_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cache_read_input_tokens BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN token_usage_daily.cache_creation_input_tokens IS '缓存创建 Token 数（首次写入缓存）';
COMMENT ON COLUMN token_usage_daily.cache_read_input_tokens IS '缓存命中 Token 数（从缓存读取）';

-- 3. 为 token_usage_daily_key 表添加缓存字段
ALTER TABLE token_usage_daily_key
    ADD COLUMN IF NOT EXISTS cache_creation_input_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cache_read_input_tokens BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN token_usage_daily_key.cache_creation_input_tokens IS '缓存创建 Token 数（首次写入缓存）';
COMMENT ON COLUMN token_usage_daily_key.cache_read_input_tokens IS '缓存命中 Token 数（从缓存读取）';
