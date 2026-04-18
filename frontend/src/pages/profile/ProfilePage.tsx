import { UserOutlined, MailOutlined, PhoneOutlined, LockOutlined, SafetyOutlined } from '@ant-design/icons';
import { Form, Input, Button, Divider, message, Descriptions, Tag, Avatar } from 'antd';
import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';

import authService from '@/services/authService';
import useAuthStore from '@/store/authStore';
import type { UserDetail } from '@/types';

const ProfilePage: React.FC = () => {
  const { t } = useTranslation();
  const fetchProfile = useAuthStore((s) => s.fetchProfile);
  const [profile, setProfile] = useState<UserDetail | null>(null);
  const [editLoading, setEditLoading] = useState(false);
  const [pwdLoading, setPwdLoading] = useState(false);
  const [editForm] = Form.useForm();
  const [pwdForm] = Form.useForm();

  useEffect(() => {
    fetchProfile().then(setProfile);
  }, [fetchProfile]);

  const handleUpdateProfile = async (values: { display_name: string; email: string; phone: string }) => {
    setEditLoading(true);
    try {
      await authService.updateProfile(values);
      message.success(t('profile.profileUpdated'));
      const updated = await fetchProfile();
      setProfile(updated);
    } catch {
      // handled by interceptor
    } finally {
      setEditLoading(false);
    }
  };

  const handleChangePassword = async (values: { old_password: string; new_password: string; confirm: string }) => {
    if (values.new_password !== values.confirm) {
      message.error(t('profile.form.passwordMismatch'));
      return;
    }
    setPwdLoading(true);
    try {
      await authService.changePassword({
        old_password: values.old_password,
        new_password: values.new_password,
      });
      message.success(t('profile.passwordChanged'));
      pwdForm.resetFields();
    } catch {
      // handled by interceptor
    } finally {
      setPwdLoading(false);
    }
  };

  const roleLabel: Record<string, { text: string; color: string; bg: string }> = {
    super_admin: { text: t('users.role.superAdmin'), color: '#FF6B6B', bg: 'rgba(255, 107, 107, 0.15)' },
    dept_manager: { text: t('users.role.deptManager'), color: '#00D9FF', bg: 'rgba(0, 217, 255, 0.15)' },
    user: { text: t('users.role.user'), color: '#00F5D4', bg: 'rgba(0, 245, 212, 0.15)' },
  };

  return (
    <div className="page-bg max-w-3xl">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        <div style={{ marginBottom: 24 }}>
          <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
            {t('profile.title')}
          </h2>
          <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
            {t('profile.pageDescription')}
          </p>
        </div>

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
            {t('profile.basicInfo')}
          </h3>
          {profile && (
            <Descriptions 
              column={2}
              labelStyle={{ color: 'rgba(255, 255, 255, 0.5)', fontSize: 14 }}
            >
              <Descriptions.Item label={t('profile.form.username')}>
                <span style={{ color: '#fff', fontSize: 15 }}>{profile.username}</span>
              </Descriptions.Item>
              <Descriptions.Item label={t('profile.form.role')}>
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
              <Descriptions.Item label={t('profile.form.department')}>
                <span style={{ color: '#fff', fontSize: 15 }}>{profile.department?.name || '-'}</span>
              </Descriptions.Item>
              <Descriptions.Item label={t('profile.form.registeredAt')}>
                <span style={{ color: '#fff', fontSize: 15 }}>
                  {new Date(profile.created_at).toLocaleDateString()}
                </span>
              </Descriptions.Item>
            </Descriptions>
          )}
        </div>

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
            {t('profile.editInfo')}
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
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{t('profile.form.displayName')}</span>} 
                rules={[{ required: true, message: t('profile.form.displayNameRequired') }]}
              >
                <Input 
                  placeholder={t('profile.form.displayNamePlaceholder')} 
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
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{t('profile.form.email')}</span>}
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
                label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{t('profile.form.phone')}</span>}
              >
                <Input 
                  placeholder={t('profile.form.phonePlaceholder')} 
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
                  {t('common.save')}
                </Button>
              </Form.Item>
            </Form>
          )}
        </div>

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
            {t('profile.changePassword')}
          </h3>
          <Form form={pwdForm} layout="vertical" onFinish={handleChangePassword}>
            <Form.Item 
              name="old_password" 
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{t('profile.form.oldPassword')}</span>} 
              rules={[{ required: true, message: t('profile.form.oldPasswordRequired') }]}
            >
              <Input.Password 
                placeholder={t('profile.form.oldPasswordPlaceholder')} 
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
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{t('profile.form.newPassword')}</span>}
              rules={[
                { required: true, message: t('profile.form.newPasswordRequired') },
                { min: 8, message: t('profile.form.passwordMinLength') },
              ]}
            >
              <Input.Password 
                placeholder={t('profile.form.newPasswordPlaceholder')} 
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
              label={<span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{t('profile.form.confirmPassword')}</span>}
              rules={[{ required: true, message: t('profile.form.confirmPasswordRequired') }]}
            >
              <Input.Password 
                placeholder={t('profile.form.confirmPasswordPlaceholder')} 
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
                {t('profile.changePassword')}
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
    </div>
  );
};

export default ProfilePage;
