-- MCP 服务管理表
-- 存储注册到 CodeMind 网关的 MCP 服务信息

CREATE TABLE IF NOT EXISTS mcp_services (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(100)  NOT NULL UNIQUE,
    display_name  VARCHAR(200)  NOT NULL,
    description   TEXT          DEFAULT '',
    endpoint_url  VARCHAR(500)  NOT NULL,
    transport_type VARCHAR(20)  NOT NULL DEFAULT 'sse',
    status        VARCHAR(20)   NOT NULL DEFAULT 'enabled',
    auth_type     VARCHAR(20)   NOT NULL DEFAULT 'none',
    auth_config   JSONB         DEFAULT '{}',
    tools_schema  JSONB         DEFAULT NULL,
    created_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP     DEFAULT NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_mcp_services_status ON mcp_services(status);
CREATE INDEX IF NOT EXISTS idx_mcp_services_deleted_at ON mcp_services(deleted_at);

-- MCP 访问控制规则表
CREATE TABLE IF NOT EXISTS mcp_access_rules (
    id            BIGSERIAL PRIMARY KEY,
    service_id    BIGINT        NOT NULL REFERENCES mcp_services(id) ON DELETE CASCADE,
    target_type   VARCHAR(20)   NOT NULL,  -- user | department | role
    target_id     BIGINT        NOT NULL DEFAULT 0,
    allowed       BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    UNIQUE(service_id, target_type, target_id)
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_mcp_access_rules_service ON mcp_access_rules(service_id);
CREATE INDEX IF NOT EXISTS idx_mcp_access_rules_target ON mcp_access_rules(target_type, target_id);

-- 注释
COMMENT ON TABLE mcp_services IS 'MCP 服务注册表';
COMMENT ON TABLE mcp_access_rules IS 'MCP 服务访问控制规则';
COMMENT ON COLUMN mcp_services.transport_type IS '传输类型: sse | streamable-http';
COMMENT ON COLUMN mcp_services.auth_type IS '认证类型: none | bearer | header';
COMMENT ON COLUMN mcp_services.tools_schema IS '工具列表缓存（定期从上游同步）';
COMMENT ON COLUMN mcp_access_rules.target_type IS '目标类型: user | department | role';
COMMENT ON COLUMN mcp_access_rules.target_id IS '目标 ID（role 类型时为 0）';
