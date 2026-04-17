import request from './request';

import type { ApiResponse, DeptTree } from '@/types';

export interface CreateDepartmentParams {
  name: string;
  description?: string;
  parent_id?: number;
  manager_id?: number;
}

const departmentService = {
  list() {
    return request.get<ApiResponse<DeptTree[]>>('/departments');
  },

  getDetail(id: number) {
    return request.get<ApiResponse<DeptTree>>(`/departments/${id}`);
  },

  create(data: CreateDepartmentParams) {
    return request.post<ApiResponse<DeptTree>>('/departments', data);
  },

  update(id: number, data: Partial<CreateDepartmentParams>) {
    return request.put<ApiResponse<null>>(`/departments/${id}`, data);
  },

  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/departments/${id}`);
  },
};

export default departmentService;
