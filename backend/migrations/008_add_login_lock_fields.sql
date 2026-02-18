-- 添加登录锁定相关字段到 users 表
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS login_fail_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS last_login_fail_at TIMESTAMP WITH TIME ZONE;

-- 添加索引优化查询
CREATE INDEX IF NOT EXISTS idx_users_locked_until ON users(locked_until) WHERE locked_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_login_fail_count ON users(login_fail_count) WHERE login_fail_count > 0;
