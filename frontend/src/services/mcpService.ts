import request from './request';
import type {
  ApiResponse,
  MCPService,
  MCPTool,
  MCPAccessRule,
  CreateMCPServiceRequest,
  UpdateMCPServiceRequest,
} from '@/types';

/** MCP 服务管理 API */
const mcpService = {
  /** 获取 MCP 服务列表 */
  listServices(status?: string) {
    return request.get<ApiResponse<MCPService[]>>('/mcp/services', { params: { status } });
  },

  /** 创建 MCP 服务 */
  createService(data: CreateMCPServiceRequest) {
    return request.post<ApiResponse<MCPService>>('/mcp/services', data);
  },

  /** 更新 MCP 服务 */
  updateService(id: number, data: UpdateMCPServiceRequest) {
    return request.put<ApiResponse<null>>(`/mcp/services/${id}`, data);
  },

  /** 删除 MCP 服务 */
  deleteService(id: number) {
    return request.delete<ApiResponse<null>>(`/mcp/services/${id}`);
  },

  /** 同步工具列表 */
  syncTools(id: number) {
    return request.post<ApiResponse<null>>(`/mcp/services/${id}/sync`);
  },

  /** 获取服务工具列表 */
  getServiceTools(id: number) {
    return request.get<ApiResponse<MCPTool[]>>(`/mcp/services/${id}/tools`);
  },

  /** 获取访问规则列表 */
  listAccessRules(serviceId?: number) {
    return request.get<ApiResponse<MCPAccessRule[]>>('/mcp/access-rules', {
      params: serviceId ? { service_id: serviceId } : undefined,
    });
  },

  /** 设置访问规则 */
  setAccessRule(data: { service_id: number; target_type: string; target_id: number; allowed: boolean }) {
    return request.post<ApiResponse<null>>('/mcp/access-rules', data);
  },

  /** 删除访问规则 */
  deleteAccessRule(id: number) {
    return request.delete<ApiResponse<null>>(`/mcp/access-rules/${id}`);
  },
};

export default mcpService;
