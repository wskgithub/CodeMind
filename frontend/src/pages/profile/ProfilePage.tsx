import { useState, useEffect } from 'react';
import { Form, Input, Button, Divider, message, Descriptions, Tag, Avatar, theme } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import type { UserDetail } from '@/types';
import useAuthStore from '@/store/authStore';
import authService from '@/services/authService';

/** 个人中心页面 — Glassmorphism 风格 */
const ProfilePage: React.FC = () => {
  const { token } = theme.useToken();
  const { fetchProfile } = useAuthStore();
  const [profile, setProfile] = useState<UserDetail | null>(null);
  const [editLoading, setEditLoading] = useState(false);
  const [pwdLoading, setPwdLoading] = useState(false);
  const [editForm] = Form.useForm();
  const [pwdForm] = Form.useForm();

  // 加载用户信息
  useEffect(() => {
    fetchProfile().then(setProfile);
  }, [fetchProfile]);

  // 更新个人信息
  const handleUpdateProfile = async (values: { display_name: string; email: string; phone: string }) => {
    setEditLoading(true);
    try {
      await authService.updateProfile(values);
      message.success('个人信息已更新');
      const updated = await fetchProfile();
      setProfile(updated);
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setEditLoading(false);
    }
  };

  // 修改密码
  const handleChangePassword = async (values: { old_password: string; new_password: string; confirm: string }) => {
    if (values.new_password !== values.confirm) {
      message.error('两次密码输入不一致');
      return;
    }
    setPwdLoading(true);
    try {
      await authService.changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success('密码修改成功');
      pwdForm.resetFields();
    } catch {
      // 错误已在拦截器中处理
    } finally {
      setPwdLoading(false);
    }
  };

  // 角色标签
  const roleLabel: Record<string, { text: string; color: string }> = {
    super_admin: { text: '超级管理员', color: 'red' },
    dept_manager: { text: '部门经理', color: 'blue' },
    user: { text: '普通用户', color: 'green' },
  };

  return (
    <div className="page-bg max-w-3xl">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        {/* 用户信息头部 — 玻璃态背景，头像与姓名 */}
        <div
          className="glass-card animate-fade-in-up p-6 mb-6"
          style={{ animationDelay: '0.05s' }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 20, flexWrap: 'wrap' }}>
            <Avatar
              size={72}
              icon={<UserOutlined />}
              style={{
                background: 'var(--gradient-primary)',
                flexShrink: 0,
              }}
            />
            <div>
              <h2 style={{ margin: 0, marginBottom: 4, color: token.colorTextHeading }}>
                {profile?.display_name || profile?.username || '-'}
              </h2>
              <span style={{ color: token.colorTextSecondary, fontSize: 14 }}>
                @{profile?.username || '-'}
              </span>
              {profile && (
                <div style={{ marginTop: 8 }}>
                  <Tag color={roleLabel[profile.role]?.color}>{roleLabel[profile.role]?.text}</Tag>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* 基本信息 — 玻璃态卡片 */}
        <div
          className="glass-card animate-fade-in-up p-6 mb-6"
          style={{ animationDelay: '0.08s' }}
        >
          <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>基本信息</h3>
          {profile && (
            <Descriptions column={2}>
              <Descriptions.Item label="用户名" labelStyle={{ color: token.colorTextSecondary }}>
                <span style={{ color: token.colorText }}>{profile.username}</span>
              </Descriptions.Item>
              <Descriptions.Item label="角色" labelStyle={{ color: token.colorTextSecondary }}>
                <Tag color={roleLabel[profile.role]?.color}>{roleLabel[profile.role]?.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="部门" labelStyle={{ color: token.colorTextSecondary }}>
                <span style={{ color: token.colorText }}>{profile.department?.name || '-'}</span>
              </Descriptions.Item>
              <Descriptions.Item label="注册时间" labelStyle={{ color: token.colorTextSecondary }}>
                <span style={{ color: token.colorText }}>
                  {new Date(profile.created_at).toLocaleDateString('zh-CN')}
                </span>
              </Descriptions.Item>
            </Descriptions>
          )}
        </div>

        {/* 编辑个人信息 — 玻璃态卡片 */}
        <div
          className="glass-card animate-fade-in-up p-6 mb-6"
          style={{ animationDelay: '0.1s' }}
        >
          <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>编辑个人信息</h3>
          {profile && (
            <Form
              form={editForm}
              layout="vertical"
              initialValues={{
                display_name: profile.display_name,
                email: profile.email || '',
                phone: profile.phone || '',
              }}
              onFinish={handleUpdateProfile}
            >
              <Form.Item name="display_name" label="显示名称" rules={[{ required: true, message: '请输入显示名称' }]}>
                <Input placeholder="您的显示名称" />
              </Form.Item>
              <Form.Item name="email" label="邮箱">
                <Input type="email" placeholder="example@email.com" />
              </Form.Item>
              <Form.Item name="phone" label="手机号">
                <Input placeholder="您的手机号" />
              </Form.Item>
              <Form.Item>
                <Button type="primary" htmlType="submit" loading={editLoading}>
                  保存
                </Button>
              </Form.Item>
            </Form>
          )}
        </div>

        {/* 修改密码 — 玻璃态卡片 */}
        <div
          className="glass-card animate-fade-in-up p-6"
          style={{ animationDelay: '0.15s' }}
        >
          <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>修改密码</h3>
          <Form form={pwdForm} layout="vertical" onFinish={handleChangePassword}>
            <Form.Item name="old_password" label="原密码" rules={[{ required: true, message: '请输入原密码' }]}>
              <Input.Password placeholder="请输入当前密码" />
            </Form.Item>
            <Divider style={{ margin: '16px 0' }} />
            <Form.Item
              name="new_password"
              label="新密码"
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码长度不能少于 8 位' },
              ]}
            >
              <Input.Password placeholder="至少 8 位字符" />
            </Form.Item>
            <Form.Item
              name="confirm"
              label="确认新密码"
              rules={[{ required: true, message: '请确认新密码' }]}
            >
              <Input.Password placeholder="再次输入新密码" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={pwdLoading}>
                修改密码
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
