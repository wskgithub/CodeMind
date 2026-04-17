import { describe, it, expect, beforeEach, vi } from 'vitest';

import useAuthStore from '@/store/authStore';

// Mock authService
vi.mock('@/services/authService', () => ({
  default: {
    login: vi.fn(),
    logout: vi.fn(),
    getProfile: vi.fn(),
  },
}));

describe('AuthStore', () => {
  beforeEach(() => {
    localStorage.clear();
    useAuthStore.setState({
      token: null,
      user: null,
      isAuthenticated: false,
      loading: false,
    });
  });

  it('initial state: not logged in', () => {
    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.token).toBeNull();
    expect(state.user).toBeNull();
  });

  it('restore: stays logged out when localStorage has no token', () => {
    useAuthStore.getState().restore();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('clear: clears all auth state', () => {
    useAuthStore.setState({
      token: 'test-token',
      user: { id: 1, username: 'admin', display_name: 'Admin', role: 'super_admin' },
      isAuthenticated: true,
    });

    useAuthStore.getState().clear();
    expect(useAuthStore.getState().token).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});
