import React from 'react';
import ReactDOM from 'react-dom/client';

import App from './App';
import './i18n'; // Initialize i18n (must be before App)
import './assets/styles/global.css';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
