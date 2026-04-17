import { useEffect, useState, useMemo } from 'react';
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
import useAppStore from '@/store/appStore';
import { useTranslation } from 'react-i18next';
import type { LLMBackend } from '@/types';

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
  0: { label: 'common.disabled', color: 'rgba(255, 255, 255, 0.5)', bg: 'rgba(255, 255, 255, 0.05)' },
  1: { label: 'common.enabled', color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
  2: { label: 'backends.status.draining', color: '#FFBE0B', bg: 'rgba(255, 190, 11, 0.15)' },
};

const FieldLabel = ({ label, tip, isDark }: { label: string; tip: string; isDark: boolean }) => (
  <Space size={4}>
    <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{label}</span>
    <Tooltip title={tip}>
      <QuestionCircleOutlined style={{ color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)', fontSize: 12 }} />
    </Tooltip>
  </Space>
);

const BackendsPage: React.FC = () => {
  const [backends, setBackends] = useState<LLMBackend[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<LLMBackend | null>(null);
  const [form] = Form.useForm();
  const { t } = useTranslation();
  
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  useEffect(() => { loadData(); }, []);

  const loadData = async () => {
    setLoading(true);
    try {
      const res = await listLLMBackends();
      setBackends(res.data.data || []);
    } catch { /* handled by interceptor */ }
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
        message.success(t('backends.nodeUpdated'));
      } else {
        await createLLMBackend(values);
        message.success(t('backends.nodeCreated'));
      }
      setModalOpen(false);
      loadData();
    } catch { /* validation failed */ }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteLLMBackend(id);
      message.success(t('backends.nodeDeleted'));
      loadData();
    } catch { /* handled by interceptor */ }
  };

  const columns = useMemo(() => [
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.table.node')}</span>,
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: LLMBackend) => (
        <div>
          <div style={{ fontWeight: 600, color: isDark ? '#fff' : '#000', fontSize: 15 }}>
            {record.display_name || name}
          </div>
          <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}>{name}</div>
        </div>
      ),
    },
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.table.serviceUrl')}</span>,
      dataIndex: 'base_url',
      key: 'base_url',
      ellipsis: true,
      render: (url: string) => (
        <span style={{ fontSize: 13, color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)', fontFamily: 'monospace' }}>
          {url}
        </span>
      ),
    },
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.table.protocol')}</span>,
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
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.table.modelPatterns')}</span>,
      dataIndex: 'model_patterns',
      key: 'model_patterns',
      width: 180,
      ellipsis: true,
      render: (p: string) => (
        <Tooltip title={p}>
          <code style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)', fontFamily: 'monospace' }}>{p}</code>
        </Tooltip>
      ),
    },
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.table.weight')}</span>,
      dataIndex: 'weight',
      key: 'weight',
      width: 70,
      align: 'center' as const,
      render: (w: number) => <span style={{ fontWeight: 600, color: isDark ? '#FFBE0B' : '#D48806' }}>{w}</span>,
    },
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.table.concurrency')}</span>,
      dataIndex: 'max_concurrency',
      key: 'max_concurrency',
      width: 70,
      align: 'center' as const,
      render: (v: number) => <span style={{ color: isDark ? '#00F5D4' : '#13C2C2' }}>{v}</span>,
    },
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('common.status')}</span>,
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (s: number) => {
        const info = statusMap[s] ?? { label: 'common.unknown', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', bg: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' };
        return (
          <Tag style={{ 
            color: info.color,
            background: info.bg,
            border: `1px solid ${info.color}40`,
            borderRadius: 6,
          }}>
            {t(info.label)}
          </Tag>
        );
      },
    },
    {
      title: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('common.actions')}</span>,
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
            {t('common.edit')}
          </Button>
          <Popconfirm
            title={t('backends.confirmDelete')}
            description={t('backends.confirmDeleteDesc')}
            onConfirm={() => handleDelete(record.id)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ], [isDark]);

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
              <PageIcon icon={<CloudServerOutlined />} />
              <div>
                <h2 style={{ margin: 0, color: isDark ? '#fff' : '#000', fontSize: 24, fontWeight: 600 }}>
                  {t('backends.pageTitle')}
                </h2>
                <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                  {t('backends.pageDescription')}
                </p>
              </div>
            </div>
            <Space>
              <Button 
                icon={<ReloadOutlined />} 
                onClick={loadData}
                style={{
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                  color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)',
                }}
              >
                {t('common.refresh')}
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
                {t('backends.create')}
              </Button>
            </Space>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Table
            columns={columns}
            dataSource={backends}
            rowKey="id"
            loading={loading}
            pagination={false}
            size="middle"
            locale={{ 
              emptyText: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{t('backends.emptyText')}</span> 
            }}
          />
        </div>

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : '#000', fontSize: 18, fontWeight: 600 }}>
              <CloudServerOutlined style={{ marginRight: 8 }} />
              {editing ? t('backends.modal.editTitle', { name: editing.display_name || editing.name }) : t('backends.modal.createTitle')}
            </span>
          }
          open={modalOpen}
          onOk={handleSubmit}
          onCancel={() => setModalOpen(false)}
          okText={editing ? t('common.save') : t('common.create')}
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

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="name"
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.form.nodeId')}</span>}
                rules={[{ required: true, message: t('backends.form.nodeIdRequired') }]}
              >
                <Input 
                  placeholder={t('backends.form.nodeIdPlaceholder')} 
                  disabled={!!editing}
                  style={{ 
                    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                    borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                    color: isDark ? '#fff' : '#000',
                  }}
                />
              </Form.Item>
              <Form.Item 
                name="display_name" 
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.form.displayName')}</span>}
              >
                <Input 
                  placeholder={t('backends.form.displayNamePlaceholder')}
                  style={{ 
                    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                    borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                    color: isDark ? '#fff' : '#000',
                  }}
                />
              </Form.Item>
            </div>

            <Form.Item
              name="base_url"
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.form.serviceUrl')}</span>}
              rules={[{ required: true, message: t('backends.form.serviceUrlRequired') }]}
            >
              <Input 
                placeholder="http://your-llm-server:8000/v1" 
                style={{ fontFamily: 'monospace', color: isDark ? '#00D9FF' : '#1890FF' }}
              />
            </Form.Item>

            <Form.Item 
              name="api_key" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>API Key</span>}
            >
              <Input.Password 
                placeholder={editing ? t('backends.form.apiKeyEditHint') : t('backends.form.apiKeyNoAuthHint')}
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                  color: isDark ? '#fff' : '#000',
                }}
              />
            </Form.Item>

            <Divider style={{ margin: '4px 0 16px', borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' }} />

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="format"
                label={<FieldLabel label={t('backends.form.protocolFormat')} tip={t('backends.form.protocolFormatTip')} isDark={isDark} />}
                rules={[{ required: true }]}
              >
                <Select 
                  options={[
                    { value: 'openai', label: t('backends.form.openaiCompat') },
                    { value: 'anthropic', label: t('backends.form.anthropicNative') },
                  ]}
                  style={{ 
                    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  }}
                />
              </Form.Item>
              <Form.Item 
                name="status" 
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('backends.form.runStatus')}</span>}
              >
                <Select 
                  options={[
                    { value: 1, label: t('common.enabled') },
                    { value: 0, label: t('common.disabled') },
                    { value: 2, label: t('backends.form.draining') },
                  ]}
                  style={{ 
                    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  }}
                />
              </Form.Item>
            </div>

            <Divider style={{ margin: '4px 0 16px', borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' }} />

            <Form.Item
              name="model_patterns"
              label={<FieldLabel label={t('backends.form.modelPatterns')} tip={t('backends.form.modelPatternsTip')} isDark={isDark} />}
            >
              <Input 
                placeholder={t('backends.form.modelPatternsPlaceholder')} 
                style={{ 
                  fontFamily: 'monospace', 
                  color: isDark ? '#fff' : '#000',
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                }}
              />
            </Form.Item>

            <Divider style={{ margin: '4px 0 16px', borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' }} />

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="weight"
                label={<FieldLabel label={t('backends.form.loadWeight')} tip={t('backends.form.loadWeightTip')} isDark={isDark} />}
              >
                <InputNumber 
                  min={1} 
                  max={10000} 
                  style={{ width: '100%' }}
                />
              </Form.Item>
              <Form.Item
                name="max_concurrency"
                label={<FieldLabel label={t('backends.form.maxConcurrency')} tip={t('backends.form.maxConcurrencyTip')} isDark={isDark} />}
              >
                <InputNumber 
                  min={1} 
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </div>

            <Divider style={{ margin: '4px 0 16px', borderColor: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)' }} />

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
              <Form.Item
                name="timeout_seconds"
                label={<FieldLabel label={t('backends.form.normalTimeout')} tip={t('backends.form.normalTimeoutTip')} isDark={isDark} />}
              >
                <InputNumber 
                  min={10} 
                  max={3600} 
                  style={{ width: '100%' }}
                />
              </Form.Item>
              <Form.Item
                name="stream_timeout_seconds"
                label={<FieldLabel label={t('backends.form.streamTimeout')} tip={t('backends.form.streamTimeoutTip')} isDark={isDark} />}
              >
                <InputNumber 
                  min={10} 
                  max={7200} 
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </div>

            <Form.Item
              name="health_check_url"
              label={<FieldLabel label={t('backends.form.healthCheckUrl')} tip={t('backends.form.healthCheckUrlTip')} isDark={isDark} />}
            >
              <Input 
                placeholder="http://your-llm-server:8000/health" 
                style={{ 
                  fontFamily: 'monospace', 
                  color: isDark ? '#fff' : '#000',
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                }}
              />
            </Form.Item>

          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default BackendsPage;
