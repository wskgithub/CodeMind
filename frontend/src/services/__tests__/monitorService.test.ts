import { describe, it, expect, beforeEach, vi } from 'vitest';

import {
  getDashboardSummary,
  getSystemMetrics,
  getRequestMetrics,
  getLLMNodeMetrics,
  healthCheck,
} from '../monitorService';
import request from '../request';

import type {
  ApiResponse,
  DashboardSummary,
  SystemMetricsSummary,
  RequestMetricsSummary,
  LLMNodeSummary,
} from '@/types';

// Mock request module
vi.mock('../request', () => ({
  default: {
    get: vi.fn(),
  },
}));

const mockRequest = request as unknown as {
  get: ReturnType<typeof vi.fn>;
};

describe('monitorService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const mockDashboardSummary: DashboardSummary = {
    system_status: {
      cpu_usage: {
        usage_percent: 45.5,
        core_count: 8,
        model_name: 'Intel(R) Core(TM) i7',
      },
      memory_usage: {
        total_gb: 32,
        used_gb: 16,
        free_gb: 16,
        usage_percent: 50,
      },
      disk_usage: [
        {
          mount_point: '/',
          device: '/dev/sda1',
          total_gb: 500,
          used_gb: 200,
          free_gb: 300,
          usage_percent: 40,
        },
      ],
      network_io: {
        interface_name: 'eth0',
        bytes_sent_mb: 1024,
        bytes_recv_mb: 2048,
        packets_sent: 10000,
        packets_recv: 20000,
      },
      load_average: {
        load_1: 1.5,
        load_5: 1.2,
        load_15: 1.0,
      },
      recorded_at: '2024-01-15T10:30:00Z',
    },
    request_metrics: {
      qps: 100.5,
      avg_response_time: 150.2,
      p95_response_time: 250.5,
      p99_response_time: 350.8,
      total_requests: 100000,
      error_rate: 0.01,
      status_codes: { 200: 95000, 500: 500 },
      time_range: {
        start: '2024-01-15T09:30:00Z',
        end: '2024-01-15T10:30:00Z',
      },
    },
    llm_nodes: [
      {
        node_id: 'node-1',
        node_name: 'LLM Node 1',
        status: 'online',
        gpu_utilization: 80,
        gpu_total_memory_gb: 24,
        gpu_used_memory_gb: 19.2,
        cpu_usage_percent: 45,
        memory_usage_percent: 60,
        requests_per_min: 120,
        avg_response_time_ms: 150,
        active_requests: 10,
        model_count: 5,
        loaded_models: ['gpt-3.5', 'gpt-4'],
        version: '1.0.0',
        last_seen_at: '2024-01-15T10:29:00Z',
      },
    ],
    active_nodes: 1,
    total_nodes: 2,
    updated_at: '2024-01-15T10:30:00Z',
  };

  describe('getDashboardSummary', () => {
    it('should get dashboard summary', async () => {
      const mockResponse: ApiResponse<DashboardSummary> = {
        code: 0,
        message: 'success',
        data: mockDashboardSummary,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getDashboardSummary();

      expect(mockRequest.get).toHaveBeenCalledWith('/monitor/dashboard');
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle empty dashboard data', async () => {
      const emptySummary: DashboardSummary = {
        system_status: {
          disk_usage: [],
          recorded_at: '',
        },
        request_metrics: {
          qps: 0,
          avg_response_time: 0,
          p95_response_time: 0,
          p99_response_time: 0,
          total_requests: 0,
          error_rate: 0,
          status_codes: {},
          time_range: { start: '', end: '' },
        },
        llm_nodes: [],
        active_nodes: 0,
        total_nodes: 0,
        updated_at: '',
      };
      const mockResponse: ApiResponse<DashboardSummary> = {
        code: 0,
        message: 'success',
        data: emptySummary,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getDashboardSummary();

      expect(result.data.data.llm_nodes).toEqual([]);
      expect(result.data.data.active_nodes).toBe(0);
    });
  });

  describe('getSystemMetrics', () => {
    it('should get system metrics', async () => {
      const mockResponse: ApiResponse<SystemMetricsSummary> = {
        code: 0,
        message: 'success',
        data: mockDashboardSummary.system_status,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getSystemMetrics();

      expect(mockRequest.get).toHaveBeenCalledWith('/monitor/system');
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getRequestMetrics', () => {
    it.each([
      { desc: 'no duration', duration: undefined },
      { desc: 'with 1h duration', duration: '1h' },
      { desc: 'with 24h duration', duration: '24h' },
      { desc: 'with 7d duration', duration: '7d' },
    ])('should get request metrics $desc', async ({ duration }) => {
      const mockResponse: ApiResponse<RequestMetricsSummary> = {
        code: 0,
        message: 'success',
        data: mockDashboardSummary.request_metrics,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getRequestMetrics(duration);

      expect(mockRequest.get).toHaveBeenCalledWith('/monitor/requests', {
        params: { duration },
      });
      expect(result.data).toEqual(mockResponse);
    });
  });

  describe('getLLMNodeMetrics', () => {
    it('should get LLM node metrics', async () => {
      const mockResponse: ApiResponse<LLMNodeSummary[]> = {
        code: 0,
        message: 'success',
        data: mockDashboardSummary.llm_nodes,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getLLMNodeMetrics();

      expect(mockRequest.get).toHaveBeenCalledWith('/monitor/llm-nodes');
      expect(result.data).toEqual(mockResponse);
    });

    it('should handle multiple nodes', async () => {
      const multipleNodes: LLMNodeSummary[] = [
        {
          node_id: 'node-1',
          node_name: 'Node 1',
          status: 'online',
          gpu_utilization: 80,
          gpu_total_memory_gb: 24,
          gpu_used_memory_gb: 19.2,
          cpu_usage_percent: 45,
          memory_usage_percent: 60,
          requests_per_min: 120,
          avg_response_time_ms: 150,
          active_requests: 10,
          model_count: 5,
          loaded_models: ['gpt-3.5'],
          version: '1.0.0',
          last_seen_at: '2024-01-15T10:29:00Z',
        },
        {
          node_id: 'node-2',
          node_name: 'Node 2',
          status: 'busy',
          gpu_utilization: 95,
          gpu_total_memory_gb: 48,
          gpu_used_memory_gb: 45.6,
          cpu_usage_percent: 80,
          memory_usage_percent: 75,
          requests_per_min: 200,
          avg_response_time_ms: 200,
          active_requests: 20,
          model_count: 3,
          loaded_models: ['gpt-4'],
          version: '1.0.0',
          last_seen_at: '2024-01-15T10:28:00Z',
        },
      ];
      const mockResponse: ApiResponse<LLMNodeSummary[]> = {
        code: 0,
        message: 'success',
        data: multipleNodes,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getLLMNodeMetrics();

      expect(result.data.data).toHaveLength(2);
      expect(result.data.data[0].status).toBe('online');
      expect(result.data.data[1].status).toBe('busy');
    });

    it('should handle offline nodes', async () => {
      const offlineNodes: LLMNodeSummary[] = [
        {
          node_id: 'node-1',
          node_name: 'Node 1',
          status: 'offline',
          gpu_utilization: 0,
          gpu_total_memory_gb: 24,
          gpu_used_memory_gb: 0,
          cpu_usage_percent: 0,
          memory_usage_percent: 0,
          requests_per_min: 0,
          avg_response_time_ms: 0,
          active_requests: 0,
          model_count: 0,
          loaded_models: [],
          version: '1.0.0',
          last_seen_at: '2024-01-15T09:00:00Z',
        },
      ];
      const mockResponse: ApiResponse<LLMNodeSummary[]> = {
        code: 0,
        message: 'success',
        data: offlineNodes,
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await getLLMNodeMetrics();

      expect(result.data.data[0].status).toBe('offline');
      expect(result.data.data[0].active_requests).toBe(0);
    });
  });

  describe('healthCheck', () => {
    it.each([
      { status: 'healthy', hostname: 'server-1', timestamp: 1705312200 },
      { status: 'degraded', hostname: 'server-2', timestamp: 1705312201 },
      { status: 'unhealthy', hostname: 'server-3', timestamp: 1705312202 },
    ])('should get health status: $status', async ({ status, hostname, timestamp }) => {
      const mockResponse: ApiResponse<{ status: string; hostname: string; timestamp: number }> = {
        code: 0,
        message: 'success',
        data: { status, hostname, timestamp },
      };
      mockRequest.get.mockResolvedValue({ data: mockResponse });

      const result = await healthCheck();

      expect(mockRequest.get).toHaveBeenCalledWith('/monitor/health');
      expect(result.data.data.status).toBe(status);
      expect(result.data.data.hostname).toBe(hostname);
      expect(result.data.data.timestamp).toBe(timestamp);
    });
  });
});
