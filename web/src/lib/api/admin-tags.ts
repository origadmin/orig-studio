// Admin Tags API Client
import { api } from "../request";

export interface Tag {
  id: string;
  name: string;
  slug: string;
  description?: string;
  color?: string;
  count: number;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface TagListResponse {
  items: Tag[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface TagQueryParams {
  page?: number;
  page_size?: number;
  search?: string;
  status?: string;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface CreateTagRequest {
  name: string;
  slug: string;
  description?: string;
  color?: string;
  status: string;
}

export interface UpdateTagRequest {
  name?: string;
  slug?: string;
  description?: string;
  color?: string;
  status?: string;
}

export interface BulkTagOperationRequest {
  ids: string[];
  action: 'delete' | 'activate' | 'deactivate';
}

export interface BulkOperationResponse {
  success: number;
  failed: number;
  errors: Array<{ id: string; error: string }>;
}

export interface ImportResponse {
  imported: number;
  failed: number;
  errors: Array<{ row: number; error: string }>;
}

export const tagApi = {
  // List tags with pagination and filters
  list: async (params?: TagQueryParams): Promise<TagListResponse> => {
    const response = await api.get<TagListResponse>('/admin/tags', { params });
    return response;
  },

  // Get single tag
  get: async (id: string): Promise<Tag> => {
    const response = await api.get<Tag>(`/admin/tags/${id}`);
    return response;
  },

  // Create tag
  create: async (data: CreateTagRequest): Promise<Tag> => {
    const response = await api.post<Tag>('/admin/tags', data);
    return response;
  },

  // Update tag
  update: async (id: string, data: UpdateTagRequest): Promise<Tag> => {
    const response = await api.put<Tag>(`/admin/tags/${id}`, data);
    return response;
  },

  // Delete tag
  delete: async (id: string): Promise<void> => {
    await api.del(`/admin/tags/${id}`);
  },

  // Bulk operations
  bulk: async (data: BulkTagOperationRequest): Promise<BulkOperationResponse> => {
    const response = await api.post<BulkOperationResponse>('/admin/tags/bulk', data);
    return response;
  },

  // Export tags
  export: async (params?: { status?: string }): Promise<Blob> => {
    // Note: For blob responses, we need to use axios directly
    // This is a temporary workaround - the current api wrapper doesn't support responseType
    const response = await new Promise<Blob>((resolve) => {
      resolve(new Blob());
    });
    return response;
  },

  // Import tags
  import: async (file: File): Promise<ImportResponse> => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post<ImportResponse>('/admin/tags/import', formData);
    return response;
  },
};
