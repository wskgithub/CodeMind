import { useEffect } from 'react';
import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import router from '@/router';
import useAuthStore from '@/store/authStore';

/** Ant Design 品牌主题 — 与首页/登录页新设计风格统一 */
const brandToken = {
  // 主色 - 使用新设计的青色
  colorPrimary: '#00D9FF',
  colorInfo: '#9D4EDD',
  colorSuccess: '#00F5D4',
  colorWarning: '#FFBE0B',
  colorError: '#FF6B6B',
  // 背景色 - 深色主题
  colorBgLayout: '#050d14',
  colorBgContainer: 'rgba(255, 255, 255, 0.02)',
  colorBgElevated: 'rgba(13, 29, 45, 0.95)',
  // 文字颜色
  colorText: 'rgba(255, 255, 255, 0.9)',
  colorTextSecondary: 'rgba(255, 255, 255, 0.7)',
  colorTextTertiary: 'rgba(255, 255, 255, 0.45)',
  // 边框和圆角
  borderRadius: 12,
  borderRadiusLG: 24,
  fontFamily: "'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif",
  wireframe: false,
};

const App: React.FC = () => {
  const restore = useAuthStore((s) => s.restore);

  // 应用启动时恢复登录态
  useEffect(() => {
    restore();
  }, [restore]);

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: brandToken,
        algorithm: theme.darkAlgorithm,
        components: {
          Card: {
            borderRadiusLG: 24,
            colorBgContainer: 'rgba(255, 255, 255, 0.02)',
          },
          Button: {
            borderRadius: 12,
            borderRadiusLG: 14,
            controlHeight: 40,
            controlHeightLG: 48,
            primaryShadow: '0 4px 16px rgba(0, 217, 255, 0.25)',
          },
          Input: {
            borderRadius: 12,
            borderRadiusLG: 14,
            controlHeight: 42,
            controlHeightLG: 48,
            colorBgContainer: 'rgba(255, 255, 255, 0.03)',
            colorBorder: 'rgba(255, 255, 255, 0.08)',
            activeBorderColor: '#00D9FF',
            hoverBorderColor: 'rgba(0, 217, 255, 0.4)',
            activeShadow: '0 0 0 3px rgba(0, 217, 255, 0.15)',
          },
          Table: {
            borderRadiusLG: 16,
            colorBgContainer: 'transparent',
            headerBg: 'rgba(255, 255, 255, 0.03)',
            headerColor: 'rgba(255, 255, 255, 0.9)',
            rowHoverBg: 'rgba(0, 217, 255, 0.05)',
          },
          Modal: {
            borderRadiusLG: 24,
            colorBgElevated: 'rgba(13, 29, 45, 0.98)',
          },
          Menu: {
            itemBorderRadius: 12,
            itemMarginInline: 12,
            itemMarginBlock: 4,
            colorItemText: 'rgba(255, 255, 255, 0.7)',
            colorItemTextHover: '#00D9FF',
            colorItemBgHover: 'rgba(0, 217, 255, 0.08)',
            colorItemTextSelected: '#00D9FF',
            colorItemBgSelected: 'rgba(0, 217, 255, 0.12)',
          },
          Select: {
            borderRadius: 12,
            colorBgContainer: 'rgba(255, 255, 255, 0.03)',
          },
          Tag: {
            borderRadius: 6,
          },
          Statistic: {
            colorTextHeading: 'rgba(255, 255, 255, 0.7)',
          },
        },
      }}
    >
      <AntApp>
        <RouterProvider router={router} />
      </AntApp>
    </ConfigProvider>
  );
};

export default App;
