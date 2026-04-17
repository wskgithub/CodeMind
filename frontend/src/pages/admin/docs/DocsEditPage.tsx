import {
  SaveOutlined,
  ArrowLeftOutlined,
  EyeOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import {
  Form,
  Input,
  Button,
  Card,
  Typography,
  message,
  Space,
  Switch,
  InputNumber,
  Skeleton,
  Alert,
  Tabs,
} from 'antd';
import React, { useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import { useNavigate, useParams } from 'react-router-dom';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';

import { documentService, CreateDocumentRequest, UpdateDocumentRequest } from '@/services/documentService';
import useAppStore from '@/store/appStore';

const { Title } = Typography;
const { TextArea } = Input;

// 自定义代码块渲染 - 代码块始终使用深色主题
const CodeBlock: React.FC<{ language: string; value: string; isDark?: boolean }> = ({ 
  language, 
  value,
  isDark = true,
}) => {
  return (
    <SyntaxHighlighter
      style={vscDarkPlus}
      language={language || 'text'}
      PreTag="div"
      customStyle={{
        margin: '16px 0',
        borderRadius: '8px',
        border: isDark ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(0,0,0,0.1)',
        fontSize: '14px',
      }}
    >
      {value}
    </SyntaxHighlighter>
  );
};

const DocsEditPage: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = !!id;
  const [form] = Form.useForm();
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
  
  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [content, setContent] = useState('');
  const [activeTab, setActiveTab] = useState('edit');

  useEffect(() => {
    if (isEdit) {
      loadDocument();
    }
  }, [id]);

  const loadDocument = async () => {
    try {
      setLoading(true);
      const doc = await documentService.getById(Number(id));
      if (doc) {
        form.setFieldsValue({
          slug: doc.slug,
          title: doc.title,
          subtitle: doc.subtitle,
          icon: doc.icon,
          content: doc.content,
          sort_order: doc.sort_order,
          is_published: doc.is_published,
        });
        setContent(doc.content);
      } else {
        message.error('文档不存在');
        navigate('/admin/docs');
      }
    } catch (error) {
      message.error('加载文档失败');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      setSaving(true);
      
      if (isEdit) {
        const data: UpdateDocumentRequest = {
          title: values.title,
          subtitle: values.subtitle || '',
          icon: values.icon || '',
          content: values.content,
          sort_order: values.sort_order || 0,
          is_published: values.is_published,
        };
        await documentService.update(Number(id), data);
        message.success('文档更新成功');
      } else {
        const data: CreateDocumentRequest = {
          slug: values.slug,
          title: values.title,
          subtitle: values.subtitle || '',
          icon: values.icon || '',
          content: values.content,
          sort_order: values.sort_order || 0,
          is_published: values.is_published ?? true,
        };
        await documentService.create(data);
        message.success('文档创建成功');
      }
      
      navigate('/admin/docs');
    } catch (error: any) {
      if (error.response?.data?.error) {
        message.error(error.response.data.error);
      } else {
        message.error(isEdit ? '更新失败' : '创建失败');
      }
    } finally {
      setSaving(false);
    }
  };

  const handleContentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setContent(value);
    // 同步更新 Form 字段，确保提交时能获取最新值
    form.setFieldValue('content', value);
  };

  const handleBack = () => {
    navigate('/admin/docs');
  };

  if (loading) {
    return (
      <div style={{ padding: 24 }}>
        <Skeleton active paragraph={{ rows: 20 }} />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
          <Title level={4} style={{ margin: 0 }}>
            <FileTextOutlined style={{ marginRight: 8 }} />
            {isEdit ? '编辑文档' : '新建文档'}
          </Title>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={handleBack}>
              返回
            </Button>
            <Button
              type="primary"
              icon={<SaveOutlined />}
              onClick={() => form.submit()}
              loading={saving}
            >
              保存
            </Button>
          </Space>
        </div>

        {!isEdit && (
          <Alert
            message="创建新文档"
            description="Slug 是文档的唯一标识，创建后无法修改。建议使用英文小写字母和连字符，如：my-tool。"
            type="info"
            showIcon
            style={{ marginBottom: 24 }}
          />
        )}

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            is_published: true,
            sort_order: 0,
          }}
        >
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 24 }}>
            {!isEdit && (
              <Form.Item
                name="slug"
                label="文档标识 (Slug)"
                rules={[
                  { required: true, message: '请输入文档标识' },
                  { pattern: /^[a-z0-9-]+$/, message: '只能使用小写字母、数字和连字符' },
                ]}
              >
                <Input placeholder="例如：claude-code" />
              </Form.Item>
            )}

            <Form.Item
              name="title"
              label="文档标题"
              rules={[{ required: true, message: '请输入文档标题' }]}
            >
              <Input placeholder="例如：Claude Code" />
            </Form.Item>

            <Form.Item
              name="icon"
              label="图标"
            >
              <Input placeholder="例如：🤖 或图标类名" />
            </Form.Item>

            <Form.Item
              name="sort_order"
              label="排序顺序"
            >
              <InputNumber min={0} style={{ width: '100%' }} placeholder="数字越小排序越靠前" />
            </Form.Item>

            <Form.Item
              name="is_published"
              label="发布状态"
              valuePropName="checked"
            >
              <Switch checkedChildren="已发布" unCheckedChildren="未发布" />
            </Form.Item>
          </div>

          <Form.Item
            name="subtitle"
            label="副标题/简介"
          >
            <Input.TextArea rows={2} placeholder="简短描述该工具的功能和特点" />
          </Form.Item>

          <Form.Item
            name="content"
            label="文档内容 (Markdown)"
            rules={[{ required: true, message: '请输入文档内容' }]}
          >
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={[
                {
                  key: 'edit',
                  label: '编辑',
                  children: (
                    <TextArea
                      rows={25}
                      placeholder="使用 Markdown 格式编写文档内容..."
                      value={content}
                      onChange={handleContentChange}
                      style={{ fontFamily: 'monospace' }}
                    />
                  ),
                },
                {
                  key: 'preview',
                  label: (
                    <span>
                      <EyeOutlined style={{ marginRight: 4 }} />
                      预览
                    </span>
                  ),
                  children: (
                    <div
                      style={{
                        border: isDark ? '1px solid rgba(255,255,255,0.1)' : '1px solid rgba(0,0,0,0.1)',
                        borderRadius: 6,
                        padding: 16,
                        minHeight: 550,
                        maxHeight: 550,
                        overflow: 'auto',
                        background: isDark ? 'rgba(20, 20, 26, 0.5)' : 'rgba(240, 240, 245, 0.5)',
                      }}
                      className="docs-markdown-preview"
                    >
                      {content ? (
                        <ReactMarkdown
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
                          }}
                        >
                          {content}
                        </ReactMarkdown>
                      ) : (
                        <div style={{ color: isDark ? 'rgba(255,255,255,0.3)' : 'rgba(0,0,0,0.3)', textAlign: 'center', paddingTop: 100 }}>
                          暂无内容，请在编辑标签页输入 Markdown 内容
                        </div>
                      )}
                    </div>
                  ),
                },
              ]}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={handleBack}>取消</Button>
              <Button type="primary" htmlType="submit" loading={saving} icon={<SaveOutlined />}>
                保存
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default DocsEditPage;
