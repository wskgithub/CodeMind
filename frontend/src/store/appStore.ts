import { create } from 'zustand';
import { persist } from 'zustand/middleware';

import { changeLanguage, type SupportedLanguage } from '@/i18n';

export type ThemeMode = 'dark' | 'light';

interface AppState {
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
  themeMode: ThemeMode;
  toggleTheme: () => void;
  setTheme: (mode: ThemeMode) => void;
  language: SupportedLanguage;
  setLanguage: (lang: SupportedLanguage) => void;
}

const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      sidebarCollapsed: false,
      themeMode: 'dark',
      language: 'zh-CN',

      toggleSidebar: () => set((s) => ({ sidebarCollapsed: !s.sidebarCollapsed })),
      toggleTheme: () => set((s) => ({ themeMode: s.themeMode === 'dark' ? 'light' : 'dark' })),
      setTheme: (mode) => set({ themeMode: mode }),
      setLanguage: (lang) => {
        changeLanguage(lang);
        set({ language: lang });
      },
    }),
    {
      name: 'codemind-app-storage',
      partialize: (state) => ({ themeMode: state.themeMode, language: state.language }),
    }
  )
);

export default useAppStore;
