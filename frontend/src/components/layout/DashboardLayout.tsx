import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Button } from 'antd';
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
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import useAuthStore from '@/store/authStore';
import useAppStore from '@/store/appStore';

const { Header, Sider, Content } = Layout;

/** 管理后台布局 — 与首页/登录页新设计风格统一 */
const DashboardLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { sidebarCollapsed, toggleSidebar } = useAppStore();

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
    {
      key: '/docs',
      icon: <BookOutlined />,
      label: '接入文档',
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
            key: '/admin/backends',
            icon: <CloudServerOutlined />,
            label: 'LLM 节点',
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
          {
            key: '/admin/monitor',
            icon: <MonitorOutlined />,
            label: '系统监控',
          },
          {
            key: '/admin/docs',
            icon: <FileTextOutlined />,
            label: '文档管理',
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
    <Layout style={{ minHeight: '100vh', background: '#050d14' }}>
      {/* 背景装饰光圈 - 新设计 */}
      <div
        style={{
          position: 'fixed',
          top: -150,
          right: -150,
          width: 500,
          height: 500,
          background: 'radial-gradient(circle, rgba(0, 217, 255, 0.1) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(60px)',
        }}
      />
      <div
        style={{
          position: 'fixed',
          bottom: -100,
          left: -100,
          width: 400,
          height: 400,
          background: 'radial-gradient(circle, rgba(157, 78, 221, 0.08) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(60px)',
        }}
      />
      <div
        style={{
          position: 'fixed',
          top: '50%',
          left: '30%',
          width: 300,
          height: 300,
          background: 'radial-gradient(circle, rgba(0, 245, 212, 0.05) 0%, transparent 70%)',
          borderRadius: '50%',
          pointerEvents: 'none',
          zIndex: 0,
          filter: 'blur(50px)',
        }}
      />

      {/* 玻璃效果侧边栏 - 新设计 */}
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
          background: 'rgba(10, 22, 40, 0.95)',
          backdropFilter: 'blur(24px)',
          WebkitBackdropFilter: 'blur(24px)',
          borderRight: '1px solid rgba(0, 217, 255, 0.08)',
        }}
      >
        {/* Logo - 新设计 */}
        <div
          style={{
            height: 72,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            borderBottom: '1px solid rgba(0, 217, 255, 0.08)',
            transition: 'all 0.3s',
          }}
          onClick={() => navigate('/dashboard')}
        >
          {!sidebarCollapsed ? (
            <div style={{ position: 'relative' }}>
              {/* 发光背景效果 */}
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

        {/* 导航菜单 - 新设计 */}
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
          theme="dark"
        />
      </Sider>

      {/* 右侧内容区 */}
      <Layout
        style={{
          marginLeft: sidebarCollapsed ? 80 : 240,
          transition: 'margin-left 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
          background: 'transparent',
          position: 'relative',
          zIndex: 1,
        }}
      >
        {/* 玻璃效果顶栏 - 新设计 */}
        <Header
          className="glass-header"
          style={{
            padding: '0 24px',
            background: 'rgba(5, 13, 20, 0.85)',
            backdropFilter: 'blur(20px)',
            WebkitBackdropFilter: 'blur(20px)',
            borderBottom: '1px solid rgba(0, 217, 255, 0.06)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            position: 'sticky',
            top: 0,
            zIndex: 15,
            height: 72,
            lineHeight: '72px',
          }}
        >
          <Button
            type="text"
            icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={toggleSidebar}
            style={{
              fontSize: 18,
              width: 44,
              height: 44,
              borderRadius: 12,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'rgba(255, 255, 255, 0.7)',
              background: 'rgba(255, 255, 255, 0.03)',
              border: '1px solid rgba(255, 255, 255, 0.06)',
            }}
          />

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  cursor: 'pointer',
                  padding: '8px 16px',
                  borderRadius: 14,
                  transition: 'all 0.2s',
                  background: 'rgba(255, 255, 255, 0.02)',
                  border: '1px solid rgba(255, 255, 255, 0.06)',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.background = 'rgba(0, 217, 255, 0.08)';
                  e.currentTarget.style.borderColor = 'rgba(0, 217, 255, 0.2)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.background = 'rgba(255, 255, 255, 0.02)';
                  e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.06)';
                }}
              >
                <Avatar
                  size={36}
                  style={{
                    background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                    boxShadow: '0 2px 12px rgba(0, 217, 255, 0.3)',
                  }}
                  icon={<UserOutlined />}
                />
                <span style={{ color: 'rgba(255, 255, 255, 0.9)', fontSize: 14, fontWeight: 500 }}>
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
