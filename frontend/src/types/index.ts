// ──────────────────────────────────
// 通用类型
// ──────────────────────────────────

/** 统一 API 响应结构 */
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
}

/** 分页信息 */
export interface Pagination {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

/** 分页响应 */
export interface PageData<T> {
  list: T[];
  pagination: Pagination;
}

// ──────────────────────────────────
// 用户相关类型
// ──────────────────────────────────

/** 用户角色 */
export type UserRole = 'super_admin' | 'dept_manager' | 'user';

/** 用户简要信息 */
export interface UserBrief {
  id: number;
  username: string;
  display_name: string;
  role: UserRole;
  department?: DeptBrief;
}

/** 用户详细信息 */
export interface UserDetail {
  id: number;
  username: string;
  display_name: string;
  email?: string;
  phone?: string;
  avatar_url?: string;
  role: UserRole;
  department_id?: number;
  department?: DeptBrief;
  status: number;
  last_login_at?: string;
  created_at: string;
}

// ──────────────────────────────────
// 部门相关类型
// ──────────────────────────────────

/** 部门简要信息 */
export interface DeptBrief {
  id: number;
  name: string;
}

/** 部门树节点 */
export interface DeptTree {
  id: number;
  name: string;
  description?: string;
  manager?: UserBrief;
  user_count: number;
  status: number;
  children: DeptTree[];
}

// ──────────────────────────────────
// 认证相关类型
// ──────────────────────────────────

/** 登录请求 */
export interface LoginParams {
  username: string;
  password: string;
}

/** 登录响应 */
export interface LoginResult {
  token: string;
  expires_at: string;
  user: UserBrief;
}

// ──────────────────────────────────
// API Key 相关类型
// ──────────────────────────────────

/** API Key 列表项 */
export interface APIKey {
  id: number;
  name: string;
  key_prefix: string;
  status: number;
  last_used_at?: string;
  expires_at?: string;
  created_at: string;
}

/** 创建 API Key 响应 */
export interface APIKeyCreateResult {
  id: number;
  name: string;
  key: string;
  key_prefix: string;
  expires_at?: string;
  created_at: string;
}

// ──────────────────────────────────
// 统计相关类型
// ──────────────────────────────────

/** 用量总览 */
export interface StatsOverview {
  today: PeriodStats;
  this_month: PeriodStats;
  total_users: number;
  total_departments: number;
  total_api_keys: number;
  system_status: string;
}

/** 时间段统计 */
export interface PeriodStats {
  total_tokens: number;
  total_requests: number;
  active_users: number;
}

/** 用量统计项 */
export interface UsageItem {
  date: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  request_count: number;
}

/** 用量统计响应 */
export interface UsageResponse {
  period: string;
  items: UsageItem[];
}

/** 排行榜项 */
export interface RankingItem {
  rank: number;
  id: number;
  name: string;
  total_tokens: number;
  request_count: number;
}

// ──────────────────────────────────
// 限额相关类型
// ──────────────────────────────────

/** 限额配置 */
export interface RateLimit {
  id: number;
  target_type: 'global' | 'department' | 'user';
  target_id: number;
  period: string;
  period_hours: number;
  max_tokens: number;
  max_requests: number;
  max_concurrency: number;
  alert_threshold: number;
  status: number;
  created_at: string;
  updated_at: string;
}

/** 限额详情 */
export interface LimitDetail {
  max_tokens: number;
  used_tokens: number;
  remaining_tokens: number;
  usage_percent: number;
}

/** 并发信息 */
export interface ConcurrencyInfo {
  max: number;
  current: number;
}

/** 个人限额响应（旧版兼容） */
export interface MyLimitResponse {
  limits: Record<string, LimitDetail>;
  concurrency: ConcurrencyInfo;
}

/** 限额进度项（新版，含重置时间） */
export interface LimitProgressItem {
  rule_id: number;
  period: string;
  period_hours: number;
  max_tokens: number;
  used_tokens: number;
  remaining_tokens: number;
  usage_percent: number;
  cycle_start_at: number | null;
  reset_at: number | null;
  reset_in_hours: number | null;
  exceeded: boolean;
}

/** 限额进度响应 */
export interface LimitProgressResponse {
  limits: LimitProgressItem[];
  concurrency: ConcurrencyInfo;
  any_exceeded: boolean;
}

/** LLM 后端节点 */
export interface LLMBackend {
  id: number;
  name: string;
  display_name: string;
  base_url: string;
  has_api_key: boolean;
  format: string;
  weight: number;
  max_concurrency: number;
  active_connections: number;
  status: number;
  healthy: boolean;
  health_check_url: string;
  timeout_seconds: number;
  stream_timeout_seconds: number;
  model_patterns: string;
  created_at: string;
  updated_at: string;
}

// ──────────────────────────────────
// 系统管理相关类型
// ──────────────────────────────────

/** 系统配置项 */
export interface SystemConfig {
  id: number;
  config_key: string;
  config_value: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

/** 公告 */
export interface Announcement {
  id: number;
  title: string;
  content: string;
  author_id: number;
  author?: UserBrief;
  status: number;
  pinned: boolean;
  created_at: string;
  updated_at: string;
}

/** 审计日志 */
export interface AuditLog {
  id: number;
  operator_id: number;
  operator?: UserBrief;
  action: string;
  target_type: string;
  target_id?: number;
  detail?: Record<string, unknown>;
  client_ip?: string;
  created_at: string;
}

// ──────────────────────────────────
// MCP 服务管理类型
// ──────────────────────────────────

/** MCP 服务 */
export interface MCPService {
  id: number;
  name: string;
  display_name: string;
  description: string;
  endpoint_url: string;
  transport_type: string; // "sse" | "streamable-http"
  status: string;         // "enabled" | "disabled"
  auth_type: string;      // "none" | "bearer" | "header"
  tools_count: number;
  connected: boolean;
  created_at: string;
  updated_at: string;
}

/** MCP 工具信息 */
export interface MCPTool {
  name: string;
  description: string;
  service_name: string;
}

/** MCP 访问规则 */
export interface MCPAccessRule {
  id: number;
  service_id: number;
  service_name: string;
  target_type: string;  // "user" | "department" | "role"
  target_id: number;
  target_name: string;
  allowed: boolean;
}

/** 创建 MCP 服务请求 */
export interface CreateMCPServiceRequest {
  name: string;
  display_name: string;
  description?: string;
  endpoint_url: string;
  transport_type: string;
  auth_type: string;
  auth_config?: Record<string, unknown>;
}

/** 更新 MCP 服务请求 */
export interface UpdateMCPServiceRequest {
  display_name?: string;
  description?: string;
  endpoint_url?: string;
  transport_type?: string;
  status?: string;
  auth_type?: string;
  auth_config?: Record<string, unknown>;
}
