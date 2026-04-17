import {
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { Progress, Tooltip, Spin } from 'antd';
import { useEffect, useState, useCallback } from 'react';

import { getMyLimitProgress } from '@/services/limitService';
import useAppStore from '@/store/appStore';
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

/** 获取进度条渐变色 - 新设计 */
function getProgressColor(percent: number, exceeded: boolean): string {
  if (exceeded) return '#FF6B6B';
  if (percent >= 80) return '#FFBE0B';
  return '#00D9FF';
}

/** 获取进度条渐变配置 - 新设计 */
function getStrokeColor(percent: number, exceeded: boolean) {
  if (exceeded) return { '0%': '#FF6B6B', '100%': '#FF8787' };
  if (percent >= 80) return { '0%': '#FFBE0B', '100%': '#FFD43B' };
  return { '0%': '#00D9FF', '100%': '#00F5D4' };
}

/** 单个限额进度卡片 - 新设计 */
const LimitCard = ({ item }: { item: LimitProgressItem }) => {
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';
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
        background: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.7)',
        border: isDark ? '1px solid rgba(255, 255, 255, 0.06)' : '1px solid rgba(0, 0, 0, 0.06)',
        borderRadius: 20,
      }}
    >
      {/* 顶部状态指示条 - 新设计 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 3,
        background: item.exceeded
          ? 'linear-gradient(90deg, #FF6B6B, #FF8787)'
          : percent >= 80
          ? 'linear-gradient(90deg, #FFBE0B, #FFD43B)'
          : 'linear-gradient(90deg, #00D9FF, #00F5D4)',
        opacity: percent > 50 ? 1 : 0.6,
      }} />

      {/* 标题行 */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{
            width: 36,
            height: 36,
            borderRadius: 10,
            background: item.exceeded
              ? 'rgba(255, 107, 107, 0.15)'
              : 'rgba(0, 217, 255, 0.15)',
            display: 'inline-flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
            border: `1px solid ${item.exceeded ? 'rgba(255, 107, 107, 0.3)' : 'rgba(0, 217, 255, 0.3)'}`,
          }}>
            <ThunderboltOutlined style={{
              fontSize: 18,
              color: item.exceeded ? '#FF6B6B' : '#00D9FF',
            }} />
          </span>
          <span style={{
            fontWeight: 600,
            fontSize: 14,
            color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
          }}>
            {formatPeriod(item.period, item.period_hours)}
          </span>
        </div>

        {item.exceeded ? (
          <Tooltip title="已达限额，等待重置">
            <WarningOutlined style={{ color: '#FF6B6B', fontSize: 16 }} />
          </Tooltip>
        ) : item.cycle_start_at ? (
          <Tooltip title="周期进行中">
            <CheckCircleOutlined style={{ color: '#00F5D4', fontSize: 16 }} />
          </Tooltip>
        ) : null}
      </div>

      {/* 进度条 - 新设计 */}
      <Progress
        percent={percent}
        strokeColor={getStrokeColor(percent, item.exceeded)}
        trailColor={isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}
        showInfo={false}
        size="small"
        style={{ marginBottom: 10 }}
      />

      {/* 数值详情 */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 6 }}>
        <span style={{ fontSize: 22, fontWeight: 700, color }}>
          {formatTokens(item.used_tokens)}
        </span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ 
            fontSize: 14, 
            fontWeight: 600, 
            color: item.exceeded ? '#FF6B6B' : percent >= 80 ? '#FFBE0B' : '#00D9FF' 
          }}>
            {item.usage_percent}%
          </span>
          <span style={{ fontSize: 12, color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)' }}>
            / {formatTokens(item.max_tokens)}
          </span>
        </div>
      </div>

      {/* 重置时间 */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        fontSize: 12,
        color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
      }}>
        <ClockCircleOutlined style={{ fontSize: 11 }} />
        {item.reset_in_hours !== null ? (
          <span>{formatResetTime(item.reset_in_hours)} 后重置</span>
        ) : (
          <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.35)' }}>等待使用后开始计时</span>
        )}
      </div>
    </div>
  );
};

/** 限额使用进度卡片组 — 与首页/登录页新设计风格统一 */
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

  // 按周期小时数排序，保证卡片顺序稳定（短周期在前，长周期在后）
  const sortedLimits = [...data.limits].sort((a, b) => a.period_hours - b.period_hours);

  // 根据卡片数量自适应 grid 列数
  const count = sortedLimits.length;
  const getGridCols = () => {
    if (count === 1) return 'repeat(1, 1fr)';
    if (count === 2) return 'repeat(2, 1fr)';
    if (count === 3) return 'repeat(3, 1fr)';
    return `repeat(auto-fill, minmax(260px, 1fr))`;
  };

  return (
    <div style={{ marginBottom: 24 }}>
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: getGridCols(),
          gap: 16,
        }}
      >
        {sortedLimits.map((item, index) => (
          <div key={item.rule_id} style={{ animationDelay: `${index * 0.05}s` }}>
            <LimitCard item={item} />
          </div>
        ))}
      </div>
    </div>
  );
};

export default UsageProgressCards;
