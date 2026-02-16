-- ============================================================
-- CodeMind Seed Data
-- ============================================================
-- Initial data for first deployment.
-- Creates default admin account and system configurations.
-- ============================================================

-- Default super admin account
-- Username: admin
-- Password: Admin@123456 (bcrypt hash, cost=12)
INSERT INTO users (username, password_hash, display_name, email, role, status)
VALUES (
    'admin',
    '$2a$12$AIv9JVUf369wHF5lIszVROJV/05hM6w4KQyha7fnFVEvB1NVcPb/W',
    'System Administrator',
    'admin@company.com',
    'super_admin',
    1
) ON CONFLICT (username) DO NOTHING;

-- Default system configurations
INSERT INTO system_configs (config_key, config_value, description) VALUES
    ('llm.base_url', '"http://llm-server:8080"', 'LLM service base URL'),
    ('llm.api_key', '""', 'LLM service API key'),
    ('llm.models', '["deepseek-coder-v2"]', 'Available LLM models'),
    ('llm.default_model', '"deepseek-coder-v2"', 'Default LLM model'),
    ('system.max_keys_per_user', '10', 'Maximum API keys per user'),
    ('system.default_concurrency', '5', 'Default max concurrent requests per user'),
    ('system.force_change_password', 'true', 'Force password change on first login')
ON CONFLICT (config_key) DO NOTHING;

-- Default global rate limits
INSERT INTO rate_limits (target_type, target_id, period, max_tokens, max_requests, max_concurrency, alert_threshold)
VALUES
    ('global', 0, 'daily', 1000000, 0, 5, 80),
    ('global', 0, 'monthly', 20000000, 0, 5, 80)
ON CONFLICT (target_type, target_id, period) DO NOTHING;
