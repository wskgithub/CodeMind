import { describe, it, expect, beforeEach, vi } from 'vitest';

import keyService from '../keyService';
import request from '../request';

import type { ApiResponse, APIKey, APIKeyCreateResult } from '@/types';

// Mock request module
vi.mock('../request', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));


const mockRequest = request as unknown as {
  get: ReturnType<typeof vi.fn>;
  post: ReturnType<typeof vi.fn>;
  put: ReturnType<typeof vi.fn>;
  delete: ReturnType<typeof vi.fn>;
};

describe('keyService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('list', () => {
    const apiKeys: APIKey[] = [
      {
        id: 1,
        name: 'Test Key 1',
        key_prefix: 'sk_test1',
        status: 1,
        created_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        name: 'Test Key 2',
        key_prefix: 'sk_test2',
        status: 0,
        last_used_at: '2024-01-15T10:30:00Z',
        expires_at: '2024-12-31T23:59:59Z',
        created_at: '2024-01-02T00:00:00Z',
      },
    ];

    it('should get API key list', async () => {
      const mockResponse: ApiResponse<APIKey[]> = {
        code: 0,
        message: 'success',
        data: apiKeys,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await keyService.list();

      expect(mockRequest.get).toHaveBeenCalledWith('/keys');
      expect(result.data).toEqual(mockResponse);
    });

    it('should return empty list when no keys', async () => {
      const mockResponse: ApiResponse<APIKey[]> = {
        code: 0,
        message: 'success',
        data: [],
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await keyService.list();

      expect(result.data.data).toEqual([]);
    });
  });

  describe('create', () => {
    const createData = {
      name: 'New API Key',
      expires_at: '2024-12-31T23:59:59Z',
    };

    const createResult: APIKeyCreateResult = {
      id: 3,
      name: 'New API Key',
      key: 'sk_full_key_xxxxxxxx',
      key_prefix: 'sk_full_ke',
      expires_at: '2024-12-31T23:59:59Z',
      created_at: '2024-01-10T00:00:00Z',
    };

    it('should create API key with correct data', async () => {
      const mockResponse: ApiResponse<APIKeyCreateResult> = {
        code: 0,
        message: 'success',
        data: createResult,
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      const result = await keyService.create(createData);

      expect(mockRequest.post).toHaveBeenCalledWith('/keys', createData);
      expect(result.data).toEqual(mockResponse);
    });

    it('should create API key without expiration', async () => {
      const dataWithoutExpire = { name: 'Permanent Key' };
      const mockResponse: ApiResponse<APIKeyCreateResult> = {
        code: 0,
        message: 'success',
        data: { ...createResult, name: 'Permanent Key', expires_at: undefined },
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      await keyService.create(dataWithoutExpire);

      expect(mockRequest.post).toHaveBeenCalledWith('/keys', dataWithoutExpire);
    });
  });

  describe('updateStatus', () => {
    it.each([
      { id: 1, status: 1, description: 'enable' },
      { id: 2, status: 0, description: 'disable' },
    ])('should $description key with id $id', async ({ id, status }) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await keyService.updateStatus(id, status);

      expect(mockRequest.put).toHaveBeenCalledWith(`/keys/${id}/status`, { status });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('delete', () => {
    it.each([1, 2, 100])('should delete key with id %i', async (id) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.delete.mockResolvedValue({ data: mockResponse });

      const result = await keyService.delete(id);

      expect(mockRequest.delete).toHaveBeenCalledWith(`/keys/${id}`);
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle delete error', async () => {
      const error = new Error('Key not found');
      mockRequest.delete.mockRejectedValue(error);

      await expect(keyService.delete(999)).rejects.toThrow('Key not found');
    });
  });
});
