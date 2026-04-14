-- ============================================================
-- Migration 017: 修复删除 API Key 时因外键约束导致的数据库错误
-- ============================================================
-- 问题：llm_training_data、token_usage_daily_key 等表的外键约束
--       未设置 ON DELETE 行为，导致删除已使用的 API Key 时失败。
-- 修复：为不同关联表设置合理的级联策略。
-- ============================================================

-- 1. token_usage_daily_key（Key 级每日用量汇总表）
--    汇总数据可从 token_usage 重新生成，允许随 API Key 级联删除
ALTER TABLE token_usage_daily_key
    DROP CONSTRAINT IF EXISTS token_usage_daily_key_api_key_id_fkey;

ALTER TABLE token_usage_daily_key
    ADD CONSTRAINT token_usage_daily_key_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE;

-- 2. llm_training_data（LLM 训练数据表）
--    训练数据具有审计和模型训练价值，删除 API Key 时不应删除数据，
--    而是将 api_key_id 置为 NULL
ALTER TABLE llm_training_data
    ALTER COLUMN api_key_id DROP NOT NULL;

ALTER TABLE llm_training_data
    DROP CONSTRAINT IF EXISTS llm_training_data_api_key_id_fkey;

ALTER TABLE llm_training_data
    ADD CONSTRAINT llm_training_data_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL;

-- 3. token_usage（Token 用量明细表）
--    明细数据具有审计价值，删除 API Key 时保留记录，将 api_key_id 置为 NULL
ALTER TABLE token_usage
    ALTER COLUMN api_key_id DROP NOT NULL;

ALTER TABLE token_usage
    DROP CONSTRAINT IF EXISTS token_usage_api_key_id_fkey;

ALTER TABLE token_usage
    ADD CONSTRAINT token_usage_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL;

-- 4. request_logs（请求日志表）
--    日志具有审计价值，删除 API Key 时保留记录，将 api_key_id 置为 NULL
ALTER TABLE request_logs
    ALTER COLUMN api_key_id DROP NOT NULL;

ALTER TABLE request_logs
    DROP CONSTRAINT IF EXISTS request_logs_api_key_id_fkey;

ALTER TABLE request_logs
    ADD CONSTRAINT request_logs_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL;
