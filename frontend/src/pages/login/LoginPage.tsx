import { useState, useEffect, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Form, Input, Button, message, ConfigProvider, ThemeConfig } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import useAuthStore from '@/store/authStore';

// 登录页输入框全局样式 - 极简玻璃态设计
const loginInputStyles = `
  /* ===== 输入框基础样式 - 极简设计 ===== */
  .login-input-wrapper {
    position: relative;
    transition: all 0.3s ease;
  }
  
  /* 输入框容器 - 完全透明背景 */
  .login-input-wrapper .ant-input-affix-wrapper {
    background: transparent !important;
    border: none !important;
    border-bottom: 1px solid rgba(255, 255, 255, 0.2) !important;
    border-radius: 0 !important;
    box-shadow: none !important;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
    padding: 0 4px 0 0 !important;
    height: 48px !important;
  }
  
  /* 输入框悬停状态 */
  .login-input-wrapper .ant-input-affix-wrapper:hover {
    border-bottom-color: rgba(255, 255, 255, 0.4) !important;
  }
  
  /* 输入框聚焦状态 - 下划线发光 */
  .login-input-wrapper .ant-input-affix-wrapper:focus,
  .login-input-wrapper .ant-input-affix-wrapper-focused {
    background: transparent !important;
    border-bottom-color: #4BA3D4 !important;
    box-shadow: 0 2px 0 0 rgba(75, 163, 212, 0.6) !important;
  }
  
  /* 输入框内部 input 元素 - 完全透明 */
  .login-input-wrapper .ant-input {
    background: transparent !important;
    border: none !important;
    box-shadow: none !important;
    color: #ffffff !important;
    font-size: 16px !important;
    font-weight: 400 !important;
    letter-spacing: 0.3px !important;
  }
  
  /* 覆盖浏览器自动填充 - 使用透明色 */
  .login-input-wrapper .ant-input:-webkit-autofill,
  .login-input-wrapper .ant-input:-webkit-autofill:hover,
  .login-input-wrapper .ant-input:-webkit-autofill:focus,
  .login-input-wrapper .ant-input:-webkit-autofill:active {
    -webkit-box-shadow: 0 0 0 1000px transparent inset !important;
    -webkit-text-fill-color: #ffffff !important;
    caret-color: #ffffff !important;
    transition: background-color 5000s ease-in-out 0s !important;
  }
  
  .login-input-wrapper .ant-input::placeholder {
    color: rgba(255, 255, 255, 0.35) !important;
    font-weight: 400 !important;
  }
  
  /* ===== 图标样式 ===== */
  .login-input-wrapper .login-input-icon {
    color: rgba(255, 255, 255, 0.5) !important;
    font-size: 18px !important;
    margin-right: 12px !important;
    transition: all 0.3s ease !important;
  }
  
  /* 聚焦时图标高亮 */
  .login-input-wrapper .ant-input-affix-wrapper-focused .login-input-icon {
    color: #4BA3D4 !important;
  }
  
  /* 密码框眼睛图标 */
  .login-input-wrapper .ant-input-suffix .anticon {
    color: rgba(255, 255, 255, 0.4) !important;
    font-size: 16px !important;
    transition: all 0.3s ease !important;
  }
  
  .login-input-wrapper .ant-input-suffix .anticon:hover {
    color: rgba(255, 255, 255, 0.7) !important;
  }
  
  /* ===== 错误状态样式 ===== */
  .login-input-wrapper.ant-form-item-has-error .ant-input-affix-wrapper {
    background: transparent !important;
    border-bottom-color: rgba(255, 120, 117, 0.8) !important;
    box-shadow: 0 2px 0 0 rgba(255, 120, 117, 0.6) !important;
  }
  
  .login-input-wrapper.ant-form-item-has-error .login-input-icon {
    color: rgba(255, 120, 117, 0.8) !important;
  }
  
  /* ===== 登录按钮样式优化 ===== */
  .login-button {
    height: 48px !important;
    border-radius: 8px !important;
    font-size: 15px !important;
    font-weight: 600 !important;
    letter-spacing: 4px !important;
    background: linear-gradient(135deg, #2B7CB3 0%, #4BA3D4 100%) !important;
    border: none !important;
    box-shadow: 
      0 4px 16px rgba(43, 124, 179, 0.35),
      inset 0 1px 0 rgba(255, 255, 255, 0.15) !important;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  }
  
  .login-button:hover {
    background: linear-gradient(135deg, #3287c2 0%, #56b0e0 100%) !important;
    box-shadow: 
      0 6px 20px rgba(43, 124, 179, 0.45),
      inset 0 1px 0 rgba(255, 255, 255, 0.2) !important;
    transform: translateY(-1px);
  }
  
  .login-button:active {
    transform: translateY(0);
    box-shadow: 
      0 2px 8px rgba(43, 124, 179, 0.3),
      inset 0 1px 0 rgba(255, 255, 255, 0.1) !important;
  }
  
  .login-button:disabled {
    background: linear-gradient(135deg, rgba(43, 124, 179, 0.5) 0%, rgba(75, 163, 212, 0.5) 100%) !important;
    box-shadow: none !important;
    transform: none !important;
  }
`;

// 登录页专用主题配置 - 极简下划线风格
const loginTheme: ThemeConfig = {
  token: {
    colorBgContainer: 'transparent',
    colorBorder: 'rgba(255, 255, 255, 0.2)',
    colorText: '#ffffff',
    colorTextPlaceholder: 'rgba(255, 255, 255, 0.35)',
    borderRadius: 0,
    controlHeight: 48,
    colorError: '#ff7875',
    colorErrorBorderHover: '#ff7875',
  },
  components: {
    Input: {
      colorBgContainer: 'transparent',
      colorBorder: 'rgba(255, 255, 255, 0.2)',
      colorText: '#ffffff',
      colorTextPlaceholder: 'rgba(255, 255, 255, 0.35)',
      colorIcon: 'rgba(255, 255, 255, 0.5)',
      colorIconHover: 'rgba(255, 255, 255, 0.7)',
      borderRadius: 0,
      controlHeight: 48,
      activeBorderColor: '#4BA3D4',
      hoverBorderColor: 'rgba(255, 255, 255, 0.4)',
      activeShadow: '0 2px 0 0 rgba(75, 163, 212, 0.6)',
      paddingInline: 0,
      colorError: '#ff7875',
      colorErrorBorder: 'rgba(255, 120, 117, 0.8)',
    },
  },
};

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
    <>
      {/* 注入全局样式 */}
      <style>{loginInputStyles}</style>
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
            background: 'rgba(255, 255, 255, 0.08)',
            backdropFilter: 'blur(24px)',
            WebkitBackdropFilter: 'blur(24px)',
            borderRadius: 24,
            border: '1px solid rgba(255, 255, 255, 0.12)',
            boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.08)',
            padding: '48px 36px 40px',
          }}
        >
          {/* 品牌标识 */}
          <div style={{ textAlign: 'center', marginBottom: 40 }}>
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
                boxShadow: '0 8px 24px rgba(43, 124, 179, 0.35)',
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
                color: 'rgba(255, 255, 255, 0.5)',
                marginTop: 6,
                letterSpacing: 2,
              }}
            >
              度影智能编码服务
            </p>
          </div>

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
                style={{ marginBottom: 28 }}
                className="login-input-wrapper"
              >
                <Input
                  prefix={
                    <span className="login-input-icon">
                      <UserOutlined />
                    </span>
                  }
                  placeholder="用户名"
                />
              </Form.Item>

              <Form.Item
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
                style={{ marginBottom: 40 }}
                className="login-input-wrapper"
              >
                <Input.Password
                  prefix={
                    <span className="login-input-icon">
                      <LockOutlined />
                    </span>
                  }
                  placeholder="密码"
                />
              </Form.Item>

              <Form.Item style={{ marginBottom: 0 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  block
                  loading={loading}
                  className="login-button"
                >
                  登 录
                </Button>
              </Form.Item>
            </Form>
          </ConfigProvider>

          {/* 底部提示 */}
          <div
            style={{
              textAlign: 'center',
              marginTop: 32,
              fontSize: 12,
              color: 'rgba(255, 255, 255, 0.35)',
            }}
          >
            默认管理员：admin / Admin@123456
          </div>
        </div>
      </div>
    </>
  );
};

export default LoginPage;
