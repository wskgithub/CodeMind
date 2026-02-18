import request from './request';
import type {
  ApiResponse,
  DashboardSummary,
  SystemMetricsSummary,
  RequestMetricsSummary,
  LLMNodeSummary,
} from '@/types';

/** 获取仪表盘汇总数据 */
export function getDashboardSummary() {
  return request.get<ApiResponse<DashboardSummary>>('/monitor/dashboard');
}

/** 获取系统资源指标 */
export function getSystemMetrics() {
  return request.get<ApiResponse<SystemMetricsSummary>>('/monitor/system');
}

/** 获取请求性能指标 */
export function getRequestMetrics(duration?: string) {
  return request.get<ApiResponse<RequestMetricsSummary>>('/monitor/requests', {
    params: { duration },
  });
}

/** 获取 LLM 节点指标 */
export function getLLMNodeMetrics() {
  return request.get<ApiResponse<LLMNodeSummary[]>>('/monitor/llm-nodes');
}

/** 健康检查 */
export function healthCheck() {
  return request.get<ApiResponse<{ status: string; hostname: string; timestamp: number }>>('/monitor/health');
}
