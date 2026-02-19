-- 文档表
-- 存储开发工具接入文档内容

CREATE TABLE IF NOT EXISTS documents (
    id            BIGSERIAL PRIMARY KEY,
    slug          VARCHAR(50)   NOT NULL UNIQUE,
    title         VARCHAR(200)  NOT NULL,
    subtitle      VARCHAR(500)  DEFAULT '',
    icon          VARCHAR(100)  DEFAULT '',
    content       TEXT          NOT NULL DEFAULT '',
    sort_order    INT           NOT NULL DEFAULT 0,
    is_published  BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMP     DEFAULT NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_documents_slug ON documents(slug);
CREATE INDEX IF NOT EXISTS idx_documents_sort_order ON documents(sort_order);
CREATE INDEX IF NOT EXISTS idx_documents_is_published ON documents(is_published);
CREATE INDEX IF NOT EXISTS idx_documents_deleted_at ON documents(deleted_at);

-- 注释
COMMENT ON TABLE documents IS '开发工具接入文档表';
COMMENT ON COLUMN documents.slug IS '文档唯一标识，如 claude, cursor';
COMMENT ON COLUMN documents.content IS 'Markdown 格式文档内容';
COMMENT ON COLUMN documents.sort_order IS '显示排序顺序';
COMMENT ON COLUMN documents.is_published IS '是否已发布';
