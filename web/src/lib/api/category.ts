// Category API
import {api} from "../request";
import {PaginatedResponse} from "./types";

export interface Category {
    id: number;
    name: string;
    slug: string;
    description?: string;
    parent_id?: number;
    order: number;
    status?: number;
    media_count?: number;
    created_at: string;
    updated_at: string;
}

export const categoryApi = {
    getAll: (params?: {page?: number; page_size?: number}) => api.get<PaginatedResponse<Category>>("/categories", {params}),
    get: (id: number | string) => api.get<Category>(`/categories/${id}`),
    create: (data: Partial<Category>) => api.post<Category>("/categories", data),
    update: (id: number | string, data: Partial<Category>) => api.put<Category>(`/categories/${id}`, data),
    delete: (id: number | string) => api.del<void>(`/categories/${id}`),
};
