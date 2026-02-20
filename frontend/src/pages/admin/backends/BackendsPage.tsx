import { useEffect, useState } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, InputNumber,
  Tag, Space, message, Popconfirm, Tooltip, Divider,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CloudServerOutlined,
  QuestionCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { listLLMBackends, createLLMBackend, updateLLMBackend, deleteLLMBackend } from '@/services/llmBackendService';
import type { LLMBackend } from '@/types';

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(157, 78, 221, 0.25)',
    }}
  >
    {icon}
  </span>
);

const statusMap: Record<number, { label: string; color: string; bg: string }> = {
  0: { label: '禁用', color: 'rgba(255, 255, 255, 0.5)', bg: 'rgba(255, 255, 255, 0.05)' },
  1: { label: '启用', color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
  2: { label: '排空', color: '#FFBE0B', bg: 'rgba(255, 190, 11, 0.15)' },
};

/** 表单字段标签 + 提示 */
const FieldLabel = ({ label, tip }: { label: string; tip: string }) => (
  <Space size={4}>
    <span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{label}</span>
    <Tooltip title={tip}>
      <QuestionCircleOutlined style={{ color: 'rgba(255, 255, 255, 0.4)', fontSize: 12 }} />
    </Tooltip>
  </Space>
);

/** LLM 后端节点管理页 — 与首页/登录页新设计风格统一 */
const BackendsPage: React.FC = () => {
  const [backends, setBackends] = useState<LLMBackend[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<LLMBackend | null>(null);
  const [form] = Form.useForm();

  useEffect(() => { loadData(); }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await listLLMBackends();
      setBackends(res.data.data || []);
    } catch { /* 错误已由拦截器处理 */ }
    finally { setLoading(false); }
  };

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({
      format: 'openai',
      weight: 100,
      max_concurrency: 100,
      model_patterns: '*',
      timeout_seconds: 300,
      stream_timeout_seconds: 600,
    });
    setModalOpen(true);
  };

  const openEdit = (record: LLMBackend) => {
    setEditing(record);
    form.setFieldsValue({
      name: record.name,
      display_name: record.display_name,
      base_url: record.base_url,
      format: record.format,
      weight: record.weight,
      max_concurrency: record.max_concurrency,
      status: record.status,
      model_patterns: record.model_patterns,
      timeout_seconds: record.timeout_seconds,
      stream_timeout_seconds: record.stream_timeout_seconds,
      health_check_url: record.health_check_url,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (!values.api_key) delete values.api_key;
      if (editing) {
        await updateLLMBackend(editing.id, values);
        message.success('节点已更新');
      } else {
        await createLLMBackend(values);
        message.success('节点已创建');
      }
      setModalOpen(false);
      loadData();
    } catch { /* 表单验证失败 */ }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteLLMBackend(id);
      message.success('节点已删除');
      loadData();
    } catch { /* 错误已由拦截器处理 */ }
  };

  // 表格列 - 新设计
  const columns = [
    {
      title: '节点',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: LLMBackend) => (
        <div>
          <div style={{ fontWeight: 600, color: '#fff', fontSize: 15 }}>
            {record.display_name || name}
          </div>
          <div style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>{name}</div>
        </div>
      ),
    },
    {
      title: '服务地址',
      dataIndex: 'base_url',
      key: 'base_url',
      ellipsis: true,
      render: (url: string) => (
        <span style={{ fontSize: 13, color: 'rgba(255, 255, 255, 0.6)', fontFamily: 'monospace' }}>
          {url}
        </span>
      ),
    },
    {
      title: '协议',
      dataIndex: 'format',
      key: 'format',
      width: 100,
      render: (f: string) => (
        <Tag style={{
          color: f === 'openai' ? '#00D9FF' : '#9D4EDD',
          background: f === 'openai' ? 'rgba(0, 217, 255, 0.15)' : 'rgba(157, 78, 221, 0.15)',
          border: `1px solid ${f === 'openai' ? 'rgba(0, 217, 255, 0.3)' : 'rgba(157, 78, 221, 0.3)'}`,
          borderRadius: 6,
        }}>
          {f}
        </Tag>
      ),
    },
    {
      title: '模型模式',
      dataIndex: 'model_patterns',
      key: 'model_patterns',
      width: 180,
      ellipsis: true,
      render: (p: string) => (
        <Tooltip title={p}>
          <code style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.6)', fontFamily: 'monospace' }}>{p}</code>
        </Tooltip>
      ),
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
      width: 70,
      align: 'center' as const,
      render: (w: number) => <span style={{ fontWeight: 600, color: '#FFBE0B' }}>{w}</span>,
    },
    {
      title: '并发',
      dataIndex: 'max_concurrency',
      key: 'max_concurrency',
      width: 70,
      align: 'center' as const,
      render: (v: number) => <span style={{ color: '#00F5D4' }}>{v}</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s: number) => {
        const info = statusMap[s] ?? { label: '未知', color: 'rgba(255, 255, 255, 0.5)', bg: 'rgba(255, 255, 255, 0.05)' };
        return (
          <Tag style={{ 
            color: info.color,
            background: info.bg,
            border: `1px solid ${info.color}40`,
            borderRadius: 6,
          }}>
            {info.label}
          </Tag>
        );
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 130,
      render: (_: unknown, record: LLMBackend) => (
        <Space size="small">
          <Button 
            type="link" 
            size="small" 
            icon={<EditOutlined />} 
            onClick={() => openEdit(record)}
            style={{ color: '#00D9FF' }}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除此节点？"
            description="删除后该节点将立即从负载均衡中移除。"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
              <PageIcon icon={<CloudServerOutlined />} />
              <div>
                <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
                  LLM 节点管理
                </h2>
                <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
                  配置后端 LLM 服务节点，每个节点独立管理地址、密钥、模型和负载策略
                </p>
              </div>
            </div>
            <Space>
              <Button 
                icon={<ReloadOutlined />} 
                onClick={loadData}
                style={{
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: 'rgba(255, 255, 255, 0.8)',
                }}
              >
                刷新
              </Button>
              <Button 
                type="primary" 
                icon={<PlusOutlined />} 
                onClick={openCreate}
                style={{
                  background: 'linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)',
                  border: 'none',
                  boxShadow: '0 4px 16px rgba(157, 78, 221, 0.25)',
                }}
              >
                添加节点
              </Button>
            </Space>
          </div>
        </div>

        {/* 节点列表 — 玻璃态卡片 - 新设计 */}
        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Table
            columns={columns}
            dataSource={backends}
            rowKey="id"
            loading={loading}
            pagination={false}
            size="middle"
            locale={{ 
              emptyText: <span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>暂无节点，点击右上角「添加节点」开始配置</span> 
            }}
          />
        </div>

        {/* 创建 / 编辑弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: '#fff', fontSize: 18, fontWeight: 600 }}>
              <CloudServerOutlined style={{ marginRight: 8 }} />
              {editing ? `编辑节点：${editing.display_name || editing.name}` : '添加后端节点'}
            </span>
          }
          open={modalOpen}
          onOk={handleSubmit}
          onCancel={() => setModalOpen(false)}
          okText={editing ? '保存' : '创建'}
          width={600}
          destroyOnClose
          okButtonProps={{
            style: {
              background: 'linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)',
              border: 'none',
              boxShadow: '0 4px 16px rgba(157, 78, 221, 0.25)',
            },
          }}
        >
          <Form form={form} layout="vertical" style={{ marginTop: 16 }}>

            {/* ── 基本信息 ── */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="name"
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>节点标识</span>}
                rules={[{ required: true, message: '请输入节点标识' }]}
              >
                <Input 
                  placeholder="如 gpu-node-01" 
                  disabled={!!editing}
                  style={{ 
                    background: 'rgba(255, 255, 255, 0.03)',
                    borderColor: 'rgba(255, 255, 255, 0.1)',
                    color: '#fff',
                  }}
                />
              </Form.Item>
              <Form.Item 
                name="display_name" 
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>显示名称</span>}
              >
                <Input 
                  placeholder="如 GPU 服务器 01"
                  style={{ 
                    background: 'rgba(255, 255, 255, 0.03)',
                    borderColor: 'rgba(255, 255, 255, 0.1)',
                    color: '#fff',
                  }}
                />
              </Form.Item>
            </div>

            <Form.Item
              name="base_url"
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>服务地址</span>}
              rules={[{ required: true, message: '请输入服务地址' }]}
            >
              <Input 
                placeholder="http://192.168.1.100:8000/v1" 
                style={{ fontFamily: 'monospace', color: '#00D9FF' }}
              />
            </Form.Item>

            <Form.Item 
              name="api_key" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>API Key</span>}
            >
              <Input.Password 
                placeholder={editing ? '留空则保持不变' : '无需认证时留空'}
                style={{ 
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>

            <Divider style={{ margin: '4px 0 16px', borderColor: 'rgba(255, 255, 255, 0.08)' }} />

            {/* ── 协议与状态 ── */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="format"
                label={<FieldLabel label="协议格式" tip="OpenAI 兼容接口或 Anthropic 原生接口" />}
                rules={[{ required: true }]}
              >
                <Select options={[
                  { value: 'openai', label: 'OpenAI 兼容' },
                  { value: 'anthropic', label: 'Anthropic 原生' },
                ]} />
              </Form.Item>
              <Form.Item 
                name="status" 
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>运行状态</span>}
              >
                <Select options={[
                  { value: 1, label: '启用' },
                  { value: 0, label: '禁用' },
                  { value: 2, label: '排空（不接新请求）' },
                ]} />
              </Form.Item>
            </div>

            <Divider style={{ margin: '4px 0 16px', borderColor: 'rgba(255, 255, 255, 0.08)' }} />

            {/* ── 模型路由 ── */}
            <Form.Item
              name="model_patterns"
              label={<FieldLabel label="支持的模型" tip="逗号分隔的模型名匹配模式，支持通配符 *。如 gpt-*,o1-* 表示只路由 GPT 和 o1 系列请求" />}
            >
              <Input 
                placeholder="* 匹配所有模型，多个模式用逗号分隔，如 gpt-*,claude-*" 
                style={{ fontFamily: 'monospace', color: '#fff' }}
              />
            </Form.Item>

            <Divider style={{ margin: '4px 0 16px', borderColor: 'rgba(255, 255, 255, 0.08)' }} />

            {/* ── 负载均衡 ── */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="weight"
                label={<FieldLabel label="负载权重" tip="相对权重，数值越大分配的请求越多。如节点 A 权重 200、节点 B 权重 100，则 A 承接约 2/3 的请求" />}
              >
                <InputNumber min={1} max={10000} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item
                name="max_concurrency"
                label={<FieldLabel label="最大并发" tip="此节点允许的最大同时请求数，超出后新请求将路由至其他节点" />}
              >
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </div>

            <Divider style={{ margin: '4px 0 16px', borderColor: 'rgba(255, 255, 255, 0.08)' }} />

            {/* ── 超时配置 ── */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="timeout_seconds"
                label={<FieldLabel label="普通请求超时（秒）" tip="非流式请求的最大等待时间" />}
              >
                <InputNumber min={10} max={3600} style={{ width: '100%' }} />
              </Form.Item>
              <Form.Item
                name="stream_timeout_seconds"
                label={<FieldLabel label="流式请求超时（秒）" tip="流式（SSE）请求的最大等待时间，建议设置较长" />}
              >
                <InputNumber min={10} max={7200} style={{ width: '100%' }} />
              </Form.Item>
            </div>

            <Form.Item
              name="health_check_url"
              label={<FieldLabel label="健康检查地址" tip="可选。系统定期 GET 此 URL，响应 2xx 则视为节点健康。留空则跳过健康检查" />}
            >
              <Input 
                placeholder="http://192.168.1.100:8000/health" 
                style={{ fontFamily: 'monospace', color: '#fff' }}
              />
            </Form.Item>

          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default BackendsPage;
