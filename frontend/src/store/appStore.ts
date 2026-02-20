import { create } from 'zustand';

/** 应用全局 UI 状态 */
interface AppState {
  /** 侧边栏折叠状态 */
  sidebarCollapsed: boolean;
  /** 切换侧边栏 */
  toggleSidebar: () => void;
}

const useAppStore = create<AppState>((set) => ({
  sidebarCollapsed: false,

  toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
}));

export default useAppStore;
