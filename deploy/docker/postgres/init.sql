-- ============================================================
-- CodeMind Database Schema Initialization
-- ============================================================
-- This script runs automatically on first PostgreSQL start.
-- It creates all required tables, indexes, and constraints.
-- ============================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ──────────────────────────────────
-- Table: departments
-- ──────────────────────────────────
CREATE TABLE departments (
    id            BIGSERIAL       PRIMARY KEY,
    name          VARCHAR(100)    NOT NULL UNIQUE,
    description   TEXT,
    manager_id    BIGINT,
    parent_id     BIGINT          REFERENCES departments(id) ON DELETE SET NULL,
    status        SMALLINT        NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE departments IS 'Department information';
COMMENT ON COLUMN departments.status IS '1=enabled, 0=disabled';

-- ──────────────────────────────────
-- Table: users
-- ──────────────────────────────────
CREATE TABLE users (
    id              BIGSERIAL       PRIMARY KEY,
    username        VARCHAR(50)     NOT NULL UNIQUE,
    password_hash   VARCHAR(255)    NOT NULL,
    display_name    VARCHAR(100)    NOT NULL,
    email           VARCHAR(255)    UNIQUE,
    phone           VARCHAR(20),
    avatar_url      VARCHAR(500),
    role            VARCHAR(20)     NOT NULL DEFAULT 'user',
    department_id   BIGINT          REFERENCES departments(id) ON DELETE SET NULL,
    status          SMALLINT        NOT NULL DEFAULT 1,
    last_login_at   TIMESTAMPTZ,
    last_login_ip   VARCHAR(45),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_username      ON users(username);
CREATE INDEX idx_users_department_id ON users(department_id);
CREATE INDEX idx_users_role          ON users(role);
CREATE INDEX idx_users_status        ON users(status);
CREATE INDEX idx_users_deleted_at    ON users(deleted_at);

COMMENT ON TABLE users IS 'User accounts';
COMMENT ON COLUMN users.role IS 'super_admin | dept_manager | user';
COMMENT ON COLUMN users.status IS '1=enabled, 0=disabled';

-- Add foreign key for departments.manager_id after users table exists
ALTER TABLE departments
    ADD CONSTRAINT fk_departments_manager
    FOREIGN KEY (manager_id) REFERENCES users(id) ON DELETE SET NULL;

-- ──────────────────────────────────
-- Table: api_keys
-- ──────────────────────────────────
CREATE TABLE api_keys (
    id            BIGSERIAL       PRIMARY KEY,
    user_id       BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          VARCHAR(100)    NOT NULL,
    key_prefix    VARCHAR(20)     NOT NULL,
    key_hash      VARCHAR(255)    NOT NULL UNIQUE,
    status        SMALLINT        NOT NULL DEFAULT 1,
    last_used_at  TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user_id         ON api_keys(user_id);
CREATE INDEX idx_api_keys_key_prefix      ON api_keys(key_prefix);

COMMENT ON TABLE api_keys IS 'User API keys for LLM service access';
COMMENT ON COLUMN api_keys.key_prefix IS 'First 8 chars of the key for display';
COMMENT ON COLUMN api_keys.key_hash IS 'SHA-256 hash of the full key';

-- ──────────────────────────────────
-- Table: token_usage
-- ──────────────────────────────────
CREATE TABLE token_usage (
    id                  BIGSERIAL       PRIMARY KEY,
    user_id             BIGINT          NOT NULL REFERENCES users(id),
    api_key_id          BIGINT          NOT NULL REFERENCES api_keys(id),
    model               VARCHAR(100)    NOT NULL,
    prompt_tokens       INTEGER         NOT NULL DEFAULT 0,
    completion_tokens   INTEGER         NOT NULL DEFAULT 0,
    total_tokens        INTEGER         NOT NULL DEFAULT 0,
    request_type        VARCHAR(30)     NOT NULL,
    duration_ms         INTEGER,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_token_usage_user_created   ON token_usage(user_id, created_at);
CREATE INDEX idx_token_usage_api_key_id     ON token_usage(api_key_id);
CREATE INDEX idx_token_usage_created_at     ON token_usage(created_at);
CREATE INDEX idx_token_usage_model          ON token_usage(model);

COMMENT ON TABLE token_usage IS 'Per-request token usage records';

-- ──────────────────────────────────
-- Table: token_usage_daily
-- ──────────────────────────────────
CREATE TABLE token_usage_daily (
    id                  BIGSERIAL       PRIMARY KEY,
    user_id             BIGINT          NOT NULL REFERENCES users(id),
    usage_date          DATE            NOT NULL,
    prompt_tokens       BIGINT          NOT NULL DEFAULT 0,
    completion_tokens   BIGINT          NOT NULL DEFAULT 0,
    total_tokens        BIGINT          NOT NULL DEFAULT 0,
    request_count       INTEGER         NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_token_usage_daily_user_date ON token_usage_daily(user_id, usage_date);
CREATE INDEX idx_token_usage_daily_date             ON token_usage_daily(usage_date);

COMMENT ON TABLE token_usage_daily IS 'Daily aggregated token usage per user';

-- ──────────────────────────────────
-- Table: rate_limits
-- ──────────────────────────────────
CREATE TABLE rate_limits (
    id              BIGSERIAL       PRIMARY KEY,
    target_type     VARCHAR(20)     NOT NULL,
    target_id       BIGINT          NOT NULL DEFAULT 0,
    period          VARCHAR(20)     NOT NULL,
    max_tokens      BIGINT          NOT NULL,
    max_requests    INTEGER         NOT NULL DEFAULT 0,
    max_concurrency INTEGER         NOT NULL DEFAULT 5,
    alert_threshold SMALLINT        NOT NULL DEFAULT 80,
    status          SMALLINT        NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_rate_limits_target ON rate_limits(target_type, target_id, period);

COMMENT ON TABLE rate_limits IS 'Token usage rate limit configuration';
COMMENT ON COLUMN rate_limits.target_type IS 'global | department | user';
COMMENT ON COLUMN rate_limits.period IS 'daily | weekly | monthly';
COMMENT ON COLUMN rate_limits.alert_threshold IS 'Alert when usage reaches this percentage';

-- ──────────────────────────────────
-- Table: request_logs
-- ──────────────────────────────────
CREATE TABLE request_logs (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL,
    api_key_id      BIGINT          NOT NULL,
    request_type    VARCHAR(30)     NOT NULL,
    model           VARCHAR(100),
    status_code     INTEGER         NOT NULL,
    error_message   TEXT,
    client_ip       VARCHAR(45),
    user_agent      VARCHAR(500),
    duration_ms     INTEGER,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_logs_user_created ON request_logs(user_id, created_at);
CREATE INDEX idx_request_logs_created_at   ON request_logs(created_at);

COMMENT ON TABLE request_logs IS 'LLM request logs';

-- ──────────────────────────────────
-- Table: announcements
-- ──────────────────────────────────
CREATE TABLE announcements (
    id          BIGSERIAL       PRIMARY KEY,
    title       VARCHAR(200)    NOT NULL,
    content     TEXT            NOT NULL,
    author_id   BIGINT          NOT NULL REFERENCES users(id),
    status      SMALLINT        NOT NULL DEFAULT 1,
    pinned      BOOLEAN         NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE announcements IS 'System announcements';
COMMENT ON COLUMN announcements.status IS '1=published, 0=draft';

-- ──────────────────────────────────
-- Table: system_configs
-- ──────────────────────────────────
CREATE TABLE system_configs (
    id            BIGSERIAL       PRIMARY KEY,
    config_key    VARCHAR(100)    NOT NULL UNIQUE,
    config_value  TEXT            NOT NULL,
    description   VARCHAR(500),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE system_configs IS 'System configuration key-value store';

-- ──────────────────────────────────
-- Table: audit_logs
-- ──────────────────────────────────
CREATE TABLE audit_logs (
    id            BIGSERIAL       PRIMARY KEY,
    operator_id   BIGINT          NOT NULL,
    action        VARCHAR(50)     NOT NULL,
    target_type   VARCHAR(50)     NOT NULL,
    target_id     BIGINT,
    detail        JSONB,
    client_ip     VARCHAR(45),
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_operator_id ON audit_logs(operator_id);
CREATE INDEX idx_audit_logs_action      ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at  ON audit_logs(created_at);

COMMENT ON TABLE audit_logs IS 'Operation audit trail';
