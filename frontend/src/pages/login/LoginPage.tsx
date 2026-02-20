import { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Form, Input, Button, message, ConfigProvider, ThemeConfig, Alert } from 'antd';
import { UserOutlined, LockOutlined, LockFilled, PlayCircleOutlined } from '@ant-design/icons';
import useAuthStore from '@/store/authStore';
import axios from 'axios';

// 登录页专用主题配置
const loginTheme: ThemeConfig = {
  token: {
    colorBgContainer: 'transparent',
    colorBorder: 'rgba(255, 255, 255, 0.15)',
    colorText: '#ffffff',
    colorTextPlaceholder: 'rgba(255, 255, 255, 0.35)',
    borderRadius: 12,
    controlHeight: 52,
    colorError: '#ff7875',
    colorErrorBorderHover: '#ff7875',
  },
  components: {
    Input: {
      colorBgContainer: 'transparent',
      colorBorder: 'rgba(255, 255, 255, 0.15)',
      colorText: '#ffffff',
      colorTextPlaceholder: 'rgba(255, 255, 255, 0.35)',
      colorIcon: 'rgba(255, 255, 255, 0.4)',
      colorIconHover: 'rgba(255, 255, 255, 0.8)',
      borderRadius: 0,
      controlHeight: 52,
      activeBorderColor: '#00D9FF',
      hoverBorderColor: 'rgba(255, 255, 255, 0.3)',
      activeShadow: 'none',
      paddingInline: 0,
      colorError: '#ff7875',
      colorErrorBorder: 'rgba(255, 120, 117, 0.8)',
    },
  },
};

/** 格式化剩余时间 */
const formatRemainingTime = (seconds: number): string => {
  if (seconds < 60) return `${seconds}秒`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}分钟`;
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (remainingMinutes > 0) return `${hours}小时${remainingMinutes}分钟`;
  return `${hours}小时`;
};

/** 星空连线粒子动效 */
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

    const count = Math.min(Math.floor((window.innerWidth * window.innerHeight) / 10000), 120);
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

    const maxDist = 180;

    const draw = () => {
      ctx.clearRect(0, 0, window.innerWidth, window.innerHeight);

      for (const p of particles) {
        p.x += p.vx;
        p.y += p.vy;
        p.pulse += 0.02;
        if (p.x < 0 || p.x > window.innerWidth) p.vx *= -1;
        if (p.y < 0 || p.y > window.innerHeight) p.vy *= -1;
      }

      // 绘制连线
      for (let i = 0; i < particles.length; i++) {
        const pi = particles[i]!;
        
        // 鼠标交互连线
        const dx = mouseRef.current.x - pi.x;
        const dy = mouseRef.current.y - pi.y;
        const mouseDist = Math.sqrt(dx * dx + dy * dy);
        if (mouseDist < 200) {
          const alpha = (1 - mouseDist / 200) * 0.3;
          const gradient = ctx.createLinearGradient(pi.x, pi.y, mouseRef.current.x, mouseRef.current.y);
          gradient.addColorStop(0, `rgba(0, 217, 255, ${alpha})`);
          gradient.addColorStop(1, `rgba(157, 78, 221, ${alpha})`);
          ctx.beginPath();
          ctx.strokeStyle = gradient;
          ctx.lineWidth = 0.6;
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
            const alpha = (1 - dist / maxDist) * 0.2;
            ctx.beginPath();
            ctx.strokeStyle = `rgba(100, 200, 255, ${alpha})`;
            ctx.lineWidth = 0.4;
            ctx.moveTo(pi.x, pi.y);
            ctx.lineTo(pj.x, pj.y);
            ctx.stroke();
          }
        }
      }

      // 绘制粒子
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

/** 登录页面 — 全新设计风格 */
const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((s) => s.login);
  const [loading, setLoading] = useState(false);
  const [lockInfo, setLockInfo] = useState<{
    locked: boolean;
    remainingTime: number;
    failCount: number;
    maxFailCount: number;
  } | null>(null);
  const [countdown, setCountdown] = useState<number>(0);
  const countdownTimerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/dashboard';

  useEffect(() => {
    return () => {
      if (countdownTimerRef.current) clearInterval(countdownTimerRef.current);
    };
  }, []);

  useEffect(() => {
    if (lockInfo?.locked && lockInfo.remainingTime > 0) {
      setCountdown(lockInfo.remainingTime);
      countdownTimerRef.current = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            setLockInfo(null);
            if (countdownTimerRef.current) clearInterval(countdownTimerRef.current);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
      return () => {
        if (countdownTimerRef.current) clearInterval(countdownTimerRef.current);
      };
    }
  }, [lockInfo]);

  const handleSubmit = async (values: { username: string; password: string }) => {
    if (lockInfo?.locked && countdown > 0) {
      message.error(`账号已被锁定，请${formatRemainingTime(countdown)}后再试`);
      return;
    }

    setLoading(true);
    setLockInfo(null);
    
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      navigate(from, { replace: true });
    } catch (error) {
      if (axios.isAxiosError(error) && error.response) {
        const { data, status } = error.response;
        if (data?.code === 40008 || (status === 403 && data?.data?.locked)) {
          const lockData = data.data || {};
          setLockInfo({
            locked: true,
            remainingTime: lockData.remaining_time || 300,
            failCount: lockData.fail_count || 5,
            maxFailCount: lockData.max_fail_count || 5,
          });
        } else if (status === 401) {
          if (data?.data?.fail_count > 0) {
            const remainingAttempts = (data.data.max_fail_count || 5) - data.data.fail_count;
            message.error(`${data.message}，还剩 ${remainingAttempts} 次机会`);
          } else {
            message.error(data?.message || '用户名或密码错误');
          }
        } else {
          message.error(data?.message || '登录失败');
        }
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        position: 'relative',
        overflow: 'hidden',
        background: `
          radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 217, 255, 0.1), transparent),
          radial-gradient(ellipse 60% 40% at 80% 80%, rgba(157, 78, 221, 0.08), transparent),
          linear-gradient(180deg, #0a1628 0%, #050d14 100%)
        `,
      }}
    >
      {/* 星空粒子背景 */}
      <StarfieldCanvas />

      {/* 动态网格背景 */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            linear-gradient(rgba(0, 217, 255, 0.02) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0, 217, 255, 0.02) 1px, transparent 1px)
          `,
          backgroundSize: '60px 60px',
          animation: 'gridMove 25s linear infinite',
          zIndex: 0,
        }}
      />

      {/* 发光球体装饰 */}
      <div
        className="glow-orb"
        style={{
          position: 'absolute',
          top: '10%',
          left: '5%',
          width: 500,
          height: 500,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(0, 217, 255, 0.12) 0%, transparent 70%)',
          filter: 'blur(80px)',
          animation: 'float 10s ease-in-out infinite',
          zIndex: 0,
        }}
      />
      <div
        className="glow-orb"
        style={{
          position: 'absolute',
          bottom: '5%',
          right: '5%',
          width: 450,
          height: 450,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(157, 78, 221, 0.1) 0%, transparent 70%)',
          filter: 'blur(80px)',
          animation: 'float 12s ease-in-out infinite reverse',
          zIndex: 0,
        }}
      />

      {/* 浮动代码片段装饰 */}
      <div
        className="code-float"
        style={{
          position: 'absolute',
          top: '20%',
          right: '8%',
          padding: '14px 18px',
          background: 'rgba(0, 217, 255, 0.04)',
          backdropFilter: 'blur(12px)',
          border: '1px solid rgba(0, 217, 255, 0.08)',
          borderRadius: 12,
          fontFamily: 'monospace',
          fontSize: 12,
          color: 'rgba(0, 217, 255, 0.6)',
          transform: 'perspective(1000px) rotateY(-20deg) rotateX(5deg)',
          animation: 'floatCode 8s ease-in-out infinite',
          zIndex: 2,
        }}
      >
        <div>{`{`}</div>
        <div style={{ paddingLeft: 10 }}>{`"model": "claude-3",`}</div>
        <div style={{ paddingLeft: 10 }}>{`"streaming": true`}</div>
        <div>{`}`}</div>
      </div>

      <div
        className="code-float"
        style={{
          position: 'absolute',
          bottom: '25%',
          left: '8%',
          padding: '14px 18px',
          background: 'rgba(157, 78, 221, 0.04)',
          backdropFilter: 'blur(12px)',
          border: '1px solid rgba(157, 78, 221, 0.08)',
          borderRadius: 12,
          fontFamily: 'monospace',
          fontSize: 12,
          color: 'rgba(157, 78, 221, 0.6)',
          transform: 'perspective(1000px) rotateY(20deg) rotateX(-5deg)',
          animation: 'floatCode 10s ease-in-out infinite reverse',
          zIndex: 2,
        }}
      >
        <div style={{ color: 'rgba(255, 255, 255, 0.4)' }}>{`// AI Gateway`}</div>
        <div>{`const ai = new CodeMind()`}</div>
        <div style={{ paddingLeft: 10 }}>{`.connect()`}</div>
      </div>

      {/* 玻璃登录卡片 */}
      <div
        style={{
          width: 440,
          position: 'relative',
          zIndex: 10,
          background: 'rgba(255, 255, 255, 0.02)',
          backdropFilter: 'blur(24px)',
          WebkitBackdropFilter: 'blur(24px)',
          borderRadius: 28,
          border: '1px solid rgba(255, 255, 255, 0.08)',
          boxShadow: `
            0 25px 80px rgba(0, 0, 0, 0.4),
            inset 0 1px 0 rgba(255, 255, 255, 0.05)
          `,
          padding: '48px 40px 40px',
          animation: 'fadeInUp 0.8s ease-out',
        }}
      >
        {/* 卡片顶部光晕 */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: '10%',
            right: '10%',
            height: 1,
            background: 'linear-gradient(90deg, transparent, rgba(0, 217, 255, 0.4), transparent)',
          }}
        />

        {/* 品牌标识 */}
        <div style={{ textAlign: 'center', marginBottom: 36 }}>
          {/* Badge */}
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 12px',
              background: 'rgba(0, 217, 255, 0.08)',
              border: '1px solid rgba(0, 217, 255, 0.15)',
              borderRadius: 50,
              fontSize: 12,
              color: '#00D9FF',
              marginBottom: 20,
            }}
          >
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: '#00D9FF', animation: 'pulse 2s infinite' }} />
            企业级 AI 编码平台
          </div>

          {/* Logo */}
          <h1
            style={{
              fontSize: 42,
              fontWeight: 900,
              margin: '0 0 8px',
              letterSpacing: -2,
              background: 'linear-gradient(135deg, #fff 0%, #00D9FF 50%, #9D4EDD 100%)',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              backgroundClip: 'text',
            }}
          >
            CodeMind
          </h1>
          <p
            style={{
              fontSize: 14,
              color: 'rgba(255, 255, 255, 0.4)',
              letterSpacing: 3,
              textTransform: 'uppercase',
            }}
          >
            智能编码新纪元
          </p>
        </div>

        {/* 锁定提示 */}
        {lockInfo?.locked && (
          <Alert
            icon={<LockFilled />}
            message="账号已被锁定"
            description={
              <div>
                <div style={{ fontSize: 13 }}>登录失败次数过多，系统已临时锁定。</div>
                {countdown > 0 && (
                  <div style={{ fontSize: 13, marginTop: 8, color: '#ffccc7' }}>
                    剩余时间：<strong>{formatRemainingTime(countdown)}</strong>
                  </div>
                )}
                <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255,255,255,0.5)' }}>
                  如需立即解锁，请联系系统管理员
                </div>
              </div>
            }
            type="error"
            showIcon
            style={{
              marginBottom: 24,
              background: 'rgba(255, 77, 79, 0.1)',
              border: '1px solid rgba(255, 77, 79, 0.2)',
              borderRadius: 12,
            }}
          />
        )}

        {/* 登录表单 */}
        <ConfigProvider theme={loginTheme}>
          <Form
            name="login"
            onFinish={handleSubmit}
            autoComplete="off"
            size="large"
          >
            <Form.Item
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
              style={{ marginBottom: 24 }}
              className="login-form-item"
            >
              <Input
                prefix={<UserOutlined className="login-input-icon" />}
                placeholder="用户名"
                size="large"
                disabled={lockInfo?.locked && countdown > 0}
                className="login-input"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
              style={{ marginBottom: 40 }}
              className="login-form-item"
            >
              <Input.Password
                prefix={<LockOutlined className="login-input-icon" />}
                placeholder="密码"
                size="large"
                disabled={lockInfo?.locked && countdown > 0}
                className="login-input"
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                type="primary"
                htmlType="submit"
                block
                loading={loading}
                disabled={lockInfo?.locked && countdown > 0}
                icon={<PlayCircleOutlined />}
                style={{
                  height: 52,
                  borderRadius: 12,
                  fontSize: 16,
                  fontWeight: 600,
                  letterSpacing: 2,
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  border: 'none',
                  boxShadow: '0 8px 32px rgba(0, 217, 255, 0.25)',
                  transition: 'all 0.3s ease',
                }}
                className="login-btn"
              >
                {lockInfo?.locked && countdown > 0 ? '账号已锁定' : '登 录'}
              </Button>
            </Form.Item>
          </Form>
        </ConfigProvider>

        {/* 底部提示 */}
        <div
          style={{
            textAlign: 'center',
            marginTop: 28,
            fontSize: 12,
            color: 'rgba(255, 255, 255, 0.25)',
          }}
        >
          默认管理员：admin / Admin@123456
        </div>
      </div>

      {/* 全局样式 */}
      <style>{`
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
        
        @keyframes float {
          0%, 100% { transform: translateY(0) scale(1); }
          50% { transform: translateY(-20px) scale(1.02); }
        }
        
        @keyframes floatCode {
          0%, 100% { transform: perspective(1000px) rotateY(-20deg) rotateX(5deg) translateY(0); }
          50% { transform: perspective(1000px) rotateY(-20deg) rotateX(5deg) translateY(-10px); }
        }
        
        @keyframes pulse {
          0%, 100% { opacity: 1; transform: scale(1); }
          50% { opacity: 0.6; transform: scale(1.1); }
        }
        
        @keyframes gridMove {
          0% { transform: translateY(0); }
          100% { transform: translateY(60px); }
        }
        
        .login-btn:hover {
          transform: translateY(-2px) !important;
          box-shadow: 0 12px 40px rgba(0, 217, 255, 0.35) !important;
        }
        
        .login-btn:active {
          transform: translateY(0) !important;
        }
        
        /* 纯色输入框样式 - 避免自动填充颜色不一致问题 */
        .login-form-item {
          margin-bottom: 24px !important;
        }
        
        /* 输入框外层容器 - 统一背景色 */
        .login-form-item .ant-input-affix-wrapper {
          background: #0d1d2d !important;
          border: 1px solid rgba(255, 255, 255, 0.12) !important;
          border-radius: 12px !important;
          padding: 12px 16px !important;
          box-shadow: 
            inset 0 1px 0 rgba(255, 255, 255, 0.05),
            0 4px 20px rgba(0, 0, 0, 0.3) !important;
          transition: all 0.3s ease !important;
        }
        
        /* hover状态 - 外层容器和内部输入框保持相同背景色 */
        .login-form-item .ant-input-affix-wrapper:hover {
          background: #0d1d2d !important;
          border-color: rgba(0, 217, 255, 0.5) !important;
          box-shadow: 
            inset 0 1px 0 rgba(255, 255, 255, 0.08),
            0 4px 24px rgba(0, 217, 255, 0.15) !important;
        }
        
        /* hover时内部输入框保持与外层一致 */
        .login-form-item .ant-input-affix-wrapper:hover .ant-input {
          background: #0d1d2d !important;
          background-color: #0d1d2d !important;
        }
        
        /* focus状态 - 统一背景色 */
        .login-form-item .ant-input-affix-wrapper:focus,
        .login-form-item .ant-input-affix-wrapper-focused {
          background: #0d1d2d !important;
          border-color: #00D9FF !important;
          box-shadow: 
            inset 0 1px 0 rgba(255, 255, 255, 0.1),
            0 0 0 3px rgba(0, 217, 255, 0.2),
            0 4px 24px rgba(0, 217, 255, 0.25) !important;
        }
        
        /* focus时内部输入框保持与外层一致 */
        .login-form-item .ant-input-affix-wrapper:focus .ant-input,
        .login-form-item .ant-input-affix-wrapper-focused .ant-input {
          background: #0d1d2d !important;
          background-color: #0d1d2d !important;
        }
        
        .login-input-icon {
          color: rgba(255, 255, 255, 0.45) !important;
          font-size: 18px !important;
          margin-right: 12px !important;
          transition: all 0.3s ease !important;
        }
        
        .login-form-item .ant-input-affix-wrapper-focused .login-input-icon {
          color: #00D9FF !important;
        }
        
        /* 内部输入框基础样式 - 与外层容器统一背景色 */
        .login-form-item .ant-input {
          background: #0d1d2d !important;
          background-color: #0d1d2d !important;
          color: #ffffff !important;
          font-size: 16px !important;
          font-weight: 400 !important;
          border: none !important;
          box-shadow: none !important;
        }
        
        .login-form-item .ant-input::placeholder {
          color: rgba(255, 255, 255, 0.35) !important;
        }
        
        /* 密码框眼睛图标 */
        .login-form-item .ant-input-suffix .anticon {
          color: rgba(255, 255, 255, 0.4) !important;
          font-size: 16px !important;
          transition: all 0.3s ease !important;
        }
        
        .login-form-item .ant-input-suffix .anticon:hover {
          color: rgba(255, 255, 255, 0.8) !important;
        }
        
        /* 覆盖浏览器自动填充 - 最强覆盖 */
        .login-form-item .ant-input-affix-wrapper input.ant-input,
        input.ant-input:-webkit-autofill,
        input.ant-input:-webkit-autofill:hover,
        input.ant-input:-webkit-autofill:focus,
        input.ant-input:-webkit-autofill:active {
          -webkit-box-shadow: 0 0 0 100px #0d1d2d inset !important;
          box-shadow: 0 0 0 100px #0d1d2d inset !important;
          background: #0d1d2d !important;
          background-color: #0d1d2d !important;
          background-image: none !important;
          -webkit-text-fill-color: #ffffff !important;
          caret-color: #ffffff !important;
          color: #ffffff !important;
        }
      `}</style>
    </div>
  );
};

export default LoginPage;
