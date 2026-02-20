import { useState, useEffect, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, Space, Tag, message, Select, TreeSelect } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, ApartmentOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { DeptTree, UserDetail } from '@/types';
import departmentService, { type CreateDepartmentParams } from '@/services/departmentService';
import userService from '@/services/userService';
import useAppStore from '@/store/appStore';

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
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

/** 部门管理页面 — 与首页/登录页新设计风格统一 */
const DepartmentsPage: React.FC = () => {
  const { themeMode } = useAppStore();
  const isDark = themeMode === 'dark';

  const [departments, setDepartments] = useState<DeptTree[]>([]);
  const [users, setUsers] = useState<UserDetail[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editingDept, setEditingDept] = useState<DeptTree | null>(null);
  const [form] = Form.useForm();

  // 加载部门列表
  const loadDepartments = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await departmentService.list();
      setDepartments(resp.data.data || []);
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setLoading(false);
    }
  }, []);

  // 加载用户列表
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
      // 错误已在拦截器中处理
    }
  }, []);

  useEffect(() => {
    loadDepartments();
    loadUsers();
  }, [loadDepartments, loadUsers]);

  // 创建部门
  const handleCreate = () => {
    setEditingDept(null);
    form.resetFields();
    setModalOpen(true);
  };

  // 编辑部门
  const handleEdit = (record: DeptTree) => {
    setEditingDept(record);
    form.setFieldsValue({
      name: record.name,
      description: record.description || '',
      manager_id: record.manager?.id,
    });
    setModalOpen(true);
  };

  // 提交表单
  const handleSubmit = async (values: CreateDepartmentParams) => {
    try {
      if (editingDept) {
        await departmentService.update(editingDept.id, values);
        message.success('部门信息已更新');
      } else {
        await departmentService.create(values);
        message.success('部门创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      loadDepartments();
    } catch {
      // 错误已在拦截器中处理
    }
  };

  // 删除部门
  const handleDelete = (record: DeptTree) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除部门 "${record.name}" 吗？请确保部门下无用户和子部门。`,
      okText: '删除',
      okType: 'danger',
      okButtonProps: {
        style: { background: '#FF6B6B', borderColor: '#FF6B6B' },
      },
      onOk: async () => {
        await departmentService.delete(record.id);
        message.success('删除成功');
        loadDepartments();
      },
    });
  };

  // 将部门列表转换为树形选择器数据
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

  // 表格列 - 新设计
  const columns: ColumnsType<DeptTree> = [
    { 
      title: '部门名称', 
      dataIndex: 'name', 
      key: 'name',
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>{text}</span>,
    },
    { 
      title: '描述', 
      dataIndex: 'description', 
      key: 'description', 
      render: (v) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.45)' }}>{v || '-'}</span>,
    },
    {
      title: '部门经理',
      key: 'manager',
      render: (_, r) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>
          {r.manager?.display_name || '-'}
        </span>
      ),
    },
    { 
      title: '用户数', 
      dataIndex: 'user_count', 
      key: 'user_count', 
      width: 100,
      render: (v) => <span style={{ color: '#00D9FF', fontWeight: 600 }}>{v}</span>,
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
            启用
          </Tag>
        ) : (
          <Tag style={{ 
            color: '#FF6B6B', 
            background: 'rgba(255, 107, 107, 0.15)',
            border: '1px solid rgba(255, 107, 107, 0.3)',
            borderRadius: 6,
          }}>
            禁用
          </Tag>
        ),
    },
    {
      title: '操作',
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
            编辑
          </Button>
          <Button 
            type="link" 
            size="small" 
            danger 
            icon={<DeleteOutlined />} 
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<ApartmentOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                部门管理
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', fontSize: 14, marginTop: 4 }}>
                管理组织架构，支持树形结构、上级部门选择及部门经理配置
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
              刷新
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
              创建部门
            </Button>
          </div>
        </div>

        {/* 树形表格 — 玻璃态卡片 - 新设计 */}
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

        {/* 创建/编辑弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {editingDept ? '编辑部门' : '创建部门'}
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>部门名称</span>} 
              rules={[{ required: true, message: '请输入部门名称' }]}
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>部门描述</span>}
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
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>上级部门</span>}
              >
                <TreeSelect
                  placeholder="选择上级部门（留空表示顶级部门）"
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
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>部门经理</span>}
            >
              <Select
                placeholder="选择部门经理（可选）"
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
                {editingDept ? '保存' : '创建'}
              </Button>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default DepartmentsPage;
