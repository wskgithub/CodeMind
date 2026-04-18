import request from './request';

import type { ApiResponse, RateLimit, MyLimitResponse, LimitProgressResponse } from '@/types';

export function listLimits(params?: { target_type?: string; target_id?: number }) {
  return request.get<ApiResponse<RateLimit[]>>('/limits', { params });
}

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

export function deleteLimit(id: number) {
  return request.delete<ApiResponse<null>>(`/limits/${id}`);
}

export function getMyLimits() {
  return request.get<ApiResponse<MyLimitResponse>>('/limits/my');
}

export function getMyLimitProgress() {
  return request.get<ApiResponse<LimitProgressResponse>>('/limits/my/progress');
}
