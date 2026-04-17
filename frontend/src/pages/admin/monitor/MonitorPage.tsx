import { useEffect, useState, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
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
import useAppStore from '@/store/appStore';
import type { DashboardSummary, SystemMetricsSummary, RequestMetricsSummary, LLMNodeSummary } from '@/types';

const { Title, Text } = Typography;

const getStatusColor = (status: string, isDark = true): string => {
  const colorMap: Record<string, string> = {
    online: '#00F5D4',
    offline: '#FF6B6B',
    busy: '#FFBE0B',
    error: '#FF6B6B',
    idle: '#00D9FF',
  };
  return colorMap[status] || (isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)');
};

const formatBytes = (gb: number): string => {
  if (gb >= 1024) {
    return `${(gb / 1024).toFixed(2)} TB`;
  }
  return `${gb.toFixed(2)} GB`;
};

const SystemResourceCard: React.FC<{
  title: string;
  icon: React.ReactNode;
  usage: number;
  details: React.ReactNode;
  gradient: string;
  colors: {
    text: string;
    textSecondary: string;
    trailColor: string;
    cardBg: string;
  };
}> = ({ title, icon, usage, details, gradient, colors }) => (
  <Card 
    className="glass-card" 
    bordered={false} 
    style={{ height: '100%', width: '100%', background: colors.cardBg }}
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
          <Text style={{ fontSize: 13, color: colors.textSecondary }}>{title}</Text>
          <div style={{ fontSize: 28, fontWeight: 700, color: colors.text }}>{usage.toFixed(1)}%</div>
        </div>
      </Space>
      <Progress
        percent={Math.min(usage, 100)}
        strokeColor={usage > 90 ? '#FF6B6B' : usage > 70 ? '#FFBE0B' : '#00F5D4'}
        showInfo={false}
        size="small"
        style={{ width: '100%' }}
        trailColor={colors.trailColor}
      />
      {details}
    </Space>
  </Card>
);

const MonitorPage: React.FC = () => {
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
  const { t } = useTranslation();
  
  const colors = {
    text: isDark ? '#fff' : '#1f2937',
    textSecondary: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
    textTertiary: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)',
    textMuted: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.7)',
    textLight: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.3)',
    trailColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
    cardBg: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.6)',
    segmentedBg: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)',
    tagDefaultBg: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)',
    tagDefaultBorder: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
    tagDefaultColor: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
  };

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dashboard, setDashboard] = useState<DashboardSummary | null>(null);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetricsSummary | null>(null);
  const [requestMetrics, setRequestMetrics] = useState<RequestMetricsSummary | null>(null);
  const [llmNodes, setLLMNodes] = useState<LLMNodeSummary[]>([]);
  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);
  const [refreshInterval, setRefreshInterval] = useState<number>(10);

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
      const errorMsg = err instanceof Error ? err.message : t('monitor.loadFailed');
      setError(errorMsg);
    } finally {
      if (showLoading) setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // auto-refresh
  useInterval(
    () => {
      loadData(false);
    },
    autoRefresh ? refreshInterval * 1000 : null
  );

  const nodeColumns = useMemo(() => [
    {
      title: t('monitor.table.node'),
      dataIndex: 'node_name',
      key: 'node_name',
      render: (name: string, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0}>
          <Text style={{ color: colors.text, fontWeight: 600 }}>{name || record.node_id}</Text>
          <Text style={{ fontSize: 12, color: colors.textTertiary }}>{record.node_id}</Text>
        </Space>
      ),
    },
    {
      title: t('common.status'),
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Badge 
          status={status === 'online' ? 'success' : status === 'error' ? 'error' : 'default'} 
          text={<span style={{ color: getStatusColor(status, isDark) }}>{t(`monitor.status.${status}`)}</span>} 
        />
      ),
    },
    {
      title: 'GPU',
      key: 'gpu',
      width: 150,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0} style={{ width: '100%' }}>
          <Text style={{ fontSize: 12, color: colors.textMuted }}>{t('monitor.gpu.utilization', { value: record.gpu_utilization?.toFixed(1) })}</Text>
          <Progress
            percent={record.gpu_utilization}
            size="small"
            strokeColor={record.gpu_utilization > 90 ? '#FF6B6B' : '#00F5D4'}
            showInfo={false}
            trailColor={colors.trailColor}
          />
          <Text style={{ fontSize: 11, color: colors.textTertiary }}>
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
          trailColor={colors.trailColor}
        />
      ),
    },
    {
      title: t('monitor.table.memory'),
      key: 'memory',
      width: 100,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Progress
          type="circle"
          percent={record.memory_usage_percent}
          size={40}
          strokeColor={record.memory_usage_percent > 80 ? '#FF6B6B' : '#9D4EDD'}
          trailColor={colors.trailColor}
        />
      ),
    },
    {
      title: t('monitor.table.requests'),
      key: 'requests',
      width: 120,
      render: (_: unknown, record: LLMNodeSummary) => (
        <Space direction="vertical" size={0}>
          <Text style={{ fontSize: 12, color: '#00D9FF' }}>{record.requests_per_min} req/min</Text>
          <Text style={{ fontSize: 11, color: colors.textTertiary }}>
            {t('monitor.node.active', { count: record.active_requests })}
          </Text>
          <Text style={{ fontSize: 11, color: colors.textTertiary }}>
            {t('monitor.node.response', { value: record.avg_response_time_ms?.toFixed(0) })}
          </Text>
        </Space>
      ),
    },
    {
      title: t('monitor.table.models'),
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
              <Tag style={{ 
                borderRadius: 6, 
                background: colors.tagDefaultBg,
                borderColor: colors.tagDefaultBorder,
                color: colors.tagDefaultColor,
              }}>+{record.loaded_models.length - 3}</Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: t('monitor.table.version'),
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
  ], [isDark, colors, t]);

  const diskColumns = useMemo(() => [
    {
      title: t('monitor.table.mountPoint'),
      dataIndex: 'mount_point',
      key: 'mount_point',
      render: (mount: string) => <code style={{ fontSize: 12, color: '#00D9FF', fontFamily: 'monospace', background: isDark ? 'transparent' : 'rgba(0, 217, 255, 0.1)', padding: '2px 4px', borderRadius: 4 }}>{mount}</code>,
    },
    {
      title: t('monitor.table.device'),
      dataIndex: 'device',
      key: 'device',
      render: (device: string) => <Text style={{ fontSize: 12, color: colors.textSecondary }}>{device}</Text>,
    },
    {
      title: t('monitor.table.usagePercent'),
      dataIndex: 'usage_percent',
      key: 'usage_percent',
      width: 150,
      render: (percent: number) => (
        <Progress
          percent={Math.round(percent * 10) / 10}
          size="small"
          strokeColor={percent > 90 ? '#FF6B6B' : percent > 70 ? '#FFBE0B' : '#00F5D4'}
          format={(p) => <span style={{ color: colors.text }}>{p}%</span>}
          trailColor={colors.trailColor}
        />
      ),
    },
    {
      title: t('monitor.table.usedTotal'),
      key: 'size',
      width: 150,
      render: (_: unknown, record: { used_gb: number; total_gb: number }) => (
        <Text style={{ fontSize: 12, color: colors.textMuted }}>
          {formatBytes(record.used_gb)} / {formatBytes(record.total_gb)}
        </Text>
      ),
    },
  ], [isDark, colors, t]);

  if (loading && !dashboard) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '60vh' }}>
        <Spin size="large" tip={t('monitor.loadingData')} />
      </div>
    );
  }

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        
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
                  <Title level={3} style={{ margin: 0, color: colors.text }}>{t('monitor.title')}</Title>
                  <Text style={{ color: colors.textSecondary }}>
                    {t('monitor.pageDescription')}
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
                  style={{ background: colors.segmentedBg }}
                />
                <Tag
                  color={autoRefresh ? 'success' : 'default'}
                  style={{ 
                    cursor: 'pointer',
                    background: autoRefresh ? 'rgba(0, 245, 212, 0.15)' : colors.tagDefaultBg,
                    border: `1px solid ${autoRefresh ? 'rgba(0, 245, 212, 0.3)' : colors.tagDefaultBorder}`,
                    color: autoRefresh ? '#00F5D4' : colors.tagDefaultColor,
                  }}
                  onClick={() => setAutoRefresh(!autoRefresh)}
                >
                  {autoRefresh ? t('monitor.autoRefresh.running') : t('monitor.autoRefresh.paused')}
                </Tag>
              </Space>
            </Col>
          </Row>
        </div>

        {error && (
          <Alert
            message={t('monitor.alert.loadFailed')}
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
                {t('common.retry')}
              </Tag>
            }
          />
        )}

        <Card
          title={
            <Space>
              <DesktopOutlined style={{ color: '#00D9FF' }} />
              <span style={{ color: colors.text, fontWeight: 600 }}>{t('monitor.section.systemResources')}</span>
            </Space>
          }
          style={{ marginBottom: 24, background: 'transparent', border: 'none' }}
          className="glass-card"
        >
          <Row gutter={[16, 16]} align="stretch">
            <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
              <SystemResourceCard
                title={t('monitor.system.cpu')}
                icon={<ThunderboltOutlined />}
                usage={systemMetrics?.cpu_usage?.usage_percent || 0}
                gradient="linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)"
                colors={colors}
                details={
                  <Space direction="vertical" size={0}>
                    <Text style={{ fontSize: 12, color: colors.textTertiary }}>
                      {t('monitor.cpu.cores', { count: systemMetrics?.cpu_usage?.core_count })} · {systemMetrics?.cpu_usage?.model_name}
                    </Text>
                    {systemMetrics?.load_average && (
                      <Text style={{ fontSize: 12, color: colors.textTertiary }}>
                        {t('monitor.cpu.load', { load1: systemMetrics.load_average.load_1.toFixed(2), load5: systemMetrics.load_average.load_5.toFixed(2), load15: systemMetrics.load_average.load_15.toFixed(2) })}
                      </Text>
                    )}
                  </Space>
                }
              />
            </Col>
            <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
              <SystemResourceCard
                title={t('monitor.system.memory')}
                icon={<ApiOutlined />}
                usage={systemMetrics?.memory_usage?.usage_percent || 0}
                gradient="linear-gradient(135deg, #9D4EDD 0%, #00D9FF 100%)"
                colors={colors}
                details={
                  <Space direction="vertical" size={0}>
                    <Text style={{ fontSize: 12, color: colors.textTertiary }}>
                      {t('monitor.memory.used', { value: formatBytes(systemMetrics?.memory_usage?.used_gb || 0) })}
                    </Text>
                    <Text style={{ fontSize: 12, color: colors.textTertiary }}>
                      {t('monitor.memory.totalFree', { total: formatBytes(systemMetrics?.memory_usage?.total_gb || 0), free: formatBytes(systemMetrics?.memory_usage?.free_gb || 0) })}
                    </Text>
                  </Space>
                }
              />
            </Col>
            <Col xs={24} sm={12} lg={8} style={{ display: 'flex' }}>
              <Card className="glass-card" bordered={false} style={{ height: '100%', width: '100%', background: colors.cardBg }}>
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
                      <Text style={{ fontSize: 13, color: colors.textSecondary }}>{t('monitor.llmNode.card')}</Text>
                      <div style={{ fontSize: 28, fontWeight: 700, color: colors.text }}>
                        {dashboard?.active_nodes || 0} / {dashboard?.total_nodes || 0}
                      </div>
                    </div>
                  </Space>
                  <Text style={{ fontSize: 12, color: colors.textTertiary }}>
                    {t('monitor.llmNode.activeTotal')}
                  </Text>
                </Space>
              </Card>
            </Col>
          </Row>

          {systemMetrics?.disk_usage && systemMetrics.disk_usage.length > 0 && (
            <div style={{ marginTop: 24 }}>
              <Title level={5} style={{ color: colors.text }}>{t('monitor.section.diskUsage')}</Title>
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

        <Card
          title={
            <Space>
              <ClockCircleOutlined style={{ color: '#FFBE0B' }} />
              <span style={{ color: colors.text, fontWeight: 600 }}>{t('monitor.section.requestPerformance')}</span>
            </Space>
          }
          style={{ marginBottom: 24, background: 'transparent', border: 'none' }}
          className="glass-card"
        >
          <Row gutter={[16, 16]}>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: colors.textSecondary }}>QPS</span>}
                value={requestMetrics?.qps?.toFixed(2) || 0}
                suffix="req/s"
                valueStyle={{ color: '#00D9FF', fontSize: 28, fontWeight: 700 }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: colors.textSecondary }}>{t('monitor.performance.avgResponseTime')}</span>}
                value={requestMetrics?.avg_response_time?.toFixed(2) || 0}
                suffix="ms"
                valueStyle={{ color: '#00F5D4', fontSize: 28, fontWeight: 700 }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: colors.textSecondary }}>{t('monitor.performance.p95ResponseTime')}</span>}
                value={requestMetrics?.p95_response_time?.toFixed(2) || 0}
                suffix="ms"
                valueStyle={{ color: '#FFBE0B', fontSize: 28, fontWeight: 700 }}
              />
            </Col>
            <Col xs={12} sm={6}>
              <Statistic
                title={<span style={{ color: colors.textSecondary }}>{t('monitor.performance.errorRate')}</span>}
                value={requestMetrics?.error_rate?.toFixed(2) || 0}
                suffix="%"
                valueStyle={{ color: (requestMetrics?.error_rate || 0) > 5 ? '#FF6B6B' : '#00F5D4', fontSize: 28, fontWeight: 700 }}
                prefix={(requestMetrics?.error_rate || 0) > 5 ? <WarningOutlined /> : null}
              />
            </Col>
          </Row>

          {requestMetrics?.status_codes && Object.keys(requestMetrics.status_codes).length > 0 && (
            <div style={{ marginTop: 16 }}>
              <Text style={{ fontSize: 12, color: colors.textTertiary }}>{t('monitor.performance.statusCodeDist')}</Text>
              <Space wrap style={{ marginTop: 8 }}>
                {Object.entries(requestMetrics.status_codes).map(([code, count]) => (
                  <Tag
                    key={code}
                    style={{
                      color: code.startsWith('2') ? '#00F5D4' : code.startsWith('4') ? '#FFBE0B' : code.startsWith('5') ? '#FF6B6B' : colors.textSecondary,
                      background: code.startsWith('2') ? 'rgba(0, 245, 212, 0.15)' : code.startsWith('4') ? 'rgba(255, 190, 11, 0.15)' : code.startsWith('5') ? 'rgba(255, 107, 107, 0.15)' : colors.tagDefaultBg,
                      border: `1px solid ${code.startsWith('2') ? 'rgba(0, 245, 212, 0.3)' : code.startsWith('4') ? 'rgba(255, 190, 11, 0.3)' : code.startsWith('5') ? 'rgba(255, 107, 107, 0.3)' : colors.tagDefaultBorder}`,
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

        <Card
          title={
            <Space>
              <CloudServerOutlined style={{ color: '#00F5D4' }} />
              <span style={{ color: colors.text, fontWeight: 600 }}>{t('monitor.nodes.title')}</span>
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
              message={t('monitor.noNodes')}
              description={t('monitor.noNodesDesc')}
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

        <div style={{ textAlign: 'center', marginTop: 24 }}>
          <Text style={{ fontSize: 12, color: colors.textLight }}>
            {t('monitor.dataUpdateTime', { time: dashboard?.updated_at ? new Date(dashboard.updated_at).toLocaleString() : '-' })}
          </Text>
        </div>
      </div>
    </div>
  );
};

export default MonitorPage;
