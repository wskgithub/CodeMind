import { useState, useEffect, useCallback } from 'react';
import {
  Table, Button, Select, Space, Tag, message, DatePicker, Drawer,
  Card, Statistic, Row, Col, Popconfirm, Tooltip,
} from 'antd';
import {
  DatabaseOutlined, DownloadOutlined, ReloadOutlined,
  CheckCircleOutlined, StopOutlined, EyeOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import type { TrainingDataItem, TrainingDataDetail, TrainingDataStats } from '@/types';
import {
  listTrainingData,
  getTrainingDataDetail,
  updateExcluded,
  getTrainingDataStats,
  exportTrainingData,
  type TrainingDataListParams,
} from '@/services/trainingDataService';
import useAppStore from '@/store/appStore';

const { RangePicker } = DatePicker;

const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
    }}
  >
    {icon}
  </span>
);

const REQUEST_TYPE_OPTIONS = [
  { label: '全部类型', value: '' },
  { label: 'Chat Completion', value: 'chat_completion' },
  { label: 'Completion', value: 'completion' },
  { label: 'Embedding', value: 'embedding' },
  { label: 'Responses', value: 'responses' },
  { label: 'Anthropic Messages', value: 'anthropic_messages' },
];

const REQUEST_TYPE_COLORS: Record<string, string> = {
  chat_completion: 'blue',
  completion: 'cyan',
  embedding: 'purple',
  responses: 'geekblue',
  anthropic_messages: 'orange',
};

const TrainingDataPage: React.FC = () => {
  const { themeMode } = useAppStore();

  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<TrainingDataItem[]>([]);
  const [total, setTotal] = useState(0);
  const [stats, setStats] = useState<TrainingDataStats | null>(null);
  const [params, setParams] = useState<TrainingDataListParams>({
    page: 1,
    page_size: 20,
  });

  // 详情抽屉
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detail, setDetail] = useState<TrainingDataDetail | null>(null);

  const [exporting, setExporting] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await listTrainingData(params);
      if (res.data.code === 0) {
        setData(res.data.data.list || []);
        setTotal(res.data.data.pagination.total);
      }
    } catch {
      message.error('加载训练数据失败');
    } finally {
      setLoading(false);
    }
  }, [params]);

  const fetchStats = useCallback(async () => {
    try {
      const res = await getTrainingDataStats();
      if (res.data.code === 0) {
        setStats(res.data.data);
      }
    } catch {
      // 静默失败
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  const handleViewDetail = async (id: number) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await getTrainingDataDetail(id);
      if (res.data.code === 0) {
        setDetail(res.data.data);
      }
    } catch {
      message.error('加载详情失败');
    } finally {
      setDetailLoading(false);
    }
  };

  const handleToggleExclude = async (id: number, currentExcluded: boolean) => {
    try {
      await updateExcluded(id, !currentExcluded);
      message.success(currentExcluded ? '已恢复到训练集' : '已从训练集排除');
      fetchData();
      fetchStats();
    } catch {
      message.error('操作失败');
    }
  };

  const handleExport = async () => {
    setExporting(true);
    try {
      await exportTrainingData({
        model: params.model,
        request_type: params.request_type,
        start_date: params.start_date,
        end_date: params.end_date,
      });
      message.success('导出成功');
    } catch {
      message.error('导出失败');
    } finally {
      setExporting(false);
    }
  };

  const columns: ColumnsType<TrainingDataItem> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: '用户',
      dataIndex: 'username',
      width: 120,
      ellipsis: true,
    },
    {
      title: '模型',
      dataIndex: 'model',
      width: 180,
      ellipsis: true,
      render: (model: string) => (
        <Tag style={{ maxWidth: 160 }} className="truncate">{model}</Tag>
      ),
    },
    {
      title: '请求类型',
      dataIndex: 'request_type',
      width: 150,
      render: (type: string) => (
        <Tag color={REQUEST_TYPE_COLORS[type] || 'default'}>{type}</Tag>
      ),
    },
    {
      title: '流式',
      dataIndex: 'is_stream',
      width: 70,
      align: 'center',
      render: (v: boolean) => v ? <Tag color="green">是</Tag> : <Tag>否</Tag>,
    },
    {
      title: 'Tokens',
      dataIndex: 'total_tokens',
      width: 100,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: '耗时',
      dataIndex: 'duration_ms',
      width: 90,
      align: 'right',
      render: (v: number | null) => v != null ? `${(v / 1000).toFixed(1)}s` : '-',
    },
    {
      title: '状态',
      dataIndex: 'is_excluded',
      width: 90,
      align: 'center',
      render: (excluded: boolean) => excluded
        ? <Tag color="red">已排除</Tag>
        : <Tag color="green">可用</Tag>,
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 170,
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      width: 140,
      fixed: 'right',
      render: (_: unknown, record: TrainingDataItem) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record.id)}
            />
          </Tooltip>
          <Popconfirm
            title={record.is_excluded ? '确定恢复到训练集？' : '确定从训练集中排除？'}
            onConfirm={() => handleToggleExclude(record.id, record.is_excluded)}
          >
            <Tooltip title={record.is_excluded ? '恢复' : '排除'}>
              <Button
                type="link"
                size="small"
                danger={!record.is_excluded}
                icon={record.is_excluded ? <CheckCircleOutlined /> : <StopOutlined />}
              />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-bg animate-fade-in-up">
      {/* 页面头部 */}
      <div className="flex items-center gap-4 mb-6">
        <PageIcon icon={<DatabaseOutlined />} />
        <div>
          <h2
            className="text-xl font-bold m-0"
            style={{ color: themeMode === 'dark' ? '#fff' : '#1a1a2e' }}
          >
            训练数据管理
          </h2>
          <p
            className="text-sm m-0 mt-1"
            style={{ color: themeMode === 'dark' ? 'rgba(255,255,255,0.5)' : 'rgba(0,0,0,0.45)' }}
          >
            查看和管理 LLM 请求/响应记录，导出为模型训练数据
          </p>
        </div>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <Row gutter={16} className="mb-4">
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic title="总记录数" value={stats.total_count} />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic title="今日新增" value={stats.today_count} />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic
                title="可用 / 已排除"
                value={stats.total_count - stats.excluded_count}
                suffix={`/ ${stats.excluded_count}`}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic
                title="模型数"
                value={stats.model_distribution?.length || 0}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* 主卡片 */}
      <div className="glass-card" style={{ padding: 24, borderRadius: 16 }}>
        {/* 筛选栏 */}
        <div className="flex items-center justify-between flex-wrap gap-3 mb-4">
          <Space wrap>
            <Select
              placeholder="模型"
              allowClear
              style={{ width: 200 }}
              value={params.model || undefined}
              onChange={(v) => setParams((p) => ({ ...p, model: v || undefined, page: 1 }))}
              options={
                stats?.model_distribution?.map((m) => ({
                  label: `${m.model} (${m.count})`,
                  value: m.model,
                })) || []
              }
            />
            <Select
              placeholder="请求类型"
              allowClear
              style={{ width: 180 }}
              value={params.request_type || undefined}
              onChange={(v) => setParams((p) => ({ ...p, request_type: v || undefined, page: 1 }))}
              options={REQUEST_TYPE_OPTIONS.filter((o) => o.value !== '')}
            />
            <RangePicker
              onChange={(dates) => {
                setParams((p) => ({
                  ...p,
                  start_date: dates?.[0]?.format('YYYY-MM-DD'),
                  end_date: dates?.[1]?.format('YYYY-MM-DD'),
                  page: 1,
                }));
              }}
            />
          </Space>

          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => { fetchData(); fetchStats(); }}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              loading={exporting}
              onClick={handleExport}
            >
              导出 JSONL
            </Button>
          </Space>
        </div>

        {/* 数据表格 */}
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            current: params.page,
            pageSize: params.page_size,
            total,
            showSizeChanger: true,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (page, pageSize) => setParams((p) => ({ ...p, page, page_size: pageSize })),
          }}
        />
      </div>

      {/* 详情抽屉 */}
      <Drawer
        title={`训练数据详情 #${detail?.id || ''}`}
        open={detailVisible}
        onClose={() => { setDetailVisible(false); setDetail(null); }}
        width={720}
        loading={detailLoading}
      >
        {detail && (
          <div className="space-y-4">
            <Row gutter={16}>
              <Col span={8}><Statistic title="模型" value={detail.model} /></Col>
              <Col span={8}><Statistic title="请求类型" value={detail.request_type} /></Col>
              <Col span={8}>
                <Statistic title="Tokens" value={detail.total_tokens} suffix={`(${detail.prompt_tokens}+${detail.completion_tokens})`} />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}><Statistic title="耗时" value={detail.duration_ms != null ? `${(detail.duration_ms / 1000).toFixed(2)}s` : '-'} /></Col>
              <Col span={8}><Statistic title="流式" value={detail.is_stream ? '是' : '否'} /></Col>
              <Col span={8}><Statistic title="状态" value={detail.is_excluded ? '已排除' : '可用'} /></Col>
            </Row>

            <div>
              <h4 style={{ marginBottom: 8, fontWeight: 600 }}>请求体 (Request Body)</h4>
              <pre
                style={{
                  background: themeMode === 'dark' ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.03)',
                  border: `1px solid ${themeMode === 'dark' ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)'}`,
                  borderRadius: 8,
                  padding: 16,
                  maxHeight: 300,
                  overflow: 'auto',
                  fontSize: 12,
                  lineHeight: 1.6,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                }}
              >
                {JSON.stringify(detail.request_body, null, 2)}
              </pre>
            </div>

            <div>
              <h4 style={{ marginBottom: 8, fontWeight: 600 }}>响应体 (Response Body)</h4>
              {detail.response_body ? (
                <pre
                  style={{
                    background: themeMode === 'dark' ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.03)',
                    border: `1px solid ${themeMode === 'dark' ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)'}`,
                    borderRadius: 8,
                    padding: 16,
                    maxHeight: 300,
                    overflow: 'auto',
                    fontSize: 12,
                    lineHeight: 1.6,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                  }}
                >
                  {JSON.stringify(detail.response_body, null, 2)}
                </pre>
              ) : (
                <p style={{ color: 'rgba(128,128,128,0.6)', fontStyle: 'italic' }}>
                  无响应体（Embedding 请求不记录响应向量）
                </p>
              )}
            </div>
          </div>
        )}
      </Drawer>
    </div>
  );
};

export default TrainingDataPage;
