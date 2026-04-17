import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import zhCN from './locales/zh-CN';
import enUS from './locales/en-US';

// 支持的语言列表（用于 UI 切换器与校验）
export const SUPPORTED_LANGUAGES = [
  { code: 'zh-CN', name: '简体中文', nativeName: '简体中文' },
  { code: 'en-US', name: 'English', nativeName: 'English' },
] as const;

export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number]['code'];

const LANG_STORAGE_KEY = 'codemind-language';

const resources = {
  'zh-CN': { translation: zhCN },
  'en-US': { translation: enUS },
};

// 从语言标签推导回退语言（例如 zh 或 zh-HK -> zh-CN）
const normalizeLanguage = (lang: string | undefined): SupportedLanguage => {
  if (!lang) return 'zh-CN';
  if (lang.toLowerCase().startsWith('zh')) return 'zh-CN';
  if (lang.toLowerCase().startsWith('en')) return 'en-US';
  return 'zh-CN';
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'zh-CN',
    supportedLngs: SUPPORTED_LANGUAGES.map((l) => l.code),
    load: 'currentOnly',
    interpolation: {
      escapeValue: false, // React 已内置 XSS 防御
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: LANG_STORAGE_KEY,
    },
    returnNull: false,
  })
  .catch(() => {
    // 初始化失败静默处理，回退到默认语言
  });

// 确保 i18n 初始语言符合支持列表
const initialLang = normalizeLanguage(i18n.language);
if (initialLang !== i18n.language) {
  i18n.changeLanguage(initialLang);
}

export const changeLanguage = (lang: SupportedLanguage) => {
  i18n.changeLanguage(lang);
  localStorage.setItem(LANG_STORAGE_KEY, lang);
  document.documentElement.setAttribute('lang', lang);
};

export default i18n;
