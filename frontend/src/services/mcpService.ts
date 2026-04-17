import request from './request';
import type {
  ApiResponse,
  MCPService,
  MCPTool,
  MCPAccessRule,
  CreateMCPServiceRequest,
  UpdateMCPServiceRequest,
} from '@/types';

const mcpService = {
  listServices(status?: string) {
    return request.get<ApiResponse<MCPService[]>>('/mcp/services', { params: { status } });
  },

  createService(data: CreateMCPServiceRequest) {
    return request.post<ApiResponse<MCPService>>('/mcp/services', data);
  },

  updateService(id: number, data: UpdateMCPServiceRequest) {
    return request.put<ApiResponse<null>>(`/mcp/services/${id}`, data);
  },

  deleteService(id: number) {
    return request.delete<ApiResponse<null>>(`/mcp/services/${id}`);
  },

  syncTools(id: number) {
    return request.post<ApiResponse<null>>(`/mcp/services/${id}/sync`);
  },

  getServiceTools(id: number) {
    return request.get<ApiResponse<MCPTool[]>>(`/mcp/services/${id}/tools`);
  },

  listAccessRules(serviceId?: number) {
    return request.get<ApiResponse<MCPAccessRule[]>>('/mcp/access-rules', {
      params: serviceId ? { service_id: serviceId } : undefined,
    });
  },

  setAccessRule(data: { service_id: number; target_type: string; target_id: number; allowed: boolean }) {
    return request.post<ApiResponse<null>>('/mcp/access-rules', data);
  },

  deleteAccessRule(id: number) {
    return request.delete<ApiResponse<null>>(`/mcp/access-rules/${id}`);
  },
};

export default mcpService;
