import request from './request';

import type { ApiResponse, StatsOverview, UsageResponse, RankingItem, KeyUsageItem } from '@/types';

export function getOverview() {
  return request.get<ApiResponse<StatsOverview>>('/stats/overview');
}

export function getUsageStats(params: {
  period: string;
  start_date?: string;
  end_date?: string;
  user_id?: number;
  department_id?: number;
}) {
  return request.get<ApiResponse<UsageResponse>>('/stats/usage', { params });
}

export function getRanking(params: {
  type: 'user' | 'department';
  period: string;
  limit?: number;
}) {
  return request.get<ApiResponse<RankingItem[]>>('/stats/ranking', { params });
}

export function getKeyUsageStats(params: {
  start_date?: string;
  end_date?: string;
}) {
  return request.get<ApiResponse<KeyUsageItem[]>>('/stats/key-usage', { params });
}

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
