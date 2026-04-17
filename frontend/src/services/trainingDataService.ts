import request from './request';

import type {
  ApiResponse,
  PageData,
  TrainingDataItem,
  TrainingDataDetail,
  TrainingDataStats,
} from '@/types';

/** 训练数据列表查询参数 */
export interface TrainingDataListParams {
  page?: number;
  page_size?: number;
  model?: string;
  request_type?: string;
  user_id?: number;
  start_date?: string;
  end_date?: string;
  is_excluded?: boolean;
}

/** 训练数据导出查询参数 */
export interface TrainingDataExportParams {
  model?: string;
  request_type?: string;
  start_date?: string;
  end_date?: string;
}

/** 获取训练数据列表 */
export function listTrainingData(params: TrainingDataListParams) {
  return request.get<ApiResponse<PageData<TrainingDataItem>>>('/training-data', { params });
}

/** 获取训练数据详情 */
export function getTrainingDataDetail(id: number) {
  return request.get<ApiResponse<TrainingDataDetail>>(`/training-data/${id}`);
}

/** 更新训练数据排除状态 */
export function updateExcluded(id: number, excluded: boolean) {
  return request.put<ApiResponse<null>>(`/training-data/${id}/exclude`, { excluded });
}

/** 获取训练数据统计 */
export function getTrainingDataStats() {
  return request.get<ApiResponse<TrainingDataStats>>('/training-data/stats');
}

/** 导出训练数据为 JSONL */
export async function exportTrainingData(params: TrainingDataExportParams) {
  const response = await request.post('/training-data/export', null, {
    params,
    responseType: 'blob',
  });

  const disposition = response.headers['content-disposition'];
  let filename = 'training_data.jsonl';
  if (disposition) {
    const match = disposition.match(/filename=(.+)/);
    if (match) filename = match[1];
  }

  const url = window.URL.createObjectURL(new Blob([response.data]));
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}
