import request from './request';
import type { ApiResponse, PageData, SystemConfig, Announcement, AuditLog } from '@/types';

export function getConfigs() {
  return request.get<ApiResponse<SystemConfig[]>>('/system/configs');
}

export function updateConfigs(configs: { key: string; value: string }[]) {
  return request.put<ApiResponse<null>>('/system/configs', { configs });
}

export function getPlatformSettings() {
  return request.get<ApiResponse<{
    service_url: string;
    openai_base_url: string;
    anthropic_base_url: string;
  }>>('/settings/platform');
}

export function listAnnouncements() {
  return request.get<ApiResponse<Announcement[]>>('/announcements');
}

export function createAnnouncement(data: {
  title: string;
  content: string;
  pinned?: boolean;
  status?: number;
}) {
  return request.post<ApiResponse<Announcement>>('/system/announcements', data);
}

export function updateAnnouncement(id: number, data: {
  title?: string;
  content?: string;
  pinned?: boolean;
  status?: number;
}) {
  return request.put<ApiResponse<null>>(`/system/announcements/${id}`, data);
}

export function deleteAnnouncement(id: number) {
  return request.delete<ApiResponse<null>>(`/system/announcements/${id}`);
}

export function listAuditLogs(params?: {
  page?: number;
  page_size?: number;
  action?: string;
  operator_id?: number;
  start_date?: string;
  end_date?: string;
}) {
  return request.get<ApiResponse<PageData<AuditLog>>>('/system/audit-logs', { params });
}
