import { useEffect, useRef, useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from 'antd';
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
} from '@ant-design/icons';
import useAuthStore from '@/store/authStore';

/** 功能特点卡片数据 */
const features = [
  {
    icon: <ApiOutlined />,
    title: 'OpenAI 兼容',
    desc: '完整兼容 OpenAI API 格式，无缝对接主流开发工具',
    color: '#00D9FF',
  },
  {
    icon: <RobotOutlined />,
    title: 'Anthropic 原生',
    desc: '原生支持 Anthropic 模型协议，提供Claude Code接入能力',
    color: '#FF6B6B',
  },
  {
    icon: <GatewayOutlined />,
    title: 'MCP 网关',
    desc: '集成 MCP 协议网关，打通 AI 与外部工具的无缝协作',
    color: '#9D4EDD',
  },
  {
    icon: <CloudServerOutlined />,
    title: '多节点负载均衡',
    desc: '支持多 LLM 后端节点配置，智能权重分配与并发控制',
    color: '#00F5D4',
  },
  {
    icon: <SafetyCertificateOutlined />,
    title: '三级权限控制',
    desc: 'Super Admin / Dept Manager / User 三层 RBAC 精细化管控',
    color: '#FFBE0B',
  },
  {
    icon: <BarChartOutlined />,
    title: '实时用量计量',
    desc: '多维度 Token 用量统计，支持小时/周/月级限额控制',
    color: '#FB5607',
  },
  {
    icon: <CodeOutlined />,
    title: '15+ IDE 生态',
    desc: '提供 Claude Code、Cursor、VS Code、JetBrains 等接入文档',
    color: '#3A86FF',
  },
  {
    icon: <ThunderboltOutlined />,
    title: 'SSE 流式响应',
    desc: '支持 Server-Sent Events 流式输出，实时呈现生成内容',
    color: '#06FFA5',
  },
  {
    icon: <TeamOutlined />,
    title: '私有化部署',
    desc: '基于本地大语言模型，数据不出域，满足企业安全合规',
    color: '#E85D04',
  },
];

/** 统计数据 */
const stats = [
  { value: 15, suffix: '+', label: 'IDE 工具支持', icon: <CodeOutlined /> },
  { value: 3, suffix: '级', label: '权限层级', icon: <SafetyCertificateOutlined /> },
  { value: 99, suffix: '%', label: '服务可用性', icon: <CloudServerOutlined /> },
  { value: 0, suffix: '.1s', label: '响应延迟', icon: <ThunderboltOutlined /> },
];

/** 支持的 IDE/工具 */
const tools = [
  'Claude Code', 'Cursor', 'VS Code', 'JetBrains', 
  'Cline', 'Roo Code', 'Kilo Code', 'TRAE', 
  'OpenCode', 'Factory Droid', 'Cherry Studio', 'Goose'
];

/** 星空连线粒子动效 - 增强版 */
const StarfieldCanvas: React.FC = () => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const mouseRef = useRef({ x: 0, y: 0 });

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

    const count = Math.min(Math.floor((window.innerWidth * window.innerHeight) / 8000), 150);
    const particles: { x: number; y: number; vx: number; vy: number; r: number; a: number; pulse: number }[] = [];

    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * window.innerWidth,
        y: Math.random() * window.innerHeight,
        vx: (Math.random() - 0.5) * 0.3,
        vy: (Math.random() - 0.5) * 0.3,
        r: Math.random() * 2 + 0.5,
        a: Math.random() * 0.5 + 0.3,
        pulse: Math.random() * Math.PI * 2,
      });
    }

    const maxDist = 200;

    const draw = () => {
      ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);

      // 更新粒子
      for (const p of particles) {
        p.x += p.vx;
        p.y += p.vy;
        p.pulse += 0.02;
        
        if (p.x < 0 || p.x > window.innerWidth) p.vx *= -1;
        if (p.y < 0 || p.y > window.innerHeight) p.vy *= -1;
      }

      // 绘制连线 - 带渐变效果
      for (let i = 0; i < particles.length; i++) {
        const pi = particles[i]!;
        
        // 鼠标交互连线
        const dx = mouseRef.current.x - pi.x;
        const dy = mouseRef.current.y - pi.y;
        const mouseDist = Math.sqrt(dx * dx + dy * dy);
        if (mouseDist < 250) {
          const alpha = (1 - mouseDist / 250) * 0.4;
          const gradient = ctx.createLinearGradient(pi.x, pi.y, mouseRef.current.x, mouseRef.current.y);
          gradient.addColorStop(0, `rgba(0, 217, 255, ${alpha})`);
          gradient.addColorStop(1, `rgba(157, 78, 221, ${alpha})`);
          ctx.beginPath();
          ctx.strokeStyle = gradient;
          ctx.lineWidth = 0.8;
          ctx.moveTo(pi.x, pi.y);
          ctx.lineTo(mouseRef.current.x, mouseRef.current.y);
          ctx.stroke();
        }

        for (let j = i + 1; j < particles.length; j++) {
          const pj = particles[j]!;
          const dx = pi.x - pj.x;
          const dy = pi.y - pj.y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < maxDist) {
            const alpha = (1 - dist / maxDist) * 0.25;
            ctx.beginPath();
            ctx.strokeStyle = `rgba(100, 200, 255, ${alpha})`;
            ctx.lineWidth = 0.5;
            ctx.moveTo(pi.x, pi.y);
            ctx.lineTo(pj.x, pj.y);
            ctx.stroke();
          }
        }
      }

      // 绘制粒子 - 带脉动效果
      for (const p of particles) {
        const pulseFactor = 1 + Math.sin(p.pulse) * 0.2;
        const gradient = ctx.createRadialGradient(p.x, p.y, 0, p.x, p.y, p.r * 3 * pulseFactor);
        gradient.addColorStop(0, `rgba(0, 217, 255, ${p.a})`);
        gradient.addColorStop(0.5, `rgba(157, 78, 221, ${p.a * 0.5})`);
        gradient.addColorStop(1, 'rgba(0, 0, 0, 0)');
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r * 3 * pulseFactor, 0, Math.PI * 2);
        ctx.fillStyle = gradient;
        ctx.fill();
        
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(255, 255, 255, ${p.a + 0.3})`;
        ctx.fill();
      }

      animId = requestAnimationFrame(draw);
    };

    draw();

    const handleMouseMove = (e: MouseEvent) => {
      mouseRef.current = { x: e.clientX, y: e.clientY };
    };
    window.addEventListener('mousemove', handleMouseMove);

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', resize);
      window.removeEventListener('mousemove', handleMouseMove);
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

/** 动态数字计数器 */
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

/** 首页 */
const HomePage: React.FC = () => {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const featuresRef = useRef<HTMLDivElement>(null);
  const [scrolled, setScrolled] = useState(false);

  // 滚动监听
  useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 50);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  // 滚动动画
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
    <div className="min-h-screen" style={{ overflow: 'hidden', background: '#050d14' }}>
      {/* ═══ Hero 区域 - 全新设计 ═══ */}
      <section
        className="hero-section"
        style={{
          position: 'relative',
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: `
            radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 217, 255, 0.15), transparent),
            radial-gradient(ellipse 60% 40% at 80% 80%, rgba(157, 78, 221, 0.1), transparent),
            radial-gradient(ellipse 50% 30% at 20% 100%, rgba(0, 245, 212, 0.08), transparent),
            linear-gradient(180deg, #0a1628 0%, #050d14 100%)
          `,
        }}
      >
        {/* 星空粒子背景 */}
        <StarfieldCanvas />

        {/* 动态网格背景 */}
        <div
          className="grid-bg"
          style={{
            position: 'absolute',
            inset: 0,
            backgroundImage: `
              linear-gradient(rgba(0, 217, 255, 0.03) 1px, transparent 1px),
              linear-gradient(90deg, rgba(0, 217, 255, 0.03) 1px, transparent 1px)
            `,
            backgroundSize: '60px 60px',
            animation: 'gridMove 20s linear infinite',
            zIndex: 0,
          }}
        />

        {/* 发光球体装饰 */}
        <div
          className="glow-orb glow-orb-1"
          style={{
            position: 'absolute',
            top: '15%',
            left: '10%',
            width: 400,
            height: 400,
            borderRadius: '50%',
            background: 'radial-gradient(circle, rgba(0, 217, 255, 0.15) 0%, transparent 70%)',
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
            background: 'radial-gradient(circle, rgba(157, 78, 221, 0.12) 0%, transparent 70%)',
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
            background: 'radial-gradient(circle, rgba(0, 245, 212, 0.1) 0%, transparent 70%)',
            filter: 'blur(50px)',
            animation: 'float 12s ease-in-out infinite',
            animationDelay: '-4s',
            zIndex: 0,
          }}
        />

        {/* 浮动代码片段装饰 */}
        <div
          className="code-snippet code-snippet-1"
          style={{
            position: 'absolute',
            top: '25%',
            right: '15%',
            padding: '16px 20px',
            background: 'rgba(0, 217, 255, 0.05)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(0, 217, 255, 0.1)',
            borderRadius: 12,
            fontFamily: 'monospace',
            fontSize: 13,
            color: 'rgba(0, 217, 255, 0.7)',
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
            background: 'rgba(157, 78, 221, 0.05)',
            backdropFilter: 'blur(10px)',
            border: '1px solid rgba(157, 78, 221, 0.1)',
            borderRadius: 12,
            fontFamily: 'monospace',
            fontSize: 13,
            color: 'rgba(157, 78, 221, 0.7)',
            transform: 'perspective(1000px) rotateY(15deg) rotateX(-5deg)',
            animation: 'floatCode 8s ease-in-out infinite reverse',
            zIndex: 2,
          }}
        >
          <div style={{ color: 'rgba(255, 255, 255, 0.5)' }}>{`// MCP Server`}</div>
          <div>{`const result = await ai`}</div>
          <div style={{ paddingLeft: 12 }}>{`.generate({ prompt })`}</div>
        </div>

        {/* 主内容 */}
        <div className="relative z-10 text-center px-4" style={{ maxWidth: 900 }}>
          {/* Badge */}
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
              animation: 'fadeInDown 0.8s ease-out',
            }}
          >
            <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#00D9FF', animation: 'pulse 2s infinite' }} />
            企业级 AI 编码平台 v0.2.0
          </div>

          {/* 主标题 */}
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

          {/* 副标题 */}
          <p
            style={{
              fontSize: 'clamp(1.25rem, 3vw, 1.75rem)',
              color: 'rgba(255, 255, 255, 0.7)',
              marginBottom: 16,
              fontWeight: 400,
              letterSpacing: 8,
              textTransform: 'uppercase',
              animation: 'fadeInUp 1s ease-out 0.4s both',
            }}
          >
            智能编码新纪元
          </p>

          {/* 描述 */}
          <p
            style={{
              fontSize: 'clamp(1rem, 2vw, 1.15rem)',
              color: 'rgba(255, 255, 255, 0.45)',
              marginBottom: 48,
              maxWidth: 560,
              marginInline: 'auto',
              lineHeight: 1.8,
              fontWeight: 300,
              animation: 'fadeInUp 1s ease-out 0.6s both',
            }}
          >
            基于本地大语言模型，为开发团队提供安全、高效、可控的智能编码辅助服务
          </p>

          {/* CTA 按钮组 */}
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
                {isAuthenticated ? '进入控制台' : '立即体验'}
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
                color: '#fff',
                borderColor: 'rgba(255, 255, 255, 0.25)',
                background: 'rgba(255, 255, 255, 0.03)',
                backdropFilter: 'blur(10px)',
              }}
            >
              探索功能
            </Button>
          </div>


        </div>

        {/* 滚动提示 */}
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
            color: 'rgba(255, 255, 255, 0.3)',
            fontSize: 12,
            animation: 'fadeIn 1s ease-out 1.5s both',
          }}
        >
          <span>向下滚动</span>
          <ArrowDownOutlined style={{ animation: 'bounce 2s infinite' }} />
        </div>
      </section>

      {/* ═══ 数据统计区域 ═══ */}
      <section
        style={{
          padding: '80px 24px',
          background: 'linear-gradient(180deg, #050d14 0%, #0a1628 100%)',
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
                  background: 'rgba(255, 255, 255, 0.02)',
                  border: '1px solid rgba(255, 255, 255, 0.06)',
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
                <div style={{ color: 'rgba(255, 255, 255, 0.4)', fontSize: 14, marginTop: 8 }}>
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ═══ 功能介绍区域 ═══ */}
      <section
        ref={featuresRef}
        style={{
          padding: '120px 24px',
          background: `
            radial-gradient(ellipse 50% 30% at 50% 0%, rgba(0, 217, 255, 0.05), transparent),
            #0a1628
          `,
          position: 'relative',
        }}
      >
        <div style={{ maxWidth: 1200, margin: '0 auto' }}>
          {/* 标题 */}
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
              }}
            >
              核心能力
            </div>
            <h2
              style={{
                fontSize: 'clamp(2rem, 4vw, 3rem)',
                fontWeight: 800,
                color: '#fff',
                marginBottom: 16,
                letterSpacing: -1,
              }}
            >
              全方位的 AI 编码服务管理
            </h2>
            <p
              style={{
                color: 'rgba(255, 255, 255, 0.4)',
                fontSize: 17,
                maxWidth: 500,
                margin: '0 auto',
              }}
            >
              从协议兼容到权限管控，从用量计量到生态集成，一应俱全
            </p>
          </div>

          {/* 功能卡片网格 */}
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
                  background: 'rgba(255, 255, 255, 0.02)',
                  border: '1px solid rgba(255, 255, 255, 0.06)',
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
                {/* 悬停光效 */}
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
                      color: '#fff',
                      marginBottom: 10,
                    }}
                  >
                    {f.title}
                  </h3>
                  <p
                    style={{
                      color: 'rgba(255, 255, 255, 0.45)',
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

      {/* ═══ 工具生态展示 ═══ */}
      <section
        style={{
          padding: '100px 24px',
          background: '#050d14',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* 背景装饰 */}
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
              color: 'rgba(255, 255, 255, 0.3)',
              textTransform: 'uppercase',
              letterSpacing: 4,
              marginBottom: 40,
            }}
          >
            兼容 15+ 主流开发工具
          </h3>

          {/* 工具标签云 */}
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              justifyContent: 'center',
              gap: 16,
            }}
          >
            {tools.map((tool, i) => (
              <div
                key={tool}
                className="tool-tag"
                style={{
                  padding: '12px 24px',
                  background: 'rgba(255, 255, 255, 0.03)',
                  border: '1px solid rgba(255, 255, 255, 0.08)',
                  borderRadius: 50,
                  fontSize: 15,
                  color: 'rgba(255, 255, 255, 0.6)',
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

      {/* ═══ CTA 区域 ═══ */}
      <section
        style={{
          padding: '120px 24px',
          background: `
            radial-gradient(ellipse 80% 50% at 50% 100%, rgba(0, 217, 255, 0.1), transparent),
            linear-gradient(180deg, #050d14 0%, #0a1628 100%)
          `,
          textAlign: 'center',
        }}
      >
        <div style={{ maxWidth: 600, margin: '0 auto' }}>
          <h2
            style={{
              fontSize: 'clamp(2rem, 4vw, 2.75rem)',
              fontWeight: 800,
              color: '#fff',
              marginBottom: 20,
              letterSpacing: -1,
            }}
          >
            开启智能编码新体验
          </h2>
          <p
            style={{
              color: 'rgba(255, 255, 255, 0.4)',
              fontSize: 17,
              marginBottom: 40,
              lineHeight: 1.7,
            }}
          >
            立即部署 CodeMind，为您的团队提供安全、高效的 AI 编码辅助服务
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
            {isAuthenticated ? '进入控制台' : '免费开始使用'}
          </Button>
        </div>
      </section>

      {/* ═══ 底部 ═══ */}
      <footer
        style={{
          padding: '48px 24px',
          textAlign: 'center',
          background: '#050d14',
          borderTop: '1px solid rgba(255, 255, 255, 0.05)',
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
        <p style={{ fontSize: 14, color: 'rgba(255, 255, 255, 0.3)', margin: 0 }}>
          度影智能编码服务 v0.2.0
        </p>
        <p style={{ fontSize: 13, color: 'rgba(255, 255, 255, 0.2)', marginTop: 8 }}>
          © {new Date().getFullYear()} RayShape. All Rights Reserved.
        </p>
      </footer>

      {/* ═══ 全局样式 ═══ */}
      <style>{`
        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }
        
        @keyframes fadeInUp {
          from {
            opacity: 0;
            transform: translateY(30px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        
        @keyframes fadeInDown {
          from {
            opacity: 0;
            transform: translateY(-20px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        
        @keyframes float {
          0%, 100% { transform: translateY(0) scale(1); }
          50% { transform: translateY(-30px) scale(1.05); }
        }
        
        @keyframes floatCode {
          0%, 100% { transform: perspective(1000px) rotateY(-15deg) rotateX(5deg) translateY(0); }
          50% { transform: perspective(1000px) rotateY(-15deg) rotateX(5deg) translateY(-15px); }
        }
        
        @keyframes pulse {
          0%, 100% { opacity: 1; transform: scale(1); }
          50% { opacity: 0.5; transform: scale(1.2); }
        }
        
        @keyframes bounce {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(8px); }
        }
        
        @keyframes gradientShift {
          0% { background-position: 0% 50%; }
          50% { background-position: 100% 50%; }
          100% { background-position: 0% 50%; }
        }
        
        @keyframes gridMove {
          0% { transform: translateY(0); }
          100% { transform: translateY(60px); }
        }
        
        .feature-card.animate-in,
        .stat-card.animate-in {
          opacity: 1 !important;
          transform: translateY(0) !important;
        }
        
        .tool-tag.animate-in {
          opacity: 1 !important;
          transform: scale(1) !important;
        }
        
        .feature-card:hover {
          transform: translateY(-4px) !important;
          border-color: rgba(255, 255, 255, 0.15) !important;
        }
        
        .feature-card:hover .card-glow {
          opacity: 1 !important;
        }
        
        .feature-card:hover .feature-icon {
          transform: scale(1.1) rotate(-5deg);
        }
        
        .stat-card:hover {
          transform: translateY(-4px) !important;
          border-color: rgba(0, 217, 255, 0.2) !important;
          background: rgba(255, 255, 255, 0.04) !important;
        }
        
        .tool-tag:hover {
          background: rgba(0, 217, 255, 0.08) !important;
          border-color: rgba(0, 217, 255, 0.2) !important;
          color: rgba(255, 255, 255, 0.9) !important;
        }
        
        .cta-primary:hover {
          transform: translateY(-2px);
          box-shadow: 0 20px 40px rgba(0, 217, 255, 0.3) !important;
        }
        
        .cta-secondary:hover {
          background: rgba(255, 255, 255, 0.08) !important;
          border-color: 'rgba(255, 255, 255, 0.4)' !important;
        }
      `}</style>
    </div>
  );
};

export default HomePage;
