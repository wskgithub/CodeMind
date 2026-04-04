-- 修复 documents 表的唯一索引，支持软删除
-- 原全局 UNIQUE 约束会在软删除后仍阻止相同 slug 的插入
-- 改为部分唯一索引，只针对未删除的记录

-- 1. 删除原全局唯一约束
ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_slug_key;

-- 2. 删除可能存在的普通索引（如果存在）
DROP INDEX IF EXISTS idx_documents_slug;

-- 3. 创建部分唯一索引（只针对未删除的记录）
CREATE UNIQUE INDEX idx_documents_slug ON documents(slug) WHERE deleted_at IS NULL;

-- 注释
COMMENT ON INDEX idx_documents_slug IS '文档 slug 唯一索引（仅未删除记录）';
