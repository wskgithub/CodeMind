import {
  PlusOutlined, DeleteOutlined, EditOutlined, BlockOutlined,
} from '@ant-design/icons';
import { Table, Button, Modal, Form, Input, InputNumber, Space, Tag, message, Select } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useState, useEffect, useCallback, useMemo } from 'react';

import modelService from '@/services/modelService';
import useAppStore from '@/store/appStore';
import type { ProviderTemplate } from '@/types';

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

const ProviderTemplatesPage: React.FC = () => {
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  const [templates, setTemplates] = useState<ProviderTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<ProviderTemplate | null>(null);
  const [form] = Form.useForm();

  const loadTemplates = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await modelService.listTemplatesAdmin();
      setTemplates(resp.data.data || []);
    } catch { /* 拦截器处理 */ }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { loadTemplates(); }, [loadTemplates]);

  const handleCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ sort_order: 0, format: 'openai' });
    setModalOpen(true);
  };

  const handleEdit = (record: ProviderTemplate) => {
    setEditing(record);
    form.setFieldsValue({
      name: record.name,
      openai_base_url: record.openai_base_url || '',
      anthropic_base_url: record.anthropic_base_url || '',
      models: record.models,
      format: record.format || 'openai',
      description: record.description,
      sort_order: record.sort_order,
    });
    setModalOpen(true);
  };

  const handleSubmit = async (values: Record<string, unknown>) => {
    try {
      if (editing) {
        await modelService.updateTemplate(editing.id, {
          name: values.name as string,
          openai_base_url: values.openai_base_url as string || '',
          anthropic_base_url: values.anthropic_base_url as string || '',
          models: values.models as string[],
          format: values.format as string,
          description: values.description as string | undefined,
          sort_order: values.sort_order as number | undefined,
        });
        message.success('更新成功');
      } else {
        await modelService.createTemplate({
          name: values.name as string,
          openai_base_url: values.openai_base_url as string || '',
          anthropic_base_url: values.anthropic_base_url as string || '',
          models: values.models as string[],
          format: values.format as string,
          description: values.description as string | undefined,
          sort_order: values.sort_order as number | undefined,
        });
        message.success('创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      loadTemplates();
    } catch { /* 拦截器处理 */ }
  };

  const handleToggleStatus = async (record: ProviderTemplate) => {
    const newStatus = record.status === 1 ? 0 : 1;
    await modelService.updateTemplate(record.id, { status: newStatus });
    message.success(newStatus === 1 ? '已启用' : '已禁用');
    loadTemplates();
  };

  const handleDelete = (record: ProviderTemplate) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除模板「${record.name}」吗？已使用此模板的用户服务不受影响。`,
      okText: '删除',
      okType: 'danger',
      okButtonProps: { style: { background: '#FF6B6B', borderColor: '#FF6B6B' } },
      onOk: async () => {
        await modelService.deleteTemplate(record.id);
        message.success('删除成功');
        loadTemplates();
      },
    });
  };

  const columns: ColumnsType<ProviderTemplate> = useMemo(() => [
    {
      title: '模板名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: ProviderTemplate) => (
        <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>
          {record.icon ? <span style={{ marginRight: 6 }}>{record.icon}</span> : null}
          {text}
        </span>
      ),
    },
    {
      title: 'Base URL',
      key: 'base_urls',
      ellipsis: true,
      render: (_: unknown, record: ProviderTemplate) => {
        const urls: { label: string; url: string }[] = [];
        if (record.openai_base_url) urls.push({ label: 'OpenAI', url: record.openai_base_url });
        if (record.anthropic_base_url) urls.push({ label: 'Anthropic', url: record.anthropic_base_url });
        if (urls.length === 0) return '-';
        return (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {urls.map(u => (
              <code key={u.label} style={{
                background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
                padding: '2px 6px', borderRadius: 6, fontSize: 12,
                color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)',
                display: 'inline-block', maxWidth: 280, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
              }}>
                {urls.length > 1 && <span style={{ color: '#00D9FF', marginRight: 4, fontSize: 11 }}>{u.label}:</span>}
                {u.url}
              </code>
            ))}
          </div>
        );
      },
    },
    {
      title: '模型列表',
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
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (v: string) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>
          {v || '-'}
        </span>
      ),
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
      title: '排序',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 70,
      render: (v: number) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{v}</span>
      ),
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
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} style={{ color: '#00D9FF' }}>编辑</Button>
          <Button type="link" size="small"
            icon={record.status === 1 ? <></> : <></>}
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
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<BlockOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                模型模板管理
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                预置第三方 LLM 服务模板，用户创建时仅需填入 API Key 即可完成配置
              </p>
            </div>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24, animationDelay: '0.05s' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <span style={{ fontWeight: 600, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 16 }}>模板列表</span>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}
              style={{
                background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                border: 'none', boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                height: 40, borderRadius: 12,
              }}
            >
              创建模板
            </Button>
          </div>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={templates}
            loading={loading}
            pagination={false}
            style={{ background: 'transparent' }}
          />
        </div>

        {/* 创建/编辑弹窗 */}
        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {editing ? '编辑模板' : '创建模板'}
            </span>
          }
          open={modalOpen}
          onCancel={() => { setModalOpen(false); form.resetFields(); }}
          footer={null}
          destroyOnClose
          width={560}
        >
          <Form form={form} layout="vertical" onFinish={handleSubmit}>
            <Form.Item name="name" label={<span style={{ color: labelColor }}>模板名称</span>}
              rules={[{ required: true, message: '请输入模板名称' }]}>
              <Input placeholder="例如：DeepSeek" style={inputStyle} />
            </Form.Item>

            <Form.Item name="format" label={<span style={{ color: labelColor }}>协议格式</span>}
              rules={[{ required: true, message: '请选择协议格式' }]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                不同协议使用不同的 Base URL，选择「全部支持」需同时配置两个 URL
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
                        rules={[{ required: true, message: '请输入 OpenAI 协议的 Base URL' }]}>
                        <Input placeholder="https://api.deepseek.com/v1" style={inputStyle} />
                      </Form.Item>
                    )}
                    {(fmt === 'anthropic' || fmt === 'all') && (
                      <Form.Item name="anthropic_base_url" label={<span style={{ color: labelColor }}>Anthropic Base URL</span>}
                        rules={[{ required: true, message: '请输入 Anthropic 协议的 Base URL' }]}>
                        <Input placeholder="https://api.example.com/v1" style={inputStyle} />
                      </Form.Item>
                    )}
                  </>
                );
              }}
            </Form.Item>

            <Form.Item name="models" label={<span style={{ color: labelColor }}>模型列表</span>}
              rules={[{ required: true, message: '请输入至少一个模型名称' }]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                输入模型名称后按回车添加
              </span>}>
              <Select mode="tags" placeholder="如 deepseek-chat, deepseek-reasoner" style={inputStyle} tokenSeparators={[',']} />
            </Form.Item>

            <Form.Item name="description" label={<span style={{ color: labelColor }}>描述（可选）</span>}>
              <Input placeholder="服务简要描述" style={inputStyle} />
            </Form.Item>

            <Form.Item name="sort_order" label={<span style={{ color: labelColor }}>排序权重</span>}>
              <InputNumber min={0} style={{ ...inputStyle, width: '100%' }} placeholder="越小越靠前" />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" block
                style={{
                  height: 44, borderRadius: 12,
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  border: 'none', boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                }}>
                {editing ? '保存' : '创建'}
              </Button>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default ProviderTemplatesPage;
