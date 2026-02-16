import { useEffect, useState } from 'react';
import { Tabs, Table, Button, Form, Input, Select, message, Modal, Switch, Popconfirm, Tag, Space, theme } from 'antd';
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

/** 页面标题图标 — 渐变圆形背景 */
const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-10 h-10 rounded-full shrink-0"
    style={{
      background: 'linear-gradient(135deg, #722ed1 0%, #b37feb 100%)',
      color: '#fff',
    }}
  >
    {icon}
  </span>
);

/** 系统管理页面 — Glassmorphism 风格，包含配置 / 公告 / 审计日志 */
const SystemPage = () => {
  const { token } = theme.useToken();

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
          <PageIcon icon={<ControlOutlined style={{ fontSize: 20 }} />} />
          <h2 style={{ margin: 0, color: token.colorTextHeading }}>系统管理</h2>
        </div>
        <p style={{ margin: 0, color: token.colorTextSecondary, fontSize: 14 }}>
          系统配置、公告管理与审计日志查看。
        </p>
      </div>

      {/* 标签页 */}
      <Tabs
        items={[
          {
            key: 'configs',
            label: (
              <span>
                <SettingOutlined /> 系统配置
              </span>
            ),
            children: <ConfigsTab />,
          },
          {
            key: 'announcements',
            label: (
              <span>
                <NotificationOutlined /> 公告管理
              </span>
            ),
            children: <AnnouncementsTab />,
          },
          {
            key: 'audit',
            label: (
              <span>
                <AuditOutlined /> 审计日志
              </span>
            ),
            children: <AuditLogsTab />,
          },
        ]}
      />
    </div>
  );
};

/** 系统配置标签页 */
const ConfigsTab = () => {
  const { token } = theme.useToken();
  const [configs, setConfigs] = useState<SystemConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  // 配置项中文标题映射
  const configLabels: Record<string, { label: string; description?: string }> = {
    'llm.api_key': { label: 'LLM API 密钥', description: '用于访问 LLM 服务的 API 密钥' },
    'llm.base_url': { label: 'LLM 服务地址', description: 'LLM 服务的基础 URL' },
    'llm.default_model': { label: '默认模型', description: '系统默认使用的 LLM 模型名称' },
    'llm.models': { label: '可用模型列表', description: '支持的模型列表（JSON 数组格式）' },
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
      const data = res.data.data || [];
      setConfigs(data);
      // 将配置项设置到表单
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
          <div style={{ padding: 40, textAlign: 'center', color: token.colorTextTertiary }}>加载中...</div>
        ) : configs.length === 0 ? (
          <div style={{ padding: 40, textAlign: 'center', color: token.colorTextTertiary }}>暂无配置项</div>
        ) : (
          configs.map((c) => {
            const config = configLabels[c.config_key] || { label: c.config_key };
            return (
              <Form.Item
                key={c.config_key}
                name={c.config_key}
                label={config.label}
                extra={config.description || c.description}
              >
                {c.config_key === 'system.force_change_password' ? (
                  <Select
                    placeholder={`当前值: ${c.config_value}`}
                    options={[
                      { label: '是', value: 'true' },
                      { label: '否', value: 'false' },
                    ]}
                  />
                ) : c.config_key === 'llm.models' ? (
                  <TextArea
                    rows={3}
                    placeholder={`当前值: ${c.config_value}`}
                  />
                ) : (
                  <Input placeholder={`当前值: ${c.config_value}`} />
                )}
              </Form.Item>
            );
          })
        )}
        {configs.length > 0 && (
          <Form.Item>
            <Button type="primary" onClick={handleSave} loading={saving}>
              保存配置
            </Button>
          </Form.Item>
        )}
      </Form>
    </div>
  );
};

/** 公告管理标签页 */
const AnnouncementsTab = () => {
  const [list, setList] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();

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

  const columns: ColumnsType<Announcement> = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (val: string, record) => (
        <Space>
          {record.pinned && <Tag color="red">置顶</Tag>}
          {val}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: number) =>
        v === 1 ? <Tag color="green">已发布</Tag> : <Tag color="default">草稿</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 120,
      render: (v: string) => v?.slice(0, 10),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" icon={<EditOutlined />} size="small" onClick={() => handleEdit(record)}>
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
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          发布公告
        </Button>
      </div>
      <Table dataSource={list} columns={columns} rowKey="id" loading={loading} pagination={false} />

      <Modal
        title={editingId ? '编辑公告' : '发布公告'}
        open={modalOpen}
        onOk={handleSubmit}
        onCancel={() => setModalOpen(false)}
        confirmLoading={submitting}
        width={600}
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input maxLength={200} />
          </Form.Item>
          <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入内容' }]}>
            <TextArea rows={6} />
          </Form.Item>
          <Space>
            <Form.Item name="pinned" label="置顶" valuePropName="checked">
              <Switch />
            </Form.Item>
            <Form.Item name="status" label="发布状态">
              <Select
                options={[
                  { label: '草稿', value: 0 },
                  { label: '已发布', value: 1 },
                ]}
                style={{ width: 120 }}
              />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </div>
  );
};

/** 审计日志标签页 */
const AuditLogsTab = () => {
  const { token } = theme.useToken();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);

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

  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (v: string) => v?.replace('T', ' ').slice(0, 19),
    },
    {
      title: '操作人',
      dataIndex: 'operator',
      key: 'operator',
      width: 120,
      render: (op: AuditLog['operator']) => op?.display_name || op?.username || '-',
    },
    {
      title: '操作',
      dataIndex: 'action',
      key: 'action',
      width: 120,
      render: (v: string) => <Tag>{actionLabels[v] || v}</Tag>,
    },
    {
      title: '目标类型',
      dataIndex: 'target_type',
      key: 'target_type',
      width: 100,
    },
    {
      title: '目标 ID',
      dataIndex: 'target_id',
      key: 'target_id',
      width: 80,
      render: (v: number) => v || '-',
    },
    {
      title: 'IP',
      dataIndex: 'client_ip',
      key: 'client_ip',
      width: 130,
    },
    {
      title: '详情',
      dataIndex: 'detail',
      key: 'detail',
      ellipsis: true,
      render: (v: Record<string, unknown>) =>
        v ? <span style={{ fontSize: 12, color: token.colorTextTertiary }}>{JSON.stringify(v)}</span> : '-',
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
