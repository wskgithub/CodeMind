import request from './request';
import type { ApiResponse, APIKey, APIKeyCreateResult } from '@/types';

const keyService = {
  list() {
    return request.get<ApiResponse<APIKey[]>>('/keys');
  },

  create(data: { name: string; expires_at?: string }) {
    return request.post<ApiResponse<APIKeyCreateResult>>('/keys', data);
  },

  copy(id: number) {
    return request.post<ApiResponse<{ key: string }>>(`/keys/${id}/copy`);
  },

  updateStatus(id: number, status: number) {
    return request.put<ApiResponse<null>>(`/keys/${id}/status`, { status });
  },

  delete(id: number) {
    return request.delete<ApiResponse<null>>(`/keys/${id}`);
  },
};

export default keyService;
