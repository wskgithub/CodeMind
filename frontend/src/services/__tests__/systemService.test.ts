import { describe, it, expect, beforeEach, vi } from 'vitest';

import request from '../request';
import {
  getConfigs,
  updateConfigs,
  listAnnouncements,
  createAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  listAuditLogs,
} from '../systemService';

import type { ApiResponse, SystemConfig, Announcement, AuditLog, PageData } from '@/types';

// Mock request module
vi.mock('../request', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));


const mockRequest = request as unknown as {
  get: ReturnType<typeof vi.fn>;
  put: ReturnType<typeof vi.fn>;
  post: ReturnType<typeof vi.fn>;
  delete: ReturnType<typeof vi.fn>;
};

describe('systemService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('System Config', () => {
    const mockConfigs: SystemConfig[] = [
      {
        id: 1,
        config_key: 'site_name',
        config_value: 'CodeMind',
        description: 'Site name',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        config_key: 'max_upload_size',
        config_value: '10485760',
        description: 'Maximum upload file size',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ];

    describe('getConfigs', () => {
      it('should get all system configs', async () => {
        const mockResponse: ApiResponse<SystemConfig[]> = {
          code: 0,
          message: 'success',
          data: mockConfigs,
        };
        mockRequest.get.mockResolvedValue({ data: mockResponse });

        const result = await getConfigs();

        expect(mockRequest.get).toHaveBeenCalledWith('/system/configs');
        expect(result.data).toEqual(mockResponse);
      });

      it('should return empty array when no configs', async () => {
        const mockResponse: ApiResponse<SystemConfig[]> = {
          code: 0,
          message: 'success',
          data: [],
        };
        mockRequest.get.mockResolvedValue({ data: mockResponse });

        const result = await getConfigs();

        expect(result.data.data).toEqual([]);
      });
    });

    describe('updateConfigs', () => {
      it.each([
        {
          desc: 'single config',
          configs: [{ key: 'site_name', value: 'New Name' }],
        },
        {
          desc: 'multiple configs',
          configs: [
            { key: 'site_name', value: 'CodeMind Pro' },
            { key: 'max_upload_size', value: '20971520' },
          ],
        },
      ])('should update $desc', async ({ configs }) => {
        const mockResponse: ApiResponse<null> = {
          code: 0,
          message: 'success',
          data: null,
        };
        mockRequest.put.mockResolvedValue({ data: mockResponse });

        const result = await updateConfigs(configs);

        expect(mockRequest.put).toHaveBeenCalledWith('/system/configs', { configs });
        expect(result.data).toEqual(mockResponse);
      });
    });
  });

  describe('Announcements', () => {
    const mockAnnouncements: Announcement[] = [
      {
        id: 1,
        title: 'System Maintenance Notice',
        content: 'System maintenance scheduled for tonight',
        author_id: 1,
        status: 1,
        pinned: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        title: 'New Feature Release',
        content: 'We released new features',
        author_id: 1,
        status: 1,
        pinned: false,
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      },
    ];

    describe('listAnnouncements', () => {
      it('should get all announcements', async () => {
        const mockResponse: ApiResponse<Announcement[]> = {
          code: 0,
          message: 'success',
          data: mockAnnouncements,
        };
        mockRequest.get.mockResolvedValue({ data: mockResponse });

        const result = await listAnnouncements();

        expect(mockRequest.get).toHaveBeenCalledWith('/announcements');
        expect(result.data).toEqual(mockResponse);
      });
    });

    describe('createAnnouncement', () => {
      it.each([
        {
          desc: 'full data',
          data: {
            title: 'New Announcement',
            content: 'Announcement content',
            pinned: true,
            status: 1,
          },
        },
        {
          desc: 'minimal data',
          data: {
            title: 'Simple Announcement',
            content: 'Content',
          },
        },
        {
          desc: 'not pinned',
          data: {
            title: 'Regular Announcement',
            content: 'Content',
            pinned: false,
          },
        },
      ])('should create announcement with $desc', async ({ data }) => {
        const mockResponse: ApiResponse<Announcement> = {
          code: 0,
          message: 'success',
          data: { ...mockAnnouncements[0], ...data, id: 3 },
        };
        mockRequest.post.mockResolvedValue({ data: mockResponse });

        const result = await createAnnouncement(data);

        expect(mockRequest.post).toHaveBeenCalledWith('/system/announcements', data);
        expect(result.data).toEqual(mockResponse);
      });
    });

    describe('updateAnnouncement', () => {
      it.each([
        { desc: 'title only', data: { title: 'New Title' } },
        { desc: 'content only', data: { content: 'New Content' } },
        { desc: 'pinned only', data: { pinned: true } },
        { desc: 'status only', data: { status: 0 } },
        { desc: 'multiple fields', data: { title: 'New Title', content: 'New Content', pinned: false } },
      ])('should update announcement with $desc', async ({ data }) => {
        const announcementId = 1;
        const mockResponse: ApiResponse<null> = {
          code: 0,
          message: 'success',
          data: null,
        };
        mockRequest.put.mockResolvedValue({ data: mockResponse });

        const result = await updateAnnouncement(announcementId, data);

        expect(mockRequest.put).toHaveBeenCalledWith(`/system/announcements/${announcementId}`, data);
        expect(result.data).toEqual(mockResponse);
      });
    });

    describe('deleteAnnouncement', () => {
      it.each([1, 2, 100])('should delete announcement with id %i', async (id) => {
        const mockResponse: ApiResponse<null> = {
          code: 0,
          message: 'success',
          data: null,
        };
        mockRequest.delete.mockResolvedValue({ data: mockResponse });

        const result = await deleteAnnouncement(id);

        expect(mockRequest.delete).toHaveBeenCalledWith(`/system/announcements/${id}`);
        expect(result.data).toEqual(mockResponse);
      });
    });
  });

  describe('Audit Logs', () => {
    const mockAuditLog: AuditLog = {
      id: 1,
      operator_id: 1,
      action: 'user.create',
      target_type: 'user',
      target_id: 2,
      detail: { username: 'newuser' },
      client_ip: '192.168.1.1',
      created_at: '2024-01-01T00:00:00Z',
    };

    it.each([
      { desc: 'no params', params: undefined },
      { desc: 'with pagination', params: { page: 1, page_size: 20 } },
      { desc: 'with action filter', params: { action: 'user.create' } },
      { desc: 'with operator filter', params: { operator_id: 1 } },
      { desc: 'with date range', params: { start_date: '2024-01-01', end_date: '2024-01-31' } },
      {
        desc: 'with combined filters',
        params: { page: 1, page_size: 10, action: 'user.create', operator_id: 1, start_date: '2024-01-01', end_date: '2024-01-31' },
      },
    ])('should list audit logs $desc', async ({ params }) => {
      const mockPageData: PageData<AuditLog> = {
        list: [mockAuditLog],
        pagination: {
          page: 1,
          page_size: 10,
          total: 1,
          total_pages: 1,
        },
      };
      const mockResponse: ApiResponse<PageData<AuditLog>> = {
        code: 0,
        message: 'success',
        data: mockPageData,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await listAuditLogs(params);

      expect(mockRequest.get).toHaveBeenCalledWith('/system/audit-logs', { params });
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle empty audit logs', async () => {
      const emptyPageData: PageData<AuditLog> = {
        list: [],
        pagination: {
          page: 1,
          page_size: 10,
          total: 0,
          total_pages: 0,
        },
      };
      const mockResponse: ApiResponse<PageData<AuditLog>> = {
        code: 0,
        message: 'success',
        data: emptyPageData,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await listAuditLogs();

      expect(result.data.data.list).toEqual([]);
      expect(result.data.data.pagination.total).toBe(0);
    });
  });
});
