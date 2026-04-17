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
  listPlatformModels() {
    return request.get<ApiResponse<PlatformModelInfo[]>>('/models/platform');
  },

  listTemplates() {
    return request.get<ApiResponse<ProviderTemplate[]>>('/models/templates');
  },

  listProviders() {
    return request.get<ApiResponse<UserThirdPartyProvider[]>>('/models/third-party');
  },

  createProvider(data: CreateThirdPartyProviderRequest) {
    return request.post<ApiResponse<UserThirdPartyProvider>>('/models/third-party', data);
  },

  updateProvider(id: number, data: UpdateThirdPartyProviderRequest) {
    return request.put<ApiResponse<null>>(`/models/third-party/${id}`, data);
  },

  updateProviderStatus(id: number, status: number) {
    return request.put<ApiResponse<null>>(`/models/third-party/${id}/status`, { status });
  },

  deleteProvider(id: number) {
    return request.delete<ApiResponse<null>>(`/models/third-party/${id}`);
  },

  // admin template management

  listTemplatesAdmin() {
    return request.get<ApiResponse<ProviderTemplate[]>>('/system/provider-templates');
  },

  createTemplate(data: CreateProviderTemplateRequest) {
    return request.post<ApiResponse<ProviderTemplate>>('/system/provider-templates', data);
  },

  updateTemplate(id: number, data: UpdateProviderTemplateRequest) {
    return request.put<ApiResponse<null>>(`/system/provider-templates/${id}`, data);
  },

  deleteTemplate(id: number) {
    return request.delete<ApiResponse<null>>(`/system/provider-templates/${id}`);
  },
};

export default modelService;
