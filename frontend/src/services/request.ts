import { message } from 'antd';
import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios';

import i18n from '@/i18n';
import type { ApiResponse } from '@/types';

const ERROR_CODE_I18N_MAP: Record<number, string> = {
  40001: 'error.invalidCredentials',
  40002: 'error.tokenExpired',
  40003: 'error.tokenInvalid',
  40004: 'error.accountDisabled',
  40005: 'error.apiKeyInvalid',
  40006: 'error.apiKeyExpired',
  40007: 'error.apiKeyDisabled',
  40008: 'error.accountLocked',
  40101: 'error.forbidden',
  40102: 'error.forbiddenUser',
  40103: 'error.forbiddenDept',
  40201: 'error.invalidParams',
  40202: 'error.missingParams',
  40301: 'error.usernameExists',
  40302: 'error.emailExists',
  40303: 'error.deptNotFound',
  40304: 'error.apiKeyLimit',
  40305: 'error.deptHasUsers',
  40306: 'error.userNotFound',
  40307: 'error.oldPasswordWrong',
  40308: 'error.apiKeyNotFound',
  40309: 'error.recordNotFound',
  40310: 'error.providerNameExists',
  40311: 'error.providerTemplateNameExists',
  40312: 'error.apiKeyNotCopyable',
  42901: 'error.tokenQuotaExceeded',
  42902: 'error.concurrencyExceeded',
  42903: 'error.rateLimitExceeded',
  50001: 'error.internal',
  50002: 'error.llmUnavailable',
  50003: 'error.database',
};

export function translateErrorCode(code: number | undefined, fallback?: string): string {
  if (code !== undefined) {
    const key = ERROR_CODE_I18N_MAP[code];
    if (key) return i18n.t(key);
  }
  return fallback || i18n.t('error.requestFailed');
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

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

request.interceptors.response.use(
  (response) => {
    if (response.config.responseType === 'blob') {
      return response;
    }
    const data = response.data as ApiResponse;
    if (data.code !== 0) {
      message.error(translateErrorCode(data.code));
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
          message.error(translateErrorCode(data?.code, i18n.t('error.invalidCredentials')));
        } else {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          message.error(i18n.t('error.tokenExpired'));
          window.location.href = '/login';
        }
        break;
      case 403:
        message.error(translateErrorCode(data?.code, i18n.t('error.forbidden')));
        break;
      case 429:
        message.error(translateErrorCode(data?.code, i18n.t('error.tooManyRequests')));
        break;
      default:
        message.error(translateErrorCode(data?.code, i18n.t('error.network')));
    }

    return Promise.reject(error);
  },
);

export default request;
