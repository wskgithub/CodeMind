import { lazy, Suspense } from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import AuthGuard from './AuthGuard';
import GuestGuard from './GuestGuard';
import DashboardLayout from '@/components/layout/DashboardLayout';

const LoginPage = lazy(() => import('@/pages/login/LoginPage'));
const HomePage = lazy(() => import('@/pages/home/HomePage'));
const DashboardPage = lazy(() => import('@/pages/dashboard/DashboardPage'));
const KeysPage = lazy(() => import('@/pages/keys/KeysPage'));
const ProfilePage = lazy(() => import('@/pages/profile/ProfilePage'));
const UsagePage = lazy(() => import('@/pages/usage/UsagePage'));
const ModelsPage = lazy(() => import('@/pages/models/ModelsPage'));
const UsersPage = lazy(() => import('@/pages/admin/users/UsersPage'));
const DepartmentsPage = lazy(() => import('@/pages/admin/departments/DepartmentsPage'));
const LimitsPage = lazy(() => import('@/pages/admin/limits/LimitsPage'));
const SystemPage = lazy(() => import('@/pages/admin/system/SystemPage'));
const McpPage = lazy(() => import('@/pages/admin/mcp/McpPage'));
const BackendsPage = lazy(() => import('@/pages/admin/backends/BackendsPage'));
const ProviderTemplatesPage = lazy(() => import('@/pages/admin/templates/ProviderTemplatesPage'));
const MonitorPage = lazy(() => import('@/pages/admin/monitor/MonitorPage'));
const DocsPage = lazy(() => import('@/pages/docs/DocsPage'));
const DocsAdminPage = lazy(() => import('@/pages/admin/docs/DocsAdminPage'));
const DocsEditPage = lazy(() => import('@/pages/admin/docs/DocsEditPage'));
const TrainingDataPage = lazy(() => import('@/pages/admin/training/TrainingDataPage'));
const PlatformSettingsPage = lazy(() => import('@/pages/admin/platform/PlatformSettingsPage'));

/** 懒加载组件包裹 Suspense */
function Lazy(Component: React.LazyExoticComponent<React.ComponentType>) {
  return (
    <Suspense fallback={null}>
      <Component />
    </Suspense>
  );
}

/** 应用路由配置 */
const router = createBrowserRouter([
  {
    path: '/',
    element: Lazy(HomePage),
  },
  {
    path: '/login',
    element: (
      <GuestGuard>
        {Lazy(LoginPage)}
      </GuestGuard>
    ),
  },
  {
    path: '/dashboard',
    element: (
      <AuthGuard>
        <DashboardLayout />
      </AuthGuard>
    ),
    children: [
      { index: true, element: Lazy(DashboardPage) },
      { path: 'keys', element: Lazy(KeysPage) },
      { path: 'usage', element: Lazy(UsagePage) },
      { path: 'models', element: Lazy(ModelsPage) },
      { path: 'profile', element: Lazy(ProfilePage) },
    ],
  },
  // 管理员和部门经理可访问的页面
  {
    path: '/admin',
    element: (
      <AuthGuard requireAdmin>
        <DashboardLayout />
      </AuthGuard>
    ),
    children: [
      { index: true, element: <Navigate to="/admin/users" replace /> },
      { path: 'users', element: Lazy(UsersPage) },
    ],
  },
  // 只有超级管理员可访问的页面
  {
    path: '/admin',
    element: (
      <AuthGuard requireSuperAdmin>
        <DashboardLayout />
      </AuthGuard>
    ),
    children: [
      { path: 'departments', element: Lazy(DepartmentsPage) },
      { path: 'limits', element: Lazy(LimitsPage) },
      { path: 'backends', element: Lazy(BackendsPage) },
      { path: 'templates', element: Lazy(ProviderTemplatesPage) },
      { path: 'mcp', element: Lazy(McpPage) },
      { path: 'system', element: Lazy(SystemPage) },
      { path: 'monitor', element: Lazy(MonitorPage) },
      { path: 'docs', element: Lazy(DocsAdminPage) },
      { path: 'docs/create', element: Lazy(DocsEditPage) },
      { path: 'docs/edit/:id', element: Lazy(DocsEditPage) },
      { path: 'training', element: Lazy(TrainingDataPage) },
      { path: 'platform', element: Lazy(PlatformSettingsPage) },
    ],
  },
  // 文档页面（需要登录）
  {
    path: '/docs',
    element: (
      <AuthGuard>
        <DashboardLayout />
      </AuthGuard>
    ),
    children: [
      { index: true, element: Lazy(DocsPage) },
      { path: ':slug', element: Lazy(DocsPage) },
    ],
  },
  // 未匹配路由重定向到首页
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

export default router;
