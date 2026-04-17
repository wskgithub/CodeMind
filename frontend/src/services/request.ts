import { message } from 'antd';
import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';

import i18n from '@/i18n';
import type { ApiResponse } from '@/types';

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// attach JWT token to requests
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// unified error handling
request.interceptors.response.use(
  (response) => {
    // pass through blob responses (file downloads)
    if (response.config.responseType === 'blob') {
      return response;
    }
    const data = response.data as ApiResponse;
    if (data.code !== 0) {
      message.error(data.message || i18n.t('error.requestFailed'));
      return Promise.reject(new Error(data.message));
    }
    return response;
  },
  (error: AxiosError<ApiResponse>) => {
    const status = error.response?.status;
    const data = error.response?.data;

    switch (status) {
      case 401:
        if (window.location.pathname === '/login') {
          message.error(data?.message || i18n.t('error.invalidCredentials'));
        } else {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          message.error(i18n.t('error.tokenExpired'));
          window.location.href = '/login';
        }
        break;
      case 403:
        message.error(data?.message || i18n.t('error.forbidden'));
        break;
      case 429:
        message.error(data?.message || i18n.t('error.tooManyRequests'));
        break;
      default:
        message.error(data?.message || i18n.t('error.network'));
    }

    return Promise.reject(error);
  },
);

export default request;
