-- ============================================================
-- Migration 016: 训练数据增强字段
-- ============================================================
-- 添加敏感信息脱敏、会话关联、去重、质量评分功能
-- ============================================================

-- 添加脱敏标记字段
ALTER TABLE llm_training_data
    ADD COLUMN IF NOT EXISTS is_sanitized BOOLEAN NOT NULL DEFAULT FALSE;

-- 添加会话关联字段
ALTER TABLE llm_training_data
    ADD COLUMN IF NOT EXISTS conversation_id VARCHAR(64);

-- 添加内容哈希字段（用于去重）
ALTER TABLE llm_training_data
    ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);

-- 添加质量评分字段
ALTER TABLE llm_training_data
    ADD COLUMN IF NOT EXISTS quality_score SMALLINT;

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_training_data_conversation
    ON llm_training_data(conversation_id) WHERE conversation_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_training_data_content_hash
    ON llm_training_data(content_hash) WHERE content_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_training_data_quality
    ON llm_training_data(quality_score) WHERE quality_score IS NOT NULL;

-- 添加字段注释
COMMENT ON COLUMN llm_training_data.is_sanitized IS '请求/响应是否已脱敏处理';
COMMENT ON COLUMN llm_training_data.conversation_id IS '会话ID，用于关联多轮对话';
COMMENT ON COLUMN llm_training_data.content_hash IS '内容哈希，用于去重（SHA256前16字节）';
COMMENT ON COLUMN llm_training_data.quality_score IS '质量评分 0-100，越高越好';

-- 添加系统配置
INSERT INTO system_configs (config_key, config_value, description) VALUES
    ('training.sanitize_enabled', 'true', '是否启用训练数据敏感信息脱敏'),
    ('training.sanitize_patterns', '["password", "passwd", "pwd", "secret", "api_key", "apikey", "token", "authorization", "credential", "private_key", "access_key", "secret_key"]', '需要脱敏的字段名模式（JSON数组）'),
    ('training.dedup_enabled', 'true', '是否启用训练数据去重'),
    ('training.quality_scoring_enabled', 'true', '是否启用自动质量评分')
ON CONFLICT (config_key) DO NOTHING;
