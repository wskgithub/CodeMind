import request from './request';
import type { ApiResponse, DeptTree } from '@/types';

/** 创建/更新部门参数 */
export interface CreateDepartmentParams {
  name: string;
  description?: string;
  parent_id?: number;
  manager_id?: number;
}

/** 部门管理 API */
const departmentService = {
  /** 获取部门树形列表 */
  list() {
    return request.get<ApiResponse<DeptTree[]>>('/departments');
  },

  /** 获取部门详情 */
  getDetail(id: number) {
    return request.get<ApiResponse<DeptTree>>(`/departments/${id}`);
  },

  /** 创建部门 */
  create(data: CreateDepartmentParams) {
    return request.post<ApiResponse<DeptTree>>('/departments', data);
  },

  /** 更新部门 */
  update(id: number, data: Partial<CreateDepartmentParams>) {
    return request.put<ApiResponse<null>>(`/departments/${id}`, data);
  },

  /** 删除部门 */
  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/departments/${id}`);
  },
};

export default departmentService;
