/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  // 避免 Tailwind 与 Ant Design 样式冲突
  corePlugins: {
    preflight: false,
  },
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#2B7CB3',
          light: '#4BA3D4',
          dark: '#1A5A8A',
        },
        accent: '#6BC5E8',
        'light-cyan': '#8FD8F0',
        'pale-cyan': '#D4F1F9',
        'dark-navy': '#1A3A5C',
        'medium-navy': '#2E5A7E',
        'light-bg': '#F0F5FA',
      },
      fontFamily: {
        sans: ['Inter', 'PingFang SC', 'Microsoft YaHei', 'sans-serif'],
      },
    },
  },
  plugins: [],
};
