import { describe, it, expect, beforeEach } from 'vitest';
import useAppStore from '@/store/appStore';

describe('AppStore', () => {
  beforeEach(() => {
    useAppStore.setState({ sidebarCollapsed: false });
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
});
