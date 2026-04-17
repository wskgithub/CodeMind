import { describe, it, expect, beforeEach, vi } from 'vitest';

import departmentService, { type CreateDepartmentParams } from '../departmentService';
import request from '../request';

import type { ApiResponse, DeptTree } from '@/types';

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

describe('departmentService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockDept: DeptTree = {
    id: 1,
    name: 'Engineering',
    description: 'Engineering and R&D department',
    user_count: 10,
    status: 1,
    children: [],
  };

  describe('list', () => {
    const mockDeptTree: DeptTree[] = [
      {
        ...mockDept,
        children: [
          {
            id: 2,
            name: 'Frontend Team',
            user_count: 5,
            status: 1,
            children: [],
          },
          {
            id: 3,
            name: 'Backend Team',
            user_count: 5,
            status: 1,
            children: [],
          },
        ],
      },
    ];

    it('should get department tree list', async () => {
      const mockResponse: ApiResponse<DeptTree[]> = {
        code: 0,
        message: 'success',
        data: mockDeptTree,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await departmentService.list();

      expect(mockRequest.get).toHaveBeenCalledWith('/departments');
      expect(result.data).toEqual(mockResponse);
    });

    it('should return empty array when no departments', async () => {
      const mockResponse: ApiResponse<DeptTree[]> = {
        code: 0,
        message: 'success',
        data: [],
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await departmentService.list();

      expect(result.data.data).toEqual([]);
    });
  });

  describe('getDetail', () => {
    it.each([1, 2, 100])('should get department detail with id %i', async (id) => {
      const mockResponse: ApiResponse<DeptTree> = {
        code: 0,
        message: 'success',
        data: { ...mockDept, id },
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await departmentService.getDetail(id);

      expect(mockRequest.get).toHaveBeenCalledWith(`/departments/${id}`);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('create', () => {
    it.each([
      {
        desc: 'full data',
        data: {
          name: 'New Product Department',
          description: 'Product R&D department',
          parent_id: 1,
          manager_id: 2,
        } as CreateDepartmentParams,
      },
      {
        desc: 'minimal data',
        data: { name: 'Simple Department' } as CreateDepartmentParams,
      },
      {
        desc: 'with description only',
        data: {
          name: 'Department with Description',
          description: 'Department description',
        } as CreateDepartmentParams,
      },
      {
        desc: 'with parent only',
        data: {
          name: 'Sub Department',
          parent_id: 1,
        } as CreateDepartmentParams,
      },
    ])('should create department with $desc', async ({ data }) => {
      const mockResponse: ApiResponse<DeptTree> = {
        code: 0,
        message: 'success',
        data: { ...mockDept, ...data, id: 10 },
      };
      mockRequest.post.mockResolvedValue({ data: mockResponse });

      const result = await departmentService.create(data);

      expect(mockRequest.post).toHaveBeenCalledWith('/departments', data);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('update', () => {
    it.each([
      { desc: 'name only', data: { name: 'New Name' } },
      { desc: 'description only', data: { description: 'New Description' } },
      { desc: 'manager only', data: { manager_id: 5 } },
      { desc: 'multiple fields', data: { name: 'New Name', description: 'New Description', manager_id: 3 } },
    ])('should update department with $desc', async ({ data }) => {
      const deptId = 1;
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.put.mockResolvedValue({ data: mockResponse });

      const result = await departmentService.update(deptId, data);

      expect(mockRequest.put).toHaveBeenCalledWith(`/departments/${deptId}`, data);
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('delete', () => {
    it.each([1, 2, 100])('should delete department with id %i', async (id) => {
      const mockResponse: ApiResponse<null> = {
        code: 0,
        message: 'success',
        data: null,
      };
      mockRequest.delete.mockResolvedValue({ data: mockResponse });

      const result = await departmentService.delete(id);

      expect(mockRequest.delete).toHaveBeenCalledWith(`/departments/${id}`);
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle delete error for department with children', async () => {
      const error = new Error('Cannot delete department with children');
      mockRequest.delete.mockRejectedValue(error);

      await expect(departmentService.delete(1)).rejects.toThrow('Cannot delete department with children');
    });
  });
});
