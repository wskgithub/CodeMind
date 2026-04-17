import { message } from 'antd';
import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';

import type { ApiResponse } from '@/types';

/** 创建 Axios 实例 */
const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

/** 请求拦截器：自动附加 JWT Token */
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

/** 响应拦截器：统一错误处理 */
request.interceptors.response.use(
  (response) => {
    // 如果是 blob 响应（文件下载），直接返回
    if (response.config.responseType === 'blob') {
      return response;
    }
    const data = response.data as ApiResponse;
    // 业务错误码非零表示业务异常
    if (data.code !== 0) {
      message.error(data.message || '请求失败');
      return Promise.reject(new Error(data.message));
    }
    return response;
  },
  (error: AxiosError<ApiResponse>) => {
    const status = error.response?.status;
    const data = error.response?.data;

    switch (status) {
      case 401:
        // 区分登录失败和 Token 过期
        if (window.location.pathname === '/login') {
          // 登录页面的 401 错误（用户名或密码错误）
          message.error(data?.message || '用户名或密码错误');
        } else {
          // 其他页面的 401 错误（Token 无效或过期）
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          message.error('登录已过期，请重新登录');
          window.location.href = '/login';
        }
        break;
      case 403:
        message.error(data?.message || '无权访问该资源');
        break;
      case 429:
        message.error(data?.message || '请求过于频繁');
        break;
      default:
        message.error(data?.message || '网络异常，请稍后重试');
    }

    return Promise.reject(error);
  },
);

export default request;
