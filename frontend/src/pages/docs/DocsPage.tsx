import { BookOutlined, EditOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import { Layout, Menu, Card, Typography, Skeleton, Empty, Alert, Button, Tooltip } from 'antd';
import React, { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { useNavigate, useParams } from 'react-router-dom';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import remarkGfm from 'remark-gfm';

import { documentService, Document, DocumentListItem } from '@/services/documentService';
import useAppStore from '@/store/appStore';
import useAuthStore from '@/store/authStore';
import '@/assets/styles/docs.css';

const { Sider, Content } = Layout;
const { Title } = Typography;

const CodeBlock = React.memo<{ language: string; value: string; isDark: boolean }>(({ 
  language, 
  value,
  isDark,
}) => {
  return (
    <SyntaxHighlighter
      style={isDark ? vscDarkPlus : oneLight}
      language={language || 'text'}
      PreTag="div"
      customStyle={{
        margin: '16px 0',
        borderRadius: '8px',
        fontSize: '14px',
        border: isDark ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(0,0,0,0.1)',
      }}
    >
      {value}
    </SyntaxHighlighter>
  );
});
CodeBlock.displayName = 'CodeBlock';

const DocsPage: React.FC = () => {
  const navigate = useNavigate();
  const { slug } = useParams<{ slug: string }>();
  const user = useAuthStore((s) => s.user);
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
  
  const [documents, setDocuments] = useState<DocumentListItem[]>([]);
  const [currentDoc, setCurrentDoc] = useState<Document | null>(null);
  const [loading, setLoading] = useState(true);
  const [contentLoading, setContentLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [collapsed, setCollapsed] = useState(false);
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);

  const isAdmin = user?.role === 'super_admin' || false;

  // 加载文档列表
  useEffect(() => {
    loadDocuments();
  }, []);

  // 当 slug 变化时加载对应文档
  useEffect(() => {
    if (slug) {
      loadDocument(slug);
      setSelectedKeys([slug]);
    } else if (documents.length > 0 && !currentDoc) {
      // 如果没有指定 slug，加载第一篇文档
      const firstDoc = documents[0];
      if (firstDoc) {
        loadDocument(firstDoc.slug);
        setSelectedKeys([firstDoc.slug]);
      }
    }
  }, [slug, documents]);

  const loadDocuments = async () => {
    try {
      setLoading(true);
      setError(null);
      const list = await documentService.list();
      setDocuments(list);
      
      // 如果没有指定 slug，默认选中第一篇
      const firstDoc = list[0];
      if (!slug && list.length > 0 && firstDoc) {
        setSelectedKeys([firstDoc.slug]);
      }
    } catch (err: any) {
      console.error('Failed to load documents:', err);
      setError('加载文档列表失败: ' + (err.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const loadDocument = async (docSlug: string) => {
    try {
      setContentLoading(true);
      setError(null);
      const doc = await documentService.getBySlug(docSlug);
      if (doc) {
        setCurrentDoc(doc);
        // 更新 URL 但不触发重新加载
        if (slug !== docSlug) {
          navigate(`/docs/${docSlug}`, { replace: true });
        }
      } else {
        setError('文档不存在或未发布');
      }
    } catch (err: any) {
      console.error('Failed to load document:', err);
      setError('加载文档内容失败: ' + (err.message || '未知错误'));
    } finally {
      setContentLoading(false);
    }
  };

  const handleMenuClick = ({ key }: { key: string }) => {
    loadDocument(key);
    setSelectedKeys([key]);
    // 在移动端自动收起侧边栏
    if (window.innerWidth < 768) {
      setCollapsed(true);
    }
  };

  const handleEditClick = () => {
    if (currentDoc) {
      navigate(`/admin/docs/edit/${currentDoc.id}`);
    }
  };

  const handleAdminClick = () => {
    navigate('/admin/docs');
  };

  if (loading) {
    return (
      <div style={{ padding: 48 }}>
        <Skeleton active paragraph={{ rows: 10 }} />
      </div>
    );
  }

  if (error && documents.length === 0) {
    return (
      <div style={{ padding: 48 }}>
        <Alert
          message="加载失败"
          description={error}
          type="error"
          showIcon
        />
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Empty
          description="暂无文档"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
        {isAdmin && (
          <Button type="primary" onClick={handleAdminClick} style={{ marginTop: 16 }}>
            管理文档
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className="docs-wrapper">
      <Layout className="docs-page" hasSider>
        <Sider
          trigger={null}
          collapsible
          collapsed={collapsed}
          collapsedWidth={0}
          width={280}
          className="docs-sider"
          theme="light"
        >
          <div className="docs-sider-header">
            <BookOutlined className="docs-sider-icon" />
            {!collapsed && <span className="docs-sider-title">接入指南</span>}
          </div>
          <Menu
            mode="inline"
            selectedKeys={selectedKeys}
            onClick={handleMenuClick}
            className="docs-menu"
            items={documents.map((doc) => ({
              key: doc.slug,
              icon: <span className="docs-menu-icon">{doc.icon || '📄'}</span>,
              label: (
                <div className="docs-menu-item">
                  <div className="docs-menu-title">{doc.title}</div>
                  {!collapsed && doc.subtitle && (
                    <div className="docs-menu-subtitle">{doc.subtitle}</div>
                  )}
                </div>
              ),
            }))}
          />
        </Sider>
        
        <Layout className="docs-content-layout">
          <Content className="docs-content">
            <div className="docs-content-header">
              <div className="docs-content-header-left">
                <Button
                  type="text"
                  icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                  onClick={() => setCollapsed(!collapsed)}
                  className="docs-collapse-btn"
                />
                {currentDoc && (
                  <div className="docs-title-wrapper">
                    <Title level={2} className="docs-title" style={{ margin: 0 }}>
                      <span className="docs-title-icon">{currentDoc.icon || '📄'}</span>
                      {currentDoc.title}
                    </Title>
                    {currentDoc.subtitle && (
                      <div className="docs-subtitle">{currentDoc.subtitle}</div>
                    )}
                  </div>
                )}
              </div>
              {isAdmin && currentDoc && (
                <div className="docs-content-actions">
                  <Tooltip title="编辑当前文档">
                    <Button
                      type="primary"
                      icon={<EditOutlined />}
                      onClick={handleEditClick}
                    >
                      编辑
                    </Button>
                  </Tooltip>
                </div>
              )}
            </div>

            <Alert
              message="当前技术预览阶段，经测试 Roo Code 具有最佳兼容性，请优先使用 Roo Code 接入平台"
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />

            {error && (
              <Alert
                message="错误"
                description={error}
                type="error"
                showIcon
                style={{ marginBottom: 16 }}
              />
            )}

            {contentLoading ? (
              <Skeleton active paragraph={{ rows: 20 }} />
            ) : currentDoc ? (
              <Card className="docs-content-card" bordered={false}>
                <div className="docs-markdown">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    components={{
                      code({ node: _node, inline, className, children, ...props }: any) {
                        const match = /language-(\w+)/.exec(className || '');
                        return !inline && match ? (
                          <CodeBlock
                            language={match[1] || 'text'}
                            value={String(children).replace(/\n$/, '')}
                            isDark={isDark}
                          />
                        ) : (
                          <code className={className} {...props}>
                            {children}
                          </code>
                        );
                      },
                      h2: ({ children }: any) => <h2 id={String(children).toLowerCase().replace(/\s+/g, '-')} className="docs-heading">{children}</h2>,
                      h3: ({ children }: any) => <h3 id={String(children).toLowerCase().replace(/\s+/g, '-')} className="docs-heading">{children}</h3>,
                    }}
                  >
                    {currentDoc.content}
                  </ReactMarkdown>
                </div>
                {currentDoc.updated_at && (
                  <div className="docs-update-time">
                    最后更新：{new Date(currentDoc.updated_at).toLocaleString('zh-CN')}
                  </div>
                )}
              </Card>
            ) : (
              <Empty description="请选择文档" />
            )}
          </Content>
        </Layout>
      </Layout>
    </div>
  );
};

export default DocsPage;
