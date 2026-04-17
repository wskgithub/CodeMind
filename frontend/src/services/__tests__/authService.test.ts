import { describe, it, expect, beforeEach, vi } from 'vitest';

import authService from '../authService';
import request from '../request';

import type { ApiResponse, LoginParams, LoginResult, UserDetail } from '@/types';

// Mock request module
vi.mock('../request', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
    put: vi.fn(),
  },
}));


const mockRequest = request as unknown as {
  post: ReturnType<typeof vi.fn>;
  get: ReturnType<typeof vi.fn>;
  put: ReturnType<typeof vi.fn>;
};

describe('authService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('login', () => {
    const loginParams: LoginParams = {
      username: 'admin',
      password: 'password123',
    };

    const loginResult: LoginResult = {
      token: 'test-token-123',
      expires_at: '2024-12-31T23:59:59Z',
      user: {
        id: 1,
        username: 'admin',
        display_name: '管理员',
        role: 'super_admin',
      },
    };

    it('should call login API with correct params', async () => {
      const mockResponse: ApiResponse<LoginResult> = {
        code: 0,
        message: 'success',
        data: loginResult,
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      const result = await authService.login(loginParams);

      expect(mockRequest.post).toHaveBeenCalledWith('/auth/login', loginParams);
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle login error', async () => {
      const error = new Error('Invalid credentials');
      mockRequest.post.mockRejectedValue(error);

      await expect(authService.login(loginParams)).rejects.toThrow('Invalid credentials');
    });
  });

  describe('logout', () => {
    it('should call logout API', async () => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      const result = await authService.logout();

      expect(mockRequest.post).toHaveBeenCalledWith('/auth/logout');
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getProfile', () => {
    const userDetail: UserDetail = {
      id: 1,
      username: 'admin',
      display_name: '管理员',
      email: 'admin@example.com',
      phone: '13800138000',
      role: 'super_admin',
      status: 1,
      login_fail_count: 0,
      created_at: '2024-01-01T00:00:00Z',
    };

    it('should get user profile', async () => {
      const mockResponse: ApiResponse<UserDetail> = {
        code: 0,
        message: 'success',
        data: userDetail,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await authService.getProfile();

      expect(mockRequest.get).toHaveBeenCalledWith('/auth/profile');
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('updateProfile', () => {
    const updateData = {
      display_name: '新名称',
      email: 'new@example.com',
      phone: '13900139000',
    };

    it('should update profile with correct data', async () => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await authService.updateProfile(updateData);

      expect(mockRequest.put).toHaveBeenCalledWith('/auth/profile', updateData);
      expect(result.data).toEqual(mockResponse);
    });

    it('should update partial profile data', async () => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const partialData = { display_name: '仅更新名称' };
      await authService.updateProfile(partialData);

      expect(mockRequest.put).toHaveBeenCalledWith('/auth/profile', partialData);
    });
  });

  describe('changePassword', () => {
    const passwordData = {
      old_password: 'oldPass123',
      new_password: 'newPass456',
    };

    it('should change password with correct data', async () => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await authService.changePassword(passwordData);

      expect(mockRequest.put).toHaveBeenCalledWith('/auth/password', passwordData);
      expect(result.data).toEqual(mockResponse);
    });
  });
});
