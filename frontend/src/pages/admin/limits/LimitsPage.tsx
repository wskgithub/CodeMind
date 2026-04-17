import { useEffect, useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Table, Button, Modal, Form, InputNumber, Select, message, Popconfirm, Tag, Space } from 'antd';
import { PlusOutlined, DeleteOutlined, SafetyOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { listLimits, upsertLimit, deleteLimit } from '@/services/limitService';
import useAppStore from '@/store/appStore';
import type { RateLimit } from '@/types';

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

const LimitsPage = () => {
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
  const { t } = useTranslation();
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
      // handled by interceptor
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
      message.success(t('limits.configSaved'));
      setModalOpen(false);
      loadLimits();
    } catch {
      // validation or request failed
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteLimit(id);
      message.success(t('limits.deleted'));
      loadLimits();
    } catch {
      // handled by interceptor
    }
  };

  const targetTypeLabel: Record<string, { text: string; color: string; bg: string }> = {
    global: { text: t('limits.filter.global'), color: '#FF6B6B', bg: isDark ? 'rgba(255, 107, 107, 0.15)' : 'rgba(255, 107, 107, 0.1)' },
    department: { text: t('limits.filter.department'), color: '#00D9FF', bg: isDark ? 'rgba(0, 217, 255, 0.15)' : 'rgba(0, 217, 255, 0.1)' },
    user: { text: t('limits.filter.user'), color: '#00F5D4', bg: isDark ? 'rgba(0, 245, 212, 0.15)' : 'rgba(0, 245, 212, 0.1)' },
  };

  const periodLabel: Record<string, string> = {
    daily: t('limits.period.daily'),
    weekly: t('limits.period.weekly'),
    monthly: t('limits.period.monthly'),
    custom: t('limits.period.custom'),
  };

  const formatPeriodHours = (hours: number): string => {
    if (hours === 24) return t('limits.formatHoursDaily');
    if (hours === 168) return t('limits.formatHoursWeekly');
    if (hours === 720) return t('limits.formatHoursMonthly');
    return t('limits.formatHours', { count: hours });
  };

  const columns: ColumnsType<RateLimit> = useMemo(() => [
    {
      title: t('limits.table.targetType'),
      dataIndex: 'target_type',
      key: 'target_type',
      width: 100,
      render: (val: string) => {
        const label = targetTypeLabel[val];
        return label ? (
          <Tag style={{ 
            color: label.color,
            background: label.bg,
            border: `1px solid ${label.color}40`,
            borderRadius: 6,
          }}>
            {label.text}
          </Tag>
        ) : val;
      },
    },
    { 
      title: t('limits.table.targetId'), 
      dataIndex: 'target_id', 
      key: 'target_id', 
      width: 80,
      render: (v) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{v}</span>,
    },
    {
      title: t('limits.table.period'),
      dataIndex: 'period',
      key: 'period',
      width: 80,
      render: (val: string) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)' }}>{periodLabel[val] || val}</span>,
    },
    {
      title: t('limits.table.periodHours'),
      dataIndex: 'period_hours',
      key: 'period_hours',
      width: 140,
      render: (v: number) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)' }}>{formatPeriodHours(v)}</span>,
    },
    {
      title: t('limits.table.maxTokens'),
      dataIndex: 'max_tokens',
      key: 'max_tokens',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00D9FF', fontWeight: 500 }}>{v.toLocaleString()}</span>,
    },
    {
      title: t('limits.table.maxRequests'),
      dataIndex: 'max_requests',
      key: 'max_requests',
      align: 'right',
      render: (v: number) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{v === 0 ? t('limits.table.noLimit') : v.toLocaleString()}</span>,
    },
    {
      title: t('limits.table.maxConcurrency'),
      dataIndex: 'max_concurrency',
      key: 'max_concurrency',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00F5D4' }}>{v}</span>,
    },
    {
      title: t('limits.table.alertThreshold'),
      dataIndex: 'alert_threshold',
      key: 'alert_threshold',
      align: 'right',
      render: (v: number) => <span style={{ color: '#FFBE0B' }}>{v}%</span>,
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: number) =>
        v === 1 ? (
          <Tag style={{ 
            color: '#00F5D4', 
            background: isDark ? 'rgba(0, 245, 212, 0.15)' : 'rgba(0, 245, 212, 0.1)',
            border: '1px solid rgba(0, 245, 212, 0.3)',
            borderRadius: 6,
          }}>
            {t('common.enabled')}
          </Tag>
        ) : (
          <Tag style={{ 
            color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', 
            background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
            border: isDark ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.1)',
            borderRadius: 6,
          }}>
            {t('common.disabled')}
          </Tag>
        ),
    },
    {
      title: t('common.actions'),
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Popconfirm 
          title={t('limits.confirmDelete')} 
          onConfirm={() => handleDelete(record.id)}
        >
          <Button type="link" danger icon={<DeleteOutlined />} size="small">
            {t('common.delete')}
          </Button>
        </Popconfirm>
      ),
    },
  ], [isDark, targetTypeLabel, periodLabel, t]);

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<SafetyOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                {t('limits.title')}
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                {t('limits.pageDescription')}
              </p>
            </div>
          </div>
          <div style={{ marginTop: 20, display: 'flex', gap: 12 }}>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={loadLimits}
              style={{
                background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)',
              }}
            >
              {t('common.refresh')}
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
              {t('limits.createLimit')}
            </Button>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Table
            dataSource={limits}
            columns={columns}
            rowKey="id"
            loading={loading}
            pagination={false}
          />
        </div>

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {t('limits.configureLimit')}
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
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.form.targetType')}</span>}
                rules={[{ required: true, message: t('limits.form.selectRequired') }]}
                style={{ width: 150 }}
              >
                <Select
                  options={[
                    { label: t('limits.filter.global'), value: 'global' },
                    { label: t('limits.filter.department'), value: 'department' },
                    { label: t('limits.filter.user'), value: 'user' },
                  ]}
                />
              </Form.Item>
              <Form.Item
                name="target_id"
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.table.targetId')}</span>}
                rules={[{ required: true, message: t('limits.form.inputRequired') }]}
                style={{ width: 120 }}
              >
                <InputNumber min={0} style={{ width: '100%' }} placeholder={`0=${t('limits.filter.global')}`} />
              </Form.Item>
              <Form.Item
                name="period"
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.form.period')}</span>}
                rules={[{ required: true, message: t('limits.form.selectRequired') }]}
                style={{ width: 120 }}
              >
                <Select
                  options={[
                    { label: t('limits.period.daily'), value: 'daily' },
                    { label: t('limits.period.weekly'), value: 'weekly' },
                    { label: t('limits.period.monthly'), value: 'monthly' },
                    { label: t('limits.period.custom'), value: 'custom' },
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
                    label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.period.customHours')}</span>}
                    rules={[{ required: true, message: t('limits.period.inputPeriodHours') }]}
                  >
                    <InputNumber min={1} style={{ width: '100%' }} placeholder={t('limits.period.inputPeriodHours')} />
                  </Form.Item>
                ) : null
              }
            </Form.Item>

            <Form.Item
              name="max_tokens"
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.table.maxTokens')}</span>}
              rules={[{ required: true, message: t('limits.form.inputRequired') }]}
            >
              <InputNumber min={0} style={{ width: '100%' }} placeholder="1000000" />
            </Form.Item>
            <Form.Item name="max_requests" label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.form.maxRequestsLabel')}</span>}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="max_concurrency" label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.form.maxConcurrencyLabel')}</span>}>
              <InputNumber min={1} max={100} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="alert_threshold" label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)' }}>{t('limits.form.alertThresholdLabel')}</span>}>
              <InputNumber min={0} max={100} style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default LimitsPage;
