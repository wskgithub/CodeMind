import request from './request';

import type { ApiResponse, RateLimit, MyLimitResponse, LimitProgressResponse } from '@/types';

/** 获取限额配置列表 */
export function listLimits(params?: { target_type?: string; target_id?: number }) {
  return request.get<ApiResponse<RateLimit[]>>('/limits', { params });
}

/** 创建或更新限额配置 */
export function upsertLimit(data: {
  target_type: string;
  target_id: number;
  period: string;
  period_hours?: number;
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

/** 获取当前用户的限额信息（旧版兼容） */
export function getMyLimits() {
  return request.get<ApiResponse<MyLimitResponse>>('/limits/my');
}

/** 获取当前用户的限额进度（新版，含重置时间） */
export function getMyLimitProgress() {
  return request.get<ApiResponse<LimitProgressResponse>>('/limits/my/progress');
}
