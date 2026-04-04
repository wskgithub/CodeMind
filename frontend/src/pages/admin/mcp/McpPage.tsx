import { useState, useEffect, useCallback, useMemo } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Tag, Space,
  Tabs, Popconfirm, Switch, message, Tooltip, Badge,
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
import useAppStore from '@/store/appStore';

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
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

/** MCP 服务管理页 — 与首页/登录页新设计风格统一 */
const McpPage: React.FC = () => {
  // 主题模式
  const themeMode = useAppStore((state) => state.themeMode);
  const isDark = themeMode === 'dark';

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

  const serviceColumns: ColumnsType<MCPService> = useMemo(() => [
    {
      title: '服务名称',
      dataIndex: 'display_name',
      key: 'display_name',
      render: (text, record) => (
        <div>
          <div style={{ fontWeight: 600, color: isDark ? '#fff' : '#1f2937', fontSize: 15 }}>{text}</div>
          <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}>{record.name}</div>
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
          <span style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontFamily: 'monospace' }}>{url}</span>
        </Tooltip>
      ),
    },
    {
      title: '传输',
      dataIndex: 'transport_type',
      key: 'transport_type',
      width: 100,
      render: (type: string) => (
        <Tag style={{
          color: type === 'sse' ? '#00D9FF' : '#00F5D4',
          background: type === 'sse' ? 'rgba(0, 217, 255, 0.15)' : 'rgba(0, 245, 212, 0.15)',
          border: `1px solid ${type === 'sse' ? 'rgba(0, 217, 255, 0.3)' : 'rgba(0, 245, 212, 0.3)'}`,
          borderRadius: 6,
        }}>
          {type.toUpperCase()}
        </Tag>
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
          style={{ color: count === 0 ? (isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.3)') : '#FFBE0B' }}
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
        <Badge 
          status={connected ? 'success' : 'default'} 
          text={<span style={{ color: connected ? '#00F5D4' : (isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)') }}>{connected ? '已连接' : '未连接'}</span>} 
        />
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
            <Button 
              type="text" 
              size="small" 
              icon={<SyncOutlined style={{ color: '#00D9FF' }} />} 
              onClick={() => handleSync(record.id)} 
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button 
              type="text" 
              size="small" 
              icon={<EditOutlined style={{ color: '#00F5D4' }} />} 
              onClick={() => handleEdit(record)} 
            />
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
  ], [isDark]);

  const ruleColumns: ColumnsType<MCPAccessRule> = useMemo(() => [
    { 
      title: '服务', 
      dataIndex: 'service_name', 
      key: 'service_name',
      render: (text) => <span style={{ color: isDark ? '#fff' : '#1f2937' }}>{text}</span>,
    },
    {
      title: '目标类型',
      dataIndex: 'target_type',
      key: 'target_type',
      render: (type: string) => {
        const map: Record<string, { text: string; color: string; bg: string }> = {
          user: { text: '用户', color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
          department: { text: '部门', color: '#9D4EDD', bg: 'rgba(157, 78, 221, 0.15)' },
          role: { text: '角色', color: '#FFBE0B', bg: 'rgba(255, 190, 11, 0.15)' },
        };
        const t = map[type] || { text: type, color: isDark ? '#fff' : '#1f2937', bg: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' };
        return (
          <Tag style={{ 
            color: t.color,
            background: t.bg,
            border: `1px solid ${t.color}40`,
            borderRadius: 6,
          }}>
            {t.text}
          </Tag>
        );
      },
    },
    { 
      title: '目标', 
      dataIndex: 'target_name', 
      key: 'target_name',
      render: (text) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{text}</span>,
    },
    {
      title: '权限',
      dataIndex: 'allowed',
      key: 'allowed',
      render: (allowed: boolean) => (
        <Tag style={{
          color: allowed ? '#00F5D4' : '#FF6B6B',
          background: allowed ? 'rgba(0, 245, 212, 0.15)' : 'rgba(255, 107, 107, 0.15)',
          border: `1px solid ${allowed ? 'rgba(0, 245, 212, 0.3)' : 'rgba(255, 107, 107, 0.3)'}`,
          borderRadius: 6,
        }}>
          {allowed ? '允许' : '拒绝'}
        </Tag>
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
  ], [isDark]);

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
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<ApiOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : '#1f2937', fontSize: 24, fontWeight: 600 }}>
                MCP 服务管理
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                管理 MCP 网关连接的后端服务和访问控制
              </p>
            </div>
          </div>
        </div>

        {/* 主内容区 — 玻璃卡片 - 新设计 */}
        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Tabs
            defaultActiveKey="services"
            items={[
              {
                key: 'services',
                label: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>服务管理</span>,
                children: (
                  <>
                    <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
                      <Button 
                        icon={<ReloadOutlined />} 
                        onClick={fetchServices}
                        style={{
                          background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#f3f4f6',
                          borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : '#d1d5db',
                          color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)',
                        }}
                      >
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
                        style={{
                          background: 'linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)',
                          border: 'none',
                          boxShadow: '0 4px 16px rgba(157, 78, 221, 0.25)',
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
                label: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>访问控制</span>,
                children: (
                  <>
                    <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
                      <Button 
                        icon={<ReloadOutlined />} 
                        onClick={fetchRules}
                        style={{
                          background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#f3f4f6',
                          borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : '#d1d5db',
                          color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)',
                        }}
                      >
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

        {/* 创建/编辑服务弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : '#1f2937', fontSize: 18, fontWeight: 600 }}>
              {editingService ? '编辑服务' : '注册 MCP 服务'}
            </span>
          }
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
          okButtonProps={{
            style: {
              background: 'linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)',
              border: 'none',
              boxShadow: '0 4px 16px rgba(157, 78, 221, 0.25)',
            },
          }}
        >
          <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
            <Form.Item
              name="name"
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>服务标识</span>}
              rules={[{ required: true, message: '请输入服务标识' }]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}>唯一标识，用于路由和配置</span>}
            >
              <Input 
                placeholder="如：code-search" 
                disabled={!!editingService}
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#f9fafb',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : '#d1d5db',
                  color: isDark ? '#fff' : '#1f2937',
                }}
              />
            </Form.Item>
            <Form.Item
              name="display_name"
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>显示名称</span>}
              rules={[{ required: true, message: '请输入显示名称' }]}
            >
              <Input 
                placeholder="如：代码搜索服务"
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#f9fafb',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : '#d1d5db',
                  color: isDark ? '#fff' : '#1f2937',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="description" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>描述</span>}
            >
              <Input.TextArea 
                placeholder="服务功能描述" 
                rows={2}
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#f9fafb',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : '#d1d5db',
                  color: isDark ? '#fff' : '#1f2937',
                }}
              />
            </Form.Item>
            <Form.Item
              name="endpoint_url"
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>服务端点</span>}
              rules={[{ required: true, message: '请输入服务端点 URL' }]}
            >
              <Input 
                placeholder="如：http://localhost:3001/sse"
                style={{ fontFamily: 'monospace', color: isDark ? '#00D9FF' : '#0891b2' }}
              />
            </Form.Item>
            <div style={{ display: 'flex', gap: 16 }}>
              <Form.Item
                name="transport_type"
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>传输类型</span>}
                rules={[{ required: true }]}
                initialValue="sse"
                style={{ flex: 1 }}
              >
                <Select
                  style={{
                    background: isDark ? 'transparent' : '#f9fafb',
                  }}
                  options={[
                    { label: 'SSE', value: 'sse' },
                    { label: 'Streamable HTTP', value: 'streamable-http' },
                  ]}
                />
              </Form.Item>
              <Form.Item
                name="auth_type"
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>认证方式</span>}
                rules={[{ required: true }]}
                initialValue="none"
                style={{ flex: 1 }}
              >
                <Select
                  style={{
                    background: isDark ? 'transparent' : '#f9fafb',
                  }}
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

        {/* 工具列表弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : '#1f2937', fontSize: 18, fontWeight: 600 }}>
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
              { 
                title: '工具名称', 
                dataIndex: 'name', 
                key: 'name', 
                width: 200,
                render: (text) => <span style={{ color: isDark ? '#fff' : '#1f2937', fontFamily: 'monospace' }}>{text}</span>,
              },
              { 
                title: '描述', 
                dataIndex: 'description', 
                key: 'description', 
                ellipsis: true,
                render: (text) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)' }}>{text}</span>,
              },
            ]}
            dataSource={tools}
            rowKey="name"
            loading={toolsLoading}
            pagination={false}
            size="small"
          />
        </Modal>
      </div>
    </div>
  );
};

export default McpPage;
