import { PlusOutlined, CopyOutlined, DeleteOutlined, StopOutlined, CheckCircleOutlined, KeyOutlined } from '@ant-design/icons';
import { Table, Button, Modal, Form, Input, Space, Tag, message, Typography, DatePicker } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';

import keyService from '@/services/keyService';
import useAppStore from '@/store/appStore';
import type { APIKey } from '@/types';
import { copyToClipboard } from '@/utils/copy';

const { Paragraph } = Typography;

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

const maskKey = (key: string | null): string => {
  if (!key) return '';
  const visible = key.slice(0, 12);
  return visible + '************************';
};

const KeysPage: React.FC = () => {
  const { t } = useTranslation();
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [form] = Form.useForm();

  const loadKeys = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await keyService.list();
      setKeys(resp.data.data || []);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadKeys();
  }, [loadKeys]);

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
      message.success(t('success.created'));
      loadKeys();
    } catch {
      // handled by interceptor
    }
  };

  const handleCopy = async (record: APIKey) => {
    try {
      const resp = await keyService.copy(record.id);
      const fullKey = resp.data.data?.key;
      if (fullKey) {
        const ok = await copyToClipboard(fullKey);
        if (ok) {
          message.success(t('success.copied'));
        } else {
          message.error(t('error.copyFailed'));
        }
      }
    } catch {
      // handled by interceptor
    }
  };

  const handleToggleStatus = async (record: APIKey) => {
    const newStatus = record.status === 1 ? 0 : 1;
    await keyService.updateStatus(record.id, newStatus);
    message.success(newStatus === 1 ? t('success.enabled') : t('success.disabled'));
    loadKeys();
  };

  const handleDelete = (record: APIKey) => {
    Modal.confirm({
      title: t('keys.deleteConfirm.title'),
      content: t('keys.deleteConfirm.content', { name: record.name }),
      okText: t('common.delete'),
      okType: 'danger',
      okButtonProps: {
        style: { background: '#FF6B6B', borderColor: '#FF6B6B' },
      },
      onOk: async () => {
        await keyService.delete(record.id);
        message.success(t('success.deleted'));
        loadKeys();
      },
    });
  };

  const columns: ColumnsType<APIKey> = useMemo(() => [
    { 
      title: t('keys.table.name'), 
      dataIndex: 'name', 
      key: 'name',
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>{text}</span>,
    },
    {
      title: t('keys.table.keyPrefix'),
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
      title: t('keys.table.status'),
      dataIndex: 'status',
      key: 'status',
      render: (v: number) =>
        v === 1 ? (
          <Tag color="success" style={{ borderRadius: 6, border: 'none' }}>{t('common.enabled')}</Tag>
        ) : (
          <Tag color="error" style={{ borderRadius: 6, border: 'none' }}>{t('common.disabled')}</Tag>
        ),
    },
    {
      title: t('keys.table.lastUsedAt'),
      dataIndex: 'last_used_at',
      key: 'last_used_at',
      render: (v: string) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)' }}>
          {v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'}
        </span>
      ),
    },
    {
      title: t('keys.table.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)' }}>
          {dayjs(v).format('YYYY-MM-DD HH:mm')}
        </span>
      ),
    },
    {
      title: t('keys.table.actions'),
      key: 'action',
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<CopyOutlined />}
            onClick={() => handleCopy(record)}
            style={{ color: '#00D9FF' }}
          >
            {t('keys.copy')}
          </Button>
          <Button
            type="link"
            size="small"
            icon={record.status === 1 ? <StopOutlined /> : <CheckCircleOutlined />}
            onClick={() => handleToggleStatus(record)}
            style={{ color: record.status === 1 ? '#FFBE0B' : '#00F5D4' }}
          >
            {record.status === 1 ? t('common.disable') : t('common.enable')}
          </Button>
          <Button
            type="link"
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            {t('common.delete')}
          </Button>
        </Space>
      ),
    },
  ], [isDark, t]);

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<KeyOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                {t('keys.title')}
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                {t('keys.description')}
              </p>
            </div>
          </div>
        </div>

        <div
          className="glass-card animate-fade-in-up"
          style={{ padding: 24, animationDelay: '0.05s' }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
            <span style={{ fontWeight: 600, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 16 }}>{t('keys.keyList')}</span>
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
              {t('keys.create')}
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

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {t('keys.create')}
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.7)' }}>{t('keys.form.name')}</span>} 
              rules={[{ required: true, message: t('keys.form.nameRequired') }]}
            >
              <Input 
                placeholder={t('keys.form.namePlaceholder')} 
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
                  color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="expires_at" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.7)' }}>{t('keys.form.expiresAt')}</span>}
            >
              <DatePicker 
                style={{ width: '100%' }} 
                placeholder={t('keys.form.expiresAtPlaceholder')}
                suffixIcon={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>📅</span>}
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
                {t('keys.form.submit')}
              </Button>
            </Form.Item>
          </Form>
        </Modal>

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {t('keys.createSuccess.title')}
            </span>
          }
          open={!!newKey}
          onCancel={() => setNewKey(null)}
          onOk={() => setNewKey(null)}
          okText={t('common.ok')}
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
            ⚠️ {t('keys.createSuccess.warning')}
          </p>
          <Paragraph
            copyable={{ text: newKey || '', icon: <CopyOutlined style={{ color: '#00D9FF' }} />, tooltips: [t('common.copy'), t('common.copied')] }}
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
            {maskKey(newKey)}
          </Paragraph>
        </Modal>
      </div>
    </div>
  );
};

export default KeysPage;
