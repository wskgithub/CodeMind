import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, InputNumber, Select, message, Popconfirm, Tag, Space, theme } from 'antd';
import { PlusOutlined, DeleteOutlined, ThunderboltOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { listLimits, upsertLimit, deleteLimit } from '@/services/limitService';
import type { RateLimit } from '@/types';

/** 页面标题图标 — 渐变圆形背景 */
const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-10 h-10 rounded-full shrink-0"
    style={{
      background: 'linear-gradient(135deg, #faad14 0%, #ffc53d 100%)',
      color: '#fff',
    }}
  >
    {icon}
  </span>
);

/** 限额管理页面 — Glassmorphism 风格，管理用户/部门/全局的 Token 限额配置 */
const LimitsPage = () => {
  const { token } = theme.useToken();
  const [limits, setLimits] = useState<RateLimit[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();

  useEffect(() => {
    loadLimits();
  }, []);

  const loadLimits = async () => {
    setLoading(true);
    try {
      const res = await listLimits();
      setLimits(res.data.data || []);
    } catch {
      // 拦截器处理
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    form.resetFields();
    form.setFieldsValue({
      max_concurrency: 5,
      alert_threshold: 80,
    });
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      await upsertLimit(values);
      message.success('限额配置已保存');
      setModalOpen(false);
      loadLimits();
    } catch {
      // 验证或请求失败
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteLimit(id);
      message.success('已删除');
      loadLimits();
    } catch {
      // 拦截器处理
    }
  };

  const targetTypeLabel: Record<string, { text: string; color: string }> = {
    global: { text: '全局', color: 'red' },
    department: { text: '部门', color: 'blue' },
    user: { text: '用户', color: 'green' },
  };

  const periodLabel: Record<string, string> = {
    daily: '每日',
    weekly: '每周',
    monthly: '每月',
    custom: '自定义',
  };

  const formatPeriodHours = (hours: number): string => {
    if (hours < 24) return `${hours} 小时`;
    if (hours === 24) return '24 小时 (每日)';
    if (hours === 168) return '168 小时 (每周)';
    if (hours === 720) return '720 小时 (每月)';
    return `${hours} 小时`;
  };

  const columns: ColumnsType<RateLimit> = [
    {
      title: '目标类型',
      dataIndex: 'target_type',
      key: 'target_type',
      width: 100,
      render: (val: string) => {
        const t = targetTypeLabel[val];
        return t ? <Tag color={t.color}>{t.text}</Tag> : val;
      },
    },
    { title: '目标 ID', dataIndex: 'target_id', key: 'target_id', width: 80 },
    {
      title: '周期',
      dataIndex: 'period',
      key: 'period',
      width: 80,
      render: (val: string) => periodLabel[val] || val,
    },
    {
      title: '周期时长',
      dataIndex: 'period_hours',
      key: 'period_hours',
      width: 140,
      render: (v: number) => formatPeriodHours(v),
    },
    {
      title: '最大 Token 数',
      dataIndex: 'max_tokens',
      key: 'max_tokens',
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: '最大请求数',
      dataIndex: 'max_requests',
      key: 'max_requests',
      align: 'right',
      render: (v: number) => (v === 0 ? '不限制' : v.toLocaleString()),
    },
    {
      title: '最大并发',
      dataIndex: 'max_concurrency',
      key: 'max_concurrency',
      align: 'right',
    },
    {
      title: '告警阈值',
      dataIndex: 'alert_threshold',
      key: 'alert_threshold',
      align: 'right',
      render: (v: number) => `${v}%`,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: number) =>
        v === 1 ? <Tag color="green">启用</Tag> : <Tag color="default">禁用</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Popconfirm title="确定删除该限额配置？" onConfirm={() => handleDelete(record.id)}>
          <Button type="link" danger icon={<DeleteOutlined />} size="small">
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div
      className="animate-fade-in-up"
      style={{
        background: 'var(--glass-bg)',
        backdropFilter: 'blur(16px)',
        WebkitBackdropFilter: 'blur(16px)',
        border: '1px solid var(--glass-border)',
        borderRadius: 16,
        boxShadow: 'var(--glass-shadow)',
        padding: 24,
      }}
    >
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <div className="flex items-center gap-3 mb-2">
          <PageIcon icon={<ThunderboltOutlined style={{ fontSize: 20 }} />} />
          <h2 style={{ margin: 0, color: token.colorTextHeading }}>限额管理</h2>
        </div>
        <p style={{ margin: 0, color: token.colorTextSecondary, fontSize: 14 }}>
          配置全局、部门及用户的 Token 限额、请求数、并发数与告警阈值。
        </p>
        <div style={{ marginTop: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新建限额
          </Button>
        </div>
      </div>

      {/* 表格区域 — 交由全局 CSS 处理行悬停 */}
      <Table
        dataSource={limits}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={false}
      />

      {/* 创建/编辑限额弹窗 */}
      <Modal
        title="配置限额"
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={submitting}
        width={520}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Space style={{ width: '100%' }} size={12}>
            <Form.Item
              name="target_type"
              label="目标类型"
              rules={[{ required: true, message: '请选择' }]}
              style={{ width: 150 }}
            >
              <Select
                options={[
                  { label: '全局', value: 'global' },
                  { label: '部门', value: 'department' },
                  { label: '用户', value: 'user' },
                ]}
              />
            </Form.Item>
            <Form.Item
              name="target_id"
              label="目标 ID"
              rules={[{ required: true, message: '请输入' }]}
              style={{ width: 120 }}
            >
              <InputNumber min={0} style={{ width: '100%' }} placeholder="0=全局" />
            </Form.Item>
            <Form.Item
              name="period"
              label="周期"
              rules={[{ required: true, message: '请选择' }]}
              style={{ width: 120 }}
            >
              <Select
                options={[
                  { label: '每日', value: 'daily' },
                  { label: '每周', value: 'weekly' },
                  { label: '每月', value: 'monthly' },
                  { label: '自定义', value: 'custom' },
                ]}
              />
            </Form.Item>
          </Space>

          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) => prev.period !== cur.period}
          >
            {({ getFieldValue }) =>
              getFieldValue('period') === 'custom' ? (
                <Form.Item
                  name="period_hours"
                  label="自定义周期（小时）"
                  rules={[{ required: true, message: '请输入周期时长' }]}
                >
                  <InputNumber min={1} style={{ width: '100%' }} placeholder="如 5 表示每 5 小时" />
                </Form.Item>
              ) : null
            }
          </Form.Item>

          <Form.Item
            name="max_tokens"
            label="最大 Token 数"
            rules={[{ required: true, message: '请输入' }]}
          >
            <InputNumber min={0} style={{ width: '100%' }} placeholder="如 1000000" />
          </Form.Item>
          <Form.Item name="max_requests" label="最大请求数（0=不限制）">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="max_concurrency" label="最大并发数">
            <InputNumber min={1} max={100} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="alert_threshold" label="告警阈值 (%)">
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default LimitsPage;
