import request from './request';
import type { ApiResponse, PageData, UserDetail } from '@/types';

export interface UserListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  department_id?: number;
  role?: string;
  status?: number;
}

export interface CreateUserParams {
  username: string;
  password: string;
  display_name: string;
  email?: string;
  phone?: string;
  role: string;
  department_id?: number;
}

const userService = {
  list(params: UserListParams) {
    return request.get<ApiResponse<PageData<UserDetail>>>('/users', { params });
  },

  getDetail(id: number) {
    return request.get<ApiResponse<UserDetail>>(`/users/${id}`);
  },

  create(data: CreateUserParams) {
    return request.post<ApiResponse<UserDetail>>('/users', data);
  },

  update(id: number, data: Partial<CreateUserParams>) {
    return request.put<ApiResponse<null>>(`/users/${id}`, data);
  },

  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/users/${id}`);
  },

  updateStatus(id: number, status: number) {
    return request.put<ApiResponse<null>>(`/users/${id}/status`, { status });
  },

  resetPassword(id: number, new_password: string) {
    return request.put<ApiResponse<null>>(`/users/${id}/reset-password`, { new_password });
  },

  unlockUser(id: number, reason?: string) {
    return request.put<ApiResponse<null>>(`/users/${id}/unlock`, { reason });
  },
};

export default userService;
