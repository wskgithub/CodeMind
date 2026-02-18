import request from './request';
import type { ApiResponse, LLMBackend } from '@/types';

/** 获取所有 LLM 后端节点 */
export function listLLMBackends() {
  return request.get<ApiResponse<LLMBackend[]>>('/system/llm-backends');
}

/** 创建 LLM 后端节点 */
export function createLLMBackend(data: {
  name: string;
  display_name?: string;
  base_url: string;
  api_key?: string;
  format: string;
  weight?: number;
  max_concurrency?: number;
  health_check_url?: string;
  timeout_seconds?: number;
  stream_timeout_seconds?: number;
  model_patterns?: string;
}) {
  return request.post<ApiResponse<null>>('/system/llm-backends', data);
}

/** 更新 LLM 后端节点 */
export function updateLLMBackend(id: number, data: Record<string, unknown>) {
  return request.put<ApiResponse<null>>(`/system/llm-backends/${id}`, data);
}

/** 删除 LLM 后端节点 */
export function deleteLLMBackend(id: number) {
  return request.delete<ApiResponse<null>>(`/system/llm-backends/${id}`);
}
