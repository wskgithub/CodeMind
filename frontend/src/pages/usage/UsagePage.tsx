import { useEffect, useState, useRef } from 'react';
import { Select, DatePicker, Space, Table, Spin, Row, Col, Segmented, Button } from 'antd';
import {
  BarChartOutlined,
  ThunderboltOutlined,
  MessageOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import * as echarts from 'echarts';
import dayjs from 'dayjs';
import { getUsageStats, getRanking, exportUsageCSV } from '@/services/statsService';
import useAuthStore from '@/store/authStore';
import type { UsageItem, RankingItem } from '@/types';
import UsageProgressCards from '@/components/common/UsageProgressCards';

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
  const isAdmin = user?.role === 'super_admin' || user?.role === 'dept_manager';

  const [period, setPeriod] = useState<string>('daily');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [usageData, setUsageData] = useState<UsageItem[]>([]);
  const [ranking, setRanking] = useState<RankingItem[]>([]);
  const [rankType, setRankType] = useState<'user' | 'department'>('user');
  const [loading, setLoading] = useState(false);
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);

  useEffect(() => {
    loadUsageData();
    if (isAdmin) loadRanking();
    return () => { chartInstance.current?.dispose(); };
  }, []);

  useEffect(() => {
    loadUsageData();
  }, [period, dateRange]);

  useEffect(() => {
    if (isAdmin) loadRanking();
  }, [rankType]);

  useEffect(() => {
    if (usageData.length > 0 && chartRef.current) renderChart();
  }, [usageData]);

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
      const res = await getUsageStats(params);
      setUsageData(res.data.data?.items || []);
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

  const renderChart = () => {
    if (!chartRef.current) return;
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current);
    }

    const dates = usageData.map((d) => d.date);
    const promptTokens = usageData.map((d) => d.prompt_tokens);
    const completionTokens = usageData.map((d) => d.completion_tokens);
    const requests = usageData.map((d) => d.request_count);

    chartInstance.current.setOption({
      backgroundColor: 'transparent',
      tooltip: { 
        trigger: 'axis',
        backgroundColor: 'rgba(13, 29, 45, 0.95)',
        borderColor: 'rgba(0, 217, 255, 0.2)',
        textStyle: { color: '#fff' },
      },
      legend: { 
        data: ['Prompt Tokens', 'Completion Tokens', '请求次数'],
        textStyle: { color: 'rgba(255, 255, 255, 0.7)' },
      },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: { 
          formatter: (v: string) => v.slice(5),
          color: 'rgba(255, 255, 255, 0.5)',
        },
        axisLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.1)' } },
      },
      yAxis: [
        { 
          type: 'value', 
          name: 'Tokens',
          nameTextStyle: { color: 'rgba(255, 255, 255, 0.5)' },
          axisLabel: { color: 'rgba(255, 255, 255, 0.5)' },
          splitLine: { lineStyle: { color: 'rgba(255, 255, 255, 0.05)' } },
        },
        { 
          type: 'value', 
          name: '请求数',
          nameTextStyle: { color: 'rgba(255, 255, 255, 0.5)' },
          axisLabel: { color: 'rgba(255, 255, 255, 0.5)' },
          splitLine: { show: false },
        },
      ],
      series: [
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
            borderRadius: [0, 0, 0, 0] 
          },
        },
        {
          name: 'Completion Tokens',
          type: 'bar',
          stack: 'tokens',
          data: completionTokens,
          itemStyle: { 
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: '#9D4EDD' },
              { offset: 1, color: '#7B2CBF' },
            ]),
            borderRadius: [4, 4, 0, 0] 
          },
        },
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
    });
  };

  // 汇总统计
  const totalTokens = usageData.reduce((sum, d) => sum + (d.total_tokens || 0), 0);
  const totalRequests = usageData.reduce((sum, d) => sum + (d.request_count || 0), 0);
  const totalPrompt = usageData.reduce((sum, d) => sum + (d.prompt_tokens || 0), 0);
  const totalCompletion = usageData.reduce((sum, d) => sum + (d.completion_tokens || 0), 0);

  const columns: ColumnsType<UsageItem> = [
    { 
      title: '日期', 
      dataIndex: 'date', 
      key: 'date', 
      width: 120,
      render: (text) => <span style={{ color: 'rgba(255, 255, 255, 0.8)' }}>{text}</span>,
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
      render: (v: number) => <strong style={{ color: '#fff' }}>{v.toLocaleString()}</strong>,
    },
    {
      title: '请求次数',
      dataIndex: 'request_count',
      key: 'request_count',
      align: 'right',
      render: (v: number) => <span style={{ color: '#00F5D4' }}>{v.toLocaleString()}</span>,
    },
  ];

  const rankColumns: ColumnsType<RankingItem> = [
    { 
      title: '排名', 
      dataIndex: 'rank', 
      key: 'rank', 
      width: 60, 
      align: 'center',
      render: (v) => (
        <span style={{ 
          color: v <= 3 ? '#FFBE0B' : 'rgba(255, 255, 255, 0.6)',
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
      render: (text) => <span style={{ color: '#fff' }}>{text}</span>,
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
  ];

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
              <h2 style={{ margin: 0, color: '#fff', fontSize: 24, fontWeight: 600 }}>
                用量统计
              </h2>
              <p style={{ margin: 0, color: 'rgba(255, 255, 255, 0.5)', fontSize: 14, marginTop: 4 }}>
                查看您的 API 使用情况和资源消耗统计
              </p>
            </div>
          </div>
        </div>

        {/* 统计汇总卡片 — stat-card 带渐变图标 - 新设计 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.05s' }}>
              <div className="flex items-center gap-4">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #00D9FF 0%, #00F5D4 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.5)' }}>总 Tokens</div>
                  <div style={{ fontWeight: 700, color: '#fff', fontSize: 20 }}>{formatNum(totalTokens)}</div>
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
              <div className="flex items-center gap-4">
                <StatIcon
                  icon={<MessageOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #9D4EDD 0%, #FF6B6B 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.5)' }}>总请求数</div>
                  <div style={{ fontWeight: 700, color: '#fff', fontSize: 20 }}>{totalRequests.toLocaleString()}</div>
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.15s' }}>
              <div className="flex items-center gap-4">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #00F5D4 0%, #00D9FF 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.5)' }}>Prompt Tokens</div>
                  <div style={{ fontWeight: 700, color: '#fff', fontSize: 20 }}>{formatNum(totalPrompt)}</div>
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
              <div className="flex items-center gap-4">
                <StatIcon
                  icon={<MessageOutlined style={{ fontSize: 22 }} />}
                  gradient="linear-gradient(135deg, #FFBE0B 0%, #FF6B6B 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: 'rgba(255, 255, 255, 0.5)' }}>Completion Tokens</div>
                  <div style={{ fontWeight: 700, color: '#fff', fontSize: 20 }}>{formatNum(totalCompletion)}</div>
                </div>
              </div>
            </div>
          </Col>
        </Row>

        {/* 筛选栏 — 玻璃态卡片 - 新设计 */}
        <div className="glass-card p-5 animate-fade-in-up" style={{ marginBottom: 24, animationDelay: '0.08s' }}>
          <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space wrap size={16}>
              <span style={{ color: 'rgba(255, 255, 255, 0.6)' }}>统计周期：</span>
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
                  background: 'rgba(255, 255, 255, 0.03)',
                }}
              />
              <span style={{ color: 'rgba(255, 255, 255, 0.6)' }}>日期范围：</span>
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

        {/* 图表 — 玻璃态容器 - 新设计 */}
        <Spin spinning={loading}>
          <div
            className="glass-card animate-fade-in-up p-6"
            style={{ marginBottom: 24, animationDelay: '0.1s' }}
          >
            <h3 style={{ 
              marginBottom: 20, 
              color: '#fff',
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
              <div style={{ textAlign: 'center', color: 'rgba(255, 255, 255, 0.5)', paddingTop: 40 }}>
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
                color: '#fff',
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
                    color: '#fff',
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
                      background: 'rgba(255, 255, 255, 0.05)',
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
