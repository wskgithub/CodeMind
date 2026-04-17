import {
  ApiOutlined,
  RobotOutlined,
  GatewayOutlined,
  ThunderboltOutlined,
  SafetyCertificateOutlined,
  BarChartOutlined,
  CloudServerOutlined,
  CodeOutlined,
  TeamOutlined,
  ArrowDownOutlined,
  PlayCircleOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons';
import { Button } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { type SupportedLanguage } from '@/i18n';
import useAppStore from '@/store/appStore';
import useAuthStore from '@/store/authStore';

// features/stats 数据移到组件内使用 t() 函数
const FEATURE_KEYS = [
  { key: 'openai', icon: <ApiOutlined />, color: '#00D9FF' },
  { key: 'anthropic', icon: <RobotOutlined />, color: '#FF6B6B' },
  { key: 'mcp', icon: <GatewayOutlined />, color: '#9D4EDD' },
  { key: 'loadBalance', icon: <CloudServerOutlined />, color: '#00F5D4' },
  { key: 'rbac', icon: <SafetyCertificateOutlined />, color: '#FFBE0B' },
  { key: 'metering', icon: <BarChartOutlined />, color: '#FB5607' },
  { key: 'ide', icon: <CodeOutlined />, color: '#3A86FF' },
  { key: 'sse', icon: <ThunderboltOutlined />, color: '#06FFA5' },
  { key: 'onprem', icon: <TeamOutlined />, color: '#E85D04' },
];

const STAT_KEYS = [
  { key: 'ideSupport', value: 15, suffix: '+', suffixKey: '', icon: <CodeOutlined /> },
  { key: 'permissionLevels', value: 3, suffix: '', suffixKey: 'levelSuffix', icon: <SafetyCertificateOutlined /> },
  { key: 'availability', value: 99, suffix: '%', suffixKey: '', icon: <CloudServerOutlined /> },
  { key: 'responseLatency', value: 0, suffix: '.1s', suffixKey: '', icon: <ThunderboltOutlined /> },
];

const TOOLS = [
  'Claude Code', 'Cursor', 'VS Code', 'JetBrains', 
  'Cline', 'Roo Code', 'Kilo Code', 'TRAE', 
  'OpenCode', 'Factory Droid', 'Cherry Studio', 'Goose'
];

const StarfieldCanvas: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const mouseRef = useRef({ x: 0, y: 0 });

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animId: number;
    let lastTime = 0;
    const FRAME_INTERVAL = 1000 / 30;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    const resize = () => {
      canvas.width = window.innerWidth * dpr;
      canvas.height = window.innerHeight * dpr;
      canvas.style.width = `${window.innerWidth}px`;
      canvas.style.height = `${window.innerHeight}px`;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };
    resize();
    window.addEventListener('resize', resize);

    // 粒子数量根据屏幕面积动态计算，上限 80（原 150）
    const count = Math.min(Math.floor((window.innerWidth * window.innerHeight) / 15000), 80);

    interface Particle {
      x: number; y: number; vx: number; vy: number;
      r: number; color: string; glowColor: string;
    }

    // 预计算颜色字符串，避免每帧重复拼接
    const particles: Particle[] = [];
    for (let i = 0; i < count; i++) {
      const a = Math.random() * 0.5 + 0.3;
      particles.push({
        x: Math.random() * window.innerWidth,
        y: Math.random() * window.innerHeight,
        vx: (Math.random() - 0.5) * 0.3,
        vy: (Math.random() - 0.5) * 0.3,
        r: Math.random() * 2 + 0.5,
        color: `rgba(255,255,255,${Math.min(a + 0.3, 1).toFixed(2)})`,
        glowColor: `rgba(0,217,255,${(a * 0.3).toFixed(2)})`,
      });
    }

    const MAX_DIST = 150;
    const MAX_DIST_SQ = MAX_DIST * MAX_DIST;
    const MOUSE_DIST = 250;
    const MOUSE_DIST_SQ = MOUSE_DIST * MOUSE_DIST;

    const draw = (timestamp: number) => {
      animId = requestAnimationFrame(draw);
      // 帧率限制 30fps + 页面不可见时暂停渲染
      if (timestamp - lastTime < FRAME_INTERVAL || document.hidden) return;
      lastTime = timestamp;

      const w = window.innerWidth;
      const h = window.innerHeight;
      ctx.clearRect(0, 0, w, h);

      for (const p of particles) {
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < 0 || p.x > w) p.vx *= -1;
        if (p.y < 0 || p.y > h) p.vy *= -1;
      }

      // 空间网格索引，减少两两遍历开销
      const grid = new Map<string, Particle[]>();
      for (const p of particles) {
        const key = `${Math.floor(p.x / MAX_DIST)},${Math.floor(p.y / MAX_DIST)}`;
        const cell = grid.get(key);
        if (cell) cell.push(p);
        else grid.set(key, [p]);
      }

      ctx.lineWidth = 0.5;
      for (const [key, cell] of grid) {
        const [cx, cy] = key.split(',').map(Number);
        for (let ddx = 0; ddx <= 1; ddx++) {
          for (let ddy = -1; ddy <= 1; ddy++) {
            if (ddx === 0 && ddy <= 0 && !(ddx === 0 && ddy === 0)) continue;
            const neighbor = grid.get(`${cx! + ddx},${cy! + ddy}`);
            if (!neighbor) continue;
            const isSame = ddx === 0 && ddy === 0;
            for (let i = 0; i < cell.length; i++) {
              for (let j = isSame ? i + 1 : 0; j < neighbor.length; j++) {
                const dx = cell[i]!.x - neighbor[j]!.x;
                const dy = cell[i]!.y - neighbor[j]!.y;
                const dSq = dx * dx + dy * dy;
                if (dSq < MAX_DIST_SQ) {
                  ctx.strokeStyle = `rgba(100,200,255,${((1 - Math.sqrt(dSq) / MAX_DIST) * 0.2).toFixed(3)})`;
                  ctx.beginPath();
                  ctx.moveTo(cell[i]!.x, cell[i]!.y);
                  ctx.lineTo(neighbor[j]!.x, neighbor[j]!.y);
                  ctx.stroke();
                }
              }
            }
          }
        }
      }

      // 鼠标附近连线，限制最多 8 条避免过度绘制
      const { x: mx, y: my } = mouseRef.current;
      if (mx > 0 && my > 0) {
        let lines = 0;
        ctx.lineWidth = 0.8;
        for (const p of particles) {
          if (lines >= 8) break;
          const dx = mx - p.x;
          const dy = my - p.y;
          const dSq = dx * dx + dy * dy;
          if (dSq < MOUSE_DIST_SQ) {
            ctx.strokeStyle = `rgba(0,217,255,${((1 - Math.sqrt(dSq) / MOUSE_DIST) * 0.3).toFixed(3)})`;
            ctx.beginPath();
            ctx.moveTo(p.x, p.y);
            ctx.lineTo(mx, my);
            ctx.stroke();
            lines++;
          }
        }
      }

      // 粒子绘制：简化为单色填充，不使用渐变以大幅提升性能
      for (const p of particles) {
        ctx.fillStyle = p.glowColor;
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r * 3, 0, Math.PI * 2);
        ctx.fill();
        ctx.fillStyle = p.color;
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fill();
      }
    };

    animId = requestAnimationFrame(draw);

    const onMouseMove = (e: MouseEvent) => {
      mouseRef.current = { x: e.clientX, y: e.clientY };
    };
    window.addEventListener('mousemove', onMouseMove, { passive: true });

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', resize);
      window.removeEventListener('mousemove', onMouseMove);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      style={{ position: 'absolute', inset: 0, zIndex: 1, pointerEvents: 'none' }}
    />
  );
};

const AnimatedCounter: React.FC<{ value: number; suffix: string; duration?: number }> = ({ 
  value, suffix, duration = 2000 
}) => {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLDivElement>(null);
  const hasAnimated = useRef(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting && !hasAnimated.current) {
            hasAnimated.current = true;
            const startTime = Date.now();
            const animate = () => {
              const elapsed = Date.now() - startTime;
              const progress = Math.min(elapsed / duration, 1);
              const easeOut = 1 - Math.pow(1 - progress, 3);
              setCount(Math.floor(easeOut * value));
              if (progress < 1) requestAnimationFrame(animate);
            };
            animate();
          }
        });
      },
      { threshold: 0.5 }
    );

    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [value, duration]);

  return (
    <div 
      ref={ref} 
      style={{ 
        fontSize: 'clamp(2.5rem, 5vw, 4rem)', 
        fontWeight: 800,
        background: 'linear-gradient(135deg, #fff 0%, #00D9FF 50%, #9D4EDD 100%)',
        WebkitBackgroundClip: 'text',
        WebkitTextFillColor: 'transparent',
        backgroundClip: 'text',
      }}
    >
      {count}{suffix}
    </div>
  );
};

const HomePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const { themeMode, toggleTheme, language, setLanguage } = useAppStore();
  const isDark = themeMode === 'dark';
  const featuresRef = useRef<HTMLDivElement>(null);

  // 使用 i18n 生成 features 数据
  const features = FEATURE_KEYS.map((f) => ({
    icon: f.icon,
    title: t(`home.features.${f.key}.title`),
    desc: t(`home.features.${f.key}.desc`),
    color: f.color,
  }));

  // 使用 i18n 生成 stats 数据
  const stats = STAT_KEYS.map((s) => ({
    value: s.value,
    suffix: s.suffixKey ? t(`home.stats.${s.suffixKey}`) : s.suffix,
    label: t(`home.stats.${s.key}`),
    icon: s.icon,
  }));

  // scroll animations
  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('animate-in');
          }
        });
      },
      { threshold: 0.1, rootMargin: '0px 0px -50px 0px' }
    );

    const cards = document.querySelectorAll('.feature-card, .stat-card, .tool-tag');
    cards.forEach((card) => observer.observe(card));

    return () => observer.disconnect();
  }, []);

  return (
    <div className="min-h-screen" style={{ overflow: 'hidden', background: isDark ? '#050d14' : '#f0f5fa' }}>
      <section
        className="hero-section"
        style={{
          position: 'relative',
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: isDark
            ? `
            radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 217, 255, 0.15), transparent),
            radial-gradient(ellipse 60% 40% at 80% 80%, rgba(157, 78, 221, 0.1), transparent),
            radial-gradient(ellipse 50% 30% at 20% 100%, rgba(0, 245, 212, 0.08), transparent),
            linear-gradient(180deg, #0a1628 0%, #050d14 100%)
          `
            : `
            radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 217, 255, 0.2), transparent),
            radial-gradient(ellipse 60% 40% at 80% 80%, rgba(157, 78, 221, 0.15), transparent),
            radial-gradient(ellipse 50% 30% at 20% 100%, rgba(0, 245, 212, 0.12), transparent),
            linear-gradient(180deg, #e8eef5 0%, #f0f5fa 100%)
          `,
        }}
      >
        <StarfieldCanvas />

        <div
          className="grid-bg"
          style={{
            position: 'absolute',
            inset: 0,
            backgroundImage: isDark
            ? `
              linear-gradient(rgba(0, 217, 255, 0.03) 1px, transparent 1px),
              linear-gradient(90deg, rgba(0, 217, 255, 0.03) 1px, transparent 1px)
            `
            : `
              linear-gradient(rgba(0, 217, 255, 0.06) 1px, transparent 1px),
              linear-gradient(90deg, rgba(0, 217, 255, 0.06) 1px, transparent 1px)
            `,
            backgroundSize: '60px 60px',
            animation: 'gridMove 20s linear infinite',
            zIndex: 0,
          }}
        />

        <div
          className="glow-orb glow-orb-1"
          style={{
            position: 'absolute',
            top: '15%',
            left: '10%',
            width: 400,
            height: 400,
            borderRadius: '50%',
            background: isDark ? 'radial-gradient(circle, rgba(0, 217, 255, 0.15) 0%, transparent 70%)' : 'radial-gradient(circle, rgba(0, 217, 255, 0.25) 0%, transparent 70%)',
            filter: 'blur(60px)',
            animation: 'float 8s ease-in-out infinite',
            zIndex: 0,
          }}
        />
        <div
          className="glow-orb glow-orb-2"
          style={{
            position: 'absolute',
            bottom: '20%',
            right: '5%',
            width: 500,
            height: 500,
            borderRadius: '50%',
            background: isDark ? 'radial-gradient(circle, rgba(157, 78, 221, 0.12) 0%, transparent 70%)' : 'radial-gradient(circle, rgba(157, 78, 221, 0.2) 0%, transparent 70%)',
            filter: 'blur(80px)',
            animation: 'float 10s ease-in-out infinite reverse',
            zIndex: 0,
          }}
        />
        <div
          className="glow-orb glow-orb-3"
          style={{
            position: 'absolute',
            top: '50%',
            right: '20%',
            width: 300,
            height: 300,
            borderRadius: '50%',
            background: isDark ? 'radial-gradient(circle, rgba(0, 245, 212, 0.1) 0%, transparent 70%)' : 'radial-gradient(circle, rgba(0, 245, 212, 0.18) 0%, transparent 70%)',
            filter: 'blur(50px)',
            animation: 'float 12s ease-in-out infinite',
            animationDelay: '-4s',
            zIndex: 0,
          }}
        />

        <div
          className="code-snippet code-snippet-1"
          style={{
            position: 'absolute',
            top: '25%',
            right: '15%',
            padding: '16px 20px',
            background: isDark ? 'rgba(0, 217, 255, 0.05)' : 'rgba(0, 217, 255, 0.08)',
            backdropFilter: 'blur(10px)',
            border: isDark ? '1px solid rgba(0, 217, 255, 0.1)' : '1px solid rgba(0, 217, 255, 0.2)',
            borderRadius: 12,
            fontFamily: 'monospace',
            fontSize: 13,
            color: isDark ? 'rgba(0, 217, 255, 0.7)' : 'rgba(0, 217, 255, 0.9)',
            transform: 'perspective(1000px) rotateY(-15deg) rotateX(5deg)',
            animation: 'floatCode 6s ease-in-out infinite',
            zIndex: 2,
          }}
        >
          <div>{`{`}</div>
          <div style={{ paddingLeft: 12 }}>{`"model": "claude-3",`}</div>
          <div style={{ paddingLeft: 12 }}>{`"streaming": true`}</div>
          <div>{`}`}</div>
        </div>

        <div
          className="code-snippet code-snippet-2"
          style={{
            position: 'absolute',
            bottom: '30%',
            left: '10%',
            padding: '16px 20px',
            background: isDark ? 'rgba(157, 78, 221, 0.05)' : 'rgba(157, 78, 221, 0.08)',
            backdropFilter: 'blur(10px)',
            border: isDark ? '1px solid rgba(157, 78, 221, 0.1)' : '1px solid rgba(157, 78, 221, 0.2)',
            borderRadius: 12,
            fontFamily: 'monospace',
            fontSize: 13,
            color: isDark ? 'rgba(157, 78, 221, 0.7)' : 'rgba(157, 78, 221, 0.9)',
            transform: 'perspective(1000px) rotateY(15deg) rotateX(-5deg)',
            animation: 'floatCode 8s ease-in-out infinite reverse',
            zIndex: 2,
          }}
        >
          <div style={{ color: 'rgba(255, 255, 255, 0.5)' }}>{`// MCP Server`}</div>
          <div>{`const result = await ai`}</div>
          <div style={{ paddingLeft: 12 }}>{`.generate({ prompt })`}</div>
        </div>

        {/* 右上角工具栏 */}
        <div style={{ position: 'fixed', top: 24, right: 24, display: 'flex', gap: 12, zIndex: 100, alignItems: 'center' }}>
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              borderRadius: 10,
              padding: 2,
              background: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(255, 255, 255, 0.8)',
              border: `1px solid ${isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)'}`,
              backdropFilter: 'blur(10px)',
              gap: 2,
            }}
          >
            {(['zh-CN', 'en-US'] as SupportedLanguage[]).map((code) => {
              const isActive = language === code;
              return (
                <button
                  key={code}
                  onClick={() => setLanguage(code)}
                  style={{
                    padding: '6px 14px',
                    fontSize: 13,
                    fontWeight: isActive ? 600 : 400,
                    color: isActive
                      ? '#00D9FF'
                      : (isDark ? 'rgba(255,255,255,0.45)' : 'rgba(0,0,0,0.4)'),
                    background: isActive
                      ? (isDark ? 'rgba(0,217,255,0.15)' : 'rgba(0,217,255,0.1)')
                      : 'transparent',
                    border: isActive
                      ? `1px solid ${isDark ? 'rgba(0,217,255,0.25)' : 'rgba(0,217,255,0.2)'}`
                      : '1px solid transparent',
                    cursor: 'pointer',
                    borderRadius: 8,
                    transition: 'all 0.25s ease',
                    whiteSpace: 'nowrap' as const,
                    lineHeight: '1.4',
                  }}
                >
                  {code === 'zh-CN' ? '中文' : 'EN'}
                </button>
              );
            })}
          </div>
          <Button
            type="text"
            shape="circle"
            icon={isDark ? <SunOutlined /> : <MoonOutlined />}
            onClick={toggleTheme}
            style={{
              width: 40,
              height: 40,
              color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)',
              background: isDark ? 'rgba(255, 255, 255, 0.06)' : 'rgba(255, 255, 255, 0.8)',
              border: isDark ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.08)',
              backdropFilter: 'blur(10px)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          />
        </div>

        <div className="relative z-10 text-center px-4" style={{ maxWidth: 900 }}>
          <div
            className="hero-badge"
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 8,
              padding: '8px 16px',
              background: 'rgba(0, 217, 255, 0.1)',
              border: '1px solid rgba(0, 217, 255, 0.2)',
              borderRadius: 50,
              fontSize: 14,
              color: '#00D9FF',
              marginBottom: 32,
              boxShadow: isDark ? 'none' : '0 4px 20px rgba(0, 217, 255, 0.15)',
              animation: 'fadeInDown 0.8s ease-out',
            }}
          >
            <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#00D9FF', animation: 'pulse 2s infinite' }} />
            {t('home.badge', { version: __APP_VERSION__ })}
          </div>

          <h1
            className="hero-title"
            style={{
              fontSize: 'clamp(4rem, 12vw, 7rem)',
              fontWeight: 900,
              margin: '0 0 24px',
              letterSpacing: -4,
              lineHeight: 1,
              animation: 'fadeInUp 1s ease-out 0.2s both',
            }}
          >
            <span
              style={{
                background: 'linear-gradient(135deg, #fff 0%, #00D9FF 30%, #9D4EDD 60%, #00F5D4 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
                backgroundClip: 'text',
                backgroundSize: '200% 200%',
                animation: 'gradientShift 8s ease infinite',
              }}
            >
              CodeMind
            </span>
          </h1>

          <p
            style={{
              fontSize: 'clamp(1.25rem, 3vw, 1.75rem)',
              color: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)',
              marginBottom: 16,
              fontWeight: 400,
              letterSpacing: language === 'zh-CN' ? 8 : 2,
              textTransform: language === 'en-US' ? 'uppercase' : 'none',
              animation: 'fadeInUp 1s ease-out 0.4s both',
            }}
          >
            {t('home.subtitle')}
          </p>

          <p
            style={{
              fontSize: 'clamp(1rem, 2vw, 1.15rem)',
              color: isDark ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.5)',
              marginBottom: 48,
              maxWidth: 560,
              marginInline: 'auto',
              lineHeight: 1.8,
              fontWeight: 300,
              animation: 'fadeInUp 1s ease-out 0.6s both',
            }}
          >
            {t('home.description')}
          </p>

          <div
            style={{
              display: 'flex',
              gap: 20,
              justifyContent: 'center',
              flexWrap: 'wrap',
              animation: 'fadeInUp 1s ease-out 0.8s both',
            }}
          >
            <Button
              type="primary"
              size="large"
              onClick={() => navigate(isAuthenticated ? '/dashboard' : '/login')}
              className="cta-primary"
              style={{
                height: 56,
                paddingInline: 40,
                borderRadius: 28,
                fontSize: 16,
                fontWeight: 600,
                background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                border: 'none',
                position: 'relative',
                overflow: 'hidden',
              }}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <PlayCircleOutlined />
                {isAuthenticated ? t('home.cta.enterConsole') : t('home.cta.tryNow')}
              </span>
            </Button>
            <Button
              size="large"
              ghost
              onClick={() => featuresRef.current?.scrollIntoView({ behavior: 'smooth' })}
              className="cta-secondary"
              style={{
                height: 56,
                paddingInline: 40,
                borderRadius: 28,
                fontSize: 16,
                color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.8)',
                borderColor: isDark ? 'rgba(255, 255, 255, 0.25)' : 'rgba(0, 0, 0, 0.15)',
                background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                backdropFilter: 'blur(10px)',
              }}
            >
              {t('home.cta.exploreFeatures')}
            </Button>
          </div>


        </div>

        <div
          className="scroll-indicator"
          style={{
            position: 'absolute',
            bottom: 40,
            left: '50%',
            transform: 'translateX(-50%)',
            zIndex: 10,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 8,
            color: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.4)',
            fontSize: 12,
            animation: 'fadeIn 1s ease-out 1.5s both',
          }}
        >
          <span>{t('home.scrollDown')}</span>
          <ArrowDownOutlined style={{ animation: 'bounce 2s infinite' }} />
        </div>
      </section>

      <section
        style={{
          padding: '80px 24px',
          background: isDark ? 'linear-gradient(180deg, #050d14 0%, #0a1628 100%)' : 'linear-gradient(180deg, #f0f5fa 0%, #e8eef5 100%)',
          position: 'relative',
        }}
      >
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: 24,
            }}
          >
            {stats.map((stat, i) => (
              <div
                key={stat.label}
                className="stat-card"
                style={{
                  background: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.8)',
                  border: isDark ? '1px solid rgba(255, 255, 255, 0.06)' : '1px solid rgba(0, 0, 0, 0.06)',
                  borderRadius: 24,
                  padding: '32px 24px',
                  textAlign: 'center',
                  opacity: 0,
                  transform: 'translateY(30px)',
                  transition: `all 0.6s cubic-bezier(0.16, 1, 0.3, 1) ${i * 0.1}s`,
                }}
              >
                <div
                  style={{
                    width: 56,
                    height: 56,
                    borderRadius: 16,
                    background: 'linear-gradient(135deg, rgba(0, 217, 255, 0.1), rgba(157, 78, 221, 0.1))',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 24,
                    color: '#00D9FF',
                    margin: '0 auto 16px',
                  }}
                >
                  {stat.icon}
                </div>
                <AnimatedCounter value={stat.value} suffix={stat.suffix} />
                <div style={{ color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 8 }}>
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section
        ref={featuresRef}
        style={{
          padding: '120px 24px',
          background: isDark
            ? `
            radial-gradient(ellipse 50% 30% at 50% 0%, rgba(0, 217, 255, 0.05), transparent),
            #0a1628
          `
            : `
            radial-gradient(ellipse 50% 30% at 50% 0%, rgba(0, 217, 255, 0.08), transparent),
            #e8eef5
          `,
          position: 'relative',
        }}
      >
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <div style={{ textAlign: 'center', marginBottom: 72 }}>
            <div
              style={{
                display: 'inline-block',
                padding: '6px 16px',
                background: 'rgba(157, 78, 221, 0.1)',
                border: '1px solid rgba(157, 78, 221, 0.2)',
                borderRadius: 50,
                fontSize: 13,
                color: '#9D4EDD',
                marginBottom: 20,
                boxShadow: isDark ? 'none' : '0 4px 20px rgba(157, 78, 221, 0.15)',
              }}
            >
              {t('home.features.sectionBadge')}
            </div>
            <h2
              style={{
                fontSize: 'clamp(2rem, 4vw, 3rem)',
                fontWeight: 800,
                color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                marginBottom: 16,
                letterSpacing: -1,
              }}
            >
              {t('home.features.sectionTitle')}
            </h2>
            <p
              style={{
                color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.5)',
                fontSize: 17,
                maxWidth: 500,
                margin: '0 auto',
              }}
            >
              {t('home.features.sectionDesc')}
            </p>
          </div>

          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(340px, 1fr))',
              gap: 24,
            }}
          >
            {features.map((f, i) => (
              <div
                key={f.title}
                className="feature-card"
                style={{
                  background: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.7)',
                  border: isDark ? '1px solid rgba(255, 255, 255, 0.06)' : '1px solid rgba(0, 0, 0, 0.06)',
                  borderRadius: 24,
                  padding: 32,
                  position: 'relative',
                  overflow: 'hidden',
                  opacity: 0,
                  transform: 'translateY(30px)',
                  transition: `all 0.6s cubic-bezier(0.16, 1, 0.3, 1) ${i * 0.08}s`,
                  cursor: 'pointer',
                }}
              >
                <div
                  className="card-glow"
                  style={{
                    position: 'absolute',
                    inset: 0,
                    background: `radial-gradient(circle at 50% 0%, ${f.color}15, transparent 60%)`,
                    opacity: 0,
                    transition: 'opacity 0.3s',
                  }}
                />
                
                <div style={{ position: 'relative', zIndex: 1 }}>
                  <div
                    style={{
                      width: 60,
                      height: 60,
                      borderRadius: 16,
                      background: `linear-gradient(135deg, ${f.color}20, ${f.color}05)`,
                      border: `1px solid ${f.color}30`,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: 28,
                      color: f.color,
                      marginBottom: 20,
                      transition: 'transform 0.3s',
                    }}
                    className="feature-icon"
                  >
                    {f.icon}
                  </div>
                  <h3
                    style={{
                      fontSize: 19,
                      fontWeight: 600,
                      color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
                      marginBottom: 10,
                    }}
                  >
                    {f.title}
                  </h3>
                  <p
                    style={{
                      color: isDark ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.55)',
                      fontSize: 15,
                      lineHeight: 1.7,
                      margin: 0,
                    }}
                  >
                    {f.desc}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section
        style={{
          padding: '100px 24px',
          background: isDark ? '#050d14' : '#f0f5fa',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: 1,
            background: 'linear-gradient(90deg, transparent, rgba(0, 217, 255, 0.3), transparent)',
          }}
        />

        <div style={{ maxWidth: 1200, margin: '0 auto', textAlign: 'center' }}>
          <h3
            style={{
              fontSize: 14,
              fontWeight: 500,
              color: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.4)',
              textTransform: 'uppercase',
              letterSpacing: 4,
              marginBottom: 40,
            }}
          >
            {t('home.tools.title')}
          </h3>

          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'center',
              gap: 16,
            }}
          >
            {TOOLS.map((tool, i) => (
              <div
                key={tool}
                className="tool-tag"
                style={{
                  padding: '12px 24px',
                  background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  border: isDark ? '1px solid rgba(255, 255, 255, 0.08)' : '1px solid rgba(0, 0, 0, 0.08)',
                  borderRadius: 50,
                  fontSize: 15,
                  color: isDark ? 'rgba(255, 255, 255, 0.6)' : 'rgba(0, 0, 0, 0.6)',
                  opacity: 0,
                  transform: 'scale(0.9)',
                  transition: `all 0.4s cubic-bezier(0.16, 1, 0.3, 1) ${i * 0.05}s`,
                  cursor: 'default',
                }}
              >
                {tool}
              </div>
            ))}
          </div>
        </div>
      </section>

      <section
        style={{
          padding: '120px 24px',
          background: isDark
            ? `
            radial-gradient(ellipse 80% 50% at 50% 100%, rgba(0, 217, 255, 0.1), transparent),
            linear-gradient(180deg, #050d14 0%, #0a1628 100%)
          `
            : `
            radial-gradient(ellipse 80% 50% at 50% 100%, rgba(0, 217, 255, 0.15), transparent),
            linear-gradient(180deg, #f0f5fa 0%, #e8eef5 100%)
          `,
          textAlign: 'center',
        }}
      >
        <div style={{ maxWidth: 600, margin: '0 auto' }}>
          <h2
            style={{
              fontSize: 'clamp(2rem, 4vw, 2.75rem)',
              fontWeight: 800,
              color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
              marginBottom: 20,
              letterSpacing: -1,
            }}
          >
            {t('home.ctaSection.title')}
          </h2>
          <p
            style={{
              color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.5)',
              fontSize: 17,
              marginBottom: 40,
              lineHeight: 1.7,
            }}
          >
            {t('home.ctaSection.desc')}
          </p>
          <Button
            type="primary"
            size="large"
            onClick={() => navigate(isAuthenticated ? '/dashboard' : '/login')}
            style={{
              height: 56,
              paddingInline: 48,
              borderRadius: 28,
              fontSize: 16,
              fontWeight: 600,
              background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
              border: 'none',
            }}
          >
            {isAuthenticated ? t('home.cta.enterConsole') : t('home.cta.startFree')}
          </Button>
        </div>
      </section>

      <footer
        style={{
          padding: '48px 24px',
          textAlign: 'center',
          background: isDark ? '#050d14' : '#f0f5fa',
          borderTop: isDark ? '1px solid rgba(255, 255, 255, 0.05)' : '1px solid rgba(0, 0, 0, 0.05)',
        }}
      >
        <div
          style={{
            fontSize: 24,
            fontWeight: 800,
            background: 'linear-gradient(135deg, #fff 0%, #00D9FF 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            marginBottom: 16,
          }}
        >
          CodeMind
        </div>
        <p style={{ fontSize: 14, color: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.4)', margin: 0 }}>
          CodeMind v{__APP_VERSION__}
        </p>
        <p style={{ fontSize: 13, color: isDark ? 'rgba(255, 255, 255, 0.2)' : 'rgba(0, 0, 0, 0.35)', marginTop: 8 }}>
          {t('home.footer.copyright', { year: new Date().getFullYear() })}
        </p>
      </footer>

    </div>
  );
};

export default HomePage;
