import request from './request';

import type { ApiResponse, LoginParams, LoginResult, UserDetail } from '@/types';

const authService = {
  login(params: LoginParams) {
    return request.post<ApiResponse<LoginResult>>('/auth/login', params);
  },

  logout() {
    return request.post<ApiResponse<null>>('/auth/logout');
  },

  getProfile() {
    return request.get<ApiResponse<UserDetail>>('/auth/profile');
  },

  updateProfile(data: { display_name?: string; email?: string; phone?: string }) {
    return request.put<ApiResponse<null>>('/auth/profile', data);
  },

  changePassword(data: { old_password: string; new_password: string }) {
    return request.put<ApiResponse<null>>('/auth/password', data);
  },
};

export default authService;
