import { useState, useEffect, useCallback } from 'react';
import { Form, Input, Button, message, Alert } from 'antd';
import { GlobalOutlined, CopyOutlined, SaveOutlined } from '@ant-design/icons';
import { getConfigs, updateConfigs } from '@/services/systemService';
import type { SystemConfig } from '@/types';
import useAppStore from '@/store/appStore';
import { copyToClipboard } from '@/utils/copy';

const PageIcon = ({ icon }: { icon: React.ReactNode }) => (
  <span
    className="flex items-center justify-center w-12 h-12 rounded-2xl shrink-0"
    style={{
      background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
      color: '#fff',
      fontSize: 22,
      boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
    }}
  >
    {icon}
  </span>
);

const PlatformSettingsPage: React.FC = () => {
  const themeMode = useAppStore((s) => s.themeMode);
  const isDark = themeMode === 'dark';

  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [serviceUrl, setServiceUrl] = useState('');

  const loadConfig = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getConfigs();
      const configs: SystemConfig[] = resp.data.data || [];
      const cfg = configs.find(c => c.config_key === 'platform.service_url');
      const url = cfg?.config_value || '';
      form.setFieldsValue({ service_url: url });
      setServiceUrl(url);
    } catch { /* 拦截器处理 */ }
    finally { setLoading(false); }
  }, [form]);

  useEffect(() => { loadConfig(); }, [loadConfig]);

  const handleSave = async (values: { service_url: string }) => {
    const url = (values.service_url || '').replace(/\/+$/, '');
    setSaving(true);
    try {
      await updateConfigs([{ key: 'platform.service_url', value: url }]);
      setServiceUrl(url);
      message.success('保存成功');
    } catch { /* 拦截器处理 */ }
    finally { setSaving(false); }
  };

  const handleCopy = async (text: string) => {
    const ok = await copyToClipboard(text);
    if (ok) message.success('已复制到剪贴板');
    else message.error('复制失败，请手动复制');
  };

  const openaiURL = serviceUrl ? `${serviceUrl}/api/openai/v1` : '';
  const anthropicURL = serviceUrl ? `${serviceUrl}/api/anthropic` : '';

  const labelColor = isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.7)';
  const inputStyle = {
    background: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
    borderColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)',
    color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)',
  };
  const codeStyle = {
    padding: '8px 14px', borderRadius: 8, fontSize: 14,
    background: isDark ? 'rgba(255, 255, 255, 0.04)' : 'rgba(0, 0, 0, 0.03)',
    color: isDark ? 'rgba(255, 255, 255, 0.8)' : 'rgba(0, 0, 0, 0.75)',
    fontFamily: 'monospace', wordBreak: 'break-all' as const,
  };

  return (
    <div className="page-bg">
      <div className="animate-fade-in-up" style={{ position: 'relative', zIndex: 1 }}>
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 8 }}>
            <PageIcon icon={<GlobalOutlined />} />
            <div>
              <h2 style={{ margin: 0, color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 24, fontWeight: 600 }}>
                平台配置
              </h2>
              <p style={{ margin: 0, color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.5)', fontSize: 14, marginTop: 4 }}>
                配置 CodeMind 平台对外服务地址，用户将使用此地址接入 LLM 服务
              </p>
            </div>
          </div>
        </div>

        <div className="glass-card animate-fade-in-up" style={{ padding: 24, animationDelay: '0.05s', maxWidth: 720 }}>
          <h3 style={{ margin: '0 0 20px 0', color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 16, fontWeight: 600 }}>
            服务地址
          </h3>

          <Form form={form} layout="vertical" onFinish={handleSave}>
            <Form.Item
              name="service_url"
              label={<span style={{ color: labelColor }}>平台服务 URL</span>}
              rules={[
                { required: true, message: '请输入平台服务 URL' },
                { pattern: /^https?:\/\/.+/, message: '请输入有效的 URL（以 http:// 或 https:// 开头）' },
              ]}
              extra={<span style={{ color: isDark ? 'rgba(255, 255, 255, 0.35)' : 'rgba(0, 0, 0, 0.35)', fontSize: 12 }}>
                用户客户端接入时使用的访问地址，例如 https://codemind.example.com（不含末尾斜杠）
              </span>}
            >
              <Input
                placeholder="https://codemind.example.com"
                style={{ ...inputStyle, height: 44, fontSize: 15 }}
                disabled={loading}
              />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SaveOutlined />}
                loading={saving}
                style={{
                  height: 44, borderRadius: 12,
                  background: 'linear-gradient(135deg, #00D9FF 0%, #9D4EDD 100%)',
                  border: 'none', boxShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
                }}
              >
                保存
              </Button>
            </Form.Item>
          </Form>

          {serviceUrl && (
            <>
              <div style={{ marginTop: 8, marginBottom: 16, height: 1, background: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.06)' }} />

              <h3 style={{ margin: '0 0 16px 0', color: isDark ? '#fff' : 'rgba(0, 0, 0, 0.85)', fontSize: 16, fontWeight: 600 }}>
                协议端点
              </h3>

              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {[
                  { label: 'OpenAI Base URL', url: openaiURL, color: '#00F5D4' },
                  { label: 'Anthropic Base URL', url: anthropicURL, color: '#9D4EDD' },
                ].map(item => (
                  <div key={item.label}>
                    <div style={{ color: item.color, fontSize: 12, fontWeight: 600, marginBottom: 4 }}>{item.label}</div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                      <code style={{ ...codeStyle, flex: 1 }}>{item.url}</code>
                      <Button
                        type="text"
                        icon={<CopyOutlined />}
                        onClick={() => handleCopy(item.url)}
                        style={{ color: isDark ? 'rgba(255, 255, 255, 0.5)' : 'rgba(0, 0, 0, 0.4)' }}
                      />
                    </div>
                  </div>
                ))}
              </div>

              <Alert
                type="info"
                showIcon
                style={{ marginTop: 20, borderRadius: 8, background: isDark ? 'rgba(0, 217, 255, 0.06)' : 'rgba(0, 217, 255, 0.04)' }}
                message="用户在编程工具中配置时，根据所使用的协议选择对应的 Base URL 即可接入。"
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default PlatformSettingsPage;
