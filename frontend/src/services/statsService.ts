import request from './request';

import type { ApiResponse, StatsOverview, UsageResponse, RankingItem, KeyUsageItem } from '@/types';

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

/** 获取 Key 用量汇总 */
export function getKeyUsageStats(params: {
  start_date?: string;
  end_date?: string;
}) {
  return request.get<ApiResponse<KeyUsageItem[]>>('/stats/key-usage', { params });
}

/** 导出租用量报表为 CSV */
export function exportUsageCSV(params: {
  period: string;
  start_date?: string;
  end_date?: string;
  user_id?: number;
  department_id?: number;
}) {
  return request.get('/stats/export/csv', {
    params,
    responseType: 'blob',
  });
}
