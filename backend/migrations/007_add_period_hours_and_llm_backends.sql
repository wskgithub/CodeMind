-- 007: 添加限额周期小时数字段 + 创建 LLM 后端服务节点表
-- 支持基于小时的灵活限额周期和多后端负载均衡

BEGIN;

-- ──────────────────────────────────
-- 1. 限额表：添加 period_hours 字段
-- ──────────────────────────────────
ALTER TABLE rate_limits
    ADD COLUMN IF NOT EXISTS period_hours INTEGER NOT NULL DEFAULT 24;

-- 为现有记录设置对应的小时数
UPDATE rate_limits SET period_hours = 24  WHERE period = 'daily'   AND period_hours = 24;
UPDATE rate_limits SET period_hours = 168 WHERE period = 'weekly';
UPDATE rate_limits SET period_hours = 720 WHERE period = 'monthly';

-- 重建唯一索引：(target_type, target_id, period_hours) 替代旧的 (target_type, target_id, period)
DROP INDEX IF EXISTS idx_rate_limits_target;
CREATE UNIQUE INDEX idx_rate_limits_target ON rate_limits(target_type, target_id, period_hours);

-- ──────────────────────────────────
-- 2. 创建 LLM 后端服务节点表
-- ──────────────────────────────────
CREATE TABLE IF NOT EXISTS llm_backends (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100)  NOT NULL UNIQUE,
    display_name    VARCHAR(200)  NOT NULL DEFAULT '',
    base_url        VARCHAR(500)  NOT NULL,
    api_key         VARCHAR(500)  NOT NULL DEFAULT '',
    format          VARCHAR(20)   NOT NULL DEFAULT 'openai',   -- openai | anthropic
    weight          INTEGER       NOT NULL DEFAULT 100,        -- 负载均衡权重
    max_concurrency INTEGER       NOT NULL DEFAULT 100,        -- 最大并发连接数
    status          SMALLINT      NOT NULL DEFAULT 1,          -- 0=禁用, 1=启用, 2=排空
    health_check_url VARCHAR(500) NOT NULL DEFAULT '',
    timeout_seconds       INTEGER NOT NULL DEFAULT 300,
    stream_timeout_seconds INTEGER NOT NULL DEFAULT 600,
    model_patterns  TEXT          NOT NULL DEFAULT '*',         -- 支持的模型模式，逗号分隔，如 'gpt-*,claude-*'
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE llm_backends IS 'LLM 后端服务节点，用于负载均衡';
COMMENT ON COLUMN llm_backends.weight IS '负载均衡权重，数值越大分配的请求越多';
COMMENT ON COLUMN llm_backends.status IS '0=禁用, 1=启用, 2=排空(不接受新请求，等待已有请求完成)';
COMMENT ON COLUMN llm_backends.model_patterns IS '该后端支持的模型匹配模式，逗号分隔，支持通配符 *';

COMMIT;
