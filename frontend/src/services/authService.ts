import request from './request';
import type { ApiResponse, LoginParams, LoginResult, UserDetail } from '@/types';

/** 认证相关 API */
const authService = {
  /** 用户登录 */
  login(params: LoginParams) {
    return request.post<ApiResponse<LoginResult>>('/auth/login', params);
  },

  /** 用户登出 */
  logout() {
    return request.post<ApiResponse<null>>('/auth/logout');
  },

  /** 获取当前用户信息 */
  getProfile() {
    return request.get<ApiResponse<UserDetail>>('/auth/profile');
  },

  /** 更新个人信息 */
  updateProfile(data: { display_name?: string; email?: string; phone?: string }) {
    return request.put<ApiResponse<null>>('/auth/profile', data);
  },

  /** 修改密码 */
  changePassword(data: { old_password: string; new_password: string }) {
    return request.put<ApiResponse<null>>('/auth/password', data);
  },
};

export default authService;
