import request from './request';

import type { ApiResponse, APIKey, APIKeyCreateResult } from '@/types';

/** API Key 管理 API */
const keyService = {
  /** 获取 Key 列表 */
  list() {
    return request.get<ApiResponse<APIKey[]>>('/keys');
  },

  /** 创建 Key */
  create(data: { name: string; expires_at?: string }) {
    return request.post<ApiResponse<APIKeyCreateResult>>('/keys', data);
  },

  /** 复制 Key（返回完整 Key） */
  copy(id: number) {
    return request.post<ApiResponse<{ key: string }>>(`/keys/${id}/copy`);
  },

  /** 切换 Key 状态 */
  updateStatus(id: number, status: number) {
    return request.put<ApiResponse<null>>(`/keys/${id}/status`, { status });
  },

  /** 删除 Key */
  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/keys/${id}`);
  },
};

export default keyService;
