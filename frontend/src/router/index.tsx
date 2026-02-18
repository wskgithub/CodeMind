import { createBrowserRouter, Navigate } from 'react-router-dom';
import AuthGuard from './AuthGuard';
import GuestGuard from './GuestGuard';
import DashboardLayout from '@/components/layout/DashboardLayout';
import LoginPage from '@/pages/login/LoginPage';
import HomePage from '@/pages/home/HomePage';
import DashboardPage from '@/pages/dashboard/DashboardPage';
import KeysPage from '@/pages/keys/KeysPage';
import ProfilePage from '@/pages/profile/ProfilePage';
import UsagePage from '@/pages/usage/UsagePage';
import UsersPage from '@/pages/admin/users/UsersPage';
import DepartmentsPage from '@/pages/admin/departments/DepartmentsPage';
import LimitsPage from '@/pages/admin/limits/LimitsPage';
import SystemPage from '@/pages/admin/system/SystemPage';
import McpPage from '@/pages/admin/mcp/McpPage';
import BackendsPage from '@/pages/admin/backends/BackendsPage';

/** 应用路由配置 */
const router = createBrowserRouter([
  {
    path: '/',
    element: <HomePage />,
  },
  {
    path: '/login',
    element: (
      <GuestGuard>
        <LoginPage />
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
      { index: true, element: <DashboardPage /> },
      { path: 'keys', element: <KeysPage /> },
      { path: 'usage', element: <UsagePage /> },
      { path: 'profile', element: <ProfilePage /> },
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
      { path: 'users', element: <UsersPage /> },
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
      { path: 'departments', element: <DepartmentsPage /> },
      { path: 'limits', element: <LimitsPage /> },
      { path: 'backends', element: <BackendsPage /> },
      { path: 'mcp', element: <McpPage /> },
      { path: 'system', element: <SystemPage /> },
    ],
  },
  // 未匹配路由重定向到首页
  {
    path: '*',
    element: <Navigate to="/" replace />,
  },
]);

export default router;
