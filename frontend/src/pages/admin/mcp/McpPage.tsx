import { useState, useEffect, useCallback } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Tag, Space,
  Tabs, Popconfirm, Switch, message, theme, Tooltip, Badge,
} from 'antd';
import {
  ApiOutlined,
  PlusOutlined,
  ReloadOutlined,
  SyncOutlined,
  DeleteOutlined,
  EditOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import mcpService from '@/services/mcpService';
import type { MCPService, MCPTool, MCPAccessRule } from '@/types';

/** MCP 服务管理页 — Glassmorphism 风格 */
const McpPage: React.FC = () => {
  const { token: themeToken } = theme.useToken();

  // 服务列表状态
  const [services, setServices] = useState<MCPService[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingService, setEditingService] = useState<MCPService | null>(null);
  const [form] = Form.useForm();

  // 工具列表状态
  const [toolsModalOpen, setToolsModalOpen] = useState(false);
  const [tools, setTools] = useState<MCPTool[]>([]);
  const [toolsLoading, setToolsLoading] = useState(false);

  // 访问规则状态
  const [rules, setRules] = useState<MCPAccessRule[]>([]);
  const [rulesLoading, setRulesLoading] = useState(false);

  // 加载服务列表
  const fetchServices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await mcpService.listServices();
      setServices(res.data.data || []);
    } catch {
      // 错误已由拦截器处理
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载访问规则
  const fetchRules = useCallback(async () => {
    setRulesLoading(true);
    try {
      const res = await mcpService.listAccessRules();
      setRules(res.data.data || []);
    } catch {
      // 错误已由拦截器处理
    } finally {
      setRulesLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchServices();
    fetchRules();
  }, [fetchServices, fetchRules]);

  // 创建/更新服务
  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      if (editingService) {
        await mcpService.updateService(editingService.id, values);
        message.success('服务更新成功');
      } else {
        await mcpService.createService(values);
        message.success('服务创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      setEditingService(null);
      fetchServices();
    } catch {
      // 验证失败
    }
  };

  // 删除服务
  const handleDelete = async (id: number) => {
    try {
      await mcpService.deleteService(id);
      message.success('服务已删除');
      fetchServices();
    } catch {
      // 错误已处理
    }
  };

  // 切换服务状态
  const handleToggleStatus = async (record: MCPService) => {
    const newStatus = record.status === 'enabled' ? 'disabled' : 'enabled';
    try {
      await mcpService.updateService(record.id, { status: newStatus });
      message.success(`服务已${newStatus === 'enabled' ? '启用' : '禁用'}`);
      fetchServices();
    } catch {
      // 错误已处理
    }
  };

  // 同步工具
  const handleSync = async (id: number) => {
    try {
      await mcpService.syncTools(id);
      message.success('工具列表同步成功');
      fetchServices();
    } catch {
      message.error('同步失败，请检查服务连接');
    }
  };

  // 查看工具
  const handleViewTools = async (record: MCPService) => {
    setToolsModalOpen(true);
    setToolsLoading(true);
    try {
      const res = await mcpService.getServiceTools(record.id);
      setTools(res.data.data || []);
    } catch {
      setTools([]);
    } finally {
      setToolsLoading(false);
    }
  };

  // 打开编辑弹窗
  const handleEdit = (record: MCPService) => {
    setEditingService(record);
    form.setFieldsValue({
      name: record.name,
      display_name: record.display_name,
      description: record.description,
      endpoint_url: record.endpoint_url,
      transport_type: record.transport_type,
      auth_type: record.auth_type,
    });
    setModalOpen(true);
  };

  // 服务列表表格列
  const serviceColumns: ColumnsType<MCPService> = [
    {
      title: '服务名称',
      dataIndex: 'display_name',
      key: 'display_name',
      render: (text, record) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          <div style={{ fontSize: 12, color: themeToken.colorTextSecondary }}>{record.name}</div>
        </div>
      ),
    },
    {
      title: '端点',
      dataIndex: 'endpoint_url',
      key: 'endpoint_url',
      ellipsis: true,
      render: (url: string) => (
        <Tooltip title={url}>
          <span style={{ fontSize: 12, color: themeToken.colorTextSecondary }}>{url}</span>
        </Tooltip>
      ),
    },
    {
      title: '传输',
      dataIndex: 'transport_type',
      key: 'transport_type',
      width: 100,
      render: (type: string) => (
        <Tag color={type === 'sse' ? 'blue' : 'cyan'}>{type.toUpperCase()}</Tag>
      ),
    },
    {
      title: '工具',
      dataIndex: 'tools_count',
      key: 'tools_count',
      width: 80,
      align: 'center',
      render: (count: number, record) => (
        <Button
          type="link"
          size="small"
          onClick={() => handleViewTools(record)}
          disabled={count === 0}
        >
          {count} 个
        </Button>
      ),
    },
    {
      title: '连接状态',
      dataIndex: 'connected',
      key: 'connected',
      width: 100,
      align: 'center',
      render: (connected: boolean) => (
        <Badge status={connected ? 'success' : 'default'} text={connected ? '已连接' : '未连接'} />
      ),
    },
    {
      title: '启用',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      align: 'center',
      render: (status: string, record) => (
        <Switch
          checked={status === 'enabled'}
          onChange={() => handleToggleStatus(record)}
          size="small"
        />
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      render: (_, record) => (
        <Space size={4}>
          <Tooltip title="同步工具">
            <Button type="text" size="small" icon={<SyncOutlined />} onClick={() => handleSync(record.id)} />
          </Tooltip>
          <Tooltip title="编辑">
            <Button type="text" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)} />
          </Tooltip>
          <Popconfirm
            title="确认删除"
            description={`确定要删除服务「${record.display_name}」吗？`}
            onConfirm={() => handleDelete(record.id)}
          >
            <Tooltip title="删除">
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 访问规则表格列
  const ruleColumns: ColumnsType<MCPAccessRule> = [
    { title: '服务', dataIndex: 'service_name', key: 'service_name' },
    {
      title: '目标类型',
      dataIndex: 'target_type',
      key: 'target_type',
      render: (type: string) => {
        const map: Record<string, string> = { user: '用户', department: '部门', role: '角色' };
        return <Tag>{map[type] || type}</Tag>;
      },
    },
    { title: '目标', dataIndex: 'target_name', key: 'target_name' },
    {
      title: '权限',
      dataIndex: 'allowed',
      key: 'allowed',
      render: (allowed: boolean) => (
        <Tag color={allowed ? 'success' : 'error'}>{allowed ? '允许' : '拒绝'}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Popconfirm title="确认删除此规则？" onConfirm={() => handleDeleteRule(record.id)}>
          <Button type="text" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  const handleDeleteRule = async (id: number) => {
    try {
      await mcpService.deleteAccessRule(id);
      message.success('规则已删除');
      fetchRules();
    } catch {
      // 错误已处理
    }
  };

  return (
    <div className="animate-fade-in-up">
      {/* 页面头部 */}
      <div style={{ marginBottom: 24, display: 'flex', alignItems: 'center', gap: 16 }}>
        <div
          style={{
            width: 44,
            height: 44,
            borderRadius: 12,
            background: 'linear-gradient(135deg, #722ed1, #b37feb)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(114, 46, 209, 0.3)',
          }}
        >
          <ApiOutlined style={{ color: '#fff', fontSize: 20 }} />
        </div>
        <div>
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600, color: themeToken.colorText }}>
            MCP 服务管理
          </h2>
          <p style={{ margin: 0, fontSize: 13, color: themeToken.colorTextSecondary }}>
            管理 MCP 网关连接的后端服务和访问控制
          </p>
        </div>
      </div>

      {/* 主内容区 — 玻璃卡片 */}
      <div
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
        <Tabs
          defaultActiveKey="services"
          items={[
            {
              key: 'services',
              label: '服务管理',
              children: (
                <>
                  <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
                    <Button icon={<ReloadOutlined />} onClick={fetchServices}>
                      刷新
                    </Button>
                    <Button
                      type="primary"
                      icon={<PlusOutlined />}
                      onClick={() => {
                        setEditingService(null);
                        form.resetFields();
                        setModalOpen(true);
                      }}
                    >
                      注册服务
                    </Button>
                  </div>
                  <Table
                    columns={serviceColumns}
                    dataSource={services}
                    rowKey="id"
                    loading={loading}
                    pagination={false}
                    size="middle"
                  />
                </>
              ),
            },
            {
              key: 'access',
              label: '访问控制',
              children: (
                <>
                  <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
                    <Button icon={<ReloadOutlined />} onClick={fetchRules}>
                      刷新
                    </Button>
                  </div>
                  <Table
                    columns={ruleColumns}
                    dataSource={rules}
                    rowKey="id"
                    loading={rulesLoading}
                    pagination={false}
                    size="middle"
                  />
                </>
              ),
            },
          ]}
        />
      </div>

      {/* 创建/编辑服务弹窗 */}
      <Modal
        title={editingService ? '编辑服务' : '注册 MCP 服务'}
        open={modalOpen}
        onOk={handleSave}
        onCancel={() => {
          setModalOpen(false);
          setEditingService(null);
          form.resetFields();
        }}
        width={560}
        okText="保存"
        cancelText="取消"
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="name"
            label="服务标识"
            rules={[{ required: true, message: '请输入服务标识' }]}
            extra="唯一标识，用于路由和配置"
          >
            <Input placeholder="如：code-search" disabled={!!editingService} />
          </Form.Item>
          <Form.Item
            name="display_name"
            label="显示名称"
            rules={[{ required: true, message: '请输入显示名称' }]}
          >
            <Input placeholder="如：代码搜索服务" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea placeholder="服务功能描述" rows={2} />
          </Form.Item>
          <Form.Item
            name="endpoint_url"
            label="服务端点"
            rules={[{ required: true, message: '请输入服务端点 URL' }]}
          >
            <Input placeholder="如：http://localhost:3001/sse" />
          </Form.Item>
          <div style={{ display: 'flex', gap: 16 }}>
            <Form.Item
              name="transport_type"
              label="传输类型"
              rules={[{ required: true }]}
              initialValue="sse"
              style={{ flex: 1 }}
            >
              <Select
                options={[
                  { label: 'SSE', value: 'sse' },
                  { label: 'Streamable HTTP', value: 'streamable-http' },
                ]}
              />
            </Form.Item>
            <Form.Item
              name="auth_type"
              label="认证方式"
              rules={[{ required: true }]}
              initialValue="none"
              style={{ flex: 1 }}
            >
              <Select
                options={[
                  { label: '无认证', value: 'none' },
                  { label: 'Bearer Token', value: 'bearer' },
                  { label: '自定义 Header', value: 'header' },
                ]}
              />
            </Form.Item>
          </div>
        </Form>
      </Modal>

      {/* 工具列表弹窗 */}
      <Modal
        title={
          <span>
            <ToolOutlined style={{ marginRight: 8 }} />
            工具列表
          </span>
        }
        open={toolsModalOpen}
        onCancel={() => setToolsModalOpen(false)}
        footer={null}
        width={600}
      >
        <Table
          columns={[
            { title: '工具名称', dataIndex: 'name', key: 'name', width: 200 },
            { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
          ]}
          dataSource={tools}
          rowKey="name"
          loading={toolsLoading}
          pagination={false}
          size="small"
        />
      </Modal>
    </div>
  );
};

export default McpPage;
