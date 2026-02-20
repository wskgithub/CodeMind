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

const { TextArea } = Input;

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
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

/** 系统管理页面 — 与首页/登录页新设计风格统一 */
const SystemPage = () => {
  const { themeMode } = useAppStore();
  const isDark = themeMode === 'dark';
  
  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<ControlOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : '#1a1a1a', fontSize: 24, fontWeight: 600 }}>
                系统管理
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                系统配置、公告管理与审计日志查看
              </p>
            </div>
          </div>
        </div>

        {/* 标签页 - 新设计 */}
        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Tabs
            items={[
              {
                key: 'configs',
                label: (
                  <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>
                    <SettingOutlined /> 系统配置
                  </span>
                ),
                children: <ConfigsTab />,
              },
              {
                key: 'announcements',
                label: (
                  <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>
                    <NotificationOutlined /> 公告管理
                  </span>
                ),
                children: <AnnouncementsTab />,
              },
              {
                key: 'audit',
                label: (
                  <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>
                    <AuditOutlined /> 审计日志
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

/** 系统配置标签页 - 新设计 */
const ConfigsTab = () => {
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();
  const { themeMode } = useAppStore();
  const isDark = themeMode === 'dark';

  const configLabels: Record<string, { label: string; description?: string }> = {
    'system.default_concurrency': { label: '默认并发数', description: '用户默认的最大并发请求数' },
    'system.force_change_password': { label: '强制修改密码', description: '用户首次登录是否强制修改密码（true/false）' },
    'system.max_keys_per_user': { label: '每用户最大密钥数', description: '每个用户可创建的最大 API Key 数量' },
    'system.site_name': { label: '站点名称', description: '系统显示的站点名称' },
    'system.site_logo': { label: '站点 Logo', description: '站点 Logo 的 URL' },
    'system.contact_email': { label: '联系邮箱', description: '系统管理员联系邮箱' },
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
      // 拦截器处理
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
      message.success('配置已保存');
      loadConfigs();
    } catch {
      // 拦截器处理
    } finally {
      setSaving(false);
    }
  };

  return (
    <div>
      <Form form={form} layout="vertical" style={{ maxWidth: 800 }}>
        {loading ? (
          <div style={{ padding: 40, textAlign: 'center', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>加载中...</div>
        ) : configs.length === 0 ? (
          <div style={{ padding: 40, textAlign: 'center', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>暂无配置项</div>
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
                    placeholder={`当前值: ${c.config_value}`}
                    options={[
                      { label: '是', value: 'true' },
                      { label: '否', value: 'false' },
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
                    placeholder={`当前值: ${c.config_value}`}
                    style={{ 
                      background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#fff',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.2)',
                      color: isDark ? '#fff' : '#1a1a1a',
                    }}
                  />
                ) : (
                  <Input 
                    placeholder={`当前值: ${c.config_value}`}
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
              保存配置
            </Button>
          </Form.Item>
        )}
      </Form>
    </div>
  );
};

/** 公告管理标签页 - 新设计 */
const AnnouncementsTab = () => {
  const [list, setList] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();
  const { themeMode } = useAppStore();
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
      // 拦截器处理
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
        message.success('公告已更新');
      } else {
        await createAnnouncement(values);
        message.success('公告已创建');
      }
      setModalOpen(false);
      loadList();
    } catch {
      // 验证或请求失败
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteAnnouncement(id);
      message.success('已删除');
      loadList();
    } catch {
      // 拦截器处理
    }
  };

  // 表格列 - 新设计
  const columns: ColumnsType<Announcement> = [
    {
      title: '标题',
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
              置顶
            </Tag>
          )}
          <span style={{ color: isDark ? '#fff' : '#1a1a1a' }}>{val}</span>
        </Space>
      ),
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
            已发布
          </Tag>
        ) : (
          <Tag style={{ 
            color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', 
            background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)',
            border: isDark ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.1)',
            borderRadius: 6,
          }}>
            草稿
          </Tag>
        ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (v: string) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>{v?.slice(0, 10)}</span>,
    },
    {
      title: '操作',
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
            编辑
          </Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" danger icon={<DeleteOutlined />} size="small">
              删除
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
          发布公告
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
            {editingId ? '编辑公告' : '发布公告'}
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
            label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>标题</span>} 
            rules={[{ required: true, message: '请输入标题' }]}
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
            label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>内容</span>} 
            rules={[{ required: true, message: '请输入内容' }]}
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>置顶</span>} 
              valuePropName="checked"
            >
              <Switch 
                checkedChildren="是" 
                unCheckedChildren="否"
                style={{ backgroundColor: isDark ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.2)' }}
              />
            </Form.Item>
            <Form.Item 
              name="status" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>发布状态</span>}
            >
              <Select
                options={[
                  { label: '草稿', value: 0 },
                  { label: '已发布', value: 1 },
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

/** 审计日志标签页 - 新设计 */
const AuditLogsTab = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const { themeMode } = useAppStore();
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
      // 拦截器处理
    } finally {
      setLoading(false);
    }
  };

  const actionLabels: Record<string, string> = {
    create_user: '创建用户',
    update_user: '更新用户',
    delete_user: '删除用户',
    disable_user: '禁用用户',
    enable_user: '启用用户',
    reset_password: '重置密码',
    create_department: '创建部门',
    update_department: '更新部门',
    delete_department: '删除部门',
    create_api_key: '创建 API Key',
    delete_api_key: '删除 API Key',
    disable_api_key: '禁用 API Key',
    enable_api_key: '启用 API Key',
    update_limit: '更新限额',
    delete_limit: '删除限额',
    update_config: '更新配置',
    create_announcement: '创建公告',
    update_announcement: '更新公告',
    delete_announcement: '删除公告',
  };

  // 表格列 - 新设计
  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v: string) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)', fontFamily: 'monospace' }}>{v?.replace('T', ' ').slice(0, 19)}</span>,
    },
    {
      title: '操作人',
      dataIndex: 'operator',
      key: 'operator',
      width: 120,
      render: (op: AuditLog['operator']) => <span style={{ color: isDark ? '#fff' : '#1a1a1a' }}>{op?.display_name || op?.username || '-'}</span>,
    },
    {
      title: '操作',
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
      title: '目标类型',
      dataIndex: 'target_type',
      key: 'target_type',
      width: 100,
      render: (v) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)' }}>{v}</span>,
    },
    {
      title: '目标 ID',
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
      title: '详情',
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
        showTotal: (t) => `共 ${t} 条`,
      }}
      size="small"
    />
  );
};

export default SystemPage;
