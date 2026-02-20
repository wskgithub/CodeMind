import { useState, useEffect, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, Space, Tag, message, Typography, DatePicker } from 'antd';
import { PlusOutlined, CopyOutlined, DeleteOutlined, StopOutlined, CheckCircleOutlined, KeyOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { APIKey } from '@/types';
import keyService from '@/services/keyService';

const { Paragraph } = Typography;

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
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

/** API Key 管理页面 — 与首页/登录页新设计风格统一 */
const KeysPage: React.FC = () => {
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [form] = Form.useForm();

  // 加载 Key 列表
  const loadKeys = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await keyService.list();
      setKeys(resp.data.data || []);
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadKeys();
  }, [loadKeys]);

  // 创建 Key
  const handleCreate = async (values: { name: string; expires_at?: dayjs.Dayjs }) => {
    try {
      const resp = await keyService.create({
        name: values.name,
        expires_at: values.expires_at?.toISOString(),
      });
      const data = resp.data.data;
      setNewKey(data.key);
      setCreateOpen(false);
      form.resetFields();
      message.success('API Key 创建成功');
      loadKeys();
    } catch {
      // 错误已在拦截器中处理
    }
  };

  // 切换 Key 状态
  const handleToggleStatus = async (record: APIKey) => {
    const newStatus = record.status === 1 ? 0 : 1;
    await keyService.updateStatus(record.id, newStatus);
    message.success(newStatus === 1 ? '已启用' : '已禁用');
    loadKeys();
  };

  // 删除 Key
  const handleDelete = (record: APIKey) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除 API Key "${record.name}" 吗？此操作不可撤销。`,
      okText: '删除',
      okType: 'danger',
      okButtonProps: {
        style: { background: '#FF6B6B', borderColor: '#FF6B6B' },
      },
      onOk: async () => {
        await keyService.delete(record.id);
        message.success('删除成功');
        loadKeys();
      },
    });
  };

  // 表格列定义 - 新设计
  const columns: ColumnsType<APIKey> = [
    { 
      title: '名称', 
      dataIndex: 'name', 
      key: 'name',
      render: (text) => <span style={{ color: '#fff', fontWeight: 500 }}>{text}</span>,
    },
    {
      title: 'Key 前缀',
      dataIndex: 'key_prefix',
      key: 'key_prefix',
      render: (v: string) => (
        <code style={{ 
          background: 'rgba(0, 217, 255, 0.1)', 
          padding: '4px 8px', 
          borderRadius: 6,
          color: '#00D9FF',
          fontFamily: 'monospace',
        }}>
          {v}...
        </code>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (v: number) =>
        v === 1 ? (
          <Tag color="success" style={{ borderRadius: 6, border: 'none' }}>启用</Tag>
        ) : (
          <Tag color="error" style={{ borderRadius: 6, border: 'none' }}>禁用</Tag>
        ),
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (v: string) => (
        <span style={{ color: 'rgba(255, 255, 255, 0.6)' }}>
          {v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'}
        </span>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => (
        <span style={{ color: 'rgba(255, 255, 255, 0.6)' }}>
          {dayjs(v).format('YYYY-MM-DD HH:mm')}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={record.status === 1 ? <StopOutlined /> : <CheckCircleOutlined />}
            onClick={() => handleToggleStatus(record)}
            style={{ color: record.status === 1 ? '#FFBE0B' : '#00F5D4' }}
          >
            {record.status === 1 ? '禁用' : '启用'}
          </Button>
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        {/* 页面标题 — 带渐变图标 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<KeyOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
                API Key 管理
              </h2>
              <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
                管理您的 API 密钥，用于接入 CodeMind AI 编码服务
              </p>
            </div>
          </div>
        </div>

        {/* 主内容区 — 玻璃态卡片包裹 - 新设计 */}
        <div
          className="glass-card animate-fade-in-up"
          style={{ padding: 24, animationDelay: '0.05s' }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <span style={{ fontWeight: 600, color: '#fff', fontSize: 16 }}>密钥列表</span>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={() => setCreateOpen(true)}
              style={{
                background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                border: 'none',
                boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                height: 40,
                borderRadius: 12,
              }}
            >
              创建 Key
            </Button>
          </div>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={keys}
            loading={loading}
            pagination={false}
            style={{ background: 'transparent' }}
          />
        </div>

        {/* 创建 Key 弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: '#fff', fontSize: 18, fontWeight: 600 }}>
              创建 API Key
            </span>
          }
          open={createOpen}
          onCancel={() => setCreateOpen(false)}
          footer={null}
          destroyOnClose
        >
          <Form form={form} layout="vertical" onFinish={handleCreate}>
            <Form.Item 
              name="name" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>Key 名称</span>} 
              rules={[{ required: true, message: '请输入名称' }]}
            >
              <Input 
                placeholder="例如：VSCode Cline 插件" 
                style={{ 
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="expires_at" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>过期时间（可选）</span>}
            >
              <DatePicker 
                style={{ width: '100%' }} 
                placeholder="留空表示永不过期"
                suffixIcon={<span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>📅</span>}
              />
            </Form.Item>
            <Form.Item>
              <Button 
                type="primary" 
                htmlType="submit" 
                block
                style={{
                  height: 44,
                  borderRadius: 12,
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  border: 'none',
                  boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                }}
              >
                创建
              </Button>
            </Form.Item>
          </Form>
        </Modal>

        {/* 显示新创建的 Key - 新设计 */}
        <Modal
          title={
            <span style={{ color: '#fff', fontSize: 18, fontWeight: 600 }}>
              API Key 已创建
            </span>
          }
          open={!!newKey}
          onCancel={() => setNewKey(null)}
          onOk={() => setNewKey(null)}
          okText="知道了"
          okButtonProps={{
            style: {
              background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
              border: 'none',
              borderRadius: 10,
            },
          }}
          cancelButtonProps={{ style: { display: 'none' } }}
        >
          <p style={{ color: '#FFBE0B', fontSize: 14, marginBottom: 16 }}>
            ⚠️ 请立即复制此 Key，关闭后将无法再次查看！
          </p>
          <Paragraph
            copyable={{ icon: <CopyOutlined style={{ color: '#00D9FF' }} />, tooltips: ['复制', '已复制'] }}
            style={{
              background: 'rgba(0, 217, 255, 0.08)',
              backdropFilter: 'blur(12px)',
              WebkitBackdropFilter: 'blur(12px)',
              border: '1px solid rgba(0, 217, 255, 0.2)',
              padding: '16px 20px',
              borderRadius: 14,
              fontFamily: 'monospace',
              wordBreak: 'break-all',
              color: '#00D9FF',
              fontSize: 14,
            }}
          >
            {newKey}
          </Paragraph>
        </Modal>
      </div>
    </div>
  );
};

export default KeysPage;
