import { useMemo } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Button, Tooltip } from 'antd';
import {
  DashboardOutlined,
  KeyOutlined,
  BarChartOutlined,
  UserOutlined,
  TeamOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  SafetyOutlined,
  ApiOutlined,
  CloudServerOutlined,
  MonitorOutlined,
  BookOutlined,
  FileTextOutlined,
  DatabaseOutlined,
  MoonOutlined,
  SunOutlined,
  AppstoreOutlined,
  BlockOutlined,
  GlobalOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import { useTranslation } from 'react-i18next';
import useAuthStore from '@/store/authStore';
import useAppStore from '@/store/appStore';
import { type SupportedLanguage } from '@/i18n';

const { Header, Sider, Content } = Layout;

const DashboardLayout: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const sidebarCollapsed = useAppStore((s) => s.sidebarCollapsed);
  const toggleSidebar = useAppStore((s) => s.toggleSidebar);
  const themeMode = useAppStore((s) => s.themeMode);
  const toggleTheme = useAppStore((s) => s.toggleTheme);
  const language = useAppStore((s) => s.language);
  const setLanguage = useAppStore((s) => s.setLanguage);

  const isSuperAdmin = user?.role === 'super_admin';
  const isDeptManager = user?.role === 'dept_manager';
  const isAdmin = isSuperAdmin || isDeptManager;

  const menuItems: MenuProps['items'] = useMemo(() => [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: t('menu.overview'),
    },
    {
      key: '/dashboard/keys',
      icon: <KeyOutlined />,
      label: t('menu.apiKey'),
    },
    {
      key: '/dashboard/usage',
      icon: <BarChartOutlined />,
      label: t('menu.usage'),
    },
    {
      key: '/dashboard/models',
      icon: <AppstoreOutlined />,
      label: t('menu.models'),
    },
    {
      key: '/docs',
      icon: <BookOutlined />,
      label: t('menu.docs'),
    },
    ...(isAdmin
      ? [
          { type: 'divider' as const },
          {
            key: '/admin/users',
            icon: <UserOutlined />,
            label: t('menu.users'),
          },
        ]
      : []),
    ...(isSuperAdmin
      ? [
          {
            key: '/admin/departments',
            icon: <TeamOutlined />,
            label: t('menu.departments'),
          },
          {
            key: '/admin/limits',
            icon: <SafetyOutlined />,
            label: t('menu.limits'),
          },
          {
            key: '/admin/backends',
            icon: <CloudServerOutlined />,
            label: t('menu.backends'),
          },
          {
            key: '/admin/templates',
            icon: <BlockOutlined />,
            label: t('menu.templates'),
          },
          {
            key: '/admin/mcp',
            icon: <ApiOutlined />,
            label: t('menu.mcp'),
          },
          {
            key: '/admin/platform',
            icon: <GlobalOutlined />,
            label: t('menu.platform'),
          },
          {
            key: '/admin/system',
            icon: <SettingOutlined />,
            label: t('menu.system'),
          },
          {
            key: '/admin/monitor',
            icon: <MonitorOutlined />,
            label: t('menu.monitor'),
          },
          {
            key: '/admin/docs',
            icon: <FileTextOutlined />,
            label: t('menu.adminDocs'),
          },
          {
            key: '/admin/training',
            icon: <DatabaseOutlined />,
            label: t('menu.training'),
          },
        ]
      : []),
  ], [isAdmin, isSuperAdmin, t]);

  const userMenuItems: MenuProps['items'] = useMemo(() => [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('userMenu.profile'),
      onClick: () => navigate('/dashboard/profile'),
    },
    {
      key: 'usage',
      icon: <BarChartOutlined />,
      label: t('userMenu.usage'),
      onClick: () => navigate('/dashboard/usage'),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('userMenu.logout'),
      danger: true,
      onClick: async () => {
        await logout();
        navigate('/login');
      },
    },
  ], [navigate, logout, t]);

  const selectedKey = location.pathname;

  return (
    <Layout style={{ height: '100vh', overflow: 'hidden', background: 'var(--bg-primary)' }}>
      <div
        style={{
          position: 'fixed',
          top: -200,
          right: -200,
          width: 800,
          height: 800,
          background: 'radial-gradient(circle, rgba(0, 217, 255, 0.08) 0%, rgba(0, 217, 255, 0.03) 30%, rgba(0, 217, 255, 0.008) 55%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
        }}
      />
      <div
        style={{
          position: 'fixed',
          bottom: -150,
          left: -150,
          width: 650,
          height: 650,
          background: 'radial-gradient(circle, rgba(157, 78, 221, 0.06) 0%, rgba(157, 78, 221, 0.025) 30%, rgba(157, 78, 221, 0.006) 55%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
        }}
      />
      <div
        style={{
          position: 'fixed',
          top: '50%',
          left: '30%',
          width: 500,
          height: 500,
          background: 'radial-gradient(circle, rgba(0, 245, 212, 0.04) 0%, rgba(0, 245, 212, 0.015) 30%, rgba(0, 245, 212, 0.004) 55%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
        }}
      />

      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        width={240}
        collapsedWidth={80}
        className="glass-sidebar"
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 20,
          background: themeMode === 'dark' ? 'rgba(10, 22, 40, 0.95)' : 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(24px)',
          WebkitBackdropFilter: 'blur(24px)',
          borderRight: `1px solid ${themeMode === 'dark' ? 'rgba(0, 217, 255, 0.08)' : 'rgba(0, 217, 255, 0.15)'}`,
        }}
      >
        <div
          style={{
            height: 72,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            borderBottom: `1px solid ${themeMode === 'dark' ? 'rgba(0, 217, 255, 0.08)' : 'rgba(0, 217, 255, 0.15)'}`,
            transition: 'all 0.3s',
          }}
          onClick={() => navigate('/dashboard')}
        >
          {!sidebarCollapsed ? (
            <div style={{ position: 'relative' }}>
              <div
                style={{
                  position: 'absolute',
                  top: '50%',
                  left: '50%',
                  transform: 'translate(-50%, -50%)',
                  width: 120,
                  height: 40,
                  background: 'radial-gradient(ellipse, rgba(0, 217, 255, 0.15) 0%, transparent 70%)',
                  filter: 'blur(8px)',
                  pointerEvents: 'none',
                }}
              />
              <span
                style={{
                  fontSize: 24,
                  fontWeight: 900,
                  background: 'linear-gradient(135deg, #ffffff 0%, #00D9FF 30%, #9D4EDD 70%, #00F5D4 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                  whiteSpace: 'nowrap',
                  letterSpacing: 1,
                  position: 'relative',
                  textShadow: '0 0 30px rgba(0, 217, 255, 0.3)',
                }}
              >
                {'<'}CodeMind{'/>'}
              </span>
            </div>
          ) : (
            <div
              style={{
                width: 44,
                height: 44,
                borderRadius: 14,
                background: 'linear-gradient(135deg, rgba(0, 217, 255, 0.15) 0%, rgba(157, 78, 221, 0.15) 100%)',
                border: '1.5px solid rgba(0, 217, 255, 0.3)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                boxShadow: '0 4px 20px rgba(0, 217, 255, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.1)',
                backdropFilter: 'blur(8px)',
              }}
            >
              <span
                style={{
                  fontSize: 18,
                  fontWeight: 800,
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                  letterSpacing: -1,
                }}
              >
                CM
              </span>
            </div>
          )}
        </div>

        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{
            borderRight: 0,
            background: 'transparent',
            padding: '12px 0',
          }}
          theme={themeMode === 'dark' ? 'dark' : 'light'}
        />
      </Sider>

      <Layout
        style={{
          marginLeft: sidebarCollapsed ? 80 : 240,
          transition: 'margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          background: 'transparent',
          position: 'relative',
          zIndex: 1,
          height: '100vh',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
        }}
      >
        <Header
          className="glass-header"
          style={{
            padding: '0 24px',
            background: themeMode === 'dark' ? 'rgba(5, 13, 20, 0.85)' : 'rgba(255, 255, 255, 0.85)',
            backdropFilter: 'blur(20px)',
            WebkitBackdropFilter: 'blur(20px)',
            borderBottom: `1px solid ${themeMode === 'dark' ? 'rgba(0, 217, 255, 0.06)' : 'rgba(0, 217, 255, 0.1)'}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            zIndex: 15,
            height: 72,
            lineHeight: '72px',
            flexShrink: 0,
          }}
        >
          <Button
            type="text"
            icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={toggleSidebar}
            style={{
              fontSize: 18,
              width: 48,
              height: 48,
              borderRadius: 12,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)',
              background: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
              border: `1px solid ${themeMode === 'dark' ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.06)'}`,
            }}
          />

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            {/* 语言切换 */}
            <div
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                borderRadius: 10,
                padding: 2,
                background: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.03)',
                border: `1px solid ${themeMode === 'dark' ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)'}`,
                gap: 2,
              }}
            >
              {(['zh-CN', 'en-US'] as SupportedLanguage[]).map((code) => {
                const isActive = language === code;
                const isDark = themeMode === 'dark';
                return (
                  <button
                    key={code}
                    onClick={() => setLanguage(code)}
                    style={{
                      padding: '4px 12px',
                      fontSize: 13,
                      fontWeight: isActive ? 600 : 400,
                      color: isActive
                        ? (isDark ? '#00D9FF' : '#0078d4')
                        : (isDark ? 'rgba(255,255,255,0.4)' : 'rgba(0,0,0,0.35)'),
                      background: isActive
                        ? (isDark ? 'rgba(0,217,255,0.12)' : 'rgba(0,217,255,0.08)')
                        : 'transparent',
                      border: isActive
                        ? `1px solid ${isDark ? 'rgba(0,217,255,0.2)' : 'rgba(0,217,255,0.15)'}`
                        : '1px solid transparent',
                      cursor: 'pointer',
                      borderRadius: 8,
                      transition: 'all 0.25s ease',
                      whiteSpace: 'nowrap' as const,
                      lineHeight: '1.4',
                    }}
                  >
                    {code === 'zh-CN' ? '中文' : 'EN'}
                  </button>
                );
              })}
            </div>

            {/* 主题切换 */}
            <Tooltip title={themeMode === 'dark' ? t('theme.toggleLight') : t('theme.toggleDark')}>
              <Button
                type="text"
                icon={themeMode === 'dark' ? <SunOutlined /> : <MoonOutlined />}
                onClick={toggleTheme}
                style={{
                  fontSize: 18,
                  width: 48,
                  height: 48,
                  borderRadius: 12,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)',
                  background: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  border: `1px solid ${themeMode === 'dark' ? 'rgba(255, 255, 255, 0.06)' : 'rgba(0, 0, 0, 0.06)'}`,
                }}
              />
            </Tooltip>
            
            <Dropdown
              menu={{ items: userMenuItems, style: { borderRadius: 0 } }}
              placement="bottomRight"
              overlayStyle={{ padding: '8px 12px', borderRadius: 0 }}
            >
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  cursor: 'pointer',
                  padding: '6px 14px',
                  borderRadius: 12,
                  transition: 'all 0.2s',
                  background: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)',
                  border: 'none',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'rgba(0, 217, 255, 0.08)';
                  e.currentTarget.style.boxShadow = '0 0 0 1px rgba(0, 217, 255, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = themeMode === 'dark' ? 'rgba(255, 255, 255, 0.02)' : 'rgba(0, 0, 0, 0.02)';
                  e.currentTarget.style.boxShadow = 'none';
                }}
              >
                <Avatar
                  size={32}
                  style={{
                    background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                    boxShadow: '0 2px 12px rgba(0, 217, 255, 0.3)',
                  }}
                  icon={<UserOutlined />}
                />
                <span style={{ color: themeMode === 'dark' ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)', fontSize: 14, fontWeight: 500 }}>
                  {user?.display_name || user?.username}
                </span>
              </div>
            </Dropdown>
          </div>
        </Header>

        <Content
          style={{
            padding: 24,
            flex: 1,
            overflowY: 'auto',
            overflowX: 'hidden',
            position: 'relative',
            zIndex: 1,
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
};

export default DashboardLayout;
