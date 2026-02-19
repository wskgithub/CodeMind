import { useEffect, useState, useRef } from 'react';
import { Select, DatePicker, Space, Table, Spin, Row, Col, Segmented, theme, Button } from 'antd';
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

const { RangePicker } = DatePicker;

/** 图标包裹层 — 渐变圆形背景 */
const StatIcon = ({ icon, gradient }: { icon: React.ReactNode; gradient: string }) => (
  <span
    className="flex items-center justify-center w-10 h-10 rounded-full shrink-0"
    style={{ background: gradient, color: '#fff' }}
  >
    {icon}
  </span>
);

/** 用量统计页面 — Glassmorphism 风格，可视化展示 Token 使用情况 */
const UsagePage = () => {
  const { token } = theme.useToken();
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
      tooltip: { trigger: 'axis' },
      legend: { data: ['Prompt Tokens', 'Completion Tokens', '请求次数'] },
      grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: { formatter: (v: string) => v.slice(5) },
      },
      yAxis: [
        { type: 'value', name: 'Tokens' },
        { type: 'value', name: '请求数' },
      ],
      series: [
        {
          name: 'Prompt Tokens',
          type: 'bar',
          stack: 'tokens',
          data: promptTokens,
          itemStyle: { color: '#1677ff', borderRadius: [0, 0, 0, 0] },
        },
        {
          name: 'Completion Tokens',
          type: 'bar',
          stack: 'tokens',
          data: completionTokens,
          itemStyle: { color: '#69b1ff', borderRadius: [4, 4, 0, 0] },
        },
        {
          name: '请求次数',
          type: 'line',
          yAxisIndex: 1,
          data: requests,
          smooth: true,
          lineStyle: { color: '#ff7a45' },
          itemStyle: { color: '#ff7a45' },
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
    { title: '日期', dataIndex: 'date', key: 'date', width: 120 },
    {
      title: 'Prompt Tokens',
      dataIndex: 'prompt_tokens',
      key: 'prompt_tokens',
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: 'Completion Tokens',
      dataIndex: 'completion_tokens',
      key: 'completion_tokens',
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: '总 Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      align: 'right',
      render: (v: number) => <strong>{v.toLocaleString()}</strong>,
    },
    {
      title: '请求次数',
      dataIndex: 'request_count',
      key: 'request_count',
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
  ];

  const rankColumns: ColumnsType<RankingItem> = [
    { title: '排名', dataIndex: 'rank', key: 'rank', width: 60, align: 'center' },
    { title: '名称', dataIndex: 'name', key: 'name' },
    {
      title: '总 Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: '请求次数',
      dataIndex: 'request_count',
      key: 'request_count',
      align: 'right',
      render: (v: number) => v.toLocaleString(),
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
      
      // 创建下载链接
      const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      
      // 从响应头获取文件名，或使用默认文件名
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
        {/* 页面标题 — 带渐变图标 */}
        <h2
          style={{
            marginBottom: 24,
            color: token.colorTextHeading,
            display: 'flex',
            alignItems: 'center',
            gap: 12,
          }}
        >
          <span
            className="flex items-center justify-center w-10 h-10 rounded-xl shrink-0"
            style={{ background: 'var(--gradient-primary)', color: '#fff' }}
          >
            <BarChartOutlined style={{ fontSize: 20 }} />
          </span>
          用量统计
        </h2>

        {/* 统计汇总卡片 — stat-card 带渐变图标 */}
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.05s' }}>
              <div className="flex items-center gap-3">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 20 }} />}
                  gradient="linear-gradient(135deg, #2B7CB3 0%, #4BA3D4 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: token.colorTextSecondary }}>总 Tokens</div>
                  <div style={{ fontWeight: 600, color: token.colorTextHeading }}>{formatNum(totalTokens)}</div>
                </div>
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
                <div>
                  <div style={{ fontSize: 12, color: token.colorTextSecondary }}>总请求数</div>
                  <div style={{ fontWeight: 600, color: token.colorTextHeading }}>{totalRequests.toLocaleString()}</div>
                </div>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <div className="stat-card animate-fade-in-up" style={{ animationDelay: '0.15s' }}>
              <div className="flex items-center gap-3">
                <StatIcon
                  icon={<ThunderboltOutlined style={{ fontSize: 20 }} />}
                  gradient="linear-gradient(135deg, #13c2c2 0%, #36cfc9 100%)"
                />
                <div>
                  <div style={{ fontSize: 12, color: token.colorTextSecondary }}>Prompt Tokens</div>
                  <div style={{ fontWeight: 600, color: token.colorTextHeading }}>{formatNum(totalPrompt)}</div>
                </div>
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
                <div>
                  <div style={{ fontSize: 12, color: token.colorTextSecondary }}>Completion Tokens</div>
                  <div style={{ fontWeight: 600, color: token.colorTextHeading }}>{formatNum(totalCompletion)}</div>
                </div>
              </div>
            </div>
          </Col>
        </Row>

        {/* 筛选栏 — 玻璃态卡片 */}
        <div className="glass-card p-4 animate-fade-in-up" style={{ marginBottom: 16, animationDelay: '0.08s' }}>
          <Space wrap style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space wrap>
              <span style={{ color: token.colorTextSecondary }}>统计周期：</span>
              <Select
                value={period}
                onChange={setPeriod}
                options={[
                  { label: '每日', value: 'daily' },
                  { label: '每周', value: 'weekly' },
                  { label: '每月', value: 'monthly' },
                ]}
                style={{ width: 120 }}
              />
              <span style={{ color: token.colorTextSecondary }}>日期范围：</span>
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
                style={{ background: 'var(--gradient-primary)' }}
              >
                导出 CSV
              </Button>
            )}
          </Space>
        </div>

        {/* 图表 — 玻璃态容器 */}
        <Spin spinning={loading}>
          <div
            className="glass-card animate-fade-in-up p-6"
            style={{ marginBottom: 16, animationDelay: '0.1s' }}
          >
            <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>用量趋势</h3>
            <div ref={chartRef} style={{ height: 400, width: '100%' }} />
            {usageData.length === 0 && !loading && (
              <div style={{ textAlign: 'center', color: token.colorTextTertiary, paddingTop: 40 }}>
                暂无数据
              </div>
            )}
          </div>
        </Spin>

        <Row gutter={16} style={{ marginTop: 16 }}>
          {/* 明细表格 — 玻璃态卡片 */}
          <Col xs={24} lg={isAdmin ? 14 : 24}>
            <div
              className="glass-card animate-fade-in-up p-6"
              style={{ animationDelay: '0.15s' }}
            >
              <h3 style={{ marginBottom: 20, color: token.colorTextHeading }}>用量明细</h3>
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

          {/* 排行榜（管理员可见）— 玻璃态卡片 */}
          {isAdmin && (
            <Col xs={24} lg={10}>
              <div
                className="glass-card animate-fade-in-up p-6"
                style={{ animationDelay: '0.2s' }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                  <h3 style={{ margin: 0, color: token.colorTextHeading }}>用量排行</h3>
                  <Segmented
                    value={rankType}
                    onChange={(v) => setRankType(v as 'user' | 'department')}
                    options={[
                      { label: '用户', value: 'user' },
                      { label: '部门', value: 'department' },
                    ]}
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
