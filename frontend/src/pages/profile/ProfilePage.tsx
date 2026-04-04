import { useState, useEffect } from 'react';
import { Form, Input, Button, Divider, message, Descriptions, Tag, Avatar } from 'antd';
import { UserOutlined, MailOutlined, PhoneOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons';
import type { UserDetail } from '@/types';
import useAuthStore from '@/store/authStore';
import authService from '@/services/authService';

/** 个人中心页面 — 与首页/登录页新设计风格统一 */
const ProfilePage: React.FC = () => {
  const fetchProfile = useAuthStore((s) => s.fetchProfile);
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

  // 角色标签 - 新设计
  const roleLabel: Record<string, { text: string; color: string; bg: string }> = {
    super_admin: { text: '超级管理员', color: '#FF6B6B', bg: 'rgba(255, 107, 107, 0.15)' },
    dept_manager: { text: '部门经理', color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
    user: { text: '普通用户', color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
  };

  return (
    <div className="page-bg max-w-3xl">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面标题 */}
        <div style={{ marginBottom: 24 }}>
          <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
            个人中心
          </h2>
          <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
            管理您的个人信息和账户安全设置
          </p>
        </div>

        {/* 用户信息头部 — 玻璃态背景，头像与姓名 - 新设计 */}
        <div
          className="glass-card animate-fade-in-up p-6 mb-6"
          style={{ animationDelay: '0.05s' }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 24, flexWrap: 'wrap' }}>
            <Avatar
              size={80}
              icon={<UserOutlined style={{ fontSize: 36 }} />}
              style={{
                background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                flexShrink: 0,
                boxShadow: '0 4px 20px rgba(0, 217, 255, 0.3)',
              }}
            />
            <div>
              <h2 style={{ margin: 0, marginBottom: 6, color: '#fff', fontSize: 24, fontWeight: 600 }}>
                {profile?.display_name || profile?.username || '-'}
              </h2>
              <span style={{ color: 'rgba(255, 255, 255, 0.5)', fontSize: 15 }}>
                @{profile?.username || '-'}
              </span>
              {profile && (
                <div style={{ marginTop: 12 }}>
                  <Tag 
                    style={{ 
                      color: roleLabel[profile.role]?.color,
                      background: roleLabel[profile.role]?.bg,
                      border: `1px solid ${roleLabel[profile.role]?.color}40`,
                      borderRadius: 8,
                      padding: '4px 12px',
                      fontSize: 13,
                    }}
                  >
                    {roleLabel[profile.role]?.text}
                  </Tag>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* 基本信息 — 玻璃态卡片 - 新设计 */}
        <div
          className="glass-card animate-fade-in-up p-6 mb-6"
          style={{ animationDelay: '0.08s' }}
        >
          <h3 style={{ 
            marginBottom: 24, 
            color: '#fff',
            fontSize: 18,
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: 8,
          }}>
            <span style={{
              width: 4,
              height: 20,
              background: 'linear-gradient(180deg, #00D9FF 0%, #9D4EDD 100%)',
              borderRadius: 2,
            }} />
            基本信息
          </h3>
          {profile && (
            <Descriptions 
              column={2}
              labelStyle={{ color: 'rgba(255, 255, 255, 0.5)', fontSize: 14 }}
            >
              <Descriptions.Item label="用户名">
                <span style={{ color: '#fff', fontSize: 15 }}>{profile.username}</span>
              </Descriptions.Item>
              <Descriptions.Item label="角色">
                <Tag 
                  style={{ 
                    color: roleLabel[profile.role]?.color,
                    background: roleLabel[profile.role]?.bg,
                    border: `1px solid ${roleLabel[profile.role]?.color}40`,
                    borderRadius: 6,
                  }}
                >
                  {roleLabel[profile.role]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="部门">
                <span style={{ color: '#fff', fontSize: 15 }}>{profile.department?.name || '-'}</span>
              </Descriptions.Item>
              <Descriptions.Item label="注册时间">
                <span style={{ color: '#fff', fontSize: 15 }}>
                  {new Date(profile.created_at).toLocaleDateString('zh-CN')}
                </span>
              </Descriptions.Item>
            </Descriptions>
          )}
        </div>

        {/* 编辑个人信息 — 玻璃态卡片 - 新设计 */}
        <div
          className="glass-card animate-fade-in-up p-6 mb-6"
          style={{ animationDelay: '0.1s' }}
        >
          <h3 style={{ 
            marginBottom: 24, 
            color: '#fff',
            fontSize: 18,
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: 8,
          }}>
            <span style={{
              width: 4,
              height: 20,
              background: 'linear-gradient(180deg, #00F5D4 0%, #00D9FF 100%)',
              borderRadius: 2,
            }} />
            编辑个人信息
          </h3>
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
              <Form.Item 
                name="display_name" 
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>显示名称</span>} 
                rules={[{ required: true, message: '请输入显示名称' }]}
              >
                <Input 
                  placeholder="您的显示名称" 
                  prefix={<UserOutlined style={{ color: 'rgba(255, 255, 255, 0.4)' }} />}
                  style={{ 
                    height: 44,
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
                  placeholder="example@email.com" 
                  prefix={<MailOutlined style={{ color: 'rgba(255, 255, 255, 0.4)' }} />}
                  style={{ 
                    height: 44,
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
                  placeholder="您的手机号" 
                  prefix={<PhoneOutlined style={{ color: 'rgba(255, 255, 255, 0.4)' }} />}
                  style={{ 
                    height: 44,
                    background: 'rgba(255, 255, 255, 0.03)',
                    borderColor: 'rgba(255, 255, 255, 0.1)',
                    color: '#fff',
                  }}
                />
              </Form.Item>
              <Form.Item>
                <Button 
                  type="primary" 
                  htmlType="submit" 
                  loading={editLoading}
                  style={{
                    height: 44,
                    borderRadius: 12,
                    background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                    border: 'none',
                    boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                    padding: '0 32px',
                  }}
                >
                  保存
                </Button>
              </Form.Item>
            </Form>
          )}
        </div>

        {/* 修改密码 — 玻璃态卡片 - 新设计 */}
        <div
          className="glass-card animate-fade-in-up p-6"
          style={{ animationDelay: '0.15s' }}
        >
          <h3 style={{ 
            marginBottom: 24, 
            color: '#fff',
            fontSize: 18,
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: 8,
          }}>
            <span style={{
              width: 4,
              height: 20,
              background: 'linear-gradient(180deg, #FFBE0B 0%, #FF6B6B 100%)',
              borderRadius: 2,
            }} />
            修改密码
          </h3>
          <Form form={pwdForm} layout="vertical" onFinish={handleChangePassword}>
            <Form.Item 
              name="old_password" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>原密码</span>} 
              rules={[{ required: true, message: '请输入原密码' }]}
            >
              <Input.Password 
                placeholder="请输入当前密码" 
                prefix={<LockOutlined style={{ color: 'rgba(255, 255, 255, 0.4)' }} />}
                style={{ 
                  height: 44,
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            <Divider style={{ margin: '20px 0', borderColor: 'rgba(255, 255, 255, 0.08)' }} />
            <Form.Item
              name="new_password"
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>新密码</span>}
              rules={[
                { required: true, message: '请输入新密码' },
                { min: 8, message: '密码长度不能少于 8 位' },
              ]}
            >
              <Input.Password 
                placeholder="至少 8 位字符" 
                prefix={<SafetyOutlined style={{ color: 'rgba(255, 255, 255, 0.4)' }} />}
                style={{ 
                  height: 44,
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            <Form.Item
              name="confirm"
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>确认新密码</span>}
              rules={[{ required: true, message: '请确认新密码' }]}
            >
              <Input.Password 
                placeholder="再次输入新密码" 
                prefix={<SafetyOutlined style={{ color: 'rgba(255, 255, 255, 0.4)' }} />}
                style={{ 
                  height: 44,
                  background: 'rgba(255, 255, 255, 0.03)',
                  borderColor: 'rgba(255, 255, 255, 0.1)',
                  color: '#fff',
                }}
              />
            </Form.Item>
            <Form.Item>
              <Button 
                type="primary" 
                htmlType="submit" 
                loading={pwdLoading}
                style={{
                  height: 44,
                  borderRadius: 12,
                  background: 'linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)',
                  border: 'none',
                  boxShadow: '0 4px 16px rgba(255, 190, 11, 0.25)',
                  padding: '0 32px',
                }}
              >
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
