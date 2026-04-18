import { create } from 'zustand';

import authService from '@/services/authService';
import type { UserBrief, UserDetail } from '@/types';

interface AuthState {
  token: string | null;
  user: UserBrief | null;
  isAuthenticated: boolean;
  loading: boolean;
  isRestored: boolean;

  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchProfile: () => Promise<UserDetail>;
  restore: () => void;
  clear: () => void;
}

const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isAuthenticated: false,
  loading: false,
  isRestored: false,

  login: async (username: string, password: string) => {
    set({ loading: true });
    try {
      const resp = await authService.login({ username, password });
      const { token, user } = resp.data.data;

      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));

      set({
        token,
        user,
        isAuthenticated: true,
        loading: false,
      });
    } catch (error) {
      set({ loading: false });
      throw error;
    }
  },

  logout: async () => {
    try {
      await authService.logout();
    } catch {
      // logout API failure shouldn't block frontend
    } finally {
      get().clear();
    }
  },

  fetchProfile: async () => {
    const resp = await authService.getProfile();
    const profile = resp.data.data;

    set({
      user: {
        id: profile.id,
        username: profile.username,
        display_name: profile.display_name,
        role: profile.role,
        department: profile.department,
      },
    });

    return profile;
  },

  restore: () => {
    const token = localStorage.getItem('token');
    const userStr = localStorage.getItem('user');

    if (token && userStr) {
      try {
        const user = JSON.parse(userStr) as UserBrief;
        set({ token, user, isAuthenticated: true, isRestored: true });
      } catch {
        // clear corrupted data
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        set({ isRestored: true });
      }
    } else {
      set({ isRestored: true });
    }
  },

  clear: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    set({ token: null, user: null, isAuthenticated: false });
  },
}));

export default useAuthStore;
