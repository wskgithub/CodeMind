import { describe, it, expect, beforeEach, vi } from 'vitest';

import request from '../request';
import {
  getOverview,
  getUsageStats,
  getRanking,
  exportUsageCSV,
} from '../statsService';

import type { ApiResponse, StatsOverview, UsageResponse, RankingItem } from '@/types';

// Mock request module
vi.mock('../request', () => ({
  default: {
    get: vi.fn(),
  },
}));


const mockRequest = request as unknown as {
  get: ReturnType<typeof vi.fn>;
};

describe('statsService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockStatsOverview: StatsOverview = {
    today: {
      total_tokens: 100000,
      total_requests: 500,
      active_users: 10,
    },
    this_month: {
      total_tokens: 2000000,
      total_requests: 10000,
      active_users: 50,
    },
    total_users: 100,
    total_departments: 10,
    total_api_keys: 20,
    system_status: 'healthy',
  };

  describe('getOverview', () => {
    it('should get stats overview', async () => {
      const mockResponse: ApiResponse<StatsOverview> = {
        code: 0,
        message: 'success',
        data: mockStatsOverview,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getOverview();

      expect(mockRequest.get).toHaveBeenCalledWith('/stats/overview');
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getUsageStats', () => {
    const mockUsageResponse: UsageResponse = {
      period: 'day',
      items: [
        {
          date: '2024-01-01',
          prompt_tokens: 50000,
          completion_tokens: 50000,
          total_tokens: 100000,
          request_count: 500,
        },
        {
          date: '2024-01-02',
          prompt_tokens: 60000,
          completion_tokens: 60000,
          total_tokens: 120000,
          request_count: 600,
        },
      ],
    };

    it.each([
      {
        desc: 'period only',
        params: { period: 'day' },
        expectedParams: { period: 'day' },
      },
      {
        desc: 'with date range',
        params: { period: 'day', start_date: '2024-01-01', end_date: '2024-01-31' },
        expectedParams: { period: 'day', start_date: '2024-01-01', end_date: '2024-01-31' },
      },
      {
        desc: 'with user filter',
        params: { period: 'week', user_id: 1 },
        expectedParams: { period: 'week', user_id: 1 },
      },
      {
        desc: 'with department filter',
        params: { period: 'month', department_id: 5 },
        expectedParams: { period: 'month', department_id: 5 },
      },
      {
        desc: 'with combined filters',
        params: { period: 'day', start_date: '2024-01-01', end_date: '2024-01-31', user_id: 1, department_id: 5 },
        expectedParams: { period: 'day', start_date: '2024-01-01', end_date: '2024-01-31', user_id: 1, department_id: 5 },
      },
    ])('should get usage stats $desc', async ({ params, expectedParams }) => {
      const mockResponse: ApiResponse<UsageResponse> = {
        code: 0,
        message: 'success',
        data: mockUsageResponse,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getUsageStats(params);

      expect(mockRequest.get).toHaveBeenCalledWith('/stats/usage', { params: expectedParams });
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle empty usage data', async () => {
      const emptyResponse: UsageResponse = {
        period: 'day',
        items: [],
      };
      const mockResponse: ApiResponse<UsageResponse> = {
        code: 0,
        message: 'success',
        data: emptyResponse,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getUsageStats({ period: 'day' });

      expect(result.data.data.items).toEqual([]);
    });
  });

  describe('getRanking', () => {
    const mockRanking: RankingItem[] = [
      {
        rank: 1,
        id: 1,
        name: '用户1',
        total_tokens: 1000000,
        request_count: 5000,
      },
      {
        rank: 2,
        id: 2,
        name: '用户2',
        total_tokens: 800000,
        request_count: 4000,
      },
      {
        rank: 3,
        id: 3,
        name: '用户3',
        total_tokens: 600000,
        request_count: 3000,
      },
    ];

    it.each([
      {
        desc: 'user ranking',
        params: { type: 'user' as const, period: 'day' },
        expectedParams: { type: 'user', period: 'day' },
      },
      {
        desc: 'department ranking',
        params: { type: 'department' as const, period: 'month' },
        expectedParams: { type: 'department', period: 'month' },
      },
      {
        desc: 'with limit',
        params: { type: 'user' as const, period: 'week', limit: 10 },
        expectedParams: { type: 'user', period: 'week', limit: 10 },
      },
    ])('should get $desc', async ({ params, expectedParams }) => {
      const mockResponse: ApiResponse<RankingItem[]> = {
        code: 0,
        message: 'success',
        data: mockRanking,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getRanking(params);

      expect(mockRequest.get).toHaveBeenCalledWith('/stats/ranking', { params: expectedParams });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('exportUsageCSV', () => {
    it.each([
      {
        desc: 'period only',
        params: { period: 'day' },
        expectedParams: { period: 'day' },
      },
      {
        desc: 'with date range',
        params: { period: 'month', start_date: '2024-01-01', end_date: '2024-01-31' },
        expectedParams: { period: 'month', start_date: '2024-01-01', end_date: '2024-01-31' },
      },
      {
        desc: 'with user filter',
        params: { period: 'week', user_id: 1 },
        expectedParams: { period: 'week', user_id: 1 },
      },
      {
        desc: 'with department filter',
        params: { period: 'day', department_id: 5 },
        expectedParams: { period: 'day', department_id: 5 },
      },
    ])('should export CSV $desc', async ({ params, expectedParams }) => {
      const blob = new Blob(['csv,data'], { type: 'text/csv' });
      mockRequest.get.mockResolvedValue({ data: blob });

      const result = await exportUsageCSV(params);

      expect(mockRequest.get).toHaveBeenCalledWith('/stats/export/csv', {
        params: expectedParams,
        responseType: 'blob',
      });
      expect(result.data).toEqual(blob);
    });
  });
});
