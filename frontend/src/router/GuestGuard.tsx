import { Navigate } from 'react-router-dom';

import useAuthStore from '@/store/authStore';

interface GuestGuardProps {
  children: React.ReactNode;
}

/** 访客守卫：已登录用户自动跳转到仪表盘 */
const GuestGuard: React.FC<GuestGuardProps> = ({ children }) => {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
};

export default GuestGuard;
