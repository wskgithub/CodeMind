-- 为 api_keys 表添加 key_encrypted 字段，用于存储加密后的完整 API Key
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_encrypted VARCHAR(255);

-- 添加注释
COMMENT ON COLUMN api_keys.key_encrypted IS 'AES-256-GCM 加密的完整 API Key，用于复制功能';
