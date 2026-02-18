import { useEffect, useState, useCallback } from 'react';
import { Progress, Tooltip, Spin, theme } from 'antd';
import {
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { getMyLimitProgress } from '@/services/limitService';
import type { LimitProgressItem, LimitProgressResponse } from '@/types';

/** 周期标签到中文映射 */
const periodLabels: Record<string, string> = {
  daily: '每日',
  weekly: '每周',
  monthly: '每月',
  custom: '自定义',
};

/** 格式化周期显示 */
function formatPeriod(period: string, hours: number): string {
  const label = periodLabels[period];
  if (label && period !== 'custom') return label + '限额';
  if (hours < 24) return `${hours} 小时限额`;
  if (hours === 24) return '每日限额';
  if (hours < 168) return `${Math.round(hours / 24)} 天限额`;
  if (hours === 168) return '每周限额';
  if (hours === 720) return '每月限额';
  return `${hours} 小时限额`;
}

/** 格式化 Token 数量 */
function formatTokens(num: number): string {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`;
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
  return num.toString();
}

/** 格式化剩余时间 */
function formatResetTime(hours: number | null): string {
  if (hours === null) return '未启动';
  if (hours <= 0) return '即将重置';
  if (hours < 1) return `${Math.ceil(hours * 60)} 分钟`;
  if (hours < 24) return `${Math.round(hours * 10) / 10} 小时`;
  const days = Math.floor(hours / 24);
  const remainHours = Math.round(hours % 24);
  if (remainHours === 0) return `${days} 天`;
  return `${days} 天 ${remainHours} 小时`;
}

/** 获取进度条渐变色 */
function getProgressColor(percent: number, exceeded: boolean): string {
  if (exceeded) return '#ff4d4f';
  if (percent >= 80) return '#faad14';
  return '#2B7CB3';
}

/** 获取进度条渐变配置 */
function getStrokeColor(percent: number, exceeded: boolean) {
  if (exceeded) return { '0%': '#ff4d4f', '100%': '#ff7875' };
  if (percent >= 80) return { '0%': '#faad14', '100%': '#ffc53d' };
  return { '0%': '#2B7CB3', '100%': '#4BA3D4' };
}

/** 单个限额进度卡片 */
const LimitCard = ({ item }: { item: LimitProgressItem }) => {
  const { token } = theme.useToken();
  const percent = Math.min(item.usage_percent, 100);
  const color = getProgressColor(percent, item.exceeded);

  return (
    <div
      className="glass-card animate-fade-in-up"
      style={{
        padding: '20px 24px',
        position: 'relative',
        overflow: 'hidden',
        minWidth: 0,
      }}
    >
      {/* 顶部状态指示条 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 3,
        background: item.exceeded
          ? 'linear-gradient(90deg, #ff4d4f, #ff7875)'
          : percent >= 80
          ? 'linear-gradient(90deg, #faad14, #ffc53d)'
          : 'var(--gradient-primary)',
        opacity: percent > 50 ? 1 : 0.6,
      }} />

      {/* 标题行 */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{
            width: 32,
            height: 32,
            borderRadius: 8,
            background: item.exceeded
              ? 'rgba(255, 77, 79, 0.1)'
              : 'rgba(43, 124, 179, 0.1)',
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}>
            <ThunderboltOutlined style={{
              fontSize: 16,
              color: item.exceeded ? '#ff4d4f' : '#2B7CB3',
            }} />
          </span>
          <span style={{
            fontWeight: 600,
            fontSize: 14,
            color: token.colorTextHeading,
          }}>
            {formatPeriod(item.period, item.period_hours)}
          </span>
        </div>

        {item.exceeded ? (
          <Tooltip title="已达限额，等待重置">
            <WarningOutlined style={{ color: '#ff4d4f', fontSize: 16 }} />
          </Tooltip>
        ) : item.cycle_start_at ? (
          <Tooltip title="周期进行中">
            <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 16 }} />
          </Tooltip>
        ) : null}
      </div>

      {/* 进度条 */}
      <Progress
        percent={percent}
        strokeColor={getStrokeColor(percent, item.exceeded)}
        trailColor="rgba(0,0,0,0.06)"
        showInfo={false}
        size="small"
        style={{ marginBottom: 10 }}
      />

      {/* 数值详情 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 6 }}>
        <span style={{ fontSize: 20, fontWeight: 700, color }}>
          {formatTokens(item.used_tokens)}
        </span>
        <span style={{ fontSize: 12, color: token.colorTextTertiary }}>
          / {formatTokens(item.max_tokens)}
        </span>
      </div>

      {/* 重置时间 */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        fontSize: 12,
        color: token.colorTextSecondary,
      }}>
        <ClockCircleOutlined style={{ fontSize: 11 }} />
        {item.reset_in_hours !== null ? (
          <span>{formatResetTime(item.reset_in_hours)} 后重置</span>
        ) : (
          <span style={{ color: token.colorTextTertiary }}>等待使用后开始计时</span>
        )}
      </div>
    </div>
  );
};

/** 限额使用进度卡片组 — 根据规则数量自适应布局 */
const UsageProgressCards: React.FC = () => {
  const [data, setData] = useState<LimitProgressResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const loadProgress = useCallback(async () => {
    try {
      const res = await getMyLimitProgress();
      setData(res.data.data);
    } catch {
      // 错误已由拦截器统一处理
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadProgress();
    // 每 60 秒刷新一次
    const timer = setInterval(loadProgress, 60_000);
    return () => clearInterval(timer);
  }, [loadProgress]);

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 16 }}>
        <Spin size="small" />
      </div>
    );
  }

  if (!data || data.limits.length === 0) {
    return null;
  }

  // 根据卡片数量自适应 grid 列数
  const count = data.limits.length;
  const getGridCols = () => {
    if (count === 1) return 'repeat(1, 1fr)';
    if (count === 2) return 'repeat(2, 1fr)';
    if (count === 3) return 'repeat(3, 1fr)';
    return `repeat(auto-fill, minmax(260px, 1fr))`;
  };

  return (
    <div style={{ marginBottom: 16 }}>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: getGridCols(),
          gap: 12,
        }}
      >
        {data.limits.map((item, index) => (
          <div key={item.rule_id} style={{ animationDelay: `${index * 0.05}s` }}>
            <LimitCard item={item} />
          </div>
        ))}
      </div>
    </div>
  );
};

export default UsageProgressCards;
