import request from './request';
import type { ApiResponse, RateLimit, MyLimitResponse } from '@/types';

/** 获取限额配置列表 */
export function listLimits(params?: { target_type?: string; target_id?: number }) {
  return request.get<ApiResponse<RateLimit[]>>('/limits', { params });
}

/** 创建或更新限额配置 */
export function upsertLimit(data: {
  target_type: string;
  target_id: number;
  period: string;
  max_tokens: number;
  max_requests?: number;
  max_concurrency?: number;
  alert_threshold?: number;
}) {
  return request.put<ApiResponse<null>>('/limits', data);
}

/** 删除限额配置 */
export function deleteLimit(id: number) {
  return request.delete<ApiResponse<null>>(`/limits/${id}`);
}

/** 获取当前用户的限额信息 */
export function getMyLimits() {
  return request.get<ApiResponse<MyLimitResponse>>('/limits/my');
}
