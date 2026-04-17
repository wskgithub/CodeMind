import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Table, Button, Modal, Form, Input, Space, Tag, message, Select, TreeSelect } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, ApartmentOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { DeptTree, UserDetail } from '@/types';
import departmentService, { type CreateDepartmentParams } from '@/services/departmentService';
import userService from '@/services/userService';
import useAppStore from '@/store/appStore';

const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(0, 245, 212, 0.25)',
    }}
  >
    {icon}
  </span>
);

const DepartmentsPage: React.FC = () => {
  const { t } = useTranslation();
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  const [departments, setDepartments] = useState<DeptTree[]>([]);
  const [users, setUsers] = useState<UserDetail[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingDept, setEditingDept] = useState<DeptTree | null>(null);
  const [form] = Form.useForm();

  const loadDepartments = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await departmentService.list();
      setDepartments(resp.data.data || []);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  }, []);

  const loadUsers = useCallback(async () => {
    try {
      const allUsers: UserDetail[] = [];
      let page = 1;
      const pageSize = 100;

      while (true) {
        const resp = await userService.list({ page, page_size: pageSize });
        const users = resp.data.data.list || [];
        allUsers.push(...users);
        if (users.length < pageSize) break;
        page++;
        if (page > 10) break;
      }
      setUsers(allUsers);
    } catch {
      // handled by interceptor
    }
  }, []);

  useEffect(() => {
    loadDepartments();
    loadUsers();
  }, [loadDepartments, loadUsers]);

  const handleCreate = () => {
    setEditingDept(null);
    form.resetFields();
    setModalOpen(true);
  };

  const handleEdit = (record: DeptTree) => {
    setEditingDept(record);
    form.setFieldsValue({
      name: record.name,
      description: record.description || '',
      manager_id: record.manager?.id,
    });
    setModalOpen(true);
  };

  const handleSubmit = async (values: CreateDepartmentParams) => {
    try {
      if (editingDept) {
        await departmentService.update(editingDept.id, values);
        message.success(t('departments.updated'));
      } else {
        await departmentService.create(values);
        message.success(t('departments.created'));
      }
      setModalOpen(false);
      form.resetFields();
      loadDepartments();
    } catch {
      // handled by interceptor
    }
  };

  const handleDelete = (record: DeptTree) => {
    Modal.confirm({
      title: t('departments.confirmDeleteTitle'),
      content: t('departments.confirmDeleteContent', { name: record.name }),
      okText: t('common.delete'),
      okType: 'danger',
      okButtonProps: {
        style: { background: '#FF6B6B', borderColor: '#FF6B6B' },
      },
      onOk: async () => {
        await departmentService.delete(record.id);
        message.success(t('success.deleted'));
        loadDepartments();
      },
    });
  };

  interface TreeSelectNode {
    title: string;
    value: number;
    children?: TreeSelectNode[];
  }
  const convertToTreeData = (depts: DeptTree[]): TreeSelectNode[] => {
    return depts.map((dept) => ({
      title: dept.name,
      value: dept.id,
      children: dept.children ? convertToTreeData(dept.children) : undefined,
    }));
  };

  const columns: ColumnsType<DeptTree> = useMemo(() => [
    { 
      title: t('departments.table.name'), 
      dataIndex: 'name', 
      key: 'name',
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>{text}</span>,
    },
    { 
      title: t('departments.form.description'), 
      dataIndex: 'description', 
      key: 'description', 
      render: (v) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.45)' }}>{v || '-'}</span>,
    },
    {
      title: t('departments.form.managerLabel'),
      key: 'manager',
      render: (_, r) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>
          {r.manager?.display_name || '-'}
        </span>
      ),
    },
    { 
      title: t('departments.table.userCount'), 
      dataIndex: 'user_count', 
      key: 'user_count', 
      width: 100,
      render: (v) => <span style={{ color: '#00D9FF', fontWeight: 600 }}>{v}</span>,
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
            {t('common.enabled')}
          </Tag>
        ) : (
          <Tag style={{ 
            color: '#FF6B6B', 
            background: 'rgba(255, 107, 107, 0.15)',
            border: '1px solid rgba(255, 107, 107, 0.3)',
            borderRadius: 6,
          }}>
            {t('common.disabled')}
          </Tag>
        ),
    },
    {
      title: t('common.actions'),
      key: 'action',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button 
            type="link" 
            size="small" 
            icon={<EditOutlined />} 
            onClick={() => handleEdit(record)}
            style={{ color: '#00D9FF' }}
          >
            {t('common.edit')}
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
  ], [isDark]);

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<ApartmentOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                {t('departments.title')}
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', fontSize: 14, marginTop: 4 }}>
                {t('departments.pageDescription')}
              </p>
            </div>
          </div>
          <div style={{ marginTop: 20, display: 'flex', gap: 12 }}>
            <Button 
              icon={<ReloadOutlined />} 
              onClick={loadDepartments}
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
                background: 'linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)',
                border: 'none',
                boxShadow: '0 4px 16px rgba(0, 245, 212, 0.25)',
              }}
            >
              {t('departments.create')}
            </Button>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={departments}
            loading={loading}
            pagination={false}
            childrenColumnName="children"
          />
        </div>

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {editingDept ? t('departments.form.editTitle') : t('departments.form.createTitle')}
            </span>
          }
          open={modalOpen}
          onCancel={() => setModalOpen(false)}
          footer={null}
          destroyOnClose
        >
          <Form form={form} layout="vertical" onFinish={handleSubmit}>
            <Form.Item 
              name="name" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('departments.form.name')}</span>} 
              rules={[{ required: true, message: t('departments.form.nameRequired') }]}
            >
              <Input 
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                  color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="description" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('departments.form.description')}</span>}
            >
              <Input.TextArea 
                rows={3}
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                  color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                }}
              />
            </Form.Item>
            {!editingDept && (
              <Form.Item 
                name="parent_id" 
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('departments.form.parent')}</span>}
              >
                <TreeSelect
                  placeholder={t('departments.form.parentPlaceholder')}
                  allowClear
                  treeData={convertToTreeData(departments)}
                  treeDefaultExpandAll
                  style={{
                    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                  }}
                />
              </Form.Item>
            )}
            <Form.Item 
              name="manager_id" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('departments.form.managerLabel')}</span>}
            >
              <Select
                placeholder={t('departments.form.managerPlaceholder')}
                allowClear
                showSearch
                filterOption={(input, option) =>
                  (option?.label?.toString() ?? '').toLowerCase().includes(input.toLowerCase())
                }
                options={users.map((u) => ({
                  label: `${u.display_name} (${u.username})`,
                  value: u.id,
                }))}
                style={{
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                }}
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
                  background: 'linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)',
                  border: 'none',
                  boxShadow: '0 4px 16px rgba(0, 245, 212, 0.25)',
                }}
              >
                {editingDept ? t('common.save') : t('common.create')}
              </Button>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default DepartmentsPage;
