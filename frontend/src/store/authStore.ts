import { create } from 'zustand';
import type { UserBrief, UserDetail } from '@/types';
import authService from '@/services/authService';

/** 认证状态 */
interface AuthState {
  token: string | null;
  user: UserBrief | null;
  isAuthenticated: boolean;
  loading: boolean;
  /** 是否已完成从本地存储恢复 */
  isRestored: boolean;

  /** 用户登录 */
  login: (username: string, password: string) => Promise<void>;
  /** 用户登出 */
  logout: () => Promise<void>;
  /** 获取用户信息 */
  fetchProfile: () => Promise<UserDetail>;
  /** 从本地存储恢复登录态 */
  restore: () => void;
  /** 清除登录状态 */
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

      // 持久化到 localStorage
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
      // 向上抛出错误，让调用方处理（错误消息已在 request 拦截器中显示）
      throw error;
    }
  },

  logout: async () => {
    try {
      await authService.logout();
    } catch {
      // 登出 API 失败不阻塞前端操作
    } finally {
      get().clear();
    }
  },

  fetchProfile: async () => {
    const resp = await authService.getProfile();
    const profile = resp.data.data;

    // 更新 store 中的用户信息
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
        // JSON 解析失败，清除脏数据
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
