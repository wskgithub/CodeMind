import { useEffect, useState, useRef } from 'react';
import { Col, Row, Statistic, Spin, Tag, Empty, theme } from 'antd';
import {
  ThunderboltOutlined,
  MessageOutlined,
  TeamOutlined,
  KeyOutlined,
  ArrowUpOutlined,
} from '@ant-design/icons';
import * as echarts from 'echarts';
import { getOverview } from '@/services/statsService';
import { getUsageStats } from '@/services/statsService';
import { listAnnouncements } from '@/services/systemService';
import useAuthStore from '@/store/authStore';
import type { StatsOverview, UsageItem, Announcement } from '@/types';
import UsageProgressCards from '@/components/common/UsageProgressCards';

/** 图标包裹层 — 渐变圆形背景 */
const StatIcon = ({
  icon,
  gradient,
}: {
  icon: React.ReactNode;
  gradient: string;
}) => (
  <span
    className="flex items-center justify-center w-10 h-10 rounded-full shrink-0"
    style={{
      background: gradient,
      color: '#fff',
    }}
  >
    {icon}
  </span>
);

/** 仪表盘页面 — Glassmorphism 风格，展示用量总览和趋势图表 */
const DashboardPage = () => {
  const { token } = theme.useToken();
  const { user } = useAuthStore();
  const [overview, setOverview] = useState<StatsOverview | null>(null);
  const [usageData, setUsageData] = useState<UsageItem[]>([]);
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [loading, setLoading] = useState(true);
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  const isAdmin = user?.role === 'super_admin' || user?.role === 'dept_manager';

  useEffect(() => {
    loadData();
    return () => {
      chartInstance.current?.dispose();
    };
  }, []);

  // 图表数据更新后渲染
  useEffect(() => {
    if (usageData.length > 0 && chartRef.current) {
      renderChart();
    }
  }, [usageData, token]);

  // 窗口 resize 时调整图表
  useEffect(() => {
    const handleResize = () => chartInstance.current?.resize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [overviewRes, usageRes, annRes] = await Promise.all([
        getOverview(),
        getUsageStats({ period: 'daily' }),
        listAnnouncements(),
      ]);
      setOverview(overviewRes.data.data);
      setUsageData(usageRes.data.data?.items || []);
      setAnnouncements(annRes.data.data || []);
    } catch {
      // 错误已由拦截器统一处理
    } finally {
      setLoading(false);
    }
  };

  const renderChart = () => {
    if (!chartRef.current) return;

    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const dates = usageData.map((d) => d.date);
    const tokens = usageData.map((d) => d.total_tokens);
    const requests = usageData.map((d) => d.request_count);

    chartInstance.current.setOption({
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
      },
      legend: { data: ['Token 用量', '请求次数'] },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: {
          formatter: (val: string) => val.slice(5), // 只显示 MM-DD
        },
      },
      yAxis: [
        {
          type: 'value',
          name: 'Token 用量',
          axisLabel: {
            formatter: (val: number) =>
              val >= 1000000 ? `${(val / 1000000).toFixed(1)}M` :
              val >= 1000 ? `${(val / 1000).toFixed(0)}K` : String(val),
          },
        },
        {
          type: 'value',
          name: '请求次数',
        },
      ],
      series: [
        {
          name: 'Token 用量',
          type: 'bar',
          data: tokens,
          itemStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: token.colorPrimary },
              { offset: 1, color: token.colorInfo },
            ]),
            borderRadius: [4, 4, 0, 0],
          },
        },
        {
          name: '请求次数',
          type: 'line',
          yAxisIndex: 1,
          data: requests,
          smooth: true,
          lineStyle: { color: token.colorSuccess, width: 2 },
          itemStyle: { color: token.colorSuccess },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(82, 196, 26, 0.25)' },
              { offset: 1, color: 'rgba(82, 196, 26, 0.02)' },
            ]),
          },
        },
      ],
    });
  };

  /** 格式化大数字 */
  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(2)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  if (loading) {
    return (
      <div
        className="flex items-center justify-center min-h-[320px]"
        style={{ color: token.colorPrimary }}
      >
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="animate-fade-in-up">
      <UsageProgressCards />
      <h2 style={{ marginBottom: 24, color: token.colorTextHeading }}>
        欢迎回来，{user?.display_name || user?.username}
      </h2>

      {/* 统计卡片 — 使用 stat-card 类 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.05s' }}>
            <div className="flex items-center gap-3">
              <StatIcon
                icon={<ThunderboltOutlined style={{ fontSize: 20 }} />}
                gradient="linear-gradient(135deg, #2B7CB3 0%, #4BA3D4 100%)"
              />
              <Statistic
                title="今日 Token 用量"
                value={overview?.today.total_tokens || 0}
                formatter={(val) => formatNumber(Number(val))}
                valueStyle={{ color: token.colorTextHeading }}
                suffix={
                  <ArrowUpOutlined style={{ fontSize: 12, color: token.colorSuccess, marginLeft: 4 }} />
                }
              />
            </div>
          </div>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
            <div className="flex items-center gap-3">
              <StatIcon
                icon={<MessageOutlined style={{ fontSize: 20 }} />}
                gradient="linear-gradient(135deg, #722ed1 0%, #b37feb 100%)"
              />
              <Statistic
                title="今日请求数"
                value={overview?.today.total_requests || 0}
                valueStyle={{ color: token.colorTextHeading }}
              />
            </div>
          </div>
        </Col>
        {isAdmin && (
          <>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.15s' }}>
                <div className="flex items-center gap-3">
                  <StatIcon
                    icon={<TeamOutlined style={{ fontSize: 20 }} />}
                    gradient="linear-gradient(135deg, #13c2c2 0%, #36cfc9 100%)"
                  />
                  <Statistic
                    title="总用户数"
                    value={overview?.total_users || 0}
                    valueStyle={{ color: token.colorTextHeading }}
                  />
                </div>
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
                <div className="flex items-center gap-3">
                  <StatIcon
                    icon={<KeyOutlined style={{ fontSize: 20 }} />}
                    gradient="linear-gradient(135deg, #faad14 0%, #ffc53d 100%)"
                  />
                  <Statistic
                    title="活跃 API Key"
                    value={overview?.total_api_keys || 0}
                    valueStyle={{ color: token.colorTextHeading }}
                  />
                </div>
              </div>
            </Col>
          </>
        )}
        {!isAdmin && (
          <>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.15s' }}>
                <div className="flex items-center gap-3">
                  <StatIcon
                    icon={<ThunderboltOutlined style={{ fontSize: 20 }} />}
                    gradient="linear-gradient(135deg, #13c2c2 0%, #36cfc9 100%)"
                  />
                  <Statistic
                    title="本月 Token 用量"
                    value={overview?.this_month.total_tokens || 0}
                    formatter={(val) => formatNumber(Number(val))}
                    valueStyle={{ color: token.colorTextHeading }}
                  />
                </div>
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
                <div className="flex items-center gap-3">
                  <StatIcon
                    icon={<MessageOutlined style={{ fontSize: 20 }} />}
                    gradient="linear-gradient(135deg, #faad14 0%, #ffc53d 100%)"
                  />
                  <Statistic
                    title="本月请求数"
                    value={overview?.this_month.total_requests || 0}
                    valueStyle={{ color: token.colorTextHeading }}
                  />
                </div>
              </div>
            </Col>
          </>
        )}
      </Row>

      {/* 图表区域 — 玻璃态卡片 */}
      <div
        className="glass-card animate-fade-in-up mt-4 p-6"
        style={{ marginTop: 16, animationDelay: '0.25s' }}
      >
        <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>近 30 天用量趋势</h3>
        {usageData.length > 0 ? (
          <div
            ref={chartRef}
            style={{
              height: 360,
              width: '100%',
              background: 'var(--glass-bg)',
              backdropFilter: 'blur(16px)',
              WebkitBackdropFilter: 'blur(16px)',
              border: '1px solid var(--glass-border)',
              borderRadius: 12,
            }}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无用量数据"
            style={{ padding: 40 }}
          />
        )}
      </div>

      {/* 公告区域 — 玻璃态卡片 */}
      {announcements.length > 0 && (
        <div
          className="glass-card animate-fade-in-up mt-4 p-6"
          style={{ marginTop: 16, animationDelay: '0.3s' }}
        >
          <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>系统公告</h3>
          <div>
            {announcements.slice(0, 5).map((ann, index) => (
              <div
                key={ann.id}
                style={{
                  paddingBottom: 16,
                  marginBottom: index < 4 ? 16 : 0,
                  borderBottom: index < 4 ? '1px solid var(--glass-border)' : 'none',
                }}
              >
                <div className="flex items-center gap-2 flex-wrap">
                  {ann.pinned && <Tag color="red">置顶</Tag>}
                  <strong style={{ color: token.colorTextHeading }}>{ann.title}</strong>
                  <span
                    style={{
                      color: token.colorTextTertiary,
                      fontSize: 12,
                      marginLeft: 'auto',
                    }}
                  >
                    {ann.created_at?.slice(0, 10)}
                  </span>
                </div>
                <div
                  style={{
                    color: token.colorTextSecondary,
                    fontSize: 13,
                    marginTop: 8,
                  }}
                >
                  {ann.content.length > 120 ? ann.content.slice(0, 120) + '...' : ann.content}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default DashboardPage;
