import { create } from 'zustand';

/** 应用全局 UI 状态 */
interface AppState {
  /** 侧边栏折叠状态 */
  sidebarCollapsed: boolean;
  /** 暗色模式 */
  darkMode: boolean;
  /** 切换侧边栏 */
  toggleSidebar: () => void;
  /** 切换暗色模式 */
  toggleDarkMode: () => void;
  /** 设置暗色模式 */
  setDarkMode: (dark: boolean) => void;
}

const useAppStore = create<AppState>((set) => ({
  sidebarCollapsed: false,
  darkMode: localStorage.getItem('codemind_dark_mode') === 'true',

  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),

  toggleDarkMode: () =>
    set((s) => {
      const next = !s.darkMode;
      localStorage.setItem('codemind_dark_mode', String(next));
      return { darkMode: next };
    }),

  setDarkMode: (dark: boolean) => {
    localStorage.setItem('codemind_dark_mode', String(dark));
    set({ darkMode: dark });
  },
}));

export default useAppStore;
