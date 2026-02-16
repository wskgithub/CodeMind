-- 修复 api_keys.key_prefix 字段长度不足问题
-- 从 VARCHAR(10) 扩展到 VARCHAR(20)

ALTER TABLE api_keys ALTER COLUMN key_prefix TYPE VARCHAR(20);

-- 验证修改
COMMENT ON COLUMN api_keys.key_prefix IS 'First 10-12 chars of the key for display (e.g., cm-48cf4808)';
