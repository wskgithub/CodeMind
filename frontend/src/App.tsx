import { useEffect } from 'react';
import { RouterProvider } from 'react-router-dom';
import { ConfigProvider, App as AntApp, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import router from '@/router';
import useAuthStore from '@/store/authStore';
import useAppStore from '@/store/appStore';

/** Ant Design 品牌主题 — Glassmorphism 风格 */
const brandToken = {
  colorPrimary: '#2B7CB3',
  colorInfo: '#4BA3D4',
  colorSuccess: '#52C41A',
  colorWarning: '#FAAD14',
  colorError: '#FF4D4F',
  borderRadius: 12,
  fontFamily: "'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif",
  wireframe: false,
};

const App: React.FC = () => {
  const restore = useAuthStore((s) => s.restore);
  const darkMode = useAppStore((s) => s.darkMode);

  // 应用启动时恢复登录态
  useEffect(() => {
    restore();
  }, [restore]);

  // 暗色模式 class 控制
  useEffect(() => {
    document.documentElement.classList.toggle('dark', darkMode);
  }, [darkMode]);

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          ...brandToken,
          colorBgLayout: darkMode ? '#0f0f13' : '#f0f4f8',
          colorBgContainer: darkMode ? 'rgba(30, 30, 36, 0.72)' : 'rgba(255, 255, 255, 0.72)',
          colorBgElevated: darkMode ? 'rgba(36, 36, 42, 0.92)' : 'rgba(255, 255, 255, 0.92)',
        },
        algorithm: darkMode ? theme.darkAlgorithm : theme.defaultAlgorithm,
        components: {
          Card: {
            borderRadiusLG: 16,
          },
          Button: {
            borderRadius: 10,
            controlHeight: 40,
          },
          Input: {
            borderRadius: 10,
            controlHeight: 42,
          },
          Table: {
            borderRadiusLG: 12,
          },
          Modal: {
            borderRadiusLG: 16,
          },
          Menu: {
            itemBorderRadius: 8,
            itemMarginInline: 8,
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
