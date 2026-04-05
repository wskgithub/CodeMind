import { useEffect, useState, useRef } from 'react';
import { Col, Row, Statistic, Spin, Tag, Empty } from 'antd';
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
import useAppStore from '@/store/appStore';
import type { StatsOverview, UsageItem, Announcement } from '@/types';
import UsageProgressCards from '@/components/common/UsageProgressCards';

/** 图标包裹层 — 渐变圆形背景 - 新设计 */
const StatIcon = ({
  icon,
  gradient,
}: {
  icon: React.ReactNode;
  gradient: string;
}) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: gradient,
      color: '#fff',
      boxShadow: '0 4px 16px rgba(0, 0, 0, 0.2)',
    }}
  >
    {icon}
  </span>
);

/** 仪表盘页面 — 与首页/登录页新设计风格统一 */
const DashboardPage = () => {
  const user = useAuthStore((s) => s.user);
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
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
  }, [usageData]);

  // 主题切换时重新渲染图表
  useEffect(() => {
    if (usageData.length > 0 && chartInstance.current) {
      renderChart();
    }
  }, [themeMode]);

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

  const hasThirdPartyData = usageData.some((d) => (d.third_party_total_tokens || 0) > 0);

  const renderChart = () => {
    if (!chartRef.current) return;

    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const dates = usageData.map((d) => d.date);
    const tokens = usageData.map((d) => d.total_tokens);
    const tpTokens = usageData.map((d) => d.third_party_total_tokens || 0);
    const requests = usageData.map((d) => (d.request_count || 0) + (d.third_party_request_count || 0));

    const legendData = hasThirdPartyData
      ? ['平台 Token', '第三方 Token', '请求次数']
      : ['Token 用量', '请求次数'];

    const barSeries: echarts.SeriesOption[] = [
      {
        name: hasThirdPartyData ? '平台 Token' : 'Token 用量',
        type: 'bar',
        stack: 'tokens',
        data: tokens,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#00D9FF' },
            { offset: 1, color: '#9D4EDD' },
          ]),
          borderRadius: hasThirdPartyData ? [0, 0, 0, 0] : [4, 4, 0, 0],
        },
      },
    ];

    if (hasThirdPartyData) {
      barSeries.push({
        name: '第三方 Token',
        type: 'bar',
        stack: 'tokens',
        data: tpTokens,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#FFBE0B' },
            { offset: 1, color: '#FF8800' },
          ]),
          borderRadius: [4, 4, 0, 0],
        },
      });
    }

    chartInstance.current.setOption({
      backgroundColor: 'transparent',
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
        backgroundColor: isDark ? 'rgba(13, 29, 45, 0.95)' : 'rgba(255, 255, 255, 0.95)',
        borderColor: 'rgba(0, 217, 255, 0.2)',
        textStyle: { color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' },
      },
      legend: {
        data: legendData,
        textStyle: { color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.6)' },
      },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: {
          formatter: (val: string) => val.slice(5),
          color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
        },
        axisLine: { lineStyle: { color: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)' } },
      },
      yAxis: [
        {
          type: 'value',
          name: 'Token 用量',
          nameTextStyle: { color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' },
          axisLabel: {
            formatter: (val: number) =>
              val >= 1000000 ? `${(val / 1000000).toFixed(1)}M` :
              val >= 1000 ? `${(val / 1000).toFixed(0)}K` : String(val),
            color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
          },
          splitLine: { lineStyle: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' } },
        },
        {
          type: 'value',
          name: '请求次数',
          nameTextStyle: { color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' },
          axisLabel: { color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' },
          splitLine: { show: false },
        },
      ],
      series: [
        ...barSeries,
        {
          name: '请求次数',
          type: 'line',
          yAxisIndex: 1,
          data: requests,
          smooth: true,
          lineStyle: { color: '#00F5D4', width: 2 },
          itemStyle: { color: '#00F5D4' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(0, 245, 212, 0.25)' },
              { offset: 1, color: 'rgba(0, 245, 212, 0.02)' },
            ]),
          },
        },
      ],
    }, true);
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
        style={{ color: '#00D9FF' }}
      >
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="animate-fade-in-up">
      <UsageProgressCards />
      
      {/* 欢迎标题 - 新设计 */}
      <h2 
        style={{ 
          marginBottom: 24, 
          color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
          fontSize: 24,
          fontWeight: 600,
        }}
      >
        欢迎回来，
        <span style={{ 
          background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
        }}>
          {user?.display_name || user?.username}
        </span>
      </h2>

      {/* 统计卡片 — 新设计 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.05s' }}>
            <div className="flex items-center gap-4 h-full">
              <StatIcon
                icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                gradient="linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)"
              />
              <div>
                <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13, marginBottom: 4 }}>今日 Token 用量</div>
                <div style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 700 }}>
                  {formatNumber((overview?.today.total_tokens || 0) + (overview?.today.third_party_total_tokens || 0))}
                  <ArrowUpOutlined style={{ fontSize: 12, color: '#00F5D4', marginLeft: 4 }} />
                </div>
                {(overview?.today.third_party_total_tokens || 0) > 0 && (
                  <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 11, marginTop: 2 }}>
                    平台 {formatNumber(overview?.today.total_tokens || 0)} · 第三方 {formatNumber(overview?.today.third_party_total_tokens || 0)}
                  </div>
                )}
              </div>
            </div>
          </div>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.1s' }}>
            <div className="flex items-center gap-4 h-full">
              <StatIcon
                icon={<MessageOutlined style={{ fontSize: 22 }} />}
                gradient="linear-gradient(135deg, #9D4EDD 0%, #FF6B6B 100%)"
              />
              <div>
                <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13, marginBottom: 4 }}>今日请求数</div>
                <div style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 700 }}>
                  {((overview?.today.total_requests || 0) + (overview?.today.third_party_total_requests || 0)).toLocaleString()}
                </div>
                {(overview?.today.third_party_total_requests || 0) > 0 && (
                  <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 11, marginTop: 2 }}>
                    平台 {(overview?.today.total_requests || 0).toLocaleString()} · 第三方 {(overview?.today.third_party_total_requests || 0).toLocaleString()}
                  </div>
                )}
              </div>
            </div>
          </div>
        </Col>
        {isAdmin && (
          <>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.15s' }}>
                <div className="flex items-center gap-4 h-full">
                  <StatIcon
                    icon={<TeamOutlined style={{ fontSize: 22 }} />}
                    gradient="linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)"
                  />
                  <Statistic
                    title={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13 }}>总用户数</span>}
                    value={overview?.total_users || 0}
                    valueStyle={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 700 }}
                  />
                </div>
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.2s' }}>
                <div className="flex items-center gap-4 h-full">
                  <StatIcon
                    icon={<KeyOutlined style={{ fontSize: 22 }} />}
                    gradient="linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)"
                  />
                  <Statistic
                    title={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13 }}>活跃 API Key</span>}
                    value={overview?.total_api_keys || 0}
                    valueStyle={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 700 }}
                  />
                </div>
              </div>
            </Col>
          </>
        )}
        {!isAdmin && (
          <>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.15s' }}>
                <div className="flex items-center gap-4 h-full">
                  <StatIcon
                    icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                    gradient="linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)"
                  />
                  <div>
                    <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13, marginBottom: 4 }}>本月 Token 用量</div>
                    <div style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 700 }}>
                      {formatNumber((overview?.this_month.total_tokens || 0) + (overview?.this_month.third_party_total_tokens || 0))}
                    </div>
                    {(overview?.this_month.third_party_total_tokens || 0) > 0 && (
                      <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 11, marginTop: 2 }}>
                        平台 {formatNumber(overview?.this_month.total_tokens || 0)} · 第三方 {formatNumber(overview?.this_month.third_party_total_tokens || 0)}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.2s' }}>
                <div className="flex items-center gap-4 h-full">
                  <StatIcon
                    icon={<MessageOutlined style={{ fontSize: 22 }} />}
                    gradient="linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)"
                  />
                  <div>
                    <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 13, marginBottom: 4 }}>本月请求数</div>
                    <div style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 700 }}>
                      {((overview?.this_month.total_requests || 0) + (overview?.this_month.third_party_total_requests || 0)).toLocaleString()}
                    </div>
                    {(overview?.this_month.third_party_total_requests || 0) > 0 && (
                      <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 11, marginTop: 2 }}>
                        平台 {(overview?.this_month.total_requests || 0).toLocaleString()} · 第三方 {(overview?.this_month.third_party_total_requests || 0).toLocaleString()}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </Col>
          </>
        )}
      </Row>

      {/* 图表区域 — 玻璃态卡片 - 新设计 */}
      <div
        className="glass-card animate-fade-in-up mt-6 p-6"
        style={{ marginTop: 24, animationDelay: '0.25s' }}
      >
        <h3 style={{ 
          marginBottom: 20, 
          color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
          fontSize: 18,
          fontWeight: 600,
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}>
          <span style={{
            width: 4,
            height: 20,
            background: 'linear-gradient(180deg, #00D9FF 0%, #9D4EDD 100%)',
            borderRadius: 2,
          }} />
          近 30 天用量趋势
        </h3>
        {usageData.length > 0 ? (
          <div
            ref={chartRef}
            style={{
              height: 360,
              width: '100%',
              background: 'transparent',
              borderRadius: 12,
            }}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' }}>暂无用量数据</span>}
            style={{ padding: 40 }}
          />
        )}
      </div>

      {/* 公告区域 — 玻璃态卡片 - 新设计 */}
      {announcements.length > 0 && (
        <div
          className="glass-card animate-fade-in-up mt-6 p-6"
          style={{ marginTop: 24, animationDelay: '0.3s' }}
        >
          <h3 style={{ 
            marginBottom: 20, 
            color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
            fontSize: 18,
            fontWeight: 600,
            display: 'flex',
            alignItems: 'center',
            gap: 8,
          }}>
            <span style={{
              width: 4,
              height: 20,
              background: 'linear-gradient(180deg, #FFBE0B 0%, #FF6B6B 100%)',
              borderRadius: 2,
            }} />
            系统公告
          </h3>
          <div>
            {announcements.slice(0, 5).map((ann, index) => (
              <div
                key={ann.id}
                style={{
                  paddingBottom: 16,
                  marginBottom: index < 4 ? 16 : 0,
                  borderBottom: index < 4 ? (isDark ? '1px solid rgba(255, 255, 255, 0.06)' : '1px solid rgba(0, 0, 0, 0.06)') : 'none',
                }}
              >
                <div className="flex items-center gap-2 flex-wrap">
                  {ann.pinned && (
                    <Tag color="error" style={{ borderRadius: 4, border: 'none' }}>置顶</Tag>
                  )}
                  <strong style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 15 }}>{ann.title}</strong>
                  <span
                    style={{
                      color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)',
                      fontSize: 12,
                      marginLeft: 'auto',
                    }}
                  >
                    {ann.created_at?.slice(0, 10)}
                  </span>
                </div>
                <div
                  style={{
                    color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)',
                    fontSize: 14,
                    marginTop: 8,
                    lineHeight: 1.6,
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
