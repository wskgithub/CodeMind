import { useEffect, useState } from 'react';
import { Tabs, Table, Button, Form, Input, Select, message, Modal, Switch, Popconfirm, Tag, Space } from 'antd';
import useAppStore from '@/store/appStore';
import {
  SettingOutlined,
  NotificationOutlined,
  AuditOutlined,
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ControlOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import {
  getConfigs,
  updateConfigs,
  listAnnouncements,
  createAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  listAuditLogs,
} from '@/services/systemService';
import type { SystemConfig, Announcement, AuditLog } from '@/types';
import { useTranslation } from 'react-i18next';

const { TextArea } = Input;

const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
    }}
  >
    {icon}
  </span>
);

const SystemPage = () => {
  const { t } = useTranslation();
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
  
  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<ControlOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : '#1a1a1a', fontSize: 24, fontWeight: 600 }}>
                {t('system.title')}
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                {t('system.pageDescription')}
              </p>
            </div>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Tabs
            items={[
              {
                key: 'configs',
                label: (
                  <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>
                    <SettingOutlined /> {t('system.tabs.config')}
                  </span>
                ),
                children: <ConfigsTab />,
              },
              {
                key: 'announcements',
                label: (
                  <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>
                    <NotificationOutlined /> {t('system.tabs.announcement')}
                  </span>
                ),
                children: <AnnouncementsTab />,
              },
              {
                key: 'audit',
                label: (
                  <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>
                    <AuditOutlined /> {t('system.tabs.audit')}
                  </span>
                ),
                children: <AuditLogsTab />,
              },
            ]}
          />
        </div>
      </div>
    </div>
  );
};

const ConfigsTab = () => {
  const { t } = useTranslation();
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  const configLabels: Record<string, { label: string; description?: string }> = {
    'system.default_concurrency': { label: t('system.config.defaultConcurrency'), description: t('system.config.defaultConcurrencyDesc') },
    'system.force_change_password': { label: t('system.config.forceChangePassword'), description: t('system.config.forceChangePasswordDesc') },
    'system.max_keys_per_user': { label: t('system.config.maxKeysPerUser'), description: t('system.config.maxKeysPerUserDesc') },
    'system.site_name': { label: t('system.config.siteName'), description: t('system.config.siteNameDesc') },
    'system.site_logo': { label: t('system.config.siteLogo'), description: t('system.config.siteLogoDesc') },
    'system.contact_email': { label: t('system.config.contactEmail'), description: t('system.config.contactEmailDesc') },
  };

  useEffect(() => {
    loadConfigs();
  }, []);

  const loadConfigs = async () => {
    setLoading(true);
    try {
      const res = await getConfigs();
      const data = (res.data.data || []).filter(
        (c) => !c.config_key.startsWith('llm.')
      );
      setConfigs(data);
      const values: Record<string, string> = {};
      data.forEach((c) => {
        values[c.config_key] = c.config_value;
      });
      form.setFieldsValue(values);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const values = form.getFieldsValue();
      const items = Object.entries(values)
        .filter(([, v]) => v !== undefined && v !== '')
        .map(([key, value]) => ({ key, value: String(value) }));

      await updateConfigs(items);
      message.success(t('system.config.saved'));
      loadConfigs();
    } catch {
      // handled by interceptor
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <Form form={form} layout="vertical" style={{ maxWidth: 800 }}>
        {loading ? (
          <div style={{ padding: 40, textAlign: 'center', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{t('common.loading')}</div>
        ) : configs.length === 0 ? (
          <div style={{ padding: 40, textAlign: 'center', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{t('system.config.noItems')}</div>
        ) : (
          configs.map((c) => {
            const config = configLabels[c.config_key] || { label: c.config_key };
            return (
              <Form.Item
                key={c.config_key}
                name={c.config_key}
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.9)', fontWeight: 500 }}>{config.label}</span>}
                extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}>{config.description || c.description}</span>}
              >
                {c.config_key === 'system.force_change_password' ? (
                  <Select
                    placeholder={t('system.config.currentValue', { value: c.config_value })}
                    options={[
                      { label: t('common.yes'), value: 'true' },
                      { label: t('common.no'), value: 'false' },
                    ]}
                    style={{
                      background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.2)',
                    }}
                    dropdownStyle={{
                      background: isDark ? 'rgba(30, 30, 40, 0.95)' : '#fff',
                    }}
                  />
                ) : c.config_key === 'llm.models' ? (
                  <TextArea
                    rows={3}
                    placeholder={t('system.config.currentValue', { value: c.config_value })}
                    style={{ 
                      background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.2)',
                      color: isDark ? '#fff' : '#1a1a1a',
                    }}
                  />
                ) : (
                  <Input 
                    placeholder={t('system.config.currentValue', { value: c.config_value })}
                    style={{ 
                      background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.2)',
                      color: isDark ? '#fff' : '#1a1a1a',
                    }}
                  />
                )}
              </Form.Item>
            );
          })
        )}
        {configs.length > 0 && (
          <Form.Item>
            <Button 
              type="primary" 
              onClick={handleSave} 
              loading={saving}
              style={{
                background: 'linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)',
                border: 'none',
                boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                height: 44,
                borderRadius: 12,
                padding: '0 32px',
              }}
            >
              {t('system.config.save')}
            </Button>
          </Form.Item>
        )}
      </Form>
    </div>
  );
};

const AnnouncementsTab = () => {
  const { t } = useTranslation();
  const [list, setList] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  useEffect(() => {
    loadList();
  }, []);

  const loadList = async () => {
    setLoading(true);
    try {
      const res = await listAnnouncements();
      setList(res.data.data || []);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    form.resetFields();
    form.setFieldsValue({ status: 1, pinned: false });
    setEditingId(null);
    setModalOpen(true);
  };

  const handleEdit = (record: Announcement) => {
    form.setFieldsValue({
      title: record.title,
      content: record.content,
      pinned: record.pinned,
      status: record.status,
    });
    setEditingId(record.id);
    setModalOpen(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      if (editingId) {
        await updateAnnouncement(editingId, values);
        message.success(t('system.announcement.updated'));
      } else {
        await createAnnouncement(values);
        message.success(t('system.announcement.created'));
      }
      setModalOpen(false);
      loadList();
    } catch {
      // validation or request failed
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteAnnouncement(id);
      message.success(t('success.deleted'));
      loadList();
    } catch {
      // handled by interceptor
    }
  };

  const columns: ColumnsType<Announcement> = [
    {
      title: t('system.announcement.title'),
      dataIndex: 'title',
      key: 'title',
      render: (val: string, record) => (
        <Space>
          {record.pinned && (
            <Tag style={{ 
              color: '#FF6B6B', 
              background: 'rgba(255, 107, 107, 0.15)',
              border: '1px solid rgba(255, 107, 107, 0.3)',
              borderRadius: 6,
            }}>
              {t('system.announcement.pinned')}
            </Tag>
          )}
          <span style={{ color: isDark ? '#fff' : '#1a1a1a' }}>{val}</span>
        </Space>
      ),
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
            background: 'rgba(0, 245, 212, 0.15)',
            border: '1px solid rgba(0, 245, 212, 0.3)',
            borderRadius: 6,
          }}>
            {t('system.announcement.published')}
          </Tag>
        ) : (
          <Tag style={{ 
            color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', 
            background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)',
            border: isDark ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.1)',
            borderRadius: 6,
          }}>
            {t('system.announcement.draft')}
          </Tag>
        ),
    },
    {
      title: t('common.createdAt'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (v: string) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{v?.slice(0, 10)}</span>,
    },
    {
      title: t('common.actions'),
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button 
            type="link" 
            icon={<EditOutlined />} 
            size="small" 
            onClick={() => handleEdit(record)}
            style={{ color: '#00D9FF' }}
          >
            {t('common.edit')}
          </Button>
          <Popconfirm title={t('system.announcement.confirmDelete')} onConfirm={() => handleDelete(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button 
          type="primary" 
          icon={<PlusOutlined />} 
          onClick={handleCreate}
          style={{
            background: 'linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)',
            border: 'none',
            boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
          }}
        >
          {t('system.announcement.publish')}
        </Button>
      </div>
      <Table 
        dataSource={list} 
        columns={columns} 
        rowKey="id" 
        loading={loading} 
        pagination={false} 
      />

      <Modal
        title={
          <span style={{ color: isDark ? '#fff' : '#1a1a1a', fontSize: 18, fontWeight: 600 }}>
            {editingId ? t('system.announcement.editTitle') : t('system.announcement.publish')}
          </span>
        }
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={submitting}
        width={600}
        okButtonProps={{
          style: {
            background: 'linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)',
            border: 'none',
            boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
          },
        }}
        style={{
          '--modal-bg': isDark ? 'rgba(30, 30, 40, 0.95)' : '#fff',
        } as React.CSSProperties}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item 
            name="title" 
            label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('system.announcement.title')}</span>} 
            rules={[{ required: true, message: t('system.announcement.titleRequired') }]}
          >
            <Input 
              maxLength={200}
              style={{ 
                background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.2)',
                color: isDark ? '#fff' : '#1a1a1a',
              }}
            />
          </Form.Item>
          <Form.Item 
            name="content" 
            label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('system.announcement.content')}</span>} 
            rules={[{ required: true, message: t('system.announcement.contentRequired') }]}
          >
            <TextArea 
              rows={6}
              style={{ 
                background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.2)',
                color: isDark ? '#fff' : '#1a1a1a',
              }}
            />
          </Form.Item>
          <Space>
            <Form.Item 
              name="pinned" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('system.announcement.pinned')}</span>} 
              valuePropName="checked"
            >
              <Switch 
                checkedChildren={t('common.yes')} 
                unCheckedChildren={t('common.no')}
                style={{ backgroundColor: isDark ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.2)' }}
              />
            </Form.Item>
            <Form.Item 
              name="status" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('system.announcement.publishStatus')}</span>}
            >
              <Select
                options={[
                  { label: t('system.announcement.draft'), value: 0 },
                  { label: t('system.announcement.published'), value: 1 },
                ]}
                style={{ width: 120, background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff' }}
                dropdownStyle={{
                  background: isDark ? 'rgba(30, 30, 40, 0.95)' : '#fff',
                }}
              />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  );
};

const AuditLogsTab = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  useEffect(() => {
    loadLogs();
  }, [page]);

  const loadLogs = async () => {
    setLoading(true);
    try {
      const res = await listAuditLogs({ page, page_size: 20 });
      const data = res.data.data;
      setLogs(data?.list || []);
      setTotal(data?.pagination?.total || 0);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  const actionLabels: Record<string, string> = {
    create_user: t('system.audit.actions.create_user'),
    update_user: t('system.audit.actions.update_user'),
    delete_user: t('system.audit.actions.delete_user'),
    disable_user: t('system.audit.actions.disable_user'),
    enable_user: t('system.audit.actions.enable_user'),
    reset_password: t('system.audit.actions.reset_password'),
    create_department: t('system.audit.actions.create_department'),
    update_department: t('system.audit.actions.update_department'),
    delete_department: t('system.audit.actions.delete_department'),
    create_api_key: t('system.audit.actions.create_api_key'),
    delete_api_key: t('system.audit.actions.delete_api_key'),
    disable_api_key: t('system.audit.actions.disable_api_key'),
    enable_api_key: t('system.audit.actions.enable_api_key'),
    update_limit: t('system.audit.actions.update_limit'),
    delete_limit: t('system.audit.actions.delete_limit'),
    update_config: t('system.audit.actions.update_config'),
    create_announcement: t('system.audit.actions.create_announcement'),
    update_announcement: t('system.audit.actions.update_announcement'),
    delete_announcement: t('system.audit.actions.delete_announcement'),
  };

  const columns: ColumnsType<AuditLog> = [
    {
      title: t('system.audit.time'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v: string) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)', fontFamily: 'monospace' }}>{v?.replace('T', ' ').slice(0, 19)}</span>,
    },
    {
      title: t('system.audit.operator'),
      dataIndex: 'operator',
      key: 'operator',
      width: 120,
      render: (op: AuditLog['operator']) => <span style={{ color: isDark ? '#fff' : '#1a1a1a' }}>{op?.display_name || op?.username || '-'}</span>,
    },
    {
      title: t('system.audit.action'),
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (v: string) => (
        <Tag style={{ 
          color: '#00D9FF', 
          background: 'rgba(0, 217, 255, 0.15)',
          border: '1px solid rgba(0, 217, 255, 0.3)',
          borderRadius: 6,
        }}>
          {actionLabels[v] || v}
        </Tag>
      ),
    },
    {
      title: t('system.audit.targetType'),
      dataIndex: 'target_type',
      key: 'target_type',
      width: 100,
      render: (v) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)' }}>{v}</span>,
    },
    {
      title: t('system.audit.targetId'),
      dataIndex: 'target_id',
      key: 'target_id',
      width: 80,
      render: (v: number) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{v || '-'}</span>,
    },
    {
      title: 'IP',
      dataIndex: 'client_ip',
      key: 'client_ip',
      width: 130,
      render: (v) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontFamily: 'monospace' }}>{v}</span>,
    },
    {
      title: t('system.audit.detail'),
      dataIndex: 'detail',
      key: 'detail',
      ellipsis: true,
      render: (v: Record<string, unknown>) =>
        v ? <span style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)', fontFamily: 'monospace' }}>{JSON.stringify(v)}</span> : '-',
    },
  ];

  return (
    <Table
      dataSource={logs}
      columns={columns}
      rowKey="id"
      loading={loading}
      pagination={{
        current: page,
        pageSize: 20,
        total,
        onChange: setPage,
        showTotal: (total) => t('common.totalRecords', { total }),
      }}
      size="small"
    />
  );
};

export default SystemPage;
