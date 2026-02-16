import request from './request';
import type { ApiResponse, StatsOverview, UsageResponse, RankingItem } from '@/types';

/** 获取用量总览 */
export function getOverview() {
  return request.get<ApiResponse<StatsOverview>>('/stats/overview');
}

/** 获取用量统计数据 */
export function getUsageStats(params: {
  period: string;
  start_date?: string;
  end_date?: string;
  user_id?: number;
  department_id?: number;
}) {
  return request.get<ApiResponse<UsageResponse>>('/stats/usage', { params });
}

/** 获取用量排行榜 */
export function getRanking(params: {
  type: 'user' | 'department';
  period: string;
  limit?: number;
}) {
  return request.get<ApiResponse<RankingItem[]>>('/stats/ranking', { params });
}
