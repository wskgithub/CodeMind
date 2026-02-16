import request from './request';
import type { ApiResponse, PageData, UserDetail } from '@/types';

/** 用户列表查询参数 */
export interface UserListParams {
  page?: number;
  page_size?: number;
  keyword?: string;
  department_id?: number;
  role?: string;
  status?: number;
}

/** 创建/更新用户参数 */
export interface CreateUserParams {
  username: string;
  password: string;
  display_name: string;
  email?: string;
  phone?: string;
  role: string;
  department_id?: number;
}

/** 用户管理 API */
const userService = {
  /** 获取用户列表 */
  list(params: UserListParams) {
    return request.get<ApiResponse<PageData<UserDetail>>>('/users', { params });
  },

  /** 获取用户详情 */
  getDetail(id: number) {
    return request.get<ApiResponse<UserDetail>>(`/users/${id}`);
  },

  /** 创建用户 */
  create(data: CreateUserParams) {
    return request.post<ApiResponse<UserDetail>>('/users', data);
  },

  /** 更新用户 */
  update(id: number, data: Partial<CreateUserParams>) {
    return request.put<ApiResponse<null>>(`/users/${id}`, data);
  },

  /** 删除用户 */
  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/users/${id}`);
  },

  /** 切换用户状态 */
  updateStatus(id: number, status: number) {
    return request.put<ApiResponse<null>>(`/users/${id}/status`, { status });
  },

  /** 重置密码 */
  resetPassword(id: number, new_password: string) {
    return request.put<ApiResponse<null>>(`/users/${id}/reset-password`, { new_password });
  },
};

export default userService;
