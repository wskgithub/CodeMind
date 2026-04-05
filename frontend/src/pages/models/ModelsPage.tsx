import { useState, useEffect, useCallback, useMemo } from 'react';
import { Table, Button, Modal, Form, Input, Space, Tag, Tabs, message, Select, Empty, Tooltip } from 'antd';
import {
  PlusOutlined, DeleteOutlined, StopOutlined, CheckCircleOutlined,
  AppstoreOutlined, EditOutlined, CloudServerOutlined, ThunderboltOutlined,
  CopyOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { PlatformModelInfo, UserThirdPartyProvider, ProviderTemplate } from '@/types';
import modelService from '@/services/modelService';
import { getPlatformSettings } from '@/services/systemService';
import useAppStore from '@/store/appStore';

/** 渲染协议格式标签 */
const FormatTags = ({ format }: { format: string }) => {
  const tags: React.ReactNode[] = [];
  if (format === 'openai' || format === 'all') {
    tags.push(
      <Tag key="openai" style={{
        color: '#00F5D4', background: 'rgba(0, 245, 212, 0.12)',
        border: '1px solid rgba(0, 245, 212, 0.3)', borderRadius: 6, fontSize: 12,
      }}>OpenAI</Tag>
    );
  }
  if (format === 'anthropic' || format === 'all') {
    tags.push(
      <Tag key="anthropic" style={{
        color: '#9D4EDD', background: 'rgba(157, 78, 221, 0.12)',
        border: '1px solid rgba(157, 78, 221, 0.3)', borderRadius: 6, fontSize: 12,
      }}>Anthropic</Tag>
    );
  }
  return <Space size={4}>{tags}</Space>;
};

const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
    }}
  >
    {icon}
  </span>
);

const ModelsPage: React.FC = () => {
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  // ── 平台服务 URL ──
  const [openaiBaseURL, setOpenaiBaseURL] = useState('/api/openai/v1');
  const [anthropicBaseURL, setAnthropicBaseURL] = useState('/api/anthropic');

  // ── CodeMind 平台模型 ──
  const [platformModels, setPlatformModels] = useState<PlatformModelInfo[]>([]);
  const [platformLoading, setPlatformLoading] = useState(false);

  // ── 第三方服务 ──
  const [providers, setProviders] = useState<UserThirdPartyProvider[]>([]);
  const [providersLoading, setProvidersLoading] = useState(false);
  const [templates, setTemplates] = useState<ProviderTemplate[]>([]);

  // ── 弹窗 ──
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<UserThirdPartyProvider | null>(null);
  const [selectedTemplateId, setSelectedTemplateId] = useState<number | undefined>();
  const [form] = Form.useForm();

  const loadPlatformModels = useCallback(async () => {
    setPlatformLoading(true);
    try {
      const resp = await modelService.listPlatformModels();
      setPlatformModels(resp.data.data || []);
    } catch { /* 拦截器处理 */ }
    finally { setPlatformLoading(false); }
  }, []);

  const loadProviders = useCallback(async () => {
    setProvidersLoading(true);
    try {
      const resp = await modelService.listProviders();
      setProviders(resp.data.data || []);
    } catch { /* 拦截器处理 */ }
    finally { setProvidersLoading(false); }
  }, []);

  const loadTemplates = useCallback(async () => {
    try {
      const resp = await modelService.listTemplates();
      setTemplates(resp.data.data || []);
    } catch { /* 拦截器处理 */ }
  }, []);

  const loadPlatformURL = useCallback(async () => {
    try {
      const resp = await getPlatformSettings();
      const data = resp.data.data;
      if (data?.openai_base_url) setOpenaiBaseURL(data.openai_base_url);
      if (data?.anthropic_base_url) setAnthropicBaseURL(data.anthropic_base_url);
    } catch { /* 拦截器处理 */ }
  }, []);

  useEffect(() => {
    loadPlatformURL();
    loadPlatformModels();
    loadProviders();
    loadTemplates();
  }, [loadPlatformURL, loadPlatformModels, loadProviders, loadTemplates]);

  // 选择模板时自动填充
  const handleTemplateSelect = (templateId: number | undefined) => {
    setSelectedTemplateId(templateId);
    if (!templateId) return;
    const tpl = templates.find(t => t.id === templateId);
    if (tpl) {
      form.setFieldsValue({
        name: tpl.name,
        openai_base_url: tpl.openai_base_url || '',
        anthropic_base_url: tpl.anthropic_base_url || '',
        models: tpl.models,
        format: tpl.format || 'openai',
      });
    }
  };

  const handleCreate = () => {
    setEditing(null);
    setSelectedTemplateId(undefined);
    form.resetFields();
    form.setFieldsValue({ format: 'openai' });
    setModalOpen(true);
  };

  const handleEdit = (record: UserThirdPartyProvider) => {
    setEditing(record);
    setSelectedTemplateId(record.template_id ?? undefined);
    form.setFieldsValue({
      name: record.name,
      openai_base_url: record.openai_base_url || '',
      anthropic_base_url: record.anthropic_base_url || '',
      models: record.models,
      format: record.format || 'openai',
      api_key: '',
    });
    setModalOpen(true);
  };

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      if (editing) {
        const updateData: Record<string, unknown> = {
          name: values.name,
          openai_base_url: values.openai_base_url || '',
          anthropic_base_url: values.anthropic_base_url || '',
          models: values.models,
          format: values.format,
        };
        if (values.api_key && (values.api_key as string).trim()) {
          updateData.api_key = values.api_key;
        }
        await modelService.updateProvider(editing.id, updateData);
        message.success('更新成功');
      } else {
        await modelService.createProvider({
          name: values.name as string,
          openai_base_url: values.openai_base_url as string || '',
          anthropic_base_url: values.anthropic_base_url as string || '',
          api_key: values.api_key as string,
          models: values.models as string[],
          format: values.format as string,
          template_id: selectedTemplateId,
        });
        message.success('添加成功');
      }
      setModalOpen(false);
      form.resetFields();
      loadProviders();
    } catch { /* 拦截器处理 */ }
  };

  const handleToggleStatus = async (record: UserThirdPartyProvider) => {
    const newStatus = record.status === 1 ? 0 : 1;
    await modelService.updateProviderStatus(record.id, newStatus);
    message.success(newStatus === 1 ? '已启用' : '已禁用');
    loadProviders();
  };

  const handleDelete = (record: UserThirdPartyProvider) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除服务「${record.name}」吗？已配置的模型将不再可用。`,
      okText: '删除',
      okType: 'danger',
      okButtonProps: { style: { background: '#FF6B6B', borderColor: '#FF6B6B' } },
      onOk: async () => {
        await modelService.deleteProvider(record.id);
        message.success('删除成功');
        loadProviders();
      },
    });
  };

  // ── 平台模型表格列 ──
  const platformColumns: ColumnsType<PlatformModelInfo> = useMemo(() => [
    {
      title: '服务名称',
      dataIndex: 'display_name',
      key: 'display_name',
      render: (text: string) => (
        <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>
          <CloudServerOutlined style={{ marginRight: 8, color: '#00D9FF' }} />
          {text}
        </span>
      ),
    },
    {
      title: '协议格式',
      dataIndex: 'format',
      key: 'format',
      width: 160,
      render: (v: string) => <FormatTags format={v} />,
    },
    {
      title: '支持模型',
      dataIndex: 'model_patterns',
      key: 'model_patterns',
      render: (v: string) => (
        <code style={{
          background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
          padding: '3px 8px', borderRadius: 6, fontSize: 13,
          color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)',
        }}>
          {v || '*'}
        </code>
      ),
    },
  ], [isDark]);

  // ── 第三方服务表格列 ──
  const providerColumns: ColumnsType<UserThirdPartyProvider> = useMemo(() => [
    {
      title: '服务名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => (
        <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>
          <ThunderboltOutlined style={{ marginRight: 8, color: '#FFBE0B' }} />
          {text}
        </span>
      ),
    },
    {
      title: 'Base URL',
      key: 'base_urls',
      ellipsis: true,
      render: (_: unknown, record: UserThirdPartyProvider) => {
        const urls: { label: string; url: string }[] = [];
        if (record.openai_base_url) urls.push({ label: 'OpenAI', url: record.openai_base_url });
        if (record.anthropic_base_url) urls.push({ label: 'Anthropic', url: record.anthropic_base_url });
        if (urls.length === 0) return '-';
        return (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {urls.map(u => (
              <Tooltip key={u.label} title={u.url}>
                <code style={{
                  background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
                  padding: '2px 6px', borderRadius: 6, fontSize: 12,
                  color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)',
                  display: 'inline-block', maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                }}>
                  {urls.length > 1 && <span style={{ color: '#00D9FF', marginRight: 4, fontSize: 11 }}>{u.label}:</span>}
                  {u.url}
                </code>
              </Tooltip>
            ))}
          </div>
        );
      },
    },
    {
      title: '模型',
      dataIndex: 'models',
      key: 'models',
      render: (models: string[]) => (
        <Space size={4} wrap>
          {(models || []).map(m => (
            <Tag key={m} style={{
              color: '#00D9FF', background: 'rgba(0, 217, 255, 0.1)',
              border: '1px solid rgba(0, 217, 255, 0.2)', borderRadius: 6,
            }}>{m}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '协议',
      dataIndex: 'format',
      key: 'format',
      width: 160,
      render: (v: string) => <FormatTags format={v || 'openai'} />,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: number) => v === 1
        ? <Tag color="success" style={{ borderRadius: 6, border: 'none' }}>启用</Tag>
        : <Tag color="error" style={{ borderRadius: 6, border: 'none' }}>禁用</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (v: string) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13 }}>
          {dayjs(v).format('YYYY-MM-DD HH:mm')}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} style={{ color: '#00D9FF' }}>编辑</Button>
          <Button type="link" size="small"
            icon={record.status === 1 ? <StopOutlined /> : <CheckCircleOutlined />}
            onClick={() => handleToggleStatus(record)}
            style={{ color: record.status === 1 ? '#FFBE0B' : '#00F5D4' }}
          >
            {record.status === 1 ? '禁用' : '启用'}
          </Button>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>删除</Button>
        </Space>
      ),
    },
  ], [isDark]);

  const labelColor = isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.7)';
  const inputStyle = {
    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
    borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
    color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
  };

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        {/* 页面标题 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<AppstoreOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                模型服务
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                查看平台内置模型和管理您的第三方模型服务
              </p>
            </div>
          </div>
        </div>

        {/* 内容区 */}
        <div className="glass-card animate-fade-in-up" style={{ padding: 24, animationDelay: '0.05s' }}>
          <Tabs items={[
            {
              key: 'platform',
              label: <span><CloudServerOutlined style={{ marginRight: 6 }} />CodeMind 模型</span>,
              children: (
                <>
                  <div style={{ marginBottom: 16 }}>
                    <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13 }}>
                      以下为平台内置的 LLM 服务节点，通过 CodeMind 统一接口访问。不同协议使用不同的 Base URL：
                    </div>
                    <div style={{
                      marginTop: 8, padding: '10px 14px', borderRadius: 8, fontSize: 13,
                      background: isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.03)',
                      display: 'flex', flexDirection: 'column', gap: 6,
                    }}>
                      {[
                        { label: 'OpenAI', url: openaiBaseURL, color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.12)', border: 'rgba(0, 245, 212, 0.3)' },
                        { label: 'Anthropic', url: anthropicBaseURL, color: '#9D4EDD', bg: 'rgba(157, 78, 221, 0.12)', border: 'rgba(157, 78, 221, 0.3)' },
                      ].map(item => (
                        <div key={item.label} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                          <Tag style={{ color: item.color, background: item.bg, border: `1px solid ${item.border}`, borderRadius: 4, minWidth: 82, textAlign: 'center' }}>{item.label}</Tag>
                          <code style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)', flex: 1 }}>{item.url}</code>
                          <Tooltip title="复制">
                            <Button type="text" size="small" icon={<CopyOutlined />}
                              style={{ color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}
                              onClick={() => { navigator.clipboard.writeText(item.url); message.success('已复制'); }} />
                          </Tooltip>
                        </div>
                      ))}
                    </div>
                  </div>
                  <Table
                    rowKey="name"
                    columns={platformColumns}
                    dataSource={platformModels}
                    loading={platformLoading}
                    pagination={false}
                    locale={{ emptyText: <Empty description="暂无可用模型服务" /> }}
                    style={{ background: 'transparent' }}
                  />
                </>
              ),
            },
            {
              key: 'third-party',
              label: <span><ThunderboltOutlined style={{ marginRight: 6 }} />第三方服务</span>,
              children: (
                <>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                    <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13 }}>
                      添加您自有的 LLM 服务（如 DeepSeek、Kimi、GLM 等），通过 CodeMind 统一代理访问。
                      <br />
                      <span style={{ color: '#FFBE0B' }}>⚠ 第三方服务的 Token 用量仅供参考，实际用量以服务提供商为准，不计入平台限额。</span>
                    </div>
                    <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}
                      style={{
                        background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                        border: 'none', boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                        height: 40, borderRadius: 12, flexShrink: 0, marginLeft: 24,
                      }}
                    >
                      添加服务
                    </Button>
                  </div>
                  <Table
                    rowKey="id"
                    columns={providerColumns}
                    dataSource={providers}
                    loading={providersLoading}
                    pagination={false}
                    locale={{ emptyText: <Empty description="暂未添加第三方服务" /> }}
                    style={{ background: 'transparent' }}
                  />
                </>
              ),
            },
          ]} />
        </div>

        {/* 添加/编辑弹窗 */}
        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {editing ? '编辑第三方服务' : '添加第三方服务'}
            </span>
          }
          open={modalOpen}
          onCancel={() => { setModalOpen(false); form.resetFields(); }}
          footer={null}
          destroyOnClose
          width={560}
        >
          <Form form={form} layout="vertical" onFinish={handleSubmit}>
            {/* 模板选择 */}
            {!editing && templates.length > 0 && (
              <Form.Item label={<span style={{ color: labelColor }}>从模板快速创建（可选）</span>}>
                <Select
                  placeholder="选择服务模板，自动填充配置"
                  allowClear
                  value={selectedTemplateId}
                  onChange={handleTemplateSelect}
                  options={templates.map(t => ({
                    label: `${t.name}${t.description ? ' — ' + t.description : ''}`,
                    value: t.id,
                  }))}
                  style={inputStyle}
                />
              </Form.Item>
            )}

            <Form.Item name="name" label={<span style={{ color: labelColor }}>服务名称</span>}
              rules={[{ required: true, message: '请输入服务名称' }]}>
              <Input placeholder="例如：DeepSeek" style={inputStyle} />
            </Form.Item>

            <Form.Item name="format" label={<span style={{ color: labelColor }}>协议格式</span>}
              rules={[{ required: true, message: '请选择协议格式' }]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                不同协议对应不同的 Base URL，选择「全部支持」需同时配置两个 URL
              </span>}>
              <Select
                options={[
                  { label: 'OpenAI 协议', value: 'openai' },
                  { label: 'Anthropic 协议', value: 'anthropic' },
                  { label: '全部支持', value: 'all' },
                ]}
                style={inputStyle}
              />
            </Form.Item>

            <Form.Item noStyle shouldUpdate={(prev, cur) => prev.format !== cur.format}>
              {({ getFieldValue }) => {
                const fmt = getFieldValue('format') || 'openai';
                return (
                  <>
                    {(fmt === 'openai' || fmt === 'all') && (
                      <Form.Item name="openai_base_url" label={<span style={{ color: labelColor }}>OpenAI Base URL</span>}
                        rules={[{ required: true, message: '请输入 OpenAI 协议的 Base URL' }]}
                        extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                          例如 https://api.deepseek.com/v1
                        </span>}>
                        <Input placeholder="https://api.example.com/v1" style={inputStyle} />
                      </Form.Item>
                    )}
                    {(fmt === 'anthropic' || fmt === 'all') && (
                      <Form.Item name="anthropic_base_url" label={<span style={{ color: labelColor }}>Anthropic Base URL</span>}
                        rules={[{ required: true, message: '请输入 Anthropic 协议的 Base URL' }]}
                        extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                          例如 https://api.example.com/v1
                        </span>}>
                        <Input placeholder="https://api.example.com/v1" style={inputStyle} />
                      </Form.Item>
                    )}
                  </>
                );
              }}
            </Form.Item>

            <Form.Item name="api_key"
              label={<span style={{ color: labelColor }}>API Key</span>}
              rules={editing ? [] : [{ required: true, message: '请输入 API Key' }]}
              extra={editing ? <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>留空表示不修改</span> : undefined}>
              <Input.Password placeholder={editing ? '留空不修改' : '输入第三方服务的 API Key'} style={inputStyle} />
            </Form.Item>

            <Form.Item name="models" label={<span style={{ color: labelColor }}>模型名称</span>}
              rules={[{ required: true, message: '请输入至少一个模型名称' }]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                输入模型名称后按回车添加，使用时以此名称请求即可路由到该服务
              </span>}>
              <Select mode="tags" placeholder="输入模型名称后回车，如 deepseek-chat" style={inputStyle}
                tokenSeparators={[',']} />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" block
                style={{
                  height: 44, borderRadius: 12,
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  border: 'none', boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                }}>
                {editing ? '保存' : '添加'}
              </Button>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default ModelsPage;
