import { useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from 'antd';
import {
  ApiOutlined,
  RobotOutlined,
  GatewayOutlined,
  ThunderboltOutlined,
  SafetyCertificateOutlined,
  BarChartOutlined,
  ArrowDownOutlined,
} from '@ant-design/icons';
import useAuthStore from '@/store/authStore';

/** 功能特点卡片数据 */
const features = [
  {
    icon: <ApiOutlined />,
    title: 'OpenAI 兼容',
    desc: '完整兼容 OpenAI API 格式，无缝对接现有开发工具与工作流',
    gradient: 'var(--gradient-primary)',
  },
  {
    icon: <RobotOutlined />,
    title: 'Anthropic 原生',
    desc: '原生支持 Anthropic 模型，提供企业级 Claude 接入能力',
    gradient: 'var(--gradient-primary)',
  },
  {
    icon: <GatewayOutlined />,
    title: 'MCP 网关',
    desc: '集成 MCP 协议网关，打通 AI 与外部工具的无缝协作',
    gradient: 'var(--gradient-primary)',
  },
  {
    icon: <ThunderboltOutlined />,
    title: 'SSE 流式',
    desc: '支持 Server-Sent Events 流式响应，实时输出模型生成内容',
    gradient: 'var(--gradient-primary)',
  },
  {
    icon: <SafetyCertificateOutlined />,
    title: '三级权限控制',
    desc: 'Super Admin / Dept Manager / User 三层 RBAC，精细化管理',
    gradient: 'var(--gradient-primary)',
  },
  {
    icon: <BarChartOutlined />,
    title: '实时用量计量',
    desc: '实时 Token 计量与用量统计，清晰的可视化报表与趋势分析',
    gradient: 'var(--gradient-primary)',
  },
];

/** 数字亮点数据 */
const highlights = [
  { value: 'OpenAI', label: '兼容 API 格式' },
  { value: 'Anthropic', label: '原生支持' },
  { value: 'MCP', label: '网关协议' },
  { value: 'SSE', label: '流式响应' },
  { value: '3 级', label: '权限控制' },
  { value: '实时', label: 'Token 计量' },
];

/** 粒子/连线背景动效 — 增强版 */
const ParticleCanvas: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const initParticles = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animId: number;
    const dpr = window.devicePixelRatio || 1;

    const resize = () => {
      canvas.width = window.innerWidth * dpr;
      canvas.height = window.innerHeight * dpr;
      canvas.style.width = `${window.innerWidth}px`;
      canvas.style.height = `${window.innerHeight}px`;
      ctx.scale(dpr, dpr);
    };
    resize();
    window.addEventListener('resize', resize);

    // 粒子数量随屏幕尺寸自适应
    const count = Math.min(
      Math.floor((window.innerWidth * window.innerHeight) / 10000),
      120
    );
    const particles: {
      x: number;
      y: number;
      vx: number;
      vy: number;
      r: number;
      a: number;
    }[] = [];

    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * window.innerWidth,
        y: Math.random() * window.innerHeight,
        vx: (Math.random() - 0.5) * 0.4,
        vy: (Math.random() - 0.5) * 0.4,
        r: Math.random() * 2.5 + 1,
        a: Math.random() * 0.4 + 0.25,
      });
    }

    const maxDist = 180;

    const draw = () => {
      ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);

      // 更新粒子位置
      for (const p of particles) {
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < 0 || p.x > window.innerWidth) p.vx *= -1;
        if (p.y < 0 || p.y > window.innerHeight) p.vy *= -1;
      }

      // 绘制连线
      for (let i = 0; i < particles.length; i++) {
        const pi = particles[i]!;
        for (let j = i + 1; j < particles.length; j++) {
          const pj = particles[j]!;
          const dx = pi.x - pj.x;
          const dy = pi.y - pj.y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < maxDist) {
            const alpha = (1 - dist / maxDist) * 0.2;
            ctx.beginPath();
            ctx.strokeStyle = `rgba(255, 255, 255, ${alpha})`;
            ctx.lineWidth = 0.6;
            ctx.moveTo(pi.x, pi.y);
            ctx.lineTo(pj.x, pj.y);
            ctx.stroke();
          }
        }
      }

      // 绘制粒子（带光晕感）
      for (const p of particles) {
        const gradient = ctx.createRadialGradient(
          p.x,
          p.y,
          0,
          p.x,
          p.y,
          p.r * 2
        );
        gradient.addColorStop(0, `rgba(255, 255, 255, ${p.a})`);
        gradient.addColorStop(1, `rgba(255, 255, 255, 0)`);
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r * 2, 0, Math.PI * 2);
        ctx.fillStyle = gradient;
        ctx.fill();
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(255, 255, 255, ${p.a + 0.2})`;
        ctx.fill();
      }

      animId = requestAnimationFrame(draw);
    };

    draw();

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', resize);
    };
  }, []);

  useEffect(() => {
    const cleanup = initParticles();
    return cleanup;
  }, [initParticles]);

  return (
    <canvas
      ref={canvasRef}
      style={{ position: 'absolute', inset: 0, zIndex: 1, pointerEvents: 'none' }}
    />
  );
};

/** 首页 */
const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const featuresRef = useRef<HTMLDivElement>(null);

  // 滚动动画 — IntersectionObserver
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

    const cards = document.querySelectorAll('.feature-card');
    cards.forEach((card) => observer.observe(card));

    return () => observer.disconnect();
  }, []);

  return (
    <div className="min-h-screen home-page-glass" style={{ overflow: 'hidden' }}>
      {/* ═══ Hero 区域 ═══ */}
      <section
        className="hero-section"
        style={{
          position: 'relative',
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'linear-gradient(135deg, #0a1628 0%, #152a47 35%, #1e4a6e 65%, #2B7CB3 100%)',
        }}
      >
        {/* 粒子动效背景 */}
        <ParticleCanvas />

        {/* 径向光晕装饰 */}
        <div
          className="hero-glow"
          style={{
            position: 'absolute',
            top: '25%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            width: '70vw',
            height: '70vw',
            borderRadius: '50%',
            background: 'var(--orb-primary)',
            zIndex: 1,
            pointerEvents: 'none',
          }}
        />
        <div
          style={{
            position: 'absolute',
            bottom: '20%',
            right: '10%',
            width: '40vw',
            height: '40vw',
            borderRadius: '50%',
            background: 'var(--orb-accent)',
            zIndex: 1,
            pointerEvents: 'none',
          }}
        />

        {/* 浮动玻璃球体装饰 */}
        <div
          className="glass-orb glass-orb-1 animate-float"
          style={{
            position: 'absolute',
            top: '20%',
            left: '15%',
            width: 120,
            height: 120,
            borderRadius: '50%',
            background: 'rgba(255, 255, 255, 0.06)',
            backdropFilter: 'blur(20px)',
            WebkitBackdropFilter: 'blur(20px)',
            border: '1px solid rgba(255, 255, 255, 0.08)',
            zIndex: 2,
          }}
        />
        <div
          className="glass-orb glass-orb-2 animate-float"
          style={{
            position: 'absolute',
            top: '60%',
            right: '20%',
            width: 80,
            height: 80,
            borderRadius: '50%',
            background: 'rgba(255, 255, 255, 0.05)',
            backdropFilter: 'blur(16px)',
            WebkitBackdropFilter: 'blur(16px)',
            border: '1px solid rgba(255, 255, 255, 0.06)',
            zIndex: 2,
            animationDelay: '-2s',
          }}
        />
        <div
          className="glass-orb glass-orb-3 animate-float"
          style={{
            position: 'absolute',
            bottom: '25%',
            left: '25%',
            width: 60,
            height: 60,
            borderRadius: '50%',
            background: 'rgba(255, 255, 255, 0.04)',
            backdropFilter: 'blur(12px)',
            WebkitBackdropFilter: 'blur(12px)',
            border: '1px solid rgba(255, 255, 255, 0.05)',
            zIndex: 2,
            animationDelay: '-4s',
          }}
        />

        {/* 主内容 */}
        <div className="relative z-10 text-center px-4" style={{ maxWidth: 800 }}>
          <h1
            className="hero-title"
            style={{
              fontSize: 'clamp(3rem, 8vw, 5rem)',
              fontWeight: 800,
              margin: '0 0 16px',
              letterSpacing: -2,
              lineHeight: 1.05,
              background: 'linear-gradient(135deg, #fff 0%, #e8f4fc 50%, #6BC5E8 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text',
            }}
          >
            CodeMind
          </h1>

          <p
            style={{
              fontSize: 'clamp(1.1rem, 2.8vw, 1.5rem)',
              color: 'rgba(255, 255, 255, 0.9)',
              marginBottom: 12,
              fontWeight: 500,
              letterSpacing: 4,
            }}
          >
            企业级 AI 编码平台
          </p>

          <p
            style={{
              fontSize: 'clamp(0.9rem, 2vw, 1.1rem)',
              color: 'rgba(255, 255, 255, 0.5)',
              marginBottom: 40,
              maxWidth: 520,
              marginInline: 'auto',
              lineHeight: 1.7,
              fontWeight: 300,
            }}
          >
            基于本地大语言模型，为开发团队提供安全、高效、可控的智能编码辅助服务
          </p>

          <div
            style={{
              display: 'flex',
              gap: 16,
              justifyContent: 'center',
              flexWrap: 'wrap',
            }}
          >
            <Button
              type="primary"
              size="large"
              onClick={() => navigate(isAuthenticated ? '/dashboard' : '/login')}
              style={{
                height: 52,
                paddingInline: 36,
                borderRadius: 26,
                fontSize: 16,
                background: 'var(--gradient-primary)',
                color: '#fff',
                border: 'none',
                fontWeight: 600,
                boxShadow: '0 8px 32px rgba(43, 124, 179, 0.4)',
              }}
            >
              {isAuthenticated ? '进入控制台' : '立即登录'}
            </Button>
            <Button
              size="large"
              ghost
              onClick={() =>
                featuresRef.current?.scrollIntoView({ behavior: 'smooth' })
              }
              style={{
                height: 52,
                paddingInline: 36,
                borderRadius: 26,
                fontSize: 16,
                color: '#fff',
                borderColor: 'rgba(255, 255, 255, 0.35)',
                background: 'rgba(255, 255, 255, 0.05)',
              }}
            >
              了解更多
            </Button>
          </div>
        </div>

        {/* 底部滚动提示 */}
        <div
          style={{
            position: 'absolute',
            bottom: 32,
            left: '50%',
            transform: 'translateX(-50%)',
            zIndex: 10,
            animation: 'hero-scroll-float 2s ease-in-out infinite',
          }}
        >
          <ArrowDownOutlined
            style={{ color: 'rgba(255, 255, 255, 0.5)', fontSize: 22 }}
          />
        </div>
      </section>

      {/* ═══ 功能介绍区域 ═══ */}
      <section
        ref={featuresRef}
        style={{
          padding: 'clamp(64px, 10vw, 120px) clamp(16px, 4vw, 48px)',
          background: 'linear-gradient(180deg, #0a1628 0%, #152a47 50%, #0f1e32 100%)',
        }}
      >
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          <h2
            style={{
              fontSize: 'clamp(1.75rem, 3.5vw, 2.5rem)',
              fontWeight: 700,
              textAlign: 'center',
              color: 'rgba(255, 255, 255, 0.95)',
              marginBottom: 12,
            }}
          >
            平台能力
          </h2>
          <p
            style={{
              textAlign: 'center',
              color: 'rgba(255, 255, 255, 0.5)',
              marginBottom: 'clamp(40px, 6vw, 72px)',
              fontSize: 16,
            }}
          >
            全方位的 AI 编码服务管理能力
          </p>

          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
              gap: 24,
            }}
          >
            {features.map((f, i) => (
              <div
                key={f.title}
                className="feature-card"
                style={{
                  background: 'rgba(255, 255, 255, 0.08)',
                  backdropFilter: 'blur(16px)',
                  WebkitBackdropFilter: 'blur(16px)',
                  border: '1px solid rgba(255, 255, 255, 0.1)',
                  borderRadius: 20,
                  padding: 32,
                  opacity: 0,
                  transform: 'translateY(30px)',
                  transition: `all 0.6s cubic-bezier(0.16, 1, 0.3, 1) ${i * 0.08}s`,
                }}
              >
                <div
                  style={{
                    width: 56,
                    height: 56,
                    borderRadius: 16,
                    background: f.gradient,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: 26,
                    color: '#fff',
                    marginBottom: 20,
                    boxShadow: '0 8px 24px rgba(43, 124, 179, 0.25)',
                  }}
                >
                  {f.icon}
                </div>
                <h3
                  style={{
                    fontSize: 18,
                    fontWeight: 600,
                    color: 'rgba(255, 255, 255, 0.95)',
                    marginBottom: 10,
                  }}
                >
                  {f.title}
                </h3>
                <p
                  style={{
                    color: 'rgba(255, 255, 255, 0.6)',
                    fontSize: 15,
                    lineHeight: 1.7,
                    margin: 0,
                  }}
                >
                  {f.desc}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 数字亮点（玻璃卡片） ═══ */}
      <section
        style={{
          padding: 'clamp(48px, 8vw, 96px) clamp(16px, 4vw, 48px)',
          background: 'linear-gradient(180deg, #0f1e32 0%, #0a1628 100%)',
        }}
      >
        <div
          style={{
            maxWidth: 1100,
            margin: '0 auto',
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
            gap: 20,
          }}
        >
          {highlights.map((item) => (
            <div
              key={item.label}
              style={{
                background: 'rgba(255, 255, 255, 0.08)',
                backdropFilter: 'blur(16px)',
                WebkitBackdropFilter: 'blur(16px)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: 20,
                padding: '28px 20px',
                textAlign: 'center',
                transition: 'all 0.3s ease',
              }}
            >
              <div
                style={{
                  fontSize: 'clamp(1.25rem, 2.5vw, 2rem)',
                  fontWeight: 800,
                  color: '#fff',
                  background: 'var(--gradient-primary)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                }}
              >
                {item.value}
              </div>
              <div
                style={{
                  color: 'rgba(255, 255, 255, 0.55)',
                  fontSize: 14,
                  marginTop: 8,
                }}
              >
                {item.label}
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* ═══ 底部 ═══ */}
      <footer
        style={{
          padding: '36px 16px',
          textAlign: 'center',
          background: '#050d14',
          color: 'rgba(255, 255, 255, 0.4)',
          borderTop: '1px solid rgba(255, 255, 255, 0.06)',
        }}
      >
        <p style={{ fontSize: 14, margin: 0 }}>度影智能编码服务 CodeMind v0.2.0</p>
        <p style={{ fontSize: 12, marginTop: 8 }}>
          © {new Date().getFullYear()} RayShape. All Rights Reserved.
        </p>
      </footer>

      {/* 内联动画样式 */}
      <style>{`
        @keyframes hero-scroll-float {
          0%, 100% { transform: translateX(-50%) translateY(0); }
          50% { transform: translateX(-50%) translateY(10px); }
        }
        .feature-card.animate-in {
          opacity: 1 !important;
          transform: translateY(0) !important;
        }
        .glass-orb:hover {
          background: rgba(255, 255, 255, 0.1) !important;
        }
      `}</style>
    </div>
  );
};

export default HomePage;
