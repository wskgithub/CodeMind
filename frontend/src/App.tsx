import { ConfigProvider, App as AntApp, theme } from 'antd';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import { useEffect, useMemo } from 'react';
import { RouterProvider } from 'react-router-dom';

import { changeLanguage } from '@/i18n';
import router from '@/router';
import useAppStore, { ThemeMode } from '@/store/appStore';
import useAuthStore from '@/store/authStore';

// Ant Design theme tokens for light/dark mode
const getThemeTokens = (themeMode: ThemeMode) => {
  const isDark = themeMode === 'dark';
  
  return {
    colorPrimary: '#00D9FF',
    colorInfo: '#9D4EDD',
    colorSuccess: '#00F5D4',
    colorWarning: '#FFBE0B',
    colorError: '#FF6B6B',
    colorBgLayout: isDark ? '#050d14' : '#f0f5fa',
    colorBgContainer: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.8)',
    colorBgElevated: isDark ? 'rgba(13, 29, 45, 0.95)' : 'rgba(255, 255, 255, 0.98)',
    colorText: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
    colorTextSecondary: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)',
    colorTextTertiary: isDark ? 'rgba(255, 255, 255, 0.45)' : 'rgba(0, 0, 0, 0.45)',
    borderRadius: 12,
    borderRadiusLG: 24,
    fontFamily: "'Inter', 'PingFang SC', 'Microsoft YaHei', sans-serif",
    wireframe: false,
  };
};

// Ant Design component-level theme tokens
const getComponentTokens = (themeMode: ThemeMode) => {
  const isDark = themeMode === 'dark';
  
  return {
    Card: {
      borderRadiusLG: 24,
      colorBgContainer: isDark ? 'rgba(255, 255, 255, 0.02)' : 'rgba(255, 255, 255, 0.8)',
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
      colorBgContainer: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
      colorBorder: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)',
      activeBorderColor: '#00D9FF',
      hoverBorderColor: 'rgba(0, 217, 255, 0.4)',
      activeShadow: '0 0 0 3px rgba(0, 217, 255, 0.15)',
      colorText: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
    },
    Table: {
      borderRadiusLG: 16,
      colorBgContainer: 'transparent',
      headerBg: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
      headerColor: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
      rowHoverBg: 'rgba(0, 217, 255, 0.05)',
      colorText: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
    },
    Modal: {
      borderRadiusLG: 24,
      colorBgElevated: isDark ? 'rgba(13, 29, 45, 0.98)' : 'rgba(255, 255, 255, 0.98)',
    },
    Menu: {
      itemBorderRadius: 12,
      itemMarginInline: 12,
      itemMarginBlock: 4,
      colorItemText: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)',
      colorItemTextHover: '#00D9FF',
      colorItemBgHover: 'rgba(0, 217, 255, 0.08)',
      colorItemTextSelected: '#00D9FF',
      colorItemBgSelected: 'rgba(0, 217, 255, 0.12)',
    },
    InputNumber: {
      borderRadius: 12,
      borderRadiusLG: 14,
      controlHeight: 42,
      controlHeightLG: 48,
      colorBgContainer: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
      colorBorder: isDark ? 'rgba(255, 255, 255, 0.08)' : 'rgba(0, 0, 0, 0.08)',
      activeBorderColor: '#00D9FF',
      hoverBorderColor: 'rgba(0, 217, 255, 0.4)',
      activeShadow: '0 0 0 3px rgba(0, 217, 255, 0.15)',
      colorText: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
    },
    Select: {
      borderRadius: 12,
      colorBgContainer: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
      colorText: isDark ? 'rgba(255, 255, 255, 0.9)' : 'rgba(0, 0, 0, 0.85)',
    },
    Tag: {
      borderRadius: 6,
    },
    Statistic: {
      colorTextHeading: isDark ? 'rgba(255, 255, 255, 0.7)' : 'rgba(0, 0, 0, 0.65)',
    },
  };
};

const App: React.FC = () => {
  const restore = useAuthStore((s) => s.restore);
  const isRestored = useAuthStore((s) => s.isRestored);
  const themeMode = useAppStore((s) => s.themeMode);
  const language = useAppStore((s) => s.language);

  // restore auth state on app start
  useEffect(() => {
    restore();
  }, [restore]);

  // sync theme to document root for CSS variables
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', themeMode);
  }, [themeMode]);

  // sync language on app start (restore from storage)
  useEffect(() => {
    changeLanguage(language);
    document.documentElement.setAttribute('lang', language);
  }, [language]);

  const themeTokens = useMemo(() => getThemeTokens(themeMode), [themeMode]);
  const componentTokens = useMemo(() => getComponentTokens(themeMode), [themeMode]);
  const antdLocale = useMemo(() => (language === 'en-US' ? enUS : zhCN), [language]);

  // wait for auth state restoration before rendering routes
  if (!isRestored) {
    return null;
  }

  return (
    <ConfigProvider
      locale={antdLocale}
      theme={{
        token: themeTokens,
        algorithm: themeMode === 'dark' ? theme.darkAlgorithm : theme.defaultAlgorithm,
        components: componentTokens,
      }}
    >
      <AntApp>
        <RouterProvider router={router} />
      </AntApp>
    </ConfigProvider>
  );
};

export default App;
