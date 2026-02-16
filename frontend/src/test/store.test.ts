import { describe, it, expect, beforeEach } from 'vitest';
import useAppStore from '@/store/appStore';

describe('AppStore', () => {
  beforeEach(() => {
    // 重置状态
    useAppStore.setState({ sidebarCollapsed: false, darkMode: false });
  });

  it('初始状态正确', () => {
    const state = useAppStore.getState();
    expect(state.sidebarCollapsed).toBe(false);
  });

  it('切换侧边栏状态', () => {
    useAppStore.getState().toggleSidebar();
    expect(useAppStore.getState().sidebarCollapsed).toBe(true);

    useAppStore.getState().toggleSidebar();
    expect(useAppStore.getState().sidebarCollapsed).toBe(false);
  });

  it('切换暗色模式', () => {
    useAppStore.getState().toggleDarkMode();
    expect(useAppStore.getState().darkMode).toBe(true);

    useAppStore.getState().toggleDarkMode();
    expect(useAppStore.getState().darkMode).toBe(false);
  });

  it('设置暗色模式', () => {
    useAppStore.getState().setDarkMode(true);
    expect(useAppStore.getState().darkMode).toBe(true);

    useAppStore.getState().setDarkMode(false);
    expect(useAppStore.getState().darkMode).toBe(false);
  });
});
