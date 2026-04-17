import {
  ThunderboltOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { Progress, Tooltip, Spin } from 'antd';
import type { TFunction } from 'i18next';
import { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

import { getMyLimitProgress } from '@/services/limitService';
import useAppStore from '@/store/appStore';
import type { LimitProgressItem, LimitProgressResponse } from '@/types';

function formatPeriod(t: TFunction, period: string, hours: number): string {
  const periodKeyMap: Record<string, string> = {
    daily: 'periodDaily',
    weekly: 'periodWeekly',
    monthly: 'periodMonthly',
    custom: 'periodCustom',
  };
  const periodKey = periodKeyMap[period];
  if (periodKey && period !== 'custom') {
    return t(`usageProgress.${periodKey}`) + t('usageProgress.quota');
  }
  if (hours < 24) return t('usageProgress.hourQuota', { hours });
  if (hours === 24) return t('usageProgress.periodDaily') + t('usageProgress.quota');
  if (hours < 168) return t('usageProgress.dayQuota', { days: Math.round(hours / 24) });
  if (hours === 168) return t('usageProgress.periodWeekly') + t('usageProgress.quota');
  if (hours === 720) return t('usageProgress.periodMonthly') + t('usageProgress.quota');
  return t('usageProgress.hourQuota', { hours });
}

function formatTokens(num: number): string {
  if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(1)}M`;
  if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
  return num.toString();
}

function formatResetTime(t: TFunction, hours: number | null): string {
  if (hours === null) return t('usageProgress.notStarted');
  if (hours <= 0) return t('usageProgress.resetSoon');
  if (hours < 1) return t('usageProgress.minutes', { count: Math.ceil(hours * 60) });
  if (hours < 24) return t('usageProgress.hours', { count: Math.round(hours * 10) / 10 });
  const days = Math.floor(hours / 24);
  const remainHours = Math.round(hours % 24);
  if (remainHours === 0) return t('usageProgress.days', { count: days });
  return t('usageProgress.daysHours', { days, hours: remainHours });
}

function getProgressColor(percent: number, exceeded: boolean): string {
  if (exceeded) return '#FF6B6B';
  if (percent >= 80) return '#FFBE0B';
  return '#00D9FF';
}

function getStrokeColor(percent: number, exceeded: boolean) {
  if (exceeded) return { '0%': '#FF6B6B', '100%': '#FF8787' };
  if (percent >= 80) return { '0%': '#FFBE0B', '100%': '#FFD43B' };
  return { '0%': '#00D9FF', '100%': '#00F5D4' };
}

const LimitCard = ({ item, t }: { item: LimitProgressItem; t: TFunction }) => {
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
            {formatPeriod(t, item.period, item.period_hours)}
          </span>
        </div>

        {item.exceeded ? (
          <Tooltip title={t('usageProgress.quotaExceeded')}>
            <WarningOutlined style={{ color: '#FF6B6B', fontSize: 16 }} />
          </Tooltip>
        ) : item.cycle_start_at ? (
          <Tooltip title={t('usageProgress.cycleInProgress')}>
            <CheckCircleOutlined style={{ color: '#00F5D4', fontSize: 16 }} />
          </Tooltip>
        ) : null}
      </div>

      <Progress
        percent={percent}
        strokeColor={getStrokeColor(percent, item.exceeded)}
        trailColor={isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}
        showInfo={false}
        size="small"
        style={{ marginBottom: 10 }}
      />

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

      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        fontSize: 12,
        color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)',
      }}>
        <ClockCircleOutlined style={{ fontSize: 11 }} />
        {item.reset_in_hours !== null ? (
          <span>{t('usageProgress.resetIn', { time: formatResetTime(t, item.reset_in_hours) })}</span>
        ) : (
          <span style={{ color: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.35)' }}>{t('usageProgress.waitingForUsage')}</span>
        )}
      </div>
    </div>
  );
};

const UsageProgressCards: React.FC = () => {
  const { t } = useTranslation();
  const [data, setData] = useState<LimitProgressResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const loadProgress = useCallback(async () => {
    try {
      const res = await getMyLimitProgress();
      setData(res.data.data);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadProgress();
    // refresh every 60s
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

  // sort by period hours for consistent ordering
  const sortedLimits = [...data.limits].sort((a, b) => a.period_hours - b.period_hours);

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
            <LimitCard item={item} t={t} />
          </div>
        ))}
      </div>
    </div>
  );
};

export default UsageProgressCards;
