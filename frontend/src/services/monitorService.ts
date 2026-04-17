import request from './request';
import type {
  ApiResponse,
  DashboardSummary,
  SystemMetricsSummary,
  RequestMetricsSummary,
  LLMNodeSummary,
} from '@/types';

export function getDashboardSummary() {
  return request.get<ApiResponse<DashboardSummary>>('/monitor/dashboard');
}

export function getSystemMetrics() {
  return request.get<ApiResponse<SystemMetricsSummary>>('/monitor/system');
}

export function getRequestMetrics(duration?: string) {
  return request.get<ApiResponse<RequestMetricsSummary>>('/monitor/requests', {
    params: { duration },
  });
}

export function getLLMNodeMetrics() {
  return request.get<ApiResponse<LLMNodeSummary[]>>('/monitor/llm-nodes');
}

export function healthCheck() {
  return request.get<ApiResponse<{ status: string; hostname: string; timestamp: number }>>('/monitor/health');
}
