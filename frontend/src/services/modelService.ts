import request from './request';

import type {
  ApiResponse,
  PlatformModelInfo,
  ProviderTemplate,
  UserThirdPartyProvider,
  CreateThirdPartyProviderRequest,
  UpdateThirdPartyProviderRequest,
  CreateProviderTemplateRequest,
  UpdateProviderTemplateRequest,
} from '@/types';

const modelService = {
  /** 获取 CodeMind 平台模型列表 */
  listPlatformModels() {
    return request.get<ApiResponse<PlatformModelInfo[]>>('/models/platform');
  },

  /** 获取可用模板列表（用户选择） */
  listTemplates() {
    return request.get<ApiResponse<ProviderTemplate[]>>('/models/templates');
  },

  /** 获取当前用户第三方服务列表 */
  listProviders() {
    return request.get<ApiResponse<UserThirdPartyProvider[]>>('/models/third-party');
  },

  /** 添加第三方服务 */
  createProvider(data: CreateThirdPartyProviderRequest) {
    return request.post<ApiResponse<UserThirdPartyProvider>>('/models/third-party', data);
  },

  /** 更新第三方服务 */
  updateProvider(id: number, data: UpdateThirdPartyProviderRequest) {
    return request.put<ApiResponse<null>>(`/models/third-party/${id}`, data);
  },

  /** 切换第三方服务状态 */
  updateProviderStatus(id: number, status: number) {
    return request.put<ApiResponse<null>>(`/models/third-party/${id}/status`, { status });
  },

  /** 删除第三方服务 */
  deleteProvider(id: number) {
    return request.delete<ApiResponse<null>>(`/models/third-party/${id}`);
  },

  // ── 管理员模板管理 ──

  /** 获取所有模板（管理员） */
  listTemplatesAdmin() {
    return request.get<ApiResponse<ProviderTemplate[]>>('/system/provider-templates');
  },

  /** 创建模板 */
  createTemplate(data: CreateProviderTemplateRequest) {
    return request.post<ApiResponse<ProviderTemplate>>('/system/provider-templates', data);
  },

  /** 更新模板 */
  updateTemplate(id: number, data: UpdateProviderTemplateRequest) {
    return request.put<ApiResponse<null>>(`/system/provider-templates/${id}`, data);
  },

  /** 删除模板 */
  deleteTemplate(id: number) {
    return request.delete<ApiResponse<null>>(`/system/provider-templates/${id}`);
  },
};

export default modelService;
