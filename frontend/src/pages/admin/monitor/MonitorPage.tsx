import { useEffect, useState, useCallback } from 'react';
import {
  Card, Row, Col, Statistic, Table, Tag, Progress, Space,
  Typography, Alert, Spin, Badge, Tooltip, Segmented,
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
import { getDashboardSummary, getSystemMetrics, getRequestMetrics, getLLMNodeMetrics } from '@/services/monitorService';
import { useInterval } from '@/hooks/useInterval';
import type { DashboardSummary, SystemMetricsSummary, RequestMetricsSummary, LLMNodeSummary } from '@/types';

const { Title, Text } = Typography;

/** 获取状态颜色 - 新设计 */
const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    online: '#00F5D4',
    offline: '#FF6B6B',
    busy: '#FFBE0B',
    error: '#FF6B6B',
    idle: '#00D9FF',
  };
  return colorMap[status] || 'rgba(255, 255, 255, 0.5)';
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

/** 系统资源卡片组件 - 新设计 */
const SystemResourceCard: React.FC<{
  title: string;
  icon: React.ReactNode;
  usage: number;
  details: React.ReactNode;
  gradient: string;
}> = ({ title, icon, usage, details, gradient }) => (
  <Card 
    className="glass-card" 
    bordered={false} 
    style={{ height: '100%', width: '100%', background: 'rgba(255, 255, 255, 0.02)' }}
  >
    <Space direction="vertical" style={{ width: '100%', height: '100%' }}>
      <Space align="center" style={{ width: '100%' }}>
        <div
          style={{
            width: 48,
            height: 48,
            borderRadius: 14,
            background: gradient,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
            fontSize: 24,
            boxShadow: '0 4px 16px rgba(0, 0, 0, 0.2)',
          }}
        >
          {icon}
        </div>
        <div style={{ flex: 1 }}>
          <Text style={{ fontSize: 13, color: 'rgba(255, 255, 255, 0.5)' }}>{title}</Text>
          <div style={{ fontSize: 28, fontWeight: 700, color: '#fff' }}>{usage.toFixed(1)}%</div>
        </div>
      </Space>
      <Progress
        percent={Math.min(usage, 100)}
        strokeColor={usage > 90 ? '#FF6B6B' : usage > 70 ? '#FFBE0B' : '#00F5D4'}
        showInfo={false}
        size="small"
        style={{ width: '100%' }}
        trailColor="rgba(255, 255, 255, 0.1)"
      />
      {details}
    </Space>
  </Card>
);

/** 监控仪表盘页面 — 与首页/登录页新设计风格统一 */
const MonitorPage: React.FC = () => {
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

  // LLM 节点表格列 - 新设计
  const nodeColumns = [
    {
      title: '节点',
      dataIndex: 'node_name',
      key: 'node_name',
      render: (name: string, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0}>
          <Text style={{ color: '#fff', fontWeight: 600 }}>{name || record.node_id}</Text>
          <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>{record.node_id}</Text>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Badge 
          status={status === 'online' ? 'success' : status === 'error' ? 'error' : 'default'} 
          text={<span style={{ color: getStatusColor(status) }}>{getStatusText(status)}</span>} 
        />
      ),
    },
    {
      title: 'GPU',
      key: 'gpu',
      width: 150,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0} style={{ width: '100%' }}>
          <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.7)' }}>利用率: {record.gpu_utilization?.toFixed(1)}%</Text>
          <Progress
            percent={record.gpu_utilization}
            size="small"
            strokeColor={record.gpu_utilization > 90 ? '#FF6B6B' : '#00F5D4'}
            showInfo={false}
            trailColor="rgba(255, 255, 255, 0.1)"
          />
          <Text style={{ fontSize: 11, color: 'rgba(255, 255, 255, 0.4)' }}>
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
          strokeColor={record.cpu_usage_percent > 80 ? '#FF6B6B' : '#00D9FF'}
          trailColor="rgba(255, 255, 255, 0.1)"
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
          strokeColor={record.memory_usage_percent > 80 ? '#FF6B6B' : '#9D4EDD'}
          trailColor="rgba(255, 255, 255, 0.1)"
        />
      ),
    },
    {
      title: '请求',
      key: 'requests',
      width: 120,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0}>
          <Text style={{ fontSize: 12, color: '#00D9FF' }}>{record.requests_per_min} req/min</Text>
          <Text style={{ fontSize: 11, color: 'rgba(255, 255, 255, 0.4)' }}>
            活跃: {record.active_requests}
          </Text>
          <Text style={{ fontSize: 11, color: 'rgba(255, 255, 255, 0.4)' }}>
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
            <Tag 
              key={model} 
              style={{ 
                fontSize: 11, 
                color: '#FFBE0B',
                background: 'rgba(255, 190, 11, 0.15)',
                border: '1px solid rgba(255, 190, 11, 0.3)',
                borderRadius: 6,
              }}
            >
              {model}
            </Tag>
          ))}
          {(record.loaded_models?.length || 0) > 3 && (
            <Tooltip title={record.loaded_models?.slice(3).join(', ')}>
              <Tag style={{ borderRadius: 6 }}>+{record.loaded_models.length - 3}</Tag>
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
        <Tag style={{ 
          fontSize: 11, 
          color: '#00F5D4',
          background: 'rgba(0, 245, 212, 0.15)',
          border: '1px solid rgba(0, 245, 212, 0.3)',
          borderRadius: 6,
        }}>
          {version || 'unknown'}
        </Tag>
      ),
    },
  ];

  // 磁盘使用表格列 - 新设计
  const diskColumns = [
    {
      title: '挂载点',
      dataIndex: 'mount_point',
      key: 'mount_point',
      render: (mount: string) => <code style={{ fontSize: 12, color: '#00D9FF', fontFamily: 'monospace' }}>{mount}</code>,
    },
    {
      title: '设备',
      dataIndex: 'device',
      key: 'device',
      render: (device: string) => <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.5)' }}>{device}</Text>,
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
          strokeColor={percent > 90 ? '#FF6B6B' : percent > 70 ? '#FFBE0B' : '#00F5D4'}
          format={(p) => <span style={{ color: '#fff' }}>{p}%</span>}
          trailColor="rgba(255, 255, 255, 0.1)"
        />
      ),
    },
    {
      title: '已用 / 总计',
      key: 'size',
      width: 150,
      render: (_: unknown, record: { used_gb: number; total_gb: number }) => (
        <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.7)' }}>
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
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
        {/* 页面头部 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <Row justify="space-between" align="middle">
            <Col>
              <Space>
                <div
                  style={{
                    width: 52,
                    height: 52,
                    borderRadius: 16,
                    background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: '#fff',
                    fontSize: 28,
                    boxShadow: '0 4px 20px rgba(0, 217, 255, 0.3)',
                  }}
                >
                  <DashboardOutlined />
                </div>
                <div>
                  <Title level={3} style={{ margin: 0, color: '#fff' }}>系统监控</Title>
                  <Text style={{ color: 'rgba(255, 255, 255, 0.5)' }}>
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
                  style={{ background: 'rgba(255, 255, 255, 0.05)' }}
                />
                <Tag
                  color={autoRefresh ? 'success' : 'default'}
                  style={{ 
                    cursor: 'pointer',
                    background: autoRefresh ? 'rgba(0, 245, 212, 0.15)' : 'rgba(255, 255, 255, 0.05)',
                    border: `1px solid ${autoRefresh ? 'rgba(0, 245, 212, 0.3)' : 'rgba(255, 255, 255, 0.1)'}`,
                    color: autoRefresh ? '#00F5D4' : 'rgba(255, 255, 255, 0.5)',
                  }}
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
            style={{ marginBottom: 24, background: 'rgba(255, 107, 107, 0.1)', borderColor: 'rgba(255, 107, 107, 0.3)' }}
            action={
              <Tag 
                icon={<ReloadOutlined />} 
                color="error" 
                style={{ cursor: 'pointer' }} 
                onClick={() => loadData()}
              >
                重试
              </Tag>
            }
          />
        )}

        {/* 系统资源概览 - 新设计 */}
        <Card
          title={
            <Space>
              <DesktopOutlined style={{ color: '#00D9FF' }} />
              <span style={{ color: '#fff', fontWeight: 600 }}>系统资源</span>
            </Space>
          }
          style={{ marginBottom: 24, background: 'transparent', border: 'none' }}
          className="glass-card"
        >
          <Row gutter={[16, 16]} align="stretch">
            <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
              <SystemResourceCard
                title="CPU 使用率"
                icon={<ThunderboltOutlined />}
                usage={systemMetrics?.cpu_usage?.usage_percent || 0}
                gradient="linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)"
                details={
                  <Space direction="vertical" size={0}>
                    <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>
                      {systemMetrics?.cpu_usage?.core_count} 核心 · {systemMetrics?.cpu_usage?.model_name}
                    </Text>
                    {systemMetrics?.load_average && (
                      <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>
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
                gradient="linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)"
                details={
                  <Space direction="vertical" size={0}>
                    <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>
                      已用: {formatBytes(systemMetrics?.memory_usage?.used_gb || 0)}
                    </Text>
                    <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>
                      总计: {formatBytes(systemMetrics?.memory_usage?.total_gb || 0)} · 空闲: {formatBytes(systemMetrics?.memory_usage?.free_gb || 0)}
                    </Text>
                  </Space>
                }
              />
            </Col>
            <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
              <Card className="glass-card" bordered={false} style={{ height: '100%', width: '100%', background: 'rgba(255, 255, 255, 0.02)' }}>
                <Space direction="vertical" style={{ width: '100%', height: '100%' }}>
                  <Space align="center" style={{ width: '100%' }}>
                    <div
                      style={{
                        width: 48,
                        height: 48,
                        borderRadius: 14,
                        background: 'linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#fff',
                        fontSize: 24,
                        boxShadow: '0 4px 16px rgba(0, 0, 0, 0.2)',
                      }}
                    >
                      <CloudServerOutlined />
                    </div>
                    <div style={{ flex: 1 }}>
                      <Text style={{ fontSize: 13, color: 'rgba(255, 255, 255, 0.5)' }}>LLM 节点</Text>
                      <div style={{ fontSize: 28, fontWeight: 700, color: '#fff' }}>
                        {dashboard?.active_nodes || 0} / {dashboard?.total_nodes || 0}
                      </div>
                    </div>
                  </Space>
                  <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>
                    活跃节点 / 总节点数
                  </Text>
                </Space>
              </Card>
            </Col>
          </Row>

          {/* 磁盘使用情况 */}
          {systemMetrics?.disk_usage && systemMetrics.disk_usage.length > 0 && (
            <div style={{ marginTop: 24 }}>
              <Title level={5} style={{ color: '#fff' }}>磁盘使用</Title>
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

        {/* 请求性能概览 - 新设计 */}
        <Card
          title={
            <Space>
              <ClockCircleOutlined style={{ color: '#FFBE0B' }} />
              <span style={{ color: '#fff', fontWeight: 600 }}>请求性能</span>
            </Space>
          }
          style={{ marginBottom: 24, background: 'transparent', border: 'none' }}
          className="glass-card"
        >
          <Row gutter={[16, 16]}>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>QPS</span>}
                value={requestMetrics?.qps?.toFixed(2) || 0}
                suffix="req/s"
                valueStyle={{ color: '#00D9FF', fontSize: 28, fontWeight: 700 }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>平均响应时间</span>}
                value={requestMetrics?.avg_response_time?.toFixed(2) || 0}
                suffix="ms"
                valueStyle={{ color: '#00F5D4', fontSize: 28, fontWeight: 700 }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>P95 响应时间</span>}
                value={requestMetrics?.p95_response_time?.toFixed(2) || 0}
                suffix="ms"
                valueStyle={{ color: '#FFBE0B', fontSize: 28, fontWeight: 700 }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: 'rgba(255, 255, 255, 0.5)' }}>错误率</span>}
                value={requestMetrics?.error_rate?.toFixed(2) || 0}
                suffix="%"
                valueStyle={{ color: (requestMetrics?.error_rate || 0) > 5 ? '#FF6B6B' : '#00F5D4', fontSize: 28, fontWeight: 700 }}
                prefix={(requestMetrics?.error_rate || 0) > 5 ? <WarningOutlined /> : null}
              />
            </Col>
          </Row>

          {/* HTTP 状态码分布 */}
          {requestMetrics?.status_codes && Object.keys(requestMetrics.status_codes).length > 0 && (
            <div style={{ marginTop: 16 }}>
              <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.4)' }}>HTTP 状态码分布</Text>
              <Space wrap style={{ marginTop: 8 }}>
                {Object.entries(requestMetrics.status_codes).map(([code, count]) => (
                  <Tag
                    key={code}
                    style={{
                      color: code.startsWith('2') ? '#00F5D4' : code.startsWith('4') ? '#FFBE0B' : code.startsWith('5') ? '#FF6B6B' : 'rgba(255, 255, 255, 0.6)',
                      background: code.startsWith('2') ? 'rgba(0, 245, 212, 0.15)' : code.startsWith('4') ? 'rgba(255, 190, 11, 0.15)' : code.startsWith('5') ? 'rgba(255, 107, 107, 0.15)' : 'rgba(255, 255, 255, 0.05)',
                      border: `1px solid ${code.startsWith('2') ? 'rgba(0, 245, 212, 0.3)' : code.startsWith('4') ? 'rgba(255, 190, 11, 0.3)' : code.startsWith('5') ? 'rgba(255, 107, 107, 0.3)' : 'rgba(255, 255, 255, 0.1)'}`,
                      borderRadius: 6,
                    }}
                  >
                    {code}: {count}
                  </Tag>
                ))}
              </Space>
            </div>
          )}
        </Card>

        {/* LLM 节点状态 - 新设计 */}
        <Card
          title={
            <Space>
              <CloudServerOutlined style={{ color: '#00F5D4' }} />
              <span style={{ color: '#fff', fontWeight: 600 }}>LLM 节点状态</span>
              {llmNodes.length > 0 && (
                <Badge
                  count={llmNodes.filter(n => n.status === 'online').length}
                  style={{ backgroundColor: '#00F5D4', color: '#000' }}
                  overflowCount={99}
                />
              )}
            </Space>
          }
          className="glass-card"
          style={{ background: 'transparent', border: 'none' }}
        >
          {llmNodes.length === 0 ? (
            <Alert
              message="暂无 LLM 节点数据"
              description="请确保 LLM 服务节点已正确配置并正在运行。节点需要通过 /api/v1/monitor/nodes/report 接口定期上报状态。"
              type="info"
              showIcon
              style={{ background: 'rgba(0, 217, 255, 0.1)', borderColor: 'rgba(0, 217, 255, 0.2)' }}
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
          <Text style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.3)' }}>
            数据更新时间: {dashboard?.updated_at ? new Date(dashboard.updated_at).toLocaleString() : '-'}
          </Text>
        </div>
      </div>
    </div>
  );
};

export default MonitorPage;
