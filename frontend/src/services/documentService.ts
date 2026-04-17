import request from './request';

export interface Document {
  id: number;
  slug: string;
  title: string;
  subtitle: string;
  icon: string;
  content: string;
  sort_order: number;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export interface DocumentListItem {
  id: number;
  slug: string;
  title: string;
  subtitle: string;
  icon: string;
  sort_order: number;
  is_published: boolean;
  updated_at: string;
}

export interface CreateDocumentRequest {
  slug: string;
  title: string;
  subtitle?: string;
  icon?: string;
  content: string;
  sort_order?: number;
  is_published?: boolean;
}

export interface UpdateDocumentRequest {
  title: string;
  subtitle?: string;
  icon?: string;
  content: string;
  sort_order?: number;
  is_published?: boolean;
}

class DocumentService {
  /** 获取已发布的文档列表 */
  async list(): Promise<DocumentListItem[]> {
    const response = await request.get('/docs');
    return response.data?.data || [];
  }

  /** 根据 slug 获取文档详情 */
  async getBySlug(slug: string): Promise<Document | null> {
    const response = await request.get(`/docs/${slug}`);
    return response.data?.data || null;
  }

  /** 获取所有文档（管理接口） */
  async listAll(): Promise<Document[]> {
    const response = await request.get('/docs/admin');
    return response.data?.data || [];
  }

  /** 根据 ID 获取文档详情（管理接口） */
  async getById(id: number): Promise<Document | null> {
    const response = await request.get(`/docs/admin/${id}`);
    return response.data?.data || null;
  }

  /** 创建文档 */
  async create(data: CreateDocumentRequest): Promise<Document> {
    const response = await request.post('/docs/admin', data);
    return response.data?.data;
  }

  /** 更新文档 */
  async update(id: number, data: UpdateDocumentRequest): Promise<Document> {
    const response = await request.put(`/docs/admin/${id}`, data);
    return response.data?.data;
  }

  /** 删除文档 */
  async delete(id: number): Promise<void> {
    await request.delete(`/docs/admin/${id}`);
  }
}

export const documentService = new DocumentService();
