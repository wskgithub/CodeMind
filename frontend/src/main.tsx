import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './i18n'; // 初始化 i18n（必须在 App 之前）
import './assets/styles/global.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
