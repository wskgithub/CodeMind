import {
  BarChartOutlined,
  ThunderboltOutlined,
  MessageOutlined,
  DownloadOutlined,
  RocketOutlined,
} from '@ant-design/icons';
import { Select, DatePicker, Space, Table, Spin, Row, Col, Segmented, Button } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import * as echarts from 'echarts';
import { useEffect, useState, useRef, useMemo } from 'react';

import UsageProgressCards from '@/components/common/UsageProgressCards';
import { getUsageStats, getRanking, exportUsageCSV, getKeyUsageStats } from '@/services/statsService';
import useAppStore from '@/store/appStore';
import useAuthStore from '@/store/authStore';
import type { UsageItem, RankingItem, KeyUsageItem } from '@/types';

const { RangePicker } = DatePicker;

/** 图标包裹层 — 渐变圆形背景 - 新设计 */
const StatIcon = ({ icon, gradient }: { icon: React.ReactNode; gradient: string }) => (
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

/** 页面标题图标 — 渐变圆形背景 - 新设计 */
const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #9D4EDD 0%, #FF6B6B 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(157, 78, 221, 0.25)',
    }}
  >
    {icon}
  </span>
);

/** 用量统计页面 — 与首页/登录页新设计风格统一 */
const UsagePage = () => {
  const { user } = useAuthStore();
  const { themeMode } = useAppStore();
  const isDark = themeMode === 'dark';
  const isAdmin = user?.role === 'super_admin' || user?.role === 'dept_manager';

  const [period, setPeriod] = useState<string>('daily');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [usageData, setUsageData] = useState<UsageItem[]>([]);
  const [keyUsageData, setKeyUsageData] = useState<KeyUsageItem[]>([]);
  const [ranking, setRanking] = useState<RankingItem[]>([]);
  const [rankType, setRankType] = useState<'user' | 'department'>('user');
  const [loading, setLoading] = useState(false);
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);
  const loadTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    loadUsageData();
    if (isAdmin) loadRanking();
    return () => { chartInstance.current?.dispose(); };
  }, []);

  useEffect(() => {
    if (loadTimeoutRef.current) {
      clearTimeout(loadTimeoutRef.current);
    }
    loadTimeoutRef.current = setTimeout(() => {
      loadUsageData();
    }, 300);
    return () => {
      if (loadTimeoutRef.current) {
        clearTimeout(loadTimeoutRef.current);
      }
    };
  }, [period, dateRange]);

  useEffect(() => {
    if (isAdmin) loadRanking();
  }, [rankType]);

  useEffect(() => {
    if (usageData.length > 0 && chartRef.current) renderChart();
  }, [usageData]);

  // 主题切换时重新渲染图表
  useEffect(() => {
    if (usageData.length > 0 && chartInstance.current) renderChart();
  }, [themeMode]);

  useEffect(() => {
    const handleResize = () => chartInstance.current?.resize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const loadUsageData = async () => {
    setLoading(true);
    try {
      const params: {
        period: string;
        start_date?: string;
        end_date?: string;
      } = { period };
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const [usageRes, keyUsageRes] = await Promise.all([
        getUsageStats(params),
        getKeyUsageStats({
          start_date: dateRange ? dateRange[0].format('YYYY-MM-DD') : undefined,
          end_date: dateRange ? dateRange[1].format('YYYY-MM-DD') : undefined,
        }),
      ]);
      setUsageData(usageRes.data.data?.items || []);
      setKeyUsageData(keyUsageRes.data.data || []);
    } catch {
      // 拦截器处理
    } finally {
      setLoading(false);
    }
  };

  const loadRanking = async () => {
    try {
      const res = await getRanking({ type: rankType, period: 'monthly', limit: 10 });
      setRanking(res.data.data || []);
    } catch {
      // 拦截器处理
    }
  };

  const hasThirdPartyData = usageData.some((d) => (d.third_party_total_tokens || 0) > 0);

  const renderChart = () => {
    if (!chartRef.current) return;
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const dates = usageData.map((d) => d.date);
    const promptTokens = usageData.map((d) => d.prompt_tokens);
    const completionTokens = usageData.map((d) => d.completion_tokens);
    const cacheReadTokens = usageData.map((d) => d.cache_read_input_tokens || 0);
    const tpTokens = usageData.map((d) => d.third_party_total_tokens || 0);
    const requests = usageData.map((d) => (d.request_count || 0) + (d.third_party_request_count || 0));

    // 是否有缓存数据
    const hasCacheData = cacheReadTokens.some((v) => v > 0);

    const legendData = hasThirdPartyData
      ? ['Prompt Tokens', '缓存命中', 'Completion Tokens', '第三方 Tokens', '请求次数']
      : ['Prompt Tokens', '缓存命中', 'Completion Tokens', '请求次数'];

    const barSeries: echarts.SeriesOption[] = [
      {
        name: 'Prompt Tokens',
        type: 'bar',
        stack: 'tokens',
        data: promptTokens,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#00D9FF' },
            { offset: 1, color: '#00A8CC' },
          ]),
          borderRadius: [0, 0, 0, 0],
        },
      },
    ];

    // 添加缓存命中系列（如果有缓存数据）
    if (hasCacheData) {
      barSeries.push({
        name: '缓存命中',
        type: 'bar',
        stack: 'tokens',
        data: cacheReadTokens,
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#00F5D4' },
            { offset: 1, color: '#00CCB3' },
          ]),
          borderRadius: [0, 0, 0, 0],
        },
      });
    }

    barSeries.push({
      name: 'Completion Tokens',
      type: 'bar',
      stack: 'tokens',
      data: completionTokens,
      itemStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: '#9D4EDD' },
          { offset: 1, color: '#7B2CBF' },
        ]),
        borderRadius: hasThirdPartyData ? [0, 0, 0, 0] : [4, 4, 0, 0],
      },
    });

    if (hasThirdPartyData) {
      barSeries.push({
        name: '第三方 Tokens',
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
        backgroundColor: isDark ? 'rgba(13, 29, 45, 0.95)' : 'rgba(255, 255, 255, 0.95)',
        borderColor: isDark ? 'rgba(0, 217, 255, 0.2)' : 'rgba(0, 0, 0, 0.1)',
        textStyle: { color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' },
      },
      legend: {
        data: legendData,
        textStyle: { color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)' },
      },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: {
          formatter: (v: string) => v.slice(5),
          color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
        },
        axisLine: { lineStyle: { color: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)' } },
      },
      yAxis: [
        {
          type: 'value',
          name: 'Tokens',
          nameTextStyle: { color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' },
          axisLabel: { color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)' },
          splitLine: { lineStyle: { color: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' } },
        },
        {
          type: 'value',
          name: '请求数',
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
        },
      ],
    }, true);
  };

  // 汇总统计（平台 + 第三方）
  const platformTokens = useMemo(() => usageData.reduce((sum, d) => sum + (d.total_tokens || 0), 0), [usageData]);
  const thirdPartyTokens = useMemo(() => usageData.reduce((sum, d) => sum + (d.third_party_total_tokens || 0), 0), [usageData]);
  const totalTokens = platformTokens + thirdPartyTokens;
  
  const platformRequests = useMemo(() => usageData.reduce((sum, d) => sum + (d.request_count || 0), 0), [usageData]);
  const thirdPartyRequests = useMemo(() => usageData.reduce((sum, d) => sum + (d.third_party_request_count || 0), 0), [usageData]);
  const totalRequests = platformRequests + thirdPartyRequests;
  
  const platformPrompt = useMemo(() => usageData.reduce((sum, d) => sum + (d.prompt_tokens || 0), 0), [usageData]);
  const thirdPartyPrompt = useMemo(() => usageData.reduce((sum, d) => sum + (d.third_party_prompt_tokens || 0), 0), [usageData]);
  const totalPrompt = platformPrompt + thirdPartyPrompt;
  
  const platformCompletion = useMemo(() => usageData.reduce((sum, d) => sum + (d.completion_tokens || 0), 0), [usageData]);
  const thirdPartyCompletion = useMemo(() => usageData.reduce((sum, d) => sum + (d.third_party_completion_tokens || 0), 0), [usageData]);
  const totalCompletion = platformCompletion + thirdPartyCompletion;
  
  // 是否有第三方数据
  const hasThirdPartyStats = thirdPartyTokens > 0 || thirdPartyRequests > 0 || thirdPartyPrompt > 0 || thirdPartyCompletion > 0;

  // 缓存相关统计
  const platformCacheRead = useMemo(() => usageData.reduce((sum, d) => sum + (d.cache_read_input_tokens || 0), 0), [usageData]);
  const thirdPartyCacheRead = useMemo(() => usageData.reduce((sum, d) => sum + (d.third_party_cache_read_input_tokens || 0), 0), [usageData]);
  const totalCacheRead = platformCacheRead + thirdPartyCacheRead;
  // 缓存命中率 = 缓存命中 Token / Prompt Tokens
  const cacheHitRate = totalPrompt > 0 ? ((totalCacheRead / totalPrompt) * 100).toFixed(1) : '0';

  const columns: ColumnsType<UsageItem> = useMemo(() => {
    const cols: ColumnsType<UsageItem> = [
      {
        title: '日期',
        dataIndex: 'date',
        key: 'date',
        width: 120,
        render: (text) => <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)' }}>{text}</span>,
      },
      {
        title: 'Prompt Tokens',
        dataIndex: 'prompt_tokens',
        key: 'prompt_tokens',
        align: 'right',
        render: (v: number) => <span style={{ color: '#00D9FF' }}>{v.toLocaleString()}</span>,
      },
      {
        title: 'Completion Tokens',
        dataIndex: 'completion_tokens',
        key: 'completion_tokens',
        align: 'right',
        render: (v: number) => <span style={{ color: '#9D4EDD' }}>{v.toLocaleString()}</span>,
      },
      {
        title: '平台 Tokens',
        dataIndex: 'total_tokens',
        key: 'total_tokens',
        align: 'right',
        render: (v: number) => <strong style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{v.toLocaleString()}</strong>,
      },
    ];

    if (hasThirdPartyData) {
      cols.push({
        title: '第三方 Tokens',
        dataIndex: 'third_party_total_tokens',
        key: 'tp_tokens',
        align: 'right',
        render: (v: number) => <span style={{ color: '#FFBE0B' }}>{(v || 0).toLocaleString()}</span>,
      });
    }

    cols.push({
      title: '请求次数',
      key: 'request_count',
      align: 'right',
      render: (_, record) => {
        const total = (record.request_count || 0) + (record.third_party_request_count || 0);
        return <span style={{ color: '#00F5D4' }}>{total.toLocaleString()}</span>;
      },
    });

    return cols;
  }, [isDark, hasThirdPartyData]);

  const rankColumns: ColumnsType<RankingItem> = useMemo(() => [
    { 
      title: '排名', 
      dataIndex: 'rank', 
      key: 'rank', 
      width: 60, 
      align: 'center',
      render: (v) => (
        <span style={{ 
          color: v <= 3 ? '#FFBE0B' : (isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.5)'),
          fontWeight: v <= 3 ? 700 : 400,
        }}>
          {v}
        </span>
      ),
    },
    { 
      title: '名称', 
      dataIndex: 'name', 
      key: 'name',
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{text}</span>,
    },
    {
      title: '总 Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00D9FF' }}>{v.toLocaleString()}</span>,
    },
    {
      title: '请求次数',
      dataIndex: 'request_count',
      key: 'request_count',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00F5D4' }}>{v.toLocaleString()}</span>,
    },
  ], [isDark]);

  const keyUsageColumns: ColumnsType<KeyUsageItem> = useMemo(() => [
    {
      title: 'Key 名称',
      dataIndex: 'name',
      key: 'name',
      render: (text) => <span style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{text}</span>,
    },
    {
      title: 'Prompt Tokens',
      dataIndex: 'prompt_tokens',
      key: 'prompt_tokens',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00D9FF' }}>{v.toLocaleString()}</span>,
    },
    {
      title: 'Completion Tokens',
      dataIndex: 'completion_tokens',
      key: 'completion_tokens',
      align: 'right',
      render: (v: number) => <span style={{ color: '#9D4EDD' }}>{v.toLocaleString()}</span>,
    },
    {
      title: '总 Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      align: 'right',
      render: (v: number) => <strong style={{ color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)' }}>{v.toLocaleString()}</strong>,
    },
    {
      title: '请求次数',
      dataIndex: 'request_count',
      key: 'request_count',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00F5D4' }}>{v.toLocaleString()}</span>,
    },
  ], [isDark]);

  const formatNum = (n: number) => {
    if (n >= 1000000) return `${(n / 1000000).toFixed(2)}M`;
    if (n >= 1000) return `${(n / 1000).toFixed(1)}K`;
    return n.toString();
  };

  const handleExportCSV = async () => {
    try {
      const params: {
        period: string;
        start_date?: string;
        end_date?: string;
      } = { period };
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await exportUsageCSV(params);
      
      const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      
      const contentDisposition = response.headers['content-disposition'];
      let filename = 'usage_report.csv';
      if (contentDisposition) {
        const match = contentDisposition.match(/filename=([^;]+)/);
        if (match) {
          filename = decodeURIComponent(match[1].trim());
        }
      }
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);
    } catch {
      // 拦截器处理错误
    }
  };

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        <UsageProgressCards />
        
        {/* 页面标题 — 带渐变图标 - 新设计 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<BarChartOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                用量统计
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                查看您的 API 使用情况和资源消耗统计
              </p>
            </div>
          </div>
        </div>

        {/* 统计汇总卡片 — 2 行 3 列布局 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 12 }}>
          {/* 第一行: 总 Tokens、总请求数、第三方 Tokens */}
          <Col xs={24} sm={8}>
            <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.05s' }}>
              <div className="flex items-center gap-4 h-full">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', marginBottom: 4 }}>总 Tokens</div>
                  <div style={{ fontWeight: 700, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 20 }}>{formatNum(totalTokens)}</div>
                  {hasThirdPartyStats && (
                    <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                      平台 {formatNum(platformTokens)} · 第三方 {formatNum(thirdPartyTokens)}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.1s' }}>
              <div className="flex items-center gap-4 h-full">
                <StatIcon
                  icon={<MessageOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #9D4EDD 0%, #FF6B6B 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', marginBottom: 4 }}>总请求数</div>
                  <div style={{ fontWeight: 700, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 20 }}>{formatNum(totalRequests)}</div>
                  {hasThirdPartyStats && (
                    <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                      平台 {formatNum(platformRequests)} · 第三方 {formatNum(thirdPartyRequests)}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.15s' }}>
              <div className="flex items-center gap-4 h-full">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #FFBE0B 0%, #FF8800 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', marginBottom: 4 }}>第三方 Tokens</div>
                  <div style={{ fontWeight: 700, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 20 }}>{formatNum(thirdPartyTokens)}</div>
                  <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                    {hasThirdPartyStats ? `${thirdPartyRequests} 次请求` : '暂无第三方请求'}
                  </div>
                </div>
              </div>
            </div>
          </Col>
        </Row>

        {/* 第二行: Prompt Tokens、Completion Tokens、缓存命中 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={8}>
            <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.2s' }}>
              <div className="flex items-center gap-4 h-full">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', marginBottom: 4 }}>Prompt Tokens</div>
                  <div style={{ fontWeight: 700, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 20 }}>{formatNum(totalPrompt)}</div>
                  {hasThirdPartyStats && (
                    <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                      平台 {formatNum(platformPrompt)} · 第三方 {formatNum(thirdPartyPrompt)}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.25s' }}>
              <div className="flex items-center gap-4 h-full">
                <StatIcon
                  icon={<MessageOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #9D4EDD 0%, #7B2CBF 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', marginBottom: 4 }}>Completion Tokens</div>
                  <div style={{ fontWeight: 700, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 20 }}>{formatNum(totalCompletion)}</div>
                  {hasThirdPartyStats && (
                    <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                      平台 {formatNum(platformCompletion)} · 第三方 {formatNum(thirdPartyCompletion)}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={8}>
            <div className="stat-card animate-fade-in-up h-full" style={{ animationDelay: '0.3s' }}>
              <div className="flex items-center gap-4 h-full">
                <StatIcon
                  icon={<RocketOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #00F5D4 0%, #00CCB3 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', marginBottom: 4 }}>缓存命中</div>
                  <div style={{ fontWeight: 700, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 20 }}>{formatNum(totalCacheRead)}</div>
                  {hasThirdPartyStats ? (
                    <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                      平台 {formatNum(platformCacheRead)} · 第三方 {formatNum(thirdPartyCacheRead)}
                    </div>
                  ) : (
                    <div style={{ fontSize: 11, color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', marginTop: 2 }}>
                      命中率 {cacheHitRate}%
                    </div>
                  )}
                </div>
              </div>
            </div>
          </Col>
        </Row>

        {/* 筛选栏 — 玻璃态卡片 - 新设计 */}
        <div className="glass-card p-5 animate-fade-in-up" style={{ marginBottom: 24, animationDelay: '0.08s' }}>
          <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space wrap size={16}>
              <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)' }}>统计周期：</span>
              <Select
                value={period}
                onChange={setPeriod}
                options={[
                  { label: '每日', value: 'daily' },
                  { label: '每周', value: 'weekly' },
                  { label: '每月', value: 'monthly' },
                ]}
                style={{ 
                  width: 120,
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.02)',
                }}
              />
              <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)' }}>日期范围：</span>
              <RangePicker
                value={dateRange}
                onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
              />
            </Space>
            {isAdmin && (
              <Button
                type="primary"
                icon={<DownloadOutlined />}
                onClick={handleExportCSV}
                style={{
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  border: 'none',
                  boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                  borderRadius: 12,
                  height: 40,
                }}
              >
                导出 CSV
              </Button>
            )}
          </Space>
        </div>

        {/* Key 用量分布 — 玻璃态卡片 - 新设计 */}
        <div
          className="glass-card animate-fade-in-up p-6"
          style={{ marginBottom: 24, animationDelay: '0.09s' }}
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
            Key 用量分布
          </h3>
          <Table
            dataSource={keyUsageData}
            columns={keyUsageColumns}
            rowKey="id"
            pagination={false}
            size="small"
            scroll={{ y: 240 }}
          />
          {keyUsageData.length === 0 && !loading && (
            <div style={{ textAlign: 'center', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', paddingTop: 40 }}>
              暂无数据
            </div>
          )}
        </div>

        {/* 图表 — 玻璃态容器 - 新设计 */}
        <Spin spinning={loading}>
          <div
            className="glass-card animate-fade-in-up p-6"
            style={{ marginBottom: 24, animationDelay: '0.1s' }}
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
              用量趋势
            </h3>
            <div ref={chartRef} style={{ height: 400, width: '100%' }} />
            {usageData.length === 0 && !loading && (
              <div style={{ textAlign: 'center', color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', paddingTop: 40 }}>
                暂无数据
              </div>
            )}
          </div>
        </Spin>

        <Row gutter={24} style={{ marginTop: 24 }}>
          {/* 明细表格 — 玻璃态卡片 - 新设计 */}
          <Col xs={24} lg={isAdmin ? 14 : 24}>
            <div
              className="glass-card animate-fade-in-up p-6"
              style={{ animationDelay: '0.15s' }}
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
                  background: 'linear-gradient(180deg, #00F5D4 0%, #00D9FF 100%)',
                  borderRadius: 2,
                }} />
                用量明细
              </h3>
              <Table
                dataSource={usageData}
                columns={columns}
                rowKey="date"
                pagination={false}
                size="small"
                scroll={{ y: 360 }}
              />
            </div>
          </Col>

          {/* 排行榜（管理员可见）— 玻璃态卡片 - 新设计 */}
          {isAdmin && (
            <Col xs={24} lg={10}>
              <div
                className="glass-card animate-fade-in-up p-6"
                style={{ animationDelay: '0.2s' }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                  <h3 style={{ 
                    margin: 0, 
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
                    用量排行
                  </h3>
                  <Segmented
                    value={rankType}
                    onChange={(v) => setRankType(v as 'user' | 'department')}
                    options={[
                      { label: '用户', value: 'user' },
                      { label: '部门', value: 'department' },
                    ]}
                    style={{
                      background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.04)',
                    }}
                  />
                </div>
                <Table
                  dataSource={ranking}
                  columns={rankColumns}
                  rowKey="id"
                  pagination={false}
                  size="small"
                />
              </div>
            </Col>
          )}
        </Row>
      </div>
    </div>
  );
};

export default UsagePage;
