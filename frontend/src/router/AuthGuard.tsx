import { Navigate, useLocation } from 'react-router-dom';
import useAuthStore from '@/store/authStore';

interface AuthGuardProps {
  children: React.ReactNode;
  /** 是否需要管理员权限（部门经理或超级管理员）*/
  requireAdmin?: boolean;
  /** 是否需要超级管理员权限 */
  requireSuperAdmin?: boolean;
}

/** 路由认证守卫：未登录跳转到登录页 */
const AuthGuard: React.FC<AuthGuardProps> = ({ 
  children, 
  requireAdmin = false,
  requireSuperAdmin = false,
}) => {
  const { isAuthenticated, user } = useAuthStore();
  const location = useLocation();

  // 未登录，重定向到登录页
  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // 需要超级管理员权限但不是超级管理员
  if (requireSuperAdmin && user?.role !== 'super_admin') {
    return <Navigate to="/dashboard" replace />;
  }

  // 需要管理员权限但不是管理员或部门经理
  if (requireAdmin && user?.role === 'user') {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
};

export default AuthGuard;
