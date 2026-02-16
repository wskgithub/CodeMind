import { useState, useEffect, useCallback } from 'react';
import { Table, Button, Modal, Form, Input, Space, Tag, message, Select, TreeSelect, theme } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined, ApartmentOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { DeptTree, UserDetail } from '@/types';
import departmentService, { type CreateDepartmentParams } from '@/services/departmentService';
import userService from '@/services/userService';

/** 页面标题图标 — 渐变圆形背景 */
const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-10 h-10 rounded-full shrink-0"
    style={{
      background: 'linear-gradient(135deg, #13c2c2 0%, #36cfc9 100%)',
      color: '#fff',
    }}
  >
    {icon}
  </span>
);

/** 部门管理页面 — Glassmorphism 风格 */
const DepartmentsPage: React.FC = () => {
  const { token } = theme.useToken();
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

  // 加载用户列表（用于选择部门经理）
  const loadUsers = useCallback(async () => {
    try {
      // 后端限制 page_size 最大为 100，这里分批加载
      const allUsers: UserDetail[] = [];
      let page = 1;
      const pageSize = 100;

      while (true) {
        const resp = await userService.list({ page, page_size: pageSize });
        const users = resp.data.data.list || [];
        allUsers.push(...users);

        // 如果返回的数据少于 pageSize，说明已经是最后一页
        if (users.length < pageSize) {
          break;
        }
        page++;

        // 安全限制：最多加载 10 页（1000 个用户）
        if (page > 10) {
          break;
        }
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

  // 表格列
  const columns: ColumnsType<DeptTree> = [
    { title: '部门名称', dataIndex: 'name', key: 'name' },
    { title: '描述', dataIndex: 'description', key: 'description', render: (v) => v || '-' },
    {
      title: '部门经理',
      key: 'manager',
      render: (_, r) => r.manager?.display_name || '-',
    },
    { title: '用户数', dataIndex: 'user_count', key: 'user_count', width: 100 },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (v: number) =>
        v === 1 ? <Tag color="success">启用</Tag> : <Tag color="error">禁用</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Button type="link" size="small" danger icon={<DeleteOutlined />} onClick={() => handleDelete(record)}>
            删除
          </Button>
        </Space>
      ),
    },
  ];

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
          <PageIcon icon={<ApartmentOutlined style={{ fontSize: 20 }} />} />
          <h2 style={{ margin: 0, color: token.colorTextHeading }}>部门管理</h2>
        </div>
        <p style={{ margin: 0, color: token.colorTextSecondary, fontSize: 14 }}>
          管理组织架构，支持树形结构、上级部门选择及部门经理配置。
        </p>
        <div style={{ marginTop: 16 }}>
          <Space wrap>
            <Button icon={<ReloadOutlined />} onClick={loadDepartments}>刷新</Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              创建部门
            </Button>
          </Space>
        </div>
      </div>

      {/* 树形表格 — 交由全局 CSS 处理行悬停 */}
      <Table
        rowKey="id"
        columns={columns}
        dataSource={departments}
        loading={loading}
        pagination={false}
        childrenColumnName="children"
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingDept ? '编辑部门' : '创建部门'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="部门名称" rules={[{ required: true, message: '请输入部门名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="部门描述">
            <Input.TextArea rows={3} />
          </Form.Item>
          {!editingDept && (
            <Form.Item name="parent_id" label="上级部门">
              <TreeSelect
                placeholder="选择上级部门（留空表示顶级部门）"
                allowClear
                treeData={convertToTreeData(departments)}
                treeDefaultExpandAll
              />
            </Form.Item>
          )}
          <Form.Item name="manager_id" label="部门经理">
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
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              {editingDept ? '保存' : '创建'}
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default DepartmentsPage;
