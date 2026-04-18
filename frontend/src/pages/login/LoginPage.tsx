import { UserOutlined, LockOutlined, LockFilled, PlayCircleOutlined, SunOutlined, MoonOutlined, TranslationOutlined } from '@ant-design/icons';
import { Form, Input, Button, message, ConfigProvider, ThemeConfig, Alert, Dropdown } from 'antd';
import axios from 'axios';
import { useState, useEffect, useRef, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useLocation } from 'react-router-dom';

import { SUPPORTED_LANGUAGES, type SupportedLanguage } from '@/i18n';
import { translateErrorCode } from '@/services/request';
import useAppStore from '@/store/appStore';
import useAuthStore from '@/store/authStore';


// theme config for login page
const getLoginTheme = (isDark: boolean): ThemeConfig => ({
  token: {
    colorBgContainer: 'transparent',
    colorBorder: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.15)',
    colorText: isDark ? '#ffffff' : 'rgba(0, 0, 0, 0.85)',
    colorTextPlaceholder: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.4)',
    borderRadius: 12,
    controlHeight: 52,
    colorError: '#ff7875',
    colorErrorBorderHover: '#ff7875',
  },
  components: {
    Input: {
      colorBgContainer: 'transparent',
      colorBorder: isDark ? 'rgba(255, 255, 255, 0.15)' : 'rgba(0, 0, 0, 0.15)',
      colorText: isDark ? '#ffffff' : 'rgba(0, 0, 0, 0.85)',
      colorTextPlaceholder: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.4)',
      colorIcon: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)',
      colorIconHover: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)',
      borderRadius: 0,
      controlHeight: 52,
      activeBorderColor: '#00D9FF',
      hoverBorderColor: isDark ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.3)',
      activeShadow: 'none',
      paddingInline: 0,
      colorError: '#ff7875',
      colorErrorBorder: 'rgba(255, 120, 117, 0.8)',
    },
  },
});

// format remaining time - hook version used inside component
type TFunction = (key: string, options?: Record<string, unknown>) => string;
const formatRemainingTime = (seconds: number, t: TFunction): string => {
  if (seconds < 60) return t('time.seconds', { count: seconds });
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return t('time.minutes', { count: minutes });
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = minutes % 60;
  if (remainingMinutes > 0) return t('time.hoursMinutes', { hours, minutes: remainingMinutes });
  return t('time.hours', { count: hours });
};

// starfield particle animation with connections
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

      // mouse interaction lines
      for (let i = 0; i < particles.length; i++) {
        const pi = particles[i]!;
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
      }

      // spatial grid for neighbor search optimization
      const cellSize = maxDist;
      const grid: Map<string, typeof particles> = new Map();
      for (const p of particles) {
        const cx = Math.floor(p.x / cellSize);
        const cy = Math.floor(p.y / cellSize);
        const key = `${cx},${cy}`;
        if (!grid.has(key)) grid.set(key, []);
        grid.get(key)!.push(p);
      }
      for (const [key, cell] of grid) {
        const parts = key.split(',').map(Number);
        const cx = parts[0]!;
        const cy = parts[1]!;
        for (let ddx = 0; ddx <= 1; ddx++) {
          for (let ddy = -1; ddy <= 1; ddy++) {
            if (ddx === 0 && ddy <= 0 && !(ddx === 0 && ddy === 0)) continue;
            const neighborKey = `${cx + ddx},${cy + ddy}`;
            const neighbor = grid.get(neighborKey);
            if (!neighbor) continue;
            const isSameCell = ddx === 0 && ddy === 0;
            for (let i = 0; i < cell.length; i++) {
              const startJ = isSameCell ? i + 1 : 0;
              for (let j = startJ; j < neighbor.length; j++) {
                const dx = cell[i]!.x - neighbor[j]!.x;
                const dy = cell[i]!.y - neighbor[j]!.y;
                const dist = Math.sqrt(dx * dx + dy * dy);
                if (dist < maxDist) {
                  const alpha = (1 - dist / maxDist) * 0.2;
                  ctx.beginPath();
                  ctx.strokeStyle = `rgba(100, 200, 255, ${alpha})`;
                  ctx.lineWidth = 0.4;
                  ctx.moveTo(cell[i]!.x, cell[i]!.y);
                  ctx.lineTo(neighbor[j]!.x, neighbor[j]!.y);
                  ctx.stroke();
                }
              }
            }
          }
        }
      }

      // draw particles
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

const LoginPage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((s) => s.login);
  const { themeMode, toggleTheme, setLanguage } = useAppStore();
  const isDark = themeMode === 'dark';

  // Language switch menu
  const languageMenuItems = SUPPORTED_LANGUAGES.map((lang) => ({
    key: lang.code,
    label: lang.nativeName,
    onClick: () => setLanguage(lang.code as SupportedLanguage),
  }));
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
      message.error(t('login.retryAfter', { time: formatRemainingTime(countdown, t) }));
      return;
    }

    setLoading(true);
    setLockInfo(null);
    
    try {
      await login(values.username, values.password);
      message.success(t('success.loginSuccess'));
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
            const translatedMsg = translateErrorCode(data?.code, t('error.invalidCredentials'));
            message.error(t('login.remainingAttempts', { message: translatedMsg, count: remainingAttempts }));
          } else {
            message.error(translateErrorCode(data?.code, t('error.invalidCredentials')));
          }
        } else {
          message.error(translateErrorCode(data?.code, t('login.loginFailed')));
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
        background: isDark
          ? `
          radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 217, 255, 0.1), transparent),
          radial-gradient(ellipse 60% 40% at 80% 80%, rgba(157, 78, 221, 0.08), transparent),
          linear-gradient(180deg, #0a1628 0%, #050d14 100%)
        `
          : `
          radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 217, 255, 0.15), transparent),
          radial-gradient(ellipse 60% 40% at 80% 80%, rgba(157, 78, 221, 0.12), transparent),
          linear-gradient(180deg, #e8eef5 0%, #f0f5fa 100%)
        `,
      }}
    >
      <StarfieldCanvas />

      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: isDark
            ? `
            linear-gradient(rgba(0, 217, 255, 0.02) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0, 217, 255, 0.02) 1px, transparent 1px)
          `
            : `
            linear-gradient(rgba(0, 217, 255, 0.05) 1px, transparent 1px),
            linear-gradient(90deg, rgba(0, 217, 255, 0.05) 1px, transparent 1px)
          `,
          backgroundSize: '60px 60px',
          animation: 'gridMove 25s linear infinite',
          zIndex: 0,
        }}
      />

      <div
        className="glow-orb"
        style={{
          position: 'absolute',
          top: '10%',
          left: '5%',
          width: 500,
          height: 500,
          borderRadius: '50%',
          background: isDark ? 'radial-gradient(circle, rgba(0, 217, 255, 0.12) 0%, transparent 70%)' : 'radial-gradient(circle, rgba(0, 217, 255, 0.22) 0%, transparent 70%)',
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
          background: isDark ? 'radial-gradient(circle, rgba(157, 78, 221, 0.1) 0%, transparent 70%)' : 'radial-gradient(circle, rgba(157, 78, 221, 0.18) 0%, transparent 70%)',
          filter: 'blur(80px)',
          animation: 'float 12s ease-in-out infinite reverse',
          zIndex: 0,
        }}
      />

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
        <div style={{ color: 'rgba(255, 255, 255, 0.4)' }}>{t('login.codeSnippetComment')}</div>
        <div>{`const ai = new CodeMind()`}</div>
        <div style={{ paddingLeft: 10 }}>{`.connect()`}</div>
      </div>

      {/* Top-right toolbar: language switch + theme toggle */}
      <div style={{ position: 'fixed', top: 24, right: 24, display: 'flex', gap: 12, zIndex: 100 }}>
        <Dropdown menu={{ items: languageMenuItems }} placement="bottomRight">
          <Button
            type="text"
            shape="circle"
            icon={<TranslationOutlined />}
            style={{
              width: 44,
              height: 44,
              color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)',
              background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(255, 255, 255, 0.8)',
              border: isDark ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.1)',
              backdropFilter: 'blur(10px)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          />
        </Dropdown>
        <Button
          type="text"
          shape="circle"
          icon={isDark ? <SunOutlined /> : <MoonOutlined />}
          onClick={toggleTheme}
          style={{
            width: 44,
            height: 44,
            color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.65)',
            background: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(255, 255, 255, 0.8)',
            border: isDark ? '1px solid rgba(255, 255, 255, 0.1)' : '1px solid rgba(0, 0, 0, 0.1)',
            backdropFilter: 'blur(10px)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        />
      </div>

      <div
        style={{
          width: 440,
          position: 'relative',
          zIndex: 10,
          background: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.8)',
          backdropFilter: 'blur(24px)',
          WebkitBackdropFilter: 'blur(24px)',
          borderRadius: 28,
          border: isDark ? '1px solid rgba(255, 255, 255, 0.08)' : '1px solid rgba(0, 0, 0, 0.08)',
          boxShadow: isDark
            ? `
            0 25px 80px rgba(0, 0, 0, 0.4),
            inset 0 1px 0 rgba(255, 255, 255, 0.05)
          `
            : `
            0 25px 80px rgba(0, 0, 0, 0.1),
            inset 0 1px 0 rgba(255, 255, 255, 0.5)
          `,
          padding: '48px 40px 40px',
          animation: 'fadeInUp 0.8s ease-out',
        }}
      >
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

        <div style={{ textAlign: 'center', marginBottom: 36 }}>
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
            boxShadow: isDark ? 'none' : '0 4px 20px rgba(0, 217, 255, 0.15)',
            }}
          >
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: '#00D9FF', animation: 'pulse 2s infinite' }} />
            {t('login.tagline')}
          </div>

          <h1
            key={isDark ? 'dark' : 'light'}
            style={{
              fontSize: 42,
              fontWeight: 900,
              margin: '0 0 8px',
              letterSpacing: -2,
              background: isDark
                ? 'linear-gradient(135deg, #fff 0%, #00D9FF 50%, #9D4EDD 100%)'
                : 'linear-gradient(135deg, #1a1a2e 0%, #00D9FF 40%, #9D4EDD 100%)',
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
              color: isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.5)',
              letterSpacing: 3,
              textTransform: 'uppercase',
            }}
          >
            {t('login.subtitle')}
          </p>
        </div>

        {lockInfo?.locked && (
          <Alert
            icon={<LockFilled />}
            message={t('login.accountLocked')}
            description={
              <div>
                <div style={{ fontSize: 13 }}>{t('login.lockDescription')}</div>
                {countdown > 0 && (
                  <div style={{ fontSize: 13, marginTop: 8, color: '#ffccc7' }}>
                    {t('login.remainingTime')}<strong>{formatRemainingTime(countdown, t)}</strong>
                  </div>
                )}
                <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(255,255,255,0.5)' }}>
                  {t('login.contactAdmin')}
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

        <ConfigProvider theme={getLoginTheme(isDark)}>
          <Form
            name="login"
            onFinish={handleSubmit}
            autoComplete="off"
            size="large"
          >
            <Form.Item
              name="username"
              rules={[{ required: true, message: t('login.usernameRequired') }]}
              style={{ marginBottom: 24 }}
              className="login-form-item"
            >
              <Input
                prefix={<UserOutlined className="login-input-icon" />}
                placeholder={t('login.username')}
                size="large"
                disabled={lockInfo?.locked && countdown > 0}
                className="login-input"
              />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: t('login.passwordRequired') }]}
              style={{ marginBottom: 40 }}
              className="login-form-item"
            >
              <Input.Password
                prefix={<LockOutlined className="login-input-icon" />}
                placeholder={t('login.password')}
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
                {lockInfo?.locked && countdown > 0 ? t('login.locked') : t('login.submit')}
              </Button>
            </Form.Item>
          </Form>
        </ConfigProvider>

      </div>

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
        
        .login-form-item {
          margin-bottom: 24px !important;
        }
        
        .login-form-item .ant-input-affix-wrapper {
          background: ${isDark ? '#0d1d2d' : '#ffffff'} !important;
          border: 1px solid ${isDark ? 'rgba(255, 255, 255, 0.12)' : 'rgba(0, 0, 0, 0.12)'} !important;
          border-radius: 12px !important;
          padding: 12px 16px !important;
          box-shadow: 
            inset 0 1px 0 ${isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.02)'},
            0 4px 20px rgba(0, 0, 0, ${isDark ? '0.3' : '0.1'}) !important;
          transition: all 0.3s ease !important;
        }
        
        .login-form-item .ant-input-affix-wrapper:hover {
          border-color: rgba(0, 217, 255, 0.5) !important;
          box-shadow: 
            inset 0 1px 0 ${isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.04)'},
            0 4px 24px rgba(0, 217, 255, 0.15) !important;
        }
        
        .login-form-item .ant-input-affix-wrapper:focus,
        .login-form-item .ant-input-affix-wrapper-focused {
          border-color: #00D9FF !important;
          box-shadow: 
            inset 0 1px 0 ${isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.04)'},
            0 0 0 3px rgba(0, 217, 255, 0.2),
            0 4px 24px rgba(0, 217, 255, 0.25) !important;
        }
        
        .login-input-icon {
          color: ${isDark ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.45)'} !important;
          font-size: 18px !important;
          margin-right: 12px !important;
          transition: all 0.3s ease !important;
        }
        
        .login-form-item .ant-input-affix-wrapper-focused .login-input-icon {
          color: #00D9FF !important;
        }
        
        .login-form-item .ant-input-suffix .anticon {
          color: ${isDark ? 'rgba(255, 255, 255, 0.4)' : 'rgba(0, 0, 0, 0.4)'} !important;
          font-size: 16px !important;
          transition: all 0.3s ease !important;
        }
        
        .login-form-item .ant-input-suffix .anticon:hover {
          color: ${isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.8)'} !important;
        }
        
        /* Input background and autofill styles defined in global.css */
      `}</style>
    </div>
  );
};

export default LoginPage;
