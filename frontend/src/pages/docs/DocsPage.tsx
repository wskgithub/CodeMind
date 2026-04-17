import React, { useEffect, useState, useMemo, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Layout, Menu, Skeleton, Empty, Button, Input, Typography } from 'antd';
import {
  BookOutlined,
  EditOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SearchOutlined,
  SettingOutlined,
} from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus, oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { useNavigate, useParams } from 'react-router-dom';
import useAuthStore from '@/store/authStore';
import useAppStore from '@/store/appStore';
import { documentService, Document, DocumentListItem } from '@/services/documentService';
import '@/assets/styles/docs.css';

const { Sider, Content } = Layout;

// 代码块渲染组件
const CodeBlock = React.memo<{ language: string; value: string; isDark: boolean }>(({
  language,
  value,
  isDark,
}) => (
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
));

// 图片渲染组件：默认居中，支持尺寸调整
const MarkdownImage: React.FC<React.ImgHTMLAttributes<HTMLImageElement>> = ({ src, alt, width, height, style, ...rest }) => (
  <span style={{ display: 'block', textAlign: 'center', margin: '16px 0' }}>
    <img
      src={src}
      alt={alt || ''}
      width={width}
      height={height}
      style={{
        maxWidth: '100%',
        height: 'auto',
        borderRadius: '8px',
        ...style,
      }}
      {...rest}
    />
  </span>
);

const DocsPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { slug } = useParams<{ slug: string }>();
  const user = useAuthStore((s) => s.user);
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  const [documents, setDocuments] = useState<DocumentListItem[]>([]);
  const [currentDoc, setCurrentDoc] = useState<Document | null>(null);
  const [loading, setLoading] = useState(true);
  const [contentLoading, setContentLoading] = useState(false);
  const [collapsed, setCollapsed] = useState(false);
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [searchText, setSearchText] = useState('');

  const isAdmin = user?.role === 'super_admin' || false;

  useEffect(() => {
    loadDocuments();
  }, []);

  useEffect(() => {
    if (slug) {
      loadDocument(slug);
      setSelectedKeys([slug]);
    } else if (documents.length > 0 && !currentDoc) {
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
      const list = await documentService.list();
      setDocuments(list);
    } catch (err) {
      console.error('加载文档列表失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadDocument = async (docSlug: string) => {
    try {
      setContentLoading(true);
      const doc = await documentService.getBySlug(docSlug);
      if (doc) {
        setCurrentDoc(doc);
        if (slug !== docSlug) {
          navigate(`/docs/${docSlug}`, { replace: true });
        }
      }
    } catch (err) {
      console.error('加载文档失败:', err);
    } finally {
      setContentLoading(false);
    }
  };

  const handleMenuClick = useCallback(({ key }: { key: string }) => {
    loadDocument(key);
    setSelectedKeys([key]);
    if (window.innerWidth < 768) {
      setCollapsed(true);
    }
  }, []);

  // 搜索过滤
  const filteredDocs = useMemo(() => {
    if (!searchText.trim()) return documents;
    const keyword = searchText.toLowerCase();
    return documents.filter(
      (doc) =>
        doc.title.toLowerCase().includes(keyword) ||
        (doc.subtitle && doc.subtitle.toLowerCase().includes(keyword))
    );
  }, [documents, searchText]);

  // 当前文档在列表中的位置，用于上/下篇导航
  const currentIndex = useMemo(() => {
    if (!currentDoc) return -1;
    return documents.findIndex((d) => d.slug === currentDoc.slug);
  }, [currentDoc, documents]);

  const prevDoc = currentIndex > 0 ? documents[currentIndex - 1] : null;
  const nextDoc = currentIndex < documents.length - 1 ? documents[currentIndex + 1] : null;

  // Markdown 组件配置
  const markdownComponents = useMemo(() => ({
    code({ inline, className, children, ...props }: any) {
      const match = /language-(\w+)/.exec(className || '');
      return !inline && match ? (
        <CodeBlock
          language={match[1] || 'text'}
          value={String(children).replace(/\n$/, '')}
          isDark={isDark}
        />
      ) : (
        <code className={className} {...props}>{children}</code>
      );
    },
    img: MarkdownImage,
    h2: ({ children }: any) => (
      <h2 id={String(children).toLowerCase().replace(/\s+/g, '-')} className="docs-heading">{children}</h2>
    ),
    h3: ({ children }: any) => (
      <h3 id={String(children).toLowerCase().replace(/\s+/g, '-')} className="docs-heading">{children}</h3>
    ),
  }), [isDark]);

  if (loading) {
    return (
      <div style={{ padding: 48 }}>
        <Skeleton active paragraph={{ rows: 10 }} />
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Empty
          description={t('docs.noDocuments')}
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
        {isAdmin && (
          <Button
            type="primary"
            onClick={() => navigate('/admin/docs')}
            style={{ marginTop: 16 }}
            icon={<SettingOutlined />}
          >
            {t('docs.manageDocuments')}
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
            {!collapsed && <span className="docs-sider-title">{t('docs.sidebarTitle')}</span>}
          </div>

          {/* 搜索框 */}
          {!collapsed && (
            <div style={{ padding: '0 12px 8px' }}>
              <Input
                prefix={<SearchOutlined style={{ color: 'var(--ant-color-text-quaternary)' }} />}
                placeholder={t('docs.sidebar.search')}
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                allowClear
                size="small"
              />
            </div>
          )}

          <Menu
            mode="inline"
            selectedKeys={selectedKeys}
            onClick={handleMenuClick}
            className="docs-menu"
            items={filteredDocs.map((doc) => ({
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

          {/* 管理员入口 */}
          {isAdmin && !collapsed && (
            <div style={{ padding: '12px 16px', borderTop: '1px solid var(--ant-color-border)' }}>
              <Button
                type="text"
                icon={<SettingOutlined />}
                onClick={() => navigate('/admin/docs')}
                block
                size="small"
                style={{ textAlign: 'left' }}
              >
                {t('docs.manageDocuments')}
              </Button>
            </div>
          )}
        </Sider>

        <Layout className="docs-content-layout">
          <Content className="docs-content">
            {/* 顶部栏 */}
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
                    <Typography.Title level={2} className="docs-title" style={{ margin: 0 }}>
                      <span className="docs-title-icon">{currentDoc.icon || '📄'}</span>
                      {currentDoc.title}
                    </Typography.Title>
                    {currentDoc.subtitle && (
                      <div className="docs-subtitle">{currentDoc.subtitle}</div>
                    )}
                  </div>
                )}
              </div>
              {isAdmin && currentDoc && (
                <Button
                  type="primary"
                  icon={<EditOutlined />}
                  onClick={() => navigate(`/admin/docs/edit/${currentDoc.id}`)}
                >
                  {t('common.edit')}
                </Button>
              )}
            </div>

            {/* 正文内容 */}
            {contentLoading ? (
              <Skeleton active paragraph={{ rows: 20 }} />
            ) : currentDoc ? (
              <div className="docs-content-card">
                <article className="docs-markdown">
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    components={markdownComponents}
                  >
                    {currentDoc.content}
                  </ReactMarkdown>
                </article>

                {/* 底部信息 */}
                <div className="docs-footer">
                  {currentDoc.updated_at && (
                    <span className="docs-update-time">
                      {t('docs.lastUpdated', { time: new Date(currentDoc.updated_at).toLocaleString('zh-CN') })}
                    </span>
                  )}
                </div>

                {/* 上/下篇导航 */}
                {(prevDoc || nextDoc) && (
                  <div className="docs-nav">
                    {prevDoc ? (
                      <Button
                        type="text"
                        onClick={() => {
                          loadDocument(prevDoc.slug);
                          setSelectedKeys([prevDoc.slug]);
                        }}
                        className="docs-nav-btn docs-nav-prev"
                      >
                        <div className="docs-nav-label">{t('docs.navigation.previous')}</div>
                        <div className="docs-nav-title">← {prevDoc.icon} {prevDoc.title}</div>
                      </Button>
                    ) : <div />}
                    {nextDoc ? (
                      <Button
                        type="text"
                        onClick={() => {
                          loadDocument(nextDoc.slug);
                          setSelectedKeys([nextDoc.slug]);
                        }}
                        className="docs-nav-btn docs-nav-next"
                      >
                        <div className="docs-nav-label">{t('docs.navigation.next')}</div>
                        <div className="docs-nav-title">{nextDoc.icon} {nextDoc.title} →</div>
                      </Button>
                    ) : <div />}
                  </div>
                )}
              </div>
            ) : (
              <Empty description={t('docs.selectDocument')} />
            )}
          </Content>
        </Layout>
      </Layout>
    </div>
  );
};

export default DocsPage;
