import request from './request';
import type { ApiResponse, LLMBackend } from '@/types';

export function listLLMBackends() {
  return request.get<ApiResponse<LLMBackend[]>>('/system/llm-backends');
}

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

export function updateLLMBackend(id: number, data: Record<string, unknown>) {
  return request.put<ApiResponse<null>>(`/system/llm-backends/${id}`, data);
}

export function deleteLLMBackend(id: number) {
  return request.delete<ApiResponse<null>>(`/system/llm-backends/${id}`);
}
