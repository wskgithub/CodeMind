import {
  ApiOutlined,
  PlusOutlined,
  ReloadOutlined,
  SyncOutlined,
  DeleteOutlined,
  EditOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import {
  Table, Button, Modal, Form, Input, Select, Tag, Space,
  Tabs, Popconfirm, Switch, message, Tooltip, Badge,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';

import mcpService from '@/services/mcpService';
import useAppStore from '@/store/appStore';
import type { MCPService, MCPTool, MCPAccessRule } from '@/types';

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

const McpPage: React.FC = () => {
  const { t } = useTranslation();
  const themeMode = useAppStore((state) => state.themeMode);
  const isDark = themeMode === 'dark';

  const [services, setServices] = useState<MCPService[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingService, setEditingService] = useState<MCPService | null>(null);
  const [form] = Form.useForm();

  const [toolsModalOpen, setToolsModalOpen] = useState(false);
  const [tools, setTools] = useState<MCPTool[]>([]);
  const [toolsLoading, setToolsLoading] = useState(false);

  const [rules, setRules] = useState<MCPAccessRule[]>([]);
  const [rulesLoading, setRulesLoading] = useState(false);

  const fetchServices = useCallback(async () => {
    setLoading(true);
    try {
      const res = await mcpService.listServices();
      setServices(res.data.data || []);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchRules = useCallback(async () => {
    setRulesLoading(true);
    try {
      const res = await mcpService.listAccessRules();
      setRules(res.data.data || []);
    } catch {
      // handled by interceptor
    } finally {
      setRulesLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchServices();
    fetchRules();
  }, [fetchServices, fetchRules]);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      if (editingService) {
        await mcpService.updateService(editingService.id, values);
        message.success(t('mcp.serviceUpdated'));
      } else {
        await mcpService.createService(values);
        message.success(t('mcp.serviceCreated'));
      }
      setModalOpen(false);
      form.resetFields();
      setEditingService(null);
      fetchServices();
    } catch {
      // validation failed
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await mcpService.deleteService(id);
      message.success(t('mcp.serviceDeleted'));
      fetchServices();
    } catch {
      // handled by interceptor
    }
  };

  const handleToggleStatus = async (record: MCPService) => {
    const newStatus = record.status === 'enabled' ? 'disabled' : 'enabled';
    try {
      await mcpService.updateService(record.id, { status: newStatus });
      message.success(newStatus === 'enabled' ? t('mcp.serviceEnabled') : t('mcp.serviceDisabled'));
      fetchServices();
    } catch {
      // handled by interceptor
    }
  };

  const handleSync = async (id: number) => {
    try {
      await mcpService.syncTools(id);
      message.success(t('mcp.syncSuccess'));
      fetchServices();
    } catch {
      message.error(t('mcp.syncFailed'));
    }
  };

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
      title: t('mcp.table.serviceName'),
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
      title: t('mcp.table.endpoint'),
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
      title: t('mcp.table.transport'),
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
      title: t('mcp.table.tools'),
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
          {t('mcp.table.toolCount', { count })}
        </Button>
      ),
    },
    {
      title: t('mcp.table.connectionStatus'),
      dataIndex: 'connected',
      key: 'connected',
      width: 100,
      align: 'center',
      render: (connected: boolean) => (
        <Badge 
          status={connected ? 'success' : 'default'} 
          text={<span style={{ color: connected ? '#00F5D4' : (isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)') }}>{connected ? t('mcp.table.connected') : t('mcp.table.disconnected')}</span>} 
        />
      ),
    },
    {
      title: t('mcp.table.enable'),
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
      title: t('common.actions'),
      key: 'actions',
      width: 180,
      render: (_, record) => (
        <Space size={4}>
          <Tooltip title={t('mcp.tooltip.syncTools')}>
            <Button 
              type="text" 
              size="small" 
              icon={<SyncOutlined style={{ color: '#00D9FF' }} />} 
              onClick={() => handleSync(record.id)} 
            />
          </Tooltip>
          <Tooltip title={t('common.edit')}>
            <Button 
              type="text" 
              size="small" 
              icon={<EditOutlined style={{ color: '#00F5D4' }} />} 
              onClick={() => handleEdit(record)} 
            />
          </Tooltip>
          <Popconfirm
            title={t('mcp.confirmDeleteService', { name: record.display_name })}
            onConfirm={() => handleDelete(record.id)}
          >
            <Tooltip title={t('common.delete')}>
              <Button type="text" size="small" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ], [isDark]);

  const ruleColumns: ColumnsType<MCPAccessRule> = useMemo(() => [
    { 
      title: t('mcp.table.service'), 
      dataIndex: 'service_name', 
      key: 'service_name',
      render: (text) => <span style={{ color: isDark ? '#fff' : '#1f2937' }}>{text}</span>,
    },
    {
      title: t('mcp.table.targetType'),
      dataIndex: 'target_type',
      key: 'target_type',
      render: (type: string) => {
        const map: Record<string, { text: string; color: string; bg: string }> = {
          user: { text: t('mcp.targetType.user'), color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
          department: { text: t('mcp.targetType.department'), color: '#9D4EDD', bg: 'rgba(157, 78, 221, 0.15)' },
          role: { text: t('mcp.targetType.role'), color: '#FFBE0B', bg: 'rgba(255, 190, 11, 0.15)' },
        };
        const typeInfo = map[type] || { text: type, color: isDark ? '#fff' : '#1f2937', bg: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' };
        return (
          <Tag style={{ 
            color: typeInfo.color,
            background: typeInfo.bg,
            border: `1px solid ${typeInfo.color}40`,
            borderRadius: 6,
          }}>
            {typeInfo.text}
          </Tag>
        );
      },
    },
    { 
      title: t('mcp.table.target'), 
      dataIndex: 'target_name', 
      key: 'target_name',
      render: (text) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{text}</span>,
    },
    {
      title: t('mcp.table.permission'),
      dataIndex: 'allowed',
      key: 'allowed',
      render: (allowed: boolean) => (
        <Tag style={{
          color: allowed ? '#00F5D4' : '#FF6B6B',
          background: allowed ? 'rgba(0, 245, 212, 0.15)' : 'rgba(255, 107, 107, 0.15)',
          border: `1px solid ${allowed ? 'rgba(0, 245, 212, 0.3)' : 'rgba(255, 107, 107, 0.3)'}`,
          borderRadius: 6,
        }}>
          {allowed ? t('mcp.table.allow') : t('mcp.table.deny')}
        </Tag>
      ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Popconfirm title={t('mcp.confirmDeleteRule')} onConfirm={() => handleDeleteRule(record.id)}>
          <Button type="text" size="small" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ], [isDark]);

  const handleDeleteRule = async (id: number) => {
    try {
      await mcpService.deleteAccessRule(id);
      message.success(t('mcp.ruleDeleted'));
      fetchRules();
    } catch {
      // handled by interceptor
    }
  };

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<ApiOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : '#1f2937', fontSize: 24, fontWeight: 600 }}>
                {t('mcp.pageTitle')}
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                {t('mcp.pageDescription')}
              </p>
            </div>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Tabs
            defaultActiveKey="services"
            items={[
              {
                key: 'services',
                label: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.tabs.services')}</span>,
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
                        {t('common.refresh')}
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
                        {t('mcp.registerService')}
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
                label: <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.tabs.access')}</span>,
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
                        {t('common.refresh')}
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

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : '#1f2937', fontSize: 18, fontWeight: 600 }}>
              {editingService ? t('mcp.modal.editTitle') : t('mcp.modal.createTitle')}
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
          okText={t('common.save')}
          cancelText={t('common.cancel')}
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.form.serviceId')}</span>}
              rules={[{ required: true, message: t('mcp.form.serviceIdRequired') }]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}>{t('mcp.form.serviceIdExtra')}</span>}
            >
              <Input 
                placeholder={t('mcp.form.serviceIdPlaceholder')} 
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.form.displayName')}</span>}
              rules={[{ required: true, message: t('mcp.form.displayNameRequired') }]}
            >
              <Input 
                placeholder={t('mcp.form.displayNamePlaceholder')}
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : '#f9fafb',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : '#d1d5db',
                  color: isDark ? '#fff' : '#1f2937',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="description" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.form.description')}</span>}
            >
              <Input.TextArea 
                placeholder={t('mcp.form.descriptionPlaceholder')} 
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.form.endpoint')}</span>}
              rules={[{ required: true, message: t('mcp.form.endpointRequired') }]}
            >
              <Input 
                placeholder={t('mcp.form.endpointPlaceholder')}
                style={{ fontFamily: 'monospace', color: isDark ? '#00D9FF' : '#0891b2' }}
              />
            </Form.Item>
            <div style={{ display: 'flex', gap: 16 }}>
              <Form.Item
                name="transport_type"
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.form.transportType')}</span>}
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
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)' }}>{t('mcp.form.authType')}</span>}
                rules={[{ required: true }]}
                initialValue="none"
                style={{ flex: 1 }}
              >
                <Select
                  style={{
                    background: isDark ? 'transparent' : '#f9fafb',
                  }}
                  options={[
                    { label: t('mcp.form.noAuth'), value: 'none' },
                    { label: 'Bearer Token', value: 'bearer' },
                    { label: t('mcp.form.customHeader'), value: 'header' },
                  ]}
                />
              </Form.Item>
            </div>
          </Form>
        </Modal>

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : '#1f2937', fontSize: 18, fontWeight: 600 }}>
              <ToolOutlined style={{ marginRight: 8 }} />
              {t('mcp.tools.title')}
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
                title: t('mcp.tools.name'), 
                dataIndex: 'name', 
                key: 'name', 
                width: 200,
                render: (text) => <span style={{ color: isDark ? '#fff' : '#1f2937', fontFamily: 'monospace' }}>{text}</span>,
              },
              { 
                title: t('mcp.tools.description'), 
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
