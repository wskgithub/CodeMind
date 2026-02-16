import { useState, useEffect, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Form, Input, Button, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import useAuthStore from '@/store/authStore';

/** 登录页面 — Glassmorphism 风格 */
const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((s) => s.login);
  const [loading, setLoading] = useState(false);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/dashboard';

  // 粒子背景动画
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animationId: number;
    const particles: { x: number; y: number; vx: number; vy: number; r: number; a: number }[] = [];
    const count = 60;

    const resize = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    resize();
    window.addEventListener('resize', resize);

    for (let i = 0; i < count; i++) {
      particles.push({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        vx: (Math.random() - 0.5) * 0.5,
        vy: (Math.random() - 0.5) * 0.5,
        r: Math.random() * 2 + 1,
        a: Math.random() * 0.5 + 0.2,
      });
    }

    const draw = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);

      particles.forEach((p) => {
        p.x += p.vx;
        p.y += p.vy;
        if (p.x < 0 || p.x > canvas.width) p.vx *= -1;
        if (p.y < 0 || p.y > canvas.height) p.vy *= -1;

        ctx.beginPath();
        ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(255, 255, 255, ${p.a})`;
        ctx.fill();
      });

      // 连线
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const pi = particles[i]!;
          const pj = particles[j]!;
          const dx = pi.x - pj.x;
          const dy = pi.y - pj.y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 120) {
            ctx.beginPath();
            ctx.strokeStyle = `rgba(255, 255, 255, ${0.08 * (1 - dist / 120)})`;
            ctx.lineWidth = 0.5;
            ctx.moveTo(pi.x, pi.y);
            ctx.lineTo(pj.x, pj.y);
            ctx.stroke();
          }
        }
      }

      animationId = requestAnimationFrame(draw);
    };

    draw();

    return () => {
      cancelAnimationFrame(animationId);
      window.removeEventListener('resize', resize);
    };
  }, []);

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      navigate(from, { replace: true });
    } catch {
      // 错误已在拦截器中处理
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
        background: 'linear-gradient(135deg, #0a1628 0%, #132f4c 30%, #1a4a6e 60%, #2B7CB3 100%)',
      }}
    >
      {/* 粒子画布 */}
      <canvas
        ref={canvasRef}
        style={{ position: 'absolute', inset: 0, zIndex: 0 }}
      />

      {/* 装饰光圈 */}
      <div
        style={{
          position: 'absolute',
          top: '10%',
          left: '15%',
          width: 400,
          height: 400,
          background: 'radial-gradient(circle, rgba(43, 124, 179, 0.2) 0%, transparent 70%)',
          borderRadius: '50%',
          filter: 'blur(40px)',
          pointerEvents: 'none',
        }}
        className="animate-float"
      />
      <div
        style={{
          position: 'absolute',
          bottom: '5%',
          right: '10%',
          width: 350,
          height: 350,
          background: 'radial-gradient(circle, rgba(107, 197, 232, 0.15) 0%, transparent 70%)',
          borderRadius: '50%',
          filter: 'blur(40px)',
          pointerEvents: 'none',
          animationDelay: '3s',
        }}
        className="animate-float"
      />

      {/* 玻璃登录卡片 */}
      <div
        className="animate-fade-in-up"
        style={{
          width: 420,
          position: 'relative',
          zIndex: 10,
          background: 'rgba(255, 255, 255, 0.1)',
          backdropFilter: 'blur(24px)',
          WebkitBackdropFilter: 'blur(24px)',
          borderRadius: 24,
          border: '1px solid rgba(255, 255, 255, 0.15)',
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.1)',
          padding: '48px 36px 40px',
        }}
      >
        {/* 品牌标识 */}
        <div style={{ textAlign: 'center', marginBottom: 36 }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 16,
              background: 'linear-gradient(135deg, #2B7CB3, #4BA3D4)',
              display: 'inline-flex',
              alignItems: 'center',
              justifyContent: 'center',
              marginBottom: 16,
              boxShadow: '0 8px 24px rgba(43, 124, 179, 0.4)',
            }}
          >
            <span style={{ color: '#fff', fontWeight: 800, fontSize: 20 }}>CM</span>
          </div>
          <h1
            style={{
              fontSize: 26,
              fontWeight: 700,
              color: '#fff',
              margin: 0,
              letterSpacing: -0.5,
            }}
          >
            CodeMind
          </h1>
          <p
            style={{
              fontSize: 13,
              color: 'rgba(255, 255, 255, 0.55)',
              marginTop: 6,
              letterSpacing: 2,
            }}
          >
            度影智能编码服务
          </p>
        </div>

        {/* 登录表单 */}
        <Form
          name="login"
          onFinish={handleSubmit}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: 'rgba(255, 255, 255, 0.45)' }} />}
              placeholder="用户名"
              style={{
                background: 'rgba(255, 255, 255, 0.08)',
                border: '1px solid rgba(255, 255, 255, 0.12)',
                borderRadius: 12,
                color: '#fff',
                height: 48,
              }}
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'rgba(255, 255, 255, 0.45)' }} />}
              placeholder="密码"
              style={{
                background: 'rgba(255, 255, 255, 0.08)',
                border: '1px solid rgba(255, 255, 255, 0.12)',
                borderRadius: 12,
                height: 48,
              }}
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <Button
              type="primary"
              htmlType="submit"
              block
              loading={loading}
              style={{
                height: 48,
                borderRadius: 12,
                fontSize: 15,
                fontWeight: 600,
                background: 'linear-gradient(135deg, #2B7CB3 0%, #4BA3D4 100%)',
                border: 'none',
                boxShadow: '0 8px 24px rgba(43, 124, 179, 0.35)',
                transition: 'all 0.3s',
              }}
            >
              登 录
            </Button>
          </Form.Item>
        </Form>

        {/* 底部提示 */}
        <div
          style={{
            textAlign: 'center',
            marginTop: 24,
            fontSize: 12,
            color: 'rgba(255, 255, 255, 0.35)',
          }}
        >
          默认管理员：admin / Admin@123456
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
