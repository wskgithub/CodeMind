import request from './request';

export interface UploadResult {
  url: string;
  filename: string;
}

class UploadService {
  /** Upload a document image */
  async uploadImage(file: File): Promise<UploadResult> {
    const formData = new FormData();
    formData.append('file', file);

    const response = await request.post('/docs/admin/upload/image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
    });

    return response.data?.data;
  }
}

export const uploadService = new UploadService();
