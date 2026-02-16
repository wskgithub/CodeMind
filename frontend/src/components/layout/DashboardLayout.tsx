import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Button, theme } from 'antd';
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
  SunOutlined,
  MoonOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import useAuthStore from '@/store/authStore';
import useAppStore from '@/store/appStore';

const { Header, Sider, Content } = Layout;

/** 管理后台布局 — Glassmorphism 风格 */
const DashboardLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { sidebarCollapsed, toggleSidebar, darkMode, toggleDarkMode } = useAppStore();
  const { token: themeToken } = theme.useToken();

  const isSuperAdmin = user?.role === 'super_admin';
  const isDeptManager = user?.role === 'dept_manager';
  const isAdmin = isSuperAdmin || isDeptManager;

  // 侧边栏菜单项
  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '总览',
    },
    {
      key: '/dashboard/keys',
      icon: <KeyOutlined />,
      label: 'API Key',
    },
    {
      key: '/dashboard/usage',
      icon: <BarChartOutlined />,
      label: '用量统计',
    },
    ...(isAdmin
      ? [
          { type: 'divider' as const },
          {
            key: '/admin/users',
            icon: <UserOutlined />,
            label: '用户管理',
          },
        ]
      : []),
    ...(isSuperAdmin
      ? [
          {
            key: '/admin/departments',
            icon: <TeamOutlined />,
            label: '部门管理',
          },
          {
            key: '/admin/limits',
            icon: <SafetyOutlined />,
            label: '限额管理',
          },
          {
            key: '/admin/mcp',
            icon: <ApiOutlined />,
            label: 'MCP 服务',
          },
          {
            key: '/admin/system',
            icon: <SettingOutlined />,
            label: '系统管理',
          },
        ]
      : []),
  ];

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人中心',
      onClick: () => navigate('/dashboard/profile'),
    },
    {
      key: 'usage',
      icon: <BarChartOutlined />,
      label: '用量统计',
      onClick: () => navigate('/dashboard/usage'),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
      onClick: async () => {
        await logout();
        navigate('/login');
      },
    },
  ];

  const selectedKey = location.pathname;

  return (
    <Layout style={{ minHeight: '100vh', background: darkMode ? '#0f0f13' : '#f0f4f8' }}>
      {/* 背景装饰光圈 */}
      <div
        style={{
          position: 'fixed',
          top: -200,
          right: -200,
          width: 600,
          height: 600,
          background: 'var(--orb-primary)',
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
          width: 500,
          height: 500,
          background: 'var(--orb-accent)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
        }}
      />

      {/* 玻璃效果侧边栏 */}
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        width={220}
        collapsedWidth={72}
        className="glass-sidebar"
        style={{
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 20,
          background: 'var(--sidebar-bg)',
          backdropFilter: 'blur(20px)',
          WebkitBackdropFilter: 'blur(20px)',
          borderRight: '1px solid var(--sidebar-border)',
        }}
      >
        {/* Logo */}
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            borderBottom: '1px solid var(--glass-border)',
            transition: 'all 0.3s',
          }}
          onClick={() => navigate('/')}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: 10,
                background: 'var(--gradient-primary)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#fff',
                fontWeight: 800,
                fontSize: 14,
                boxShadow: '0 4px 12px rgba(43, 124, 179, 0.3)',
              }}
            >
              CM
            </div>
            {!sidebarCollapsed && (
              <span
                style={{
                  fontSize: 16,
                  fontWeight: 700,
                  background: 'var(--gradient-primary)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  whiteSpace: 'nowrap',
                }}
              >
                CodeMind
              </span>
            )}
          </div>
        </div>

        {/* 导航菜单 */}
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{
            borderRight: 0,
            background: 'transparent',
            padding: '8px 0',
          }}
        />
      </Sider>

      {/* 右侧内容区 */}
      <Layout
        style={{
          marginLeft: sidebarCollapsed ? 72 : 220,
          transition: 'margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          background: 'transparent',
          position: 'relative',
          zIndex: 1,
        }}
      >
        {/* 玻璃效果顶栏 */}
        <Header
          className="glass-header"
          style={{
            padding: '0 24px',
            background: 'var(--header-bg)',
            backdropFilter: 'blur(var(--header-blur))',
            WebkitBackdropFilter: 'blur(var(--header-blur))',
            borderBottom: '1px solid var(--glass-border)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            position: 'sticky',
            top: 0,
            zIndex: 15,
            height: 64,
            lineHeight: '64px',
          }}
        >
          <Button
            type="text"
            icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={toggleSidebar}
            style={{
              fontSize: 16,
              width: 40,
              height: 40,
              borderRadius: 10,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          />

          <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            {/* 暗色模式切换 */}
            <Button
              type="text"
              icon={darkMode ? <SunOutlined /> : <MoonOutlined />}
              onClick={toggleDarkMode}
              style={{
                fontSize: 16,
                width: 40,
                height: 40,
                borderRadius: 10,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
              title={darkMode ? '切换到亮色模式' : '切换到暗色模式'}
            />

            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 8,
                  cursor: 'pointer',
                  padding: '6px 12px',
                  borderRadius: 12,
                  transition: 'all 0.2s',
                  background: 'transparent',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'var(--gradient-card-hover)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'transparent';
                }}
              >
                <Avatar
                  size={34}
                  style={{
                    background: 'var(--gradient-primary)',
                    boxShadow: '0 2px 8px rgba(43, 124, 179, 0.25)',
                  }}
                  icon={<UserOutlined />}
                />
                <span style={{ color: themeToken.colorText, fontSize: 14, fontWeight: 500 }}>
                  {user?.display_name || user?.username}
                </span>
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* 页面内容区 */}
        <Content
          style={{
            margin: 24,
            minHeight: 280,
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
