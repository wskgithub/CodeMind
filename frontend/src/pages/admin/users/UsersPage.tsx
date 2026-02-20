import { useState, useEffect, useCallback } from 'react';
import {
  Table, Button, Modal, Form, Input, Select, Space, Tag, message,
} from 'antd';
import { 
  PlusOutlined, EditOutlined, DeleteOutlined, StopOutlined, 
  CheckCircleOutlined, ReloadOutlined, UserOutlined, UnlockOutlined, LockOutlined 
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { UserDetail, DeptTree } from '@/types';
import userService, { type UserListParams, type CreateUserParams } from '@/services/userService';
import departmentService from '@/services/departmentService';
import useAuthStore from '@/store/authStore';

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
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

/** 判断用户是否被锁定 */
const isUserLocked = (user: UserDetail): boolean => {
  if (!user.locked_until) return false;
  return dayjs(user.locked_until).isAfter(dayjs());
};

/** 格式化锁定时间 */
const formatLockTime = (lockedUntil?: string): string => {
  if (!lockedUntil) return '';
  const lockTime = dayjs(lockedUntil);
  if (lockTime.isBefore(dayjs())) return '';
  
  const diffMinutes = lockTime.diff(dayjs(), 'minute');
  if (diffMinutes < 60) {
    return `${diffMinutes}分钟`;
  }
  
  const diffHours = lockTime.diff(dayjs(), 'hour');
  const remainingMinutes = diffMinutes % 60;
  if (remainingMinutes > 0) {
    return `${diffHours}小时${remainingMinutes}分钟`;
  }
  return `${diffHours}小时`;
};

/** 用户管理页面 — 与首页/登录页新设计风格统一 */
const UsersPage: React.FC = () => {
  const currentUser = useAuthStore((s) => s.user);
  const isSuperAdmin = currentUser?.role === 'super_admin';
  const isDeptManager = currentUser?.role === 'dept_manager';

  const [users, setUsers] = useState<UserDetail[]>([]);
  const [departments, setDepartments] = useState<DeptTree[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [params, setParams] = useState<UserListParams>(() => {
    const initialParams: UserListParams = { page: 1, page_size: 20 };
    if (isDeptManager && currentUser?.department?.id) {
      initialParams.department_id = currentUser.department.id;
    }
    return initialParams;
  });
  const [modalOpen, setModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<UserDetail | null>(null);
  const [form] = Form.useForm();

  // 加载用户列表
  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await userService.list(params);
      const data = resp.data.data;
      setUsers(data.list || []);
      setTotal(data.pagination.total);
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setLoading(false);
    }
  }, [params]);

  // 加载部门列表
  const loadDepartments = useCallback(async () => {
    try {
      const resp = await departmentService.list();
      setDepartments(resp.data.data || []);
    } catch {
      // 错误已在拦截器中处理
    }
  }, []);

  useEffect(() => {
    loadUsers();
    loadDepartments();
  }, [loadUsers, loadDepartments]);

  // 创建用户
  const handleCreate = () => {
    setEditingUser(null);
    form.resetFields();
    if (isDeptManager && currentUser?.department?.id) {
      form.setFieldsValue({ department_id: currentUser.department.id });
    }
    setModalOpen(true);
  };

  // 编辑用户
  const handleEdit = (record: UserDetail) => {
    setEditingUser(record);
    form.setFieldsValue({
      display_name: record.display_name,
      email: record.email || '',
      phone: record.phone || '',
      role: record.role,
      department_id: record.department_id,
    });
    setModalOpen(true);
  };

  // 提交表单
  const handleSubmit = async (values: CreateUserParams) => {
    try {
      if (editingUser) {
        await userService.update(editingUser.id, values);
        message.success('用户信息已更新');
      } else {
        await userService.create(values);
        message.success('用户创建成功');
      }
      setModalOpen(false);
      form.resetFields();
      loadUsers();
    } catch {
      // 错误已在拦截器中处理
    }
  };

  // 切换状态
  const handleToggleStatus = async (record: UserDetail) => {
    const newStatus = record.status === 1 ? 0 : 1;
    await userService.updateStatus(record.id, newStatus);
    message.success(newStatus === 1 ? '已启用' : '已禁用');
    loadUsers();
  };

  // 删除用户
  const handleDelete = (record: UserDetail) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除用户 "${record.display_name}" 吗？`,
      okText: '删除',
      okType: 'danger',
      okButtonProps: {
        style: { background: '#FF6B6B', borderColor: '#FF6B6B' },
      },
      onOk: async () => {
        await userService.delete(record.id);
        message.success('删除成功');
        loadUsers();
      },
    });
  };

  // 解锁用户
  const handleUnlock = (record: UserDetail) => {
    const locked = isUserLocked(record);
    const hasFailCount = record.login_fail_count > 0;
    
    if (!locked && !hasFailCount) {
      message.info('该用户未被锁定');
      return;
    }

    Modal.confirm({
      title: '确认解锁',
      content: (
        <div>
          <p>确定要解锁用户 "{record.display_name}" 吗？</p>
          {locked && (
            <p style={{ color: '#FF6B6B', fontSize: 13 }}>
              当前账号被锁定，剩余锁定时间：{formatLockTime(record.locked_until)}
            </p>
          )}
          {hasFailCount && !locked && (
            <p style={{ color: '#FFBE0B', fontSize: 13 }}>
              该用户有 {record.login_fail_count} 次登录失败记录
            </p>
          )}
        </div>
      ),
      okText: '解锁',
      okButtonProps: { 
        type: 'primary',
        style: {
          background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
          border: 'none',
        },
      },
      onOk: async () => {
        try {
          await userService.unlockUser(record.id, '管理员手动解锁');
          message.success('账号已解锁');
          loadUsers();
        } catch {
          // 错误已在拦截器中处理
        }
      },
    });
  };

  // 角色标签 - 新设计
  const roleTag = (role: string) => {
    const map: Record<string, { text: string; color: string; bg: string }> = {
      super_admin: { text: '超级管理员', color: '#FF6B6B', bg: 'rgba(255, 107, 107, 0.15)' },
      dept_manager: { text: '部门经理', color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
      user: { text: '普通用户', color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
    };
    const r = map[role] || { text: role, color: 'default', bg: 'transparent' };
    return (
      <Tag 
        style={{ 
          color: r.color,
          background: r.bg,
          border: `1px solid ${r.color}40`,
          borderRadius: 6,
        }}
      >
        {r.text}
      </Tag>
    );
  };

  // 状态标签 - 新设计
  const statusTag = (record: UserDetail) => {
    const locked = isUserLocked(record);
    
    if (record.status !== 1) {
      return (
        <Tag style={{ 
          color: '#FF6B6B', 
          background: 'rgba(255, 107, 107, 0.15)',
          border: '1px solid rgba(255, 107, 107, 0.3)',
          borderRadius: 6,
        }}>
          禁用
        </Tag>
      );
    }
    
    if (locked) {
      return (
        <Tag 
          icon={<LockOutlined />}
          style={{ 
            color: '#FFBE0B', 
            background: 'rgba(255, 190, 11, 0.15)',
            border: '1px solid rgba(255, 190, 11, 0.3)',
            borderRadius: 6,
          }}
        >
          锁定 {formatLockTime(record.locked_until)}
        </Tag>
      );
    }
    
    if (record.login_fail_count > 0) {
      return (
        <Tag style={{ 
          color: '#FFBE0B', 
          background: 'rgba(255, 190, 11, 0.15)',
          border: '1px solid rgba(255, 190, 11, 0.3)',
          borderRadius: 6,
        }}>
          启用 ({record.login_fail_count}次失败)
        </Tag>
      );
    }
    
    return (
      <Tag style={{ 
        color: '#00F5D4', 
        background: 'rgba(0, 245, 212, 0.15)',
        border: '1px solid rgba(0, 245, 212, 0.3)',
        borderRadius: 6,
      }}>
        启用
      </Tag>
    );
  };

  // 扁平化部门树为选项列表
  const flattenDepartments = (depts: DeptTree[], prefix = ''): { label: string; value: number }[] => {
    const result: { label: string; value: number }[] = [];
    depts.forEach((dept) => {
      result.push({
        label: prefix + dept.name,
        value: dept.id,
      });
      if (dept.children && dept.children.length > 0) {
        result.push(...flattenDepartments(dept.children, prefix + dept.name + ' / '));
      }
    });
    return result;
  };

  // 表格列 - 新设计
  const columns: ColumnsType<UserDetail> = [
    { 
      title: '用户名', 
      dataIndex: 'username', 
      key: 'username', 
      width: 120,
      render: (text) => <span style={{ color: '#fff', fontWeight: 500 }}>{text}</span>,
    },
    { 
      title: '姓名', 
      dataIndex: 'display_name', 
      key: 'display_name', 
      width: 120,
      render: (text) => <span style={{ color: '#fff' }}>{text}</span>,
    },
    { 
      title: '邮箱', 
      dataIndex: 'email', 
      key: 'email', 
      width: 180, 
      render: (v) => <span style={{ color: 'rgba(255, 255, 255, 0.7)' }}>{v || '-'}</span>,
    },
    { title: '角色', dataIndex: 'role', key: 'role', width: 120, render: roleTag },
    {
      title: '部门', 
      key: 'department', 
      width: 120,
      render: (_, r) => <span style={{ color: 'rgba(255, 255, 255, 0.7)' }}>{r.department?.name || '-'}</span>,
    },
    {
      title: '状态', 
      key: 'status', 
      width: 160,
      render: (_, record) => statusTag(record),
    },
    {
      title: '最后登录', 
      dataIndex: 'last_login_at', 
      key: 'last_login_at', 
      width: 160,
      render: (v: string) => (
        <span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>
          {v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'}
        </span>
      ),
    },
    {
      title: '操作', 
      key: 'action', 
      width: 280, 
      fixed: 'right',
      render: (_, record) => {
        const locked = isUserLocked(record);
        const hasFailCount = record.login_fail_count > 0;
        
        return (
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
              icon={record.status === 1 ? <StopOutlined /> : <CheckCircleOutlined />}
              onClick={() => handleToggleStatus(record)}
              style={{ color: record.status === 1 ? '#FFBE0B' : '#00F5D4' }}
            >
              {record.status === 1 ? '禁用' : '启用'}
            </Button>
            {(locked || hasFailCount) && (
              <Button 
                type="link" 
                size="small" 
                icon={<UnlockOutlined />}
                onClick={() => handleUnlock(record)}
                style={{ color: '#FF6B6B' }}
              >
                解锁
              </Button>
            )}
            {isSuperAdmin && (
              <Button 
                type="link" 
                size="small" 
                danger 
                icon={<DeleteOutlined />} 
                onClick={() => handleDelete(record)}
              >
                删除
              </Button>
            )}
          </Space>
        );
      },
    },
  ];

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<UserOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
                用户管理
              </h2>
              <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
                管理系统用户，支持创建、编辑、启用/禁用及角色分配
              </p>
            </div>
          </div>
          <div style={{ marginTop: 20, display: 'flex', gap: 12 }}>
            <Input.Search
              placeholder="搜索用户名/姓名/邮箱"
              allowClear
              onSearch={(v) => setParams((p) => ({ ...p, keyword: v, page: 1 }))}
              style={{ width: 280 }}
            />
            <Button 
              icon={<ReloadOutlined />} 
              onClick={loadUsers}
              style={{
                background: 'rgba(255, 255, 255, 0.03)',
                borderColor: 'rgba(255, 255, 255, 0.1)',
                color: 'rgba(255, 255, 255, 0.8)',
              }}
            >
              刷新
            </Button>
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={handleCreate}
              style={{
                background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                border: 'none',
                boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
              }}
            >
              创建用户
            </Button>
          </div>
        </div>

        {/* 表格区域 — 玻璃态卡片 - 新设计 */}
        <div className="glass-card animate-fade-in-up" style={{ padding: 24 }}>
          <Table
            rowKey="id"
            columns={columns}
            dataSource={users}
            loading={loading}
            scroll={{ x: 1200 }}
            pagination={{
              current: params.page,
              pageSize: params.page_size,
              total,
              showSizeChanger: true,
              showTotal: (t) => `共 ${t} 条`,
              onChange: (page, pageSize) => setParams((p) => ({ ...p, page, page_size: pageSize })),
            }}
          />
        </div>

        {/* 创建/编辑弹窗 - 新设计 */}
        <Modal
          title={
            <span style={{ color: '#fff', fontSize: 18, fontWeight: 600 }}>
              {editingUser ? '编辑用户' : '创建用户'}
            </span>
          }
          open={modalOpen}
          onCancel={() => setModalOpen(false)}
          footer={null}
          destroyOnClose
          width={520}
        >
          <Form form={form} layout="vertical" onFinish={handleSubmit}>
            {!editingUser && (
              <>
                <Form.Item 
                  name="username" 
                  label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>用户名</span>} 
                  rules={[{ required: true, message: '请输入用户名' }]}
                >
                  <Input 
                    placeholder="2-50 位字母、数字、下划线"
                    style={{ 
                      background: 'rgba(255, 255, 255, 0.03)',
                      borderColor: 'rgba(255, 255, 255, 0.1)',
                      color: '#fff',
                    }}
                  />
                </Form.Item>
                <Form.Item 
                  name="password" 
                  label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>初始密码</span>} 
                  rules={[{ required: true, message: '请设置初始密码' }, { min: 8, message: '不少于 8 位' }]}
                >
                  <Input.Password 
                    placeholder="至少 8 位，含大小写字母和数字"
                    style={{ 
                      background: 'rgba(255, 255, 255, 0.03)',
                      borderColor: 'rgba(255, 255, 255, 0.1)',
                      color: '#fff',
                    }}
                  />
                </Form.Item>
              </>
            )}
            <Form.Item 
              name="display_name" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>姓名</span>} 
              rules={[{ required: true, message: '请输入姓名' }]}
            >
              <Input 
                style={{ 
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="email" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>邮箱</span>}
            >
              <Input 
                type="email"
                style={{ 
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="phone" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>手机号</span>}
            >
              <Input 
                style={{ 
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            {!editingUser && (
              <Form.Item 
                name="role" 
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>角色</span>} 
                rules={[{ required: true, message: '请选择角色' }]}
              >
                <Select>
                  {isSuperAdmin && <Select.Option value="super_admin">超级管理员</Select.Option>}
                  {isSuperAdmin && <Select.Option value="dept_manager">部门经理</Select.Option>}
                  <Select.Option value="user">普通用户</Select.Option>
                </Select>
              </Form.Item>
            )}
            <Form.Item
              name="department_id"
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>所属部门</span>}
              rules={isDeptManager ? [{ required: true, message: '请选择所属部门' }] : undefined}
            >
              <Select
                placeholder={isDeptManager ? '所属部门' : '选择所属部门（可选）'}
                allowClear={!isDeptManager}
                disabled={isDeptManager}
                showSearch
                filterOption={(input, option) =>
                  (option?.label?.toString() ?? '').toLowerCase().includes(input.toLowerCase())
                }
                options={flattenDepartments(departments)}
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
                {editingUser ? '保存' : '创建'}
              </Button>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default UsersPage;
