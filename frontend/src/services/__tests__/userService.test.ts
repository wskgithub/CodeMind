import { describe, it, expect, beforeEach, vi } from 'vitest';

import request from '../request';
import userService, { type UserListParams, type CreateUserParams } from '../userService';

import type { ApiResponse, PageData, UserDetail } from '@/types';

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

describe('userService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockUser: UserDetail = {
    id: 1,
    username: 'testuser',
    display_name: 'Test User',
    email: 'test@example.com',
    phone: '13800138000',
    role: 'user',
    department_id: 1,
    status: 1,
    login_fail_count: 0,
    created_at: '2024-01-01T00:00:00Z',
  };

  describe('list', () => {
    const mockPageData: PageData<UserDetail> = {
      list: [mockUser],
      pagination: {
        page: 1,
        page_size: 10,
        total: 1,
        total_pages: 1,
      },
    };

    it.each([
      {
        desc: 'no params',
        params: {},
        expectedParams: {},
      },
      {
        desc: 'with pagination',
        params: { page: 2, page_size: 20 },
        expectedParams: { page: 2, page_size: 20 },
      },
      {
        desc: 'with keyword filter',
        params: { keyword: 'test' },
        expectedParams: { keyword: 'test' },
      },
      {
        desc: 'with department filter',
        params: { department_id: 5 },
        expectedParams: { department_id: 5 },
      },
      {
        desc: 'with role filter',
        params: { role: 'super_admin' },
        expectedParams: { role: 'super_admin' },
      },
      {
        desc: 'with status filter',
        params: { status: 1 },
        expectedParams: { status: 1 },
      },
      {
        desc: 'with combined filters',
        params: { page: 1, page_size: 10, keyword: 'admin', role: 'super_admin', status: 1 },
        expectedParams: { page: 1, page_size: 10, keyword: 'admin', role: 'super_admin', status: 1 },
      },
    ])('should get user list $desc', async ({ params, expectedParams }) => {
      const mockResponse: ApiResponse<PageData<UserDetail>> = {
        code: 0,
        message: 'success',
        data: mockPageData,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await userService.list(params as UserListParams);

      expect(mockRequest.get).toHaveBeenCalledWith('/users', { params: expectedParams });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getDetail', () => {
    it.each([1, 2, 100])('should get user detail with id %i', async (id) => {
      const mockResponse: ApiResponse<UserDetail> = {
        code: 0,
        message: 'success',
        data: { ...mockUser, id },
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await userService.getDetail(id);

      expect(mockRequest.get).toHaveBeenCalledWith(`/users/${id}`);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('create', () => {
    const createParams: CreateUserParams = {
      username: 'newuser',
      password: 'password123',
      display_name: 'New User',
      email: 'new@example.com',
      phone: '13900139000',
      role: 'user',
      department_id: 1,
    };

    it('should create user with correct data', async () => {
      const mockResponse: ApiResponse<UserDetail> = {
        code: 0,
        message: 'success',
        data: { ...mockUser, ...createParams },
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      const result = await userService.create(createParams);

      expect(mockRequest.post).toHaveBeenCalledWith('/users', createParams);
      expect(result.data).toEqual(mockResponse);
    });

    it('should create user with minimal data', async () => {
      const minimalParams: CreateUserParams = {
        username: 'simpleuser',
        password: 'pass123',
        display_name: 'Simple User',
        role: 'user',
      };
      const mockResponse: ApiResponse<UserDetail> = {
        code: 0,
        message: 'success',
        data: { ...mockUser, ...minimalParams },
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      await userService.create(minimalParams);

      expect(mockRequest.post).toHaveBeenCalledWith('/users', minimalParams);
    });
  });

  describe('update', () => {
    it.each([
      { desc: 'full update', data: { display_name: 'New Name', email: 'new@example.com' } },
      { desc: 'partial update', data: { display_name: 'Name Only Update' } },
      { desc: 'update role', data: { role: 'dept_manager' } },
      { desc: 'update department', data: { department_id: 2 } },
    ])('should update user with $desc', async ({ data }) => {
      const userId = 1;
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await userService.update(userId, data);

      expect(mockRequest.put).toHaveBeenCalledWith(`/users/${userId}`, data);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('delete', () => {
    it.each([1, 2, 100])('should delete user with id %i', async (id) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.delete.mockResolvedValue({ data: mockResponse });

      const result = await userService.delete(id);

      expect(mockRequest.delete).toHaveBeenCalledWith(`/users/${id}`);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('updateStatus', () => {
    it.each([
      { id: 1, status: 1, desc: 'enable' },
      { id: 2, status: 0, desc: 'disable' },
      { id: 3, status: 2, desc: 'lock' },
    ])('should $desc user with id $id', async ({ id, status }) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await userService.updateStatus(id, status);

      expect(mockRequest.put).toHaveBeenCalledWith(`/users/${id}/status`, { status });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('resetPassword', () => {
    it('should reset password with correct params', async () => {
      const userId = 1;
      const newPassword = 'newPassword123';
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await userService.resetPassword(userId, newPassword);

      expect(mockRequest.put).toHaveBeenCalledWith(`/users/${userId}/reset-password`, { new_password: newPassword });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('unlockUser', () => {
    it.each([
      { desc: 'with reason', reason: 'Admin unlock' },
      { desc: 'without reason', reason: undefined },
    ])('should unlock user $desc', async ({ reason }) => {
      const userId = 1;
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await userService.unlockUser(userId, reason);

      expect(mockRequest.put).toHaveBeenCalledWith(`/users/${userId}/unlock`, { reason });
      expect(result.data).toEqual(mockResponse);
    });
  });
});
