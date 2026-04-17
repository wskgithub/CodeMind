import { useState, useEffect, useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
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
  { label: 'All Types', value: '' },
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
  const themeMode = useAppStore((s) => s.themeMode);
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<TrainingDataItem[]>([]);
  const [total, setTotal] = useState(0);
  const [stats, setStats] = useState<TrainingDataStats | null>(null);
  const [params, setParams] = useState<TrainingDataListParams>({
    page: 1,
    page_size: 20,
  });

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
      message.error(t('training.loadFailed'));
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
      // silently fail
    }
  }, []);

  useEffect(() => {
    Promise.all([fetchData(), fetchStats()]);
  }, [fetchData, fetchStats]);

  const handleViewDetail = async (id: number) => {
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await getTrainingDataDetail(id);
      if (res.data.code === 0) {
        setDetail(res.data.data);
      }
    } catch {
      message.error(t('training.loadDetailFailed'));
    } finally {
      setDetailLoading(false);
    }
  };

  const handleToggleExclude = async (id: number, currentExcluded: boolean) => {
    try {
      await updateExcluded(id, !currentExcluded);
      message.success(currentExcluded ? t('training.restoredToTraining') : t('training.excludedFromTraining'));
      fetchData();
      fetchStats();
    } catch {
      message.error(t('training.operationFailed'));
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
      message.success(t('training.exportSuccess'));
    } catch {
      message.error(t('training.exportFailed'));
    } finally {
      setExporting(false);
    }
  };

  const columns: ColumnsType<TrainingDataItem> = useMemo(() => [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('training.table.user'),
      dataIndex: 'username',
      width: 120,
      ellipsis: true,
    },
    {
      title: t('training.table.model'),
      dataIndex: 'model',
      width: 180,
      ellipsis: true,
      render: (model: string) => (
        <Tag style={{ maxWidth: 160 }} className="truncate">{model}</Tag>
      ),
    },
    {
      title: t('training.table.requestType'),
      dataIndex: 'request_type',
      width: 150,
      render: (type: string) => (
        <Tag color={REQUEST_TYPE_COLORS[type] || 'default'}>{type}</Tag>
      ),
    },
    {
      title: t('training.table.stream'),
      dataIndex: 'is_stream',
      width: 70,
      align: 'center',
      render: (v: boolean) => v ? <Tag color="green">{t('common.yes')}</Tag> : <Tag>{t('common.no')}</Tag>,
    },
    {
      title: 'Tokens',
      dataIndex: 'total_tokens',
      width: 100,
      align: 'right',
      render: (v: number) => v.toLocaleString(),
    },
    {
      title: t('training.table.duration'),
      dataIndex: 'duration_ms',
      width: 90,
      align: 'right',
      render: (v: number | null) => v != null ? `${(v / 1000).toFixed(1)}s` : '-',
    },
    {
      title: t('common.status'),
      dataIndex: 'is_excluded',
      width: 90,
      align: 'center',
      render: (excluded: boolean) => excluded
        ? <Tag color="red">{t('training.table.excluded')}</Tag>
        : <Tag color="green">{t('training.table.available')}</Tag>,
    },
    {
      title: t('training.table.createdAt'),
      dataIndex: 'created_at',
      width: 170,
      render: (val: string) => dayjs(val).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: t('common.actions'),
      width: 140,
      fixed: 'right',
      render: (_: unknown, record: TrainingDataItem) => (
        <Space size="small">
          <Tooltip title={t('training.viewDetail')}>
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record.id)}
            />
          </Tooltip>
          <Popconfirm
            title={record.is_excluded ? t('training.confirmRestore') : t('training.confirmExclude')}
            onConfirm={() => handleToggleExclude(record.id, record.is_excluded)}
          >
            <Tooltip title={record.is_excluded ? t('training.restore') : t('training.exclude')}>
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
  ], [t]);

  return (
    <div className="page-bg animate-fade-in-up">
      <div className="flex items-center gap-4 mb-6">
        <PageIcon icon={<DatabaseOutlined />} />
        <div>
          <h2
            className="text-xl font-bold m-0"
            style={{ color: themeMode === 'dark' ? '#fff' : '#1a1a2e' }}
          >
            {t('training.pageTitle')}
          </h2>
          <p
            className="text-sm m-0 mt-1"
            style={{ color: themeMode === 'dark' ? 'rgba(255,255,255,0.5)' : 'rgba(0,0,0,0.45)' }}
          >
            {t('training.pageDescription')}
          </p>
        </div>
      </div>

      {stats && (
        <Row gutter={16} className="mb-4">
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic title={t('training.stats.totalRecords')} value={stats.total_count} />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic title={t('training.stats.todayNew')} value={stats.today_count} />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic
                title={t('training.stats.availableExcluded')}
                value={stats.total_count - stats.excluded_count}
                suffix={`/ ${stats.excluded_count}`}
              />
            </Card>
          </Col>
          <Col span={6}>
            <Card className="glass-card" size="small">
              <Statistic
                title={t('training.stats.modelCount')}
                value={stats.model_distribution?.length || 0}
              />
            </Card>
          </Col>
        </Row>
      )}

      <div className="glass-card" style={{ padding: 24, borderRadius: 16 }}>
        <div className="flex items-center justify-between flex-wrap gap-3 mb-4">
          <Space wrap>
            <Select
              placeholder={t('training.filter.model')}
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
              placeholder={t('training.filter.requestType')}
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
              {t('common.refresh')}
            </Button>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              loading={exporting}
              onClick={handleExport}
            >
              {t('training.exportJsonl')}
            </Button>
          </Space>
        </div>

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
            showTotal: (total) => t('training.pagination.total', { total }),
            onChange: (page, pageSize) => setParams((p) => ({ ...p, page, page_size: pageSize })),
          }}
        />
      </div>

      <Drawer
        title={t('training.detailTitle', { id: detail?.id || '' })}
        open={detailVisible}
        onClose={() => { setDetailVisible(false); setDetail(null); }}
        width={720}
        loading={detailLoading}
      >
        {detail && (
          <div className="space-y-4">
            <Row gutter={16}>
              <Col span={8}><Statistic title={t('training.table.model')} value={detail.model} /></Col>
              <Col span={8}><Statistic title={t('training.table.requestType')} value={detail.request_type} /></Col>
              <Col span={8}>
                <Statistic title="Tokens" value={detail.total_tokens} suffix={`(${detail.prompt_tokens}+${detail.completion_tokens})`} />
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={8}><Statistic title={t('training.table.duration')} value={detail.duration_ms != null ? `${(detail.duration_ms / 1000).toFixed(2)}s` : '-'} /></Col>
              <Col span={8}><Statistic title={t('training.table.stream')} value={detail.is_stream ? t('common.yes') : t('common.no')} /></Col>
              <Col span={8}><Statistic title={t('common.status')} value={detail.is_excluded ? t('training.table.excluded') : t('training.table.available')} /></Col>
            </Row>

            <div>
              <h4 style={{ marginBottom: 8, fontWeight: 600 }}>{t('training.detail.requestBody')}</h4>
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
              <h4 style={{ marginBottom: 8, fontWeight: 600 }}>{t('training.detail.responseBody')}</h4>
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
                  {t('training.detail.noResponse')}
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
