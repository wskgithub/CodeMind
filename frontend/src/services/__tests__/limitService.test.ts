import { describe, it, expect, beforeEach, vi } from 'vitest';

import {
  listLimits,
  upsertLimit,
  deleteLimit,
  getMyLimits,
  getMyLimitProgress,
} from '../limitService';
import request from '../request';

import type { ApiResponse, RateLimit, MyLimitResponse, LimitProgressResponse } from '@/types';

// Mock request module
vi.mock('../request', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));


const mockRequest = request as unknown as {
  get: ReturnType<typeof vi.fn>;
  put: ReturnType<typeof vi.fn>;
  delete: ReturnType<typeof vi.fn>;
};

describe('limitService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockRateLimit: RateLimit = {
    id: 1,
    target_type: 'user',
    target_id: 1,
    period: 'day',
    period_hours: 24,
    max_tokens: 100000,
    max_requests: 1000,
    max_concurrency: 10,
    alert_threshold: 80,
    status: 1,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  describe('listLimits', () => {
    it.each([
      { desc: 'no params', params: undefined, expectedParams: undefined },
      { desc: 'with target_type', params: { target_type: 'user' }, expectedParams: { target_type: 'user' } },
      { desc: 'with target_id', params: { target_id: 1 }, expectedParams: { target_id: 1 } },
      { desc: 'with both params', params: { target_type: 'department', target_id: 5 }, expectedParams: { target_type: 'department', target_id: 5 } },
    ])('should list limits $desc', async ({ params, expectedParams }) => {
      const mockResponse: ApiResponse<RateLimit[]> = {
        code: 0,
        message: 'success',
        data: [mockRateLimit],
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await listLimits(params);

      expect(mockRequest.get).toHaveBeenCalledWith('/limits', { params: expectedParams });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('upsertLimit', () => {
    it.each([
      {
        desc: 'full data',
        data: {
          target_type: 'user',
          target_id: 1,
          period: 'day',
          period_hours: 24,
          max_tokens: 100000,
          max_requests: 1000,
          max_concurrency: 10,
          alert_threshold: 80,
        },
      },
      {
        desc: 'minimal data',
        data: {
          target_type: 'global',
          target_id: 0,
          period: 'hour',
          max_tokens: 50000,
        },
      },
      {
        desc: 'department limit',
        data: {
          target_type: 'department',
          target_id: 5,
          period: 'month',
          period_hours: 720,
          max_tokens: 1000000,
          max_requests: 5000,
          max_concurrency: 20,
          alert_threshold: 90,
        },
      },
    ])('should upsert limit with $desc', async ({ data }) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await upsertLimit(data);

      expect(mockRequest.put).toHaveBeenCalledWith('/limits', data);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('deleteLimit', () => {
    it.each([1, 2, 100])('should delete limit with id %i', async (id) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.delete.mockResolvedValue({ data: mockResponse });

      const result = await deleteLimit(id);

      expect(mockRequest.delete).toHaveBeenCalledWith(`/limits/${id}`);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getMyLimits', () => {
    const mockMyLimits: MyLimitResponse = {
      limits: {
        day: {
          max_tokens: 100000,
          used_tokens: 50000,
          remaining_tokens: 50000,
          usage_percent: 50,
        },
        hour: {
          max_tokens: 10000,
          used_tokens: 2000,
          remaining_tokens: 8000,
          usage_percent: 20,
        },
      },
      concurrency: {
        max: 10,
        current: 3,
      },
    };

    it('should get my limits (legacy)', async () => {
      const mockResponse: ApiResponse<MyLimitResponse> = {
        code: 0,
        message: 'success',
        data: mockMyLimits,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getMyLimits();

      expect(mockRequest.get).toHaveBeenCalledWith('/limits/my');
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getMyLimitProgress', () => {
    const mockLimitProgress: LimitProgressResponse = {
      limits: [
        {
          rule_id: 1,
          period: 'day',
          period_hours: 24,
          max_tokens: 100000,
          used_tokens: 50000,
          remaining_tokens: 50000,
          usage_percent: 50,
          cycle_start_at: 1704067200,
          reset_at: 1704153600,
          reset_in_hours: 12,
          exceeded: false,
        },
        {
          rule_id: 2,
          period: 'hour',
          period_hours: 1,
          max_tokens: 10000,
          used_tokens: 10000,
          remaining_tokens: 0,
          usage_percent: 100,
          cycle_start_at: 1704110400,
          reset_at: 1704114000,
          reset_in_hours: 0.5,
          exceeded: true,
        },
      ],
      concurrency: {
        max: 10,
        current: 3,
      },
      any_exceeded: true,
    };

    it('should get my limit progress (new version)', async () => {
      const mockResponse: ApiResponse<LimitProgressResponse> = {
        code: 0,
        message: 'success',
        data: mockLimitProgress,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getMyLimitProgress();

      expect(mockRequest.get).toHaveBeenCalledWith('/limits/my/progress');
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle empty limits', async () => {
      const emptyProgress: LimitProgressResponse = {
        limits: [],
        concurrency: { max: 0, current: 0 },
        any_exceeded: false,
      };
      const mockResponse: ApiResponse<LimitProgressResponse> = {
        code: 0,
        message: 'success',
        data: emptyProgress,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getMyLimitProgress();

      expect(result.data.data.limits).toEqual([]);
      expect(result.data.data.any_exceeded).toBe(false);
    });
  });
});
