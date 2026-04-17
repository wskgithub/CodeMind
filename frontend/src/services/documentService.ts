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
  /** Get list of published documents */
  async list(): Promise<DocumentListItem[]> {
    const response = await request.get('/docs');
    return response.data?.data || [];
  }

  /** Get document details by slug */
  async getBySlug(slug: string): Promise<Document | null> {
    const response = await request.get(`/docs/${slug}`);
    return response.data?.data || null;
  }

  /** Get all documents (admin API) */
  async listAll(): Promise<Document[]> {
    const response = await request.get('/docs/admin');
    return response.data?.data || [];
  }

  /** Get document details by ID (admin API) */
  async getById(id: number): Promise<Document | null> {
    const response = await request.get(`/docs/admin/${id}`);
    return response.data?.data || null;
  }

  /** Create a document */
  async create(data: CreateDocumentRequest): Promise<Document> {
    const response = await request.post('/docs/admin', data);
    return response.data?.data;
  }

  /** Update a document */
  async update(id: number, data: UpdateDocumentRequest): Promise<Document> {
    const response = await request.put(`/docs/admin/${id}`, data);
    return response.data?.data;
  }

  /** Delete a document */
  async delete(id: number): Promise<void> {
    await request.delete(`/docs/admin/${id}`);
  }
}

export const documentService = new DocumentService();
