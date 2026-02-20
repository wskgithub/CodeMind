import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, InputNumber, Select, message, Popconfirm, Tag, Space } from 'antd';
import { PlusOutlined, DeleteOutlined, SafetyOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { listLimits, upsertLimit, deleteLimit } from '@/services/limitService';
import type { RateLimit } from '@/types';

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(255, 190, 11, 0.25)',
    }}
  >
    {icon}
  </span>
);

/** 限额管理页面 — 与首页/登录页新设计风格统一 */
const LimitsPage = () => {
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

  // 目标类型标签 - 新设计
  const targetTypeLabel: Record<string, { text: string; color: string; bg: string }> = {
    global: { text: '全局', color: '#FF6B6B', bg: 'rgba(255, 107, 107, 0.15)' },
    department: { text: '部门', color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
    user: { text: '用户', color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
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

  // 表格列 - 新设计
  const columns: ColumnsType<RateLimit> = [
    {
      title: '目标类型',
      dataIndex: 'target_type',
      key: 'target_type',
      width: 100,
      render: (val: string) => {
        const t = targetTypeLabel[val];
        return t ? (
          <Tag style={{ 
            color: t.color,
            background: t.bg,
            border: `1px solid ${t.color}40`,
            borderRadius: 6,
          }}>
            {t.text}
          </Tag>
        ) : val;
      },
    },
    { 
      title: '目标 ID', 
      dataIndex: 'target_id', 
      key: 'target_id', 
      width: 80,
      render: (v) => <span style={{ color: '#fff' }}>{v}</span>,
    },
    {
      title: '周期',
      dataIndex: 'period',
      key: 'period',
      width: 80,
      render: (val: string) => <span style={{ color: 'rgba(255, 255, 255, 0.7)' }}>{periodLabel[val] || val}</span>,
    },
    {
      title: '周期时长',
      dataIndex: 'period_hours',
      key: 'period_hours',
      width: 140,
      render: (v: number) => <span style={{ color: 'rgba(255, 255, 255, 0.7)' }}>{formatPeriodHours(v)}</span>,
    },
    {
      title: '最大 Token 数',
      dataIndex: 'max_tokens',
      key: 'max_tokens',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00D9FF', fontWeight: 500 }}>{v.toLocaleString()}</span>,
    },
    {
      title: '最大请求数',
      dataIndex: 'max_requests',
      key: 'max_requests',
      align: 'right',
      render: (v: number) => <span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{v === 0 ? '不限制' : v.toLocaleString()}</span>,
    },
    {
      title: '最大并发',
      dataIndex: 'max_concurrency',
      key: 'max_concurrency',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00F5D4' }}>{v}</span>,
    },
    {
      title: '告警阈值',
      dataIndex: 'alert_threshold',
      key: 'alert_threshold',
      align: 'right',
      render: (v: number) => <span style={{ color: '#FFBE0B' }}>{v}%</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: number) =>
        v === 1 ? (
          <Tag style={{ 
            color: '#00F5D4', 
            background: 'rgba(0, 245, 212, 0.15)',
            border: '1px solid rgba(0, 245, 212, 0.3)',
            borderRadius: 6,
          }}>
            启用
          </Tag>
        ) : (
          <Tag style={{ 
            color: 'rgba(255, 255, 255, 0.5)', 
            background: 'rgba(255, 255, 255, 0.05)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            borderRadius: 6,
          }}>
            禁用
          </Tag>
        ),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Popconfirm 
          title="确定删除该限额配置？" 
          onConfirm={() => handleDelete(record.id)}
        >
          <Button type="link" danger icon={<DeleteOutlined />} size="small">
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<SafetyOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
                限额管理
              </h2>
              <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
                配置全局、部门及用户的 Token 限额、请求数、并发数与告警阈值
              </p>
            </div>
          </div>
          <div style={{ marginTop: 20, display: 'flex', gap: 12 }}>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={loadLimits}
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
              onClick={handleCreate}
              style={{
                background: 'linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)',
                border: 'none',
                boxShadow: '0 4px 16px rgba(255, 190, 11, 0.25)',
              }}
            >
              新建限额
            </Button>
          </div>
        </div>

        {/* 表格区域 — 玻璃态卡片 - 新设计 */}
        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Table
            dataSource={limits}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={false}
          />
        </div>

        {/* 创建/编辑限额弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: '#fff', fontSize: 18, fontWeight: 600 }}>
              配置限额
            </span>
          }
          open={modalOpen}
          onOk={handleSubmit}
          onCancel={() => setModalOpen(false)}
          confirmLoading={submitting}
          width={520}
          okButtonProps={{
            style: {
              background: 'linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)',
              border: 'none',
              boxShadow: '0 4px 16px rgba(255, 190, 11, 0.25)',
            },
          }}
        >
          <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
            <Space style={{ width: '100%' }} size={12}>
              <Form.Item
                name="target_type"
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>目标类型</span>}
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
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>目标 ID</span>}
                rules={[{ required: true, message: '请输入' }]}
                style={{ width: 120 }}
              >
                <InputNumber min={0} style={{ width: '100%' }} placeholder="0=全局" />
              </Form.Item>
              <Form.Item
                name="period"
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>周期</span>}
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
                    label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>自定义周期（小时）</span>}
                    rules={[{ required: true, message: '请输入周期时长' }]}
                  >
                    <InputNumber min={1} style={{ width: '100%' }} placeholder="如 5 表示每 5 小时" />
                  </Form.Item>
                ) : null
              }
            </Form.Item>

            <Form.Item
              name="max_tokens"
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>最大 Token 数</span>}
              rules={[{ required: true, message: '请输入' }]}
            >
              <InputNumber min={0} style={{ width: '100%' }} placeholder="如 1000000" />
            </Form.Item>
            <Form.Item name="max_requests" label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>最大请求数（0=不限制）</span>}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="max_concurrency" label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>最大并发数</span>}>
              <InputNumber min={1} max={100} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="alert_threshold" label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>告警阈值 (%)</span>}>
              <InputNumber min={0} max={100} style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default LimitsPage;
