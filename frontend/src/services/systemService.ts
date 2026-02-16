import request from './request';
import type { ApiResponse, PageData, SystemConfig, Announcement, AuditLog } from '@/types';

// ──────────────────────────────────
// 系统配置
// ──────────────────────────────────

/** 获取系统配置 */
export function getConfigs() {
  return request.get<ApiResponse<SystemConfig[]>>('/system/configs');
}

/** 更新系统配置 */
export function updateConfigs(configs: { key: string; value: string }[]) {
  return request.put<ApiResponse<null>>('/system/configs', { configs });
}

// ──────────────────────────────────
// 公告管理
// ──────────────────────────────────

/** 获取公告列表 */
export function listAnnouncements() {
  return request.get<ApiResponse<Announcement[]>>('/announcements');
}

/** 创建公告 */
export function createAnnouncement(data: {
  title: string;
  content: string;
  pinned?: boolean;
  status?: number;
}) {
  return request.post<ApiResponse<Announcement>>('/system/announcements', data);
}

/** 更新公告 */
export function updateAnnouncement(id: number, data: {
  title?: string;
  content?: string;
  pinned?: boolean;
  status?: number;
}) {
  return request.put<ApiResponse<null>>(`/system/announcements/${id}`, data);
}

/** 删除公告 */
export function deleteAnnouncement(id: number) {
  return request.delete<ApiResponse<null>>(`/system/announcements/${id}`);
}

// ──────────────────────────────────
// 审计日志
// ──────────────────────────────────

/** 获取审计日志 */
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
