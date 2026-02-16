import { useState, useEffect, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, Space, Tag, message, Typography, DatePicker, theme } from 'antd';
import { PlusOutlined, CopyOutlined, DeleteOutlined, StopOutlined, CheckCircleOutlined, KeyOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { APIKey } from '@/types';
import keyService from '@/services/keyService';

const { Paragraph } = Typography;

/** API Key 管理页面 — Glassmorphism 风格 */
const KeysPage: React.FC = () => {
  const { token } = theme.useToken();
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
      onOk: async () => {
        await keyService.delete(record.id);
        message.success('删除成功');
        loadKeys();
      },
    });
  };

  // 表格列定义
  const columns: ColumnsType<APIKey> = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    {
      title: 'Key 前缀',
      dataIndex: 'key_prefix',
      key: 'key_prefix',
      render: (v: string) => <code>{v}...</code>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (v: number) =>
        v === 1 ? <Tag color="success">启用</Tag> : <Tag color="error">禁用</Tag>,
    },
    {
      title: '最后使用',
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (v: string) => (v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm'),
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
        {/* 页面标题 — 带渐变图标 */}
        <h2
          style={{
            marginBottom: 24,
            color: token.colorTextHeading,
            display: 'flex',
            alignItems: 'center',
            gap: 12,
          }}
        >
          <span
            className="flex items-center justify-center w-10 h-10 rounded-xl shrink-0"
            style={{
              background: 'var(--gradient-primary)',
              color: '#fff',
            }}
          >
            <KeyOutlined style={{ fontSize: 20 }} />
          </span>
          API Key 管理
        </h2>

        {/* 主内容区 — 玻璃态卡片包裹 */}
        <div
          className="glass-card animate-fade-in-up"
          style={{ padding: 24, animationDelay: '0.05s' }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <span style={{ fontWeight: 600, color: token.colorTextHeading }}>密钥列表</span>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
              创建 Key
            </Button>
          </div>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={keys}
            loading={loading}
            pagination={false}
          />
        </div>

        {/* 创建 Key 弹窗 */}
        <Modal
          title="创建 API Key"
          open={createOpen}
          onCancel={() => setCreateOpen(false)}
          footer={null}
          destroyOnClose
        >
          <Form form={form} layout="vertical" onFinish={handleCreate}>
            <Form.Item name="name" label="Key 名称" rules={[{ required: true, message: '请输入名称' }]}>
              <Input placeholder="例如：VSCode Cline 插件" />
            </Form.Item>
            <Form.Item name="expires_at" label="过期时间（可选）">
              <DatePicker style={{ width: '100%' }} placeholder="留空表示永不过期" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" block>
                创建
              </Button>
            </Form.Item>
          </Form>
        </Modal>

        {/* 显示新创建的 Key */}
        <Modal
          title="API Key 已创建"
          open={!!newKey}
          onCancel={() => setNewKey(null)}
          onOk={() => setNewKey(null)}
          okText="知道了"
          cancelButtonProps={{ style: { display: 'none' } }}
        >
          <p className="text-orange-500 text-sm mb-4">
            请立即复制此 Key，关闭后将无法再次查看！
          </p>
          <Paragraph
            copyable={{ icon: <CopyOutlined />, tooltips: ['复制', '已复制'] }}
            style={{
              background: 'var(--glass-bg)',
              backdropFilter: 'blur(12px)',
              WebkitBackdropFilter: 'blur(12px)',
              border: '1px solid var(--glass-border)',
              padding: '12px 16px',
              borderRadius: 12,
              fontFamily: 'monospace',
              wordBreak: 'break-all',
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
