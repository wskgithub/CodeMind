import { useEffect, useState, useCallback } from 'react';
import {
  Card, Row, Col, Statistic, Table, Tag, Progress, Space,
  Typography, Alert, Spin, Badge, Tooltip, theme, Segmented,
} from 'antd';
import {
  DashboardOutlined,
  DesktopOutlined,
  CloudServerOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  ApiOutlined,
  WarningOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
// 图表组件（如需使用可取消注释）
// import {
//   AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip,
//   ResponsiveContainer, LineChart, Line, Legend,
// } from 'recharts';
import { getDashboardSummary, getSystemMetrics, getRequestMetrics, getLLMNodeMetrics } from '@/services/monitorService';
import { useInterval } from '@/hooks/useInterval';
import type { DashboardSummary, SystemMetricsSummary, RequestMetricsSummary, LLMNodeSummary } from '@/types';

const { Title, Text } = Typography;

/** 获取状态颜色 */
const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    online: 'green',
    offline: 'red',
    busy: 'orange',
    error: 'red',
    idle: 'blue',
  };
  return colorMap[status] || 'default';
};

/** 获取状态文本 */
const getStatusText = (status: string): string => {
  const textMap: Record<string, string> = {
    online: '在线',
    offline: '离线',
    busy: '忙碌',
    error: '错误',
    idle: '空闲',
  };
  return textMap[status] || status;
};

/** 格式化字节大小 */
const formatBytes = (gb: number): string => {
  if (gb >= 1024) {
    return `${(gb / 1024).toFixed(2)} TB`;
  }
  return `${gb.toFixed(2)} GB`;
};

/** 系统资源卡片组件 */
const SystemResourceCard: React.FC<{
  title: string;
  icon: React.ReactNode;
  usage: number;
  details: React.ReactNode;
  color: string;
}> = ({ title, icon, usage, details, color }) => (
  <Card className="glass-card" bordered={false} style={{ height: '100%', width: '100%' }}>
    <Space direction="vertical" style={{ width: '100%', height: '100%' }}>
      <Space align="center" style={{ width: '100%' }}>
        <div
          style={{
            width: 40,
            height: 40,
            borderRadius: 10,
            background: color,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontSize: 20,
          }}
        >
          {icon}
        </div>
        <div style={{ flex: 1 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>{title}</Text>
          <div style={{ fontSize: 24, fontWeight: 700 }}>{usage.toFixed(1)}%</div>
        </div>
      </Space>
      <Progress
        percent={Math.min(usage, 100)}
        strokeColor={usage > 90 ? '#ff4d4f' : usage > 70 ? '#faad14' : '#52c41a'}
        showInfo={false}
        size="small"
        style={{ width: '100%' }}
      />
      {details}
    </Space>
  </Card>
);

/** 监控仪表盘页面 */
const MonitorPage: React.FC = () => {
  const { token: themeToken } = theme.useToken();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dashboard, setDashboard] = useState<DashboardSummary | null>(null);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetricsSummary | null>(null);
  const [requestMetrics, setRequestMetrics] = useState<RequestMetricsSummary | null>(null);
  const [llmNodes, setLLMNodes] = useState<LLMNodeSummary[]>([]);
  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);
  const [refreshInterval, setRefreshInterval] = useState<number>(10);

  // 加载数据
  const loadData = useCallback(async (showLoading = true) => {
    if (showLoading) setLoading(true);
    setError(null);

    try {
      const [dashboardRes, systemRes, requestRes, nodesRes] = await Promise.all([
        getDashboardSummary(),
        getSystemMetrics(),
        getRequestMetrics(),
        getLLMNodeMetrics(),
      ]);

      setDashboard(dashboardRes.data.data);
      setSystemMetrics(systemRes.data.data);
      setRequestMetrics(requestRes.data.data);
      setLLMNodes(nodesRes.data.data || []);
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : '加载数据失败';
      setError(errorMsg);
    } finally {
      if (showLoading) setLoading(false);
    }
  }, []);

  // 初始加载
  useEffect(() => {
    loadData();
  }, [loadData]);

  // 自动刷新
  useInterval(
    () => {
      loadData(false);
    },
    autoRefresh ? refreshInterval * 1000 : null
  );

  // LLM 节点表格列
  const nodeColumns = [
    {
      title: '节点',
      dataIndex: 'node_name',
      key: 'node_name',
      render: (name: string, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0}>
          <Text strong>{name || record.node_id}</Text>
          <Text type="secondary" style={{ fontSize: 12 }}>{record.node_id}</Text>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Badge status={getStatusColor(status) as any} text={getStatusText(status)} />
      ),
    },
    {
      title: 'GPU',
      key: 'gpu',
      width: 150,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0} style={{ width: '100%' }}>
          <Text style={{ fontSize: 12 }}>利用率: {record.gpu_utilization?.toFixed(1)}%</Text>
          <Progress
            percent={record.gpu_utilization}
            size="small"
            strokeColor={record.gpu_utilization > 90 ? '#ff4d4f' : '#52c41a'}
            showInfo={false}
          />
          <Text type="secondary" style={{ fontSize: 11 }}>
            {formatBytes(record.gpu_used_memory_gb)} / {formatBytes(record.gpu_total_memory_gb)}
          </Text>
        </Space>
      ),
    },
    {
      title: 'CPU',
      key: 'cpu',
      width: 100,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Progress
          type="circle"
          percent={record.cpu_usage_percent}
          size={40}
          strokeColor={record.cpu_usage_percent > 80 ? '#ff4d4f' : '#52c41a'}
        />
      ),
    },
    {
      title: '内存',
      key: 'memory',
      width: 100,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Progress
          type="circle"
          percent={record.memory_usage_percent}
          size={40}
          strokeColor={record.memory_usage_percent > 80 ? '#ff4d4f' : '#52c41a'}
        />
      ),
    },
    {
      title: '请求',
      key: 'requests',
      width: 120,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0}>
          <Text style={{ fontSize: 12 }}>{record.requests_per_min} req/min</Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            活跃: {record.active_requests}
          </Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            响应: {record.avg_response_time_ms?.toFixed(0)}ms
          </Text>
        </Space>
      ),
    },
    {
      title: '模型',
      key: 'models',
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space wrap>
          {record.loaded_models?.slice(0, 3).map((model) => (
            <Tag key={model} style={{ fontSize: 11 }}>
              {model}
            </Tag>
          ))}
          {(record.loaded_models?.length || 0) > 3 && (
            <Tooltip title={record.loaded_models?.slice(3).join(', ')}>
              <Tag>+{record.loaded_models.length - 3}</Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 100,
      render: (version: string) => (
        <Tag color="blue" style={{ fontSize: 11 }}>{version || 'unknown'}</Tag>
      ),
    },
  ];

  // 磁盘使用表格列
  const diskColumns = [
    {
      title: '挂载点',
      dataIndex: 'mount_point',
      key: 'mount_point',
      render: (mount: string) => <code style={{ fontSize: 12 }}>{mount}</code>,
    },
    {
      title: '设备',
      dataIndex: 'device',
      key: 'device',
      render: (device: string) => <Text type="secondary" style={{ fontSize: 12 }}>{device}</Text>,
    },
    {
      title: '使用率',
      dataIndex: 'usage_percent',
      key: 'usage_percent',
      width: 150,
      render: (percent: number) => (
        <Progress
          percent={Math.round(percent * 10) / 10}
          size="small"
          strokeColor={percent > 90 ? '#ff4d4f' : percent > 70 ? '#faad14' : '#52c41a'}
          format={(p) => `${p}%`}
        />
      ),
    },
    {
      title: '已用 / 总计',
      key: 'size',
      width: 150,
      render: (_: unknown, record: { used_gb: number; total_gb: number }) => (
        <Text style={{ fontSize: 12 }}>
          {formatBytes(record.used_gb)} / {formatBytes(record.total_gb)}
        </Text>
      ),
    },
  ];

  if (loading && !dashboard) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '60vh' }}>
        <Spin size="large" tip="加载监控数据中..." />
      </div>
    );
  }

  return (
    <div className="animate-fade-in-up">
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <div
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: 12,
                  background: 'linear-gradient(135deg, #2B7CB3 0%, #4BA3D4 100%)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: '#fff',
                  fontSize: 24,
                }}
              >
                <DashboardOutlined />
              </div>
              <div>
                <Title level={3} style={{ margin: 0 }}>系统监控</Title>
                <Text type="secondary">
                  实时监控服务器资源、请求性能和 LLM 节点状态
                </Text>
              </div>
            </Space>
          </Col>
          <Col>
            <Space>
              <Segmented
                options={[
                  { label: '5s', value: 5 },
                  { label: '10s', value: 10 },
                  { label: '30s', value: 30 },
                ]}
                value={refreshInterval}
                onChange={(value) => setRefreshInterval(value as number)}
                disabled={!autoRefresh}
              />
              <Tag
                color={autoRefresh ? 'green' : 'default'}
                style={{ cursor: 'pointer' }}
                onClick={() => setAutoRefresh(!autoRefresh)}
              >
                {autoRefresh ? '自动刷新中' : '自动刷新已暂停'}
              </Tag>
            </Space>
          </Col>
        </Row>
      </div>

      {error && (
        <Alert
          message="加载失败"
          description={error}
          type="error"
          showIcon
          style={{ marginBottom: 24 }}
          action={
            <Tag icon={<ReloadOutlined />} color="error" style={{ cursor: 'pointer' }} onClick={() => loadData()}>
              重试
            </Tag>
          }
        />
      )}

      {/* 系统资源概览 */}
      <Card
        title={
          <Space>
            <DesktopOutlined />
            系统资源
          </Space>
        }
        style={{ marginBottom: 24 }}
        className="glass-card"
      >
        <Row gutter={[16, 16]} align="stretch">
          <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
            <SystemResourceCard
              title="CPU 使用率"
              icon={<ThunderboltOutlined />}
              usage={systemMetrics?.cpu_usage?.usage_percent || 0}
              color="#1677ff"
              details={
                <Space direction="vertical" size={0}>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {systemMetrics?.cpu_usage?.core_count} 核心 · {systemMetrics?.cpu_usage?.model_name}
                  </Text>
                  {systemMetrics?.load_average && (
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      负载: {systemMetrics.load_average.load_1.toFixed(2)} / {systemMetrics.load_average.load_5.toFixed(2)} / {systemMetrics.load_average.load_15.toFixed(2)}
                    </Text>
                  )}
                </Space>
              }
            />
          </Col>
          <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
            <SystemResourceCard
              title="内存使用率"
              icon={<ApiOutlined />}
              usage={systemMetrics?.memory_usage?.usage_percent || 0}
              color="#52c41a"
              details={
                <Space direction="vertical" size={0}>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    已用: {formatBytes(systemMetrics?.memory_usage?.used_gb || 0)}
                  </Text>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    总计: {formatBytes(systemMetrics?.memory_usage?.total_gb || 0)} · 空闲: {formatBytes(systemMetrics?.memory_usage?.free_gb || 0)}
                  </Text>
                </Space>
              }
            />
          </Col>
          <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
            <Card className="glass-card" bordered={false} style={{ height: '100%', width: '100%' }}>
              <Space direction="vertical" style={{ width: '100%', height: '100%' }}>
                <Space align="center" style={{ width: '100%' }}>
                  <div
                    style={{
                      width: 40,
                      height: 40,
                      borderRadius: 10,
                      background: '#722ed1',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#fff',
                      fontSize: 20,
                    }}
                  >
                    <CloudServerOutlined />
                  </div>
                  <div style={{ flex: 1 }}>
                    <Text type="secondary" style={{ fontSize: 12 }}>LLM 节点</Text>
                    <div style={{ fontSize: 24, fontWeight: 700 }}>
                      {dashboard?.active_nodes || 0} / {dashboard?.total_nodes || 0}
                    </div>
                  </div>
                </Space>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  活跃节点 / 总节点数
                </Text>
              </Space>
            </Card>
          </Col>
        </Row>

        {/* 磁盘使用情况 */}
        {systemMetrics?.disk_usage && systemMetrics.disk_usage.length > 0 && (
          <div style={{ marginTop: 24 }}>
            <Title level={5}>磁盘使用</Title>
            <Table
              columns={diskColumns}
              dataSource={[...systemMetrics.disk_usage].sort((a, b) => a.mount_point.localeCompare(b.mount_point))}
              rowKey="mount_point"
              size="small"
              pagination={false}
              bordered={false}
            />
          </div>
        )}
      </Card>

      {/* 请求性能概览 */}
      <Card
        title={
          <Space>
            <ClockCircleOutlined />
            请求性能
          </Space>
        }
        style={{ marginBottom: 24 }}
        className="glass-card"
      >
        <Row gutter={[16, 16]}>
          <Col xs={12} sm={6}>
            <Statistic
              title="QPS"
              value={requestMetrics?.qps?.toFixed(2) || 0}
              suffix="req/s"
              valueStyle={{ color: themeToken.colorPrimary }}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Statistic
              title="平均响应时间"
              value={requestMetrics?.avg_response_time?.toFixed(2) || 0}
              suffix="ms"
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Statistic
              title="P95 响应时间"
              value={requestMetrics?.p95_response_time?.toFixed(2) || 0}
              suffix="ms"
              valueStyle={{ color: '#faad14' }}
            />
          </Col>
          <Col xs={12} sm={6}>
            <Statistic
              title="错误率"
              value={requestMetrics?.error_rate?.toFixed(2) || 0}
              suffix="%"
              valueStyle={{ color: (requestMetrics?.error_rate || 0) > 5 ? '#ff4d4f' : '#52c41a' }}
              prefix={(requestMetrics?.error_rate || 0) > 5 ? <WarningOutlined /> : null}
            />
          </Col>
        </Row>

        {/* HTTP 状态码分布 */}
        {requestMetrics?.status_codes && Object.keys(requestMetrics.status_codes).length > 0 && (
          <div style={{ marginTop: 16 }}>
            <Text type="secondary" style={{ fontSize: 12 }}>HTTP 状态码分布</Text>
            <Space wrap style={{ marginTop: 8 }}>
              {Object.entries(requestMetrics.status_codes).map(([code, count]) => (
                <Tag
                  key={code}
                  color={code.startsWith('2') ? 'success' : code.startsWith('4') ? 'warning' : code.startsWith('5') ? 'error' : 'default'}
                >
                  {code}: {count}
                </Tag>
              ))}
            </Space>
          </div>
        )}
      </Card>

      {/* LLM 节点状态 */}
      <Card
        title={
          <Space>
            <CloudServerOutlined />
            LLM 节点状态
            {llmNodes.length > 0 && (
              <Badge
                count={llmNodes.filter(n => n.status === 'online').length}
                style={{ backgroundColor: '#52c41a' }}
                overflowCount={99}
              />
            )}
          </Space>
        }
        className="glass-card"
      >
        {llmNodes.length === 0 ? (
          <Alert
            message="暂无 LLM 节点数据"
            description="请确保 LLM 服务节点已正确配置并正在运行。节点需要通过 /api/v1/monitor/nodes/report 接口定期上报状态。"
            type="info"
            showIcon
          />
        ) : (
          <Table
            columns={nodeColumns}
            dataSource={llmNodes}
            rowKey="node_id"
            size="middle"
            scroll={{ x: 'max-content' }}
            pagination={{ pageSize: 10 }}
          />
        )}
      </Card>

      {/* 数据更新时间 */}
      <div style={{ textAlign: 'center', marginTop: 24 }}>
        <Text type="secondary" style={{ fontSize: 12 }}>
          数据更新时间: {dashboard?.updated_at ? new Date(dashboard.updated_at).toLocaleString() : '-'}
        </Text>
      </div>
    </div>
  );
};

export default MonitorPage;
