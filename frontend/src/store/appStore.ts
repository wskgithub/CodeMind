import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type ThemeMode = 'dark' | 'light';

/** 应用全局 UI 状态 */
interface AppState {
  /** 侧边栏折叠状态 */
  sidebarCollapsed: boolean;
  /** 切换侧边栏 */
  toggleSidebar: () => void;
  /** 主题模式 */
  themeMode: ThemeMode;
  /** 切换主题 */
  toggleTheme: () => void;
  /** 设置主题 */
  setTheme: (mode: ThemeMode) => void;
}

const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      sidebarCollapsed: false,
      themeMode: 'dark',

      toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
      toggleTheme: () => set((s) => ({ themeMode: s.themeMode === 'dark' ? 'light' : 'dark' })),
      setTheme: (mode) => set({ themeMode: mode }),
    }),
    {
      name: 'codemind-app-storage',
      partialize: (state) => ({ themeMode: state.themeMode }),
    }
  )
);

export default useAppStore;
