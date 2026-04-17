import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
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
import useAppStore from '@/store/appStore';

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

const isUserLocked = (user: UserDetail): boolean => {
  if (!user.locked_until) return false;
  return dayjs(user.locked_until).isAfter(dayjs());
};

const UsersPage: React.FC = () => {
  const { t } = useTranslation();
  const currentUser = useAuthStore((s) => s.user);
  const isSuperAdmin = currentUser?.role === 'super_admin';
  const isDeptManager = currentUser?.role === 'dept_manager';

  const themeMode = useAppStore((state) => state.themeMode);
  const isDark = themeMode === 'dark';

  const formatLockTime = useCallback((lockedUntil?: string): string => {
    if (!lockedUntil) return '';
    const lockTime = dayjs(lockedUntil);
    if (lockTime.isBefore(dayjs())) return '';
    
    const diffMinutes = lockTime.diff(dayjs(), 'minute');
    if (diffMinutes < 60) {
      return t('time.minutes', { count: diffMinutes });
    }
    
    const diffHours = lockTime.diff(dayjs(), 'hour');
    const remainingMinutes = diffMinutes % 60;
    if (remainingMinutes > 0) {
      return t('time.hoursMinutes', { hours: diffHours, minutes: remainingMinutes });
    }
    return t('time.hours', { count: diffHours });
  }, [t]);

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

  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await userService.list(params);
      const data = resp.data.data;
      setUsers(data.list || []);
      setTotal(data.pagination.total);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  }, [params]);

  const loadDepartments = useCallback(async () => {
    try {
      const resp = await departmentService.list();
      setDepartments(resp.data.data || []);
    } catch {
      // handled by interceptor
    }
  }, []);

  useEffect(() => {
    loadUsers();
    loadDepartments();
  }, [loadUsers, loadDepartments]);

  const handleCreate = () => {
    setEditingUser(null);
    form.resetFields();
    if (isDeptManager && currentUser?.department?.id) {
      form.setFieldsValue({ department_id: currentUser.department.id });
    }
    setModalOpen(true);
  };

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

  const handleSubmit = async (values: CreateUserParams) => {
    try {
      if (editingUser) {
        await userService.update(editingUser.id, values);
        message.success(t('users.updateSuccess'));
      } else {
        await userService.create(values);
        message.success(t('users.createSuccess'));
      }
      setModalOpen(false);
      form.resetFields();
      loadUsers();
    } catch {
      // handled by interceptor
    }
  };

  const handleToggleStatus = async (record: UserDetail) => {
    const newStatus = record.status === 1 ? 0 : 1;
    await userService.updateStatus(record.id, newStatus);
    message.success(newStatus === 1 ? t('success.enabled') : t('success.disabled'));
    loadUsers();
  };

  const handleDelete = (record: UserDetail) => {
    Modal.confirm({
      title: t('users.confirmDeleteTitle'),
      content: t('users.confirmDeleteContent', { name: record.display_name }),
      okText: t('common.delete'),
      okType: 'danger',
      okButtonProps: {
        style: { background: '#FF6B6B', borderColor: '#FF6B6B' },
      },
      onOk: async () => {
        await userService.delete(record.id);
        message.success(t('success.deleted'));
        loadUsers();
      },
    });
  };

  const handleUnlock = (record: UserDetail) => {
    const locked = isUserLocked(record);
    const hasFailCount = record.login_fail_count > 0;
    
    if (!locked && !hasFailCount) {
      message.info(t('users.notLocked'));
      return;
    }

    Modal.confirm({
      title: t('users.confirmUnlockTitle'),
      content: (
        <div>
          <p>{t('users.confirmUnlock', { name: record.display_name })}</p>
          {locked && (
            <p style={{ color: '#FF6B6B', fontSize: 13 }}>
              {t('users.lockRemainingTime', { time: formatLockTime(record.locked_until) })}
            </p>
          )}
          {hasFailCount && !locked && (
            <p style={{ color: '#FFBE0B', fontSize: 13 }}>
              {t('users.failedLoginCount', { count: record.login_fail_count })}
            </p>
          )}
        </div>
      ),
      okText: t('users.actions.unlock'),
      okButtonProps: { 
        type: 'primary',
        style: {
          background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
          border: 'none',
        },
      },
      onOk: async () => {
        try {
          await userService.unlockUser(record.id, t('users.unlockReason'));
          message.success(t('users.unlocked'));
          loadUsers();
        } catch {
          // handled by interceptor
        }
      },
    });
  };

  const roleTag = (role: string) => {
    const map: Record<string, { textKey: string; color: string; bg: string }> = {
      super_admin: { textKey: 'users.role.superAdmin', color: '#FF6B6B', bg: 'rgba(255, 107, 107, 0.15)' },
      dept_manager: { textKey: 'users.role.deptManager', color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
      user: { textKey: 'users.role.user', color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
    };
    const r = map[role] || { textKey: '', color: 'default', bg: 'transparent' };
    return (
      <Tag 
        style={{ 
          color: r.color,
          background: r.bg,
          border: `1px solid ${r.color}40`,
          borderRadius: 6,
        }}
      >
        {r.textKey ? t(r.textKey) : role}
      </Tag>
    );
  };

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
          {t('users.status.disabled')}
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
          {t('users.status.locked')} {formatLockTime(record.locked_until)}
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
          {t('users.status.enabledWithFailures', { count: record.login_fail_count })}
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
        {t('users.status.enabled')}
      </Tag>
    );
  };

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

  const columns: ColumnsType<UserDetail> = useMemo(() => [
    { 
      title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.username')}</span>, 
      dataIndex: 'username', 
      key: 'username', 
      width: 120,
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontWeight: 500 }}>{text}</span>,
    },
    { 
      title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.displayName')}</span>, 
      dataIndex: 'display_name', 
      key: 'display_name', 
      width: 120,
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{text}</span>,
    },
    { 
      title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.email')}</span>, 
      dataIndex: 'email', 
      key: 'email', 
      width: 180, 
      render: (v) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)' }}>{v || '-'}</span>,
    },
    { title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.role')}</span>, dataIndex: 'role', key: 'role', width: 120, render: roleTag },
    {
      title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.department')}</span>, 
      key: 'department', 
      width: 120,
      render: (_, r) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)' }}>{r.department?.name || '-'}</span>,
    },
    {
      title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.status')}</span>, 
      key: 'status', 
      width: 160,
      render: (_, record) => statusTag(record),
    },
    {
      title: <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{t('users.table.lastLogin')}</span>, 
      dataIndex: 'last_login_at', 
      key: 'last_login_at', 
      width: 160,
      render: (v: string) => (
        <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)' }}>
          {v ? dayjs(v).format('YYYY-MM-DD HH:mm') : '-'}
        </span>
      ),
    },
    {
      title: t('users.table.actions'), 
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
              {t('common.edit')}
            </Button>
            <Button
              type="link"
              size="small"
              icon={record.status === 1 ? <StopOutlined /> : <CheckCircleOutlined />}
              onClick={() => handleToggleStatus(record)}
              style={{ color: record.status === 1 ? '#FFBE0B' : '#00F5D4' }}
            >
              {record.status === 1 ? t('common.disable') : t('common.enable')}
            </Button>
            {(locked || hasFailCount) && (
              <Button 
                type="link" 
                size="small" 
                icon={<UnlockOutlined />}
                onClick={() => handleUnlock(record)}
                style={{ color: '#FF6B6B' }}
              >
                {t('users.actions.unlock')}
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
                {t('common.delete')}
              </Button>
            )}
          </Space>
        );
      },
    },
  ], [isDark, isSuperAdmin, t]);

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<UserOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                {t('users.title')}
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.45)', fontSize: 14, marginTop: 4 }}>
                {t('users.pageDescription')}
              </p>
            </div>
          </div>
          <div style={{ marginTop: 20, display: 'flex', gap: 12 }}>
            <Input.Search
              placeholder={t('users.search')}
              allowClear
              onSearch={(v) => setParams((p) => ({ ...p, keyword: v, page: 1 }))}
              style={{ width: 280 }}
            />
            <Button 
              icon={<ReloadOutlined />} 
              onClick={loadUsers}
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
                background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                border: 'none',
                boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
              }}
            >
              {t('users.createUser')}
            </Button>
          </div>
        </div>

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
              showTotal: (total) => t('common.totalRecords', { total }),
              onChange: (page, pageSize) => setParams((p) => ({ ...p, page, page_size: pageSize })),
            }}
          />
        </div>

        <Modal
          title={
            <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 18, fontWeight: 600 }}>
              {editingUser ? t('users.form.editTitle') : t('users.form.createTitle')}
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
                  label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.username')}</span>} 
                  rules={[{ required: true, message: t('users.form.usernameRequired') }]}
                >
                  <Input 
                    placeholder={t('users.form.usernamePlaceholder')}
                    style={{ 
                      background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                      color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                    }}
                  />
                </Form.Item>
                <Form.Item 
                  name="password" 
                  label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.initialPassword')}</span>} 
                  rules={[{ required: true, message: t('users.form.initialPasswordRequired') }, { min: 8, message: t('users.form.passwordMinLength') }]}
                >
                  <Input.Password 
                    placeholder={t('users.form.passwordPlaceholder')}
                    style={{ 
                      background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                      borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                      color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                    }}
                  />
                </Form.Item>
              </>
            )}
            <Form.Item 
              name="display_name" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.displayName')}</span>} 
              rules={[{ required: true, message: t('users.form.displayNameRequired') }]}
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
              name="email" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.email')}</span>}
            >
              <Input 
                type="email"
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                  color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                }}
              />
            </Form.Item>
            <Form.Item 
              name="phone" 
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.phone')}</span>}
            >
              <Input 
                style={{ 
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                  borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.15)',
                  color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                }}
              />
            </Form.Item>
            {!editingUser && (
              <Form.Item 
                name="role" 
                label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.role')}</span>} 
                rules={[{ required: true, message: t('users.form.roleRequired') }]}
              >
                <Select>
                  {isSuperAdmin && <Select.Option value="super_admin">{t('users.role.superAdmin')}</Select.Option>}
                  {isSuperAdmin && <Select.Option value="dept_manager">{t('users.role.deptManager')}</Select.Option>}
                  <Select.Option value="user">{t('users.role.user')}</Select.Option>
                </Select>
              </Form.Item>
            )}
            <Form.Item
              name="department_id"
              label={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{t('users.form.department')}</span>}
              rules={isDeptManager ? [{ required: true, message: t('users.form.departmentRequired') }] : undefined}
            >
              <Select
                placeholder={isDeptManager ? t('users.form.department') : t('users.form.departmentOptional')}
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
                {editingUser ? t('common.save') : t('common.create')}
              </Button>
            </Form.Item>
          </Form>
        </Modal>
      </div>
    </div>
  );
};

export default UsersPage;
