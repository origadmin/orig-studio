// Category API
import {api} from "../request";

export interface Category {
    id: string;
    name: string;
    slug: string;
    description?: string;
    parent_id?: string;
    order: number;
    created_at: string;
    updated_at: string;
}

export const categoryApi = {
    getAll: () => api.get<Category[]>("/categories"),
    get: (id: string) => api.get<Category>(`/categories/${id}`),
    create: (data: Partial<Category>) => api.post<Category>("/categories", data),
    update: (id: string, data: Partial<Category>) => api.put<Category>(`/categories/${id}`, data),
    delete: (id: string) => api.del<void>(`/categories/${id}`),
};
