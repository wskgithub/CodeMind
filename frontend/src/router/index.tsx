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

function Lazy(Component: React.LazyExoticComponent<React.ComponentType>) {
  return (
    <Suspense fallback={null}>
      <Component />
    </Suspense>
  );
}

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
  // admin and dept manager pages
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
  // super admin only pages
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
  // docs pages (requires auth)
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
  // fallback to home
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

export default router;
