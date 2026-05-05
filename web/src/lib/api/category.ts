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
    create_time: string;
    update_time: string;
}

export const categoryApi = {
    getAll: (params?: {page?: number; page_size?: number}) => api.get<PaginatedResponse<Category>>("/categories", params),
    get: (id: number | string) => api.get<Category>(`/categories/${id}`),
    create: (data: Partial<Category>) => api.post<Category>("/categories", data),
    update: (id: number | string, data: Partial<Category>) => api.put<Category>(`/categories/${id}`, data),
    delete: (id: number | string) => api.del<void>(`/categories/${id}`),
};

// ==================== Admin Category API (requires JWT + Admin) ====================
export const adminCategoryApi = {
    // List all categories (Admin, includes all statuses)
    list: (params?: {page?: number; page_size?: number}) =>
        api.get<PaginatedResponse<Category>>("/admin/categories", params),

    // Get category detail (Admin)
    get: (id: number | string) =>
        api.get<Category>(`/admin/categories/${id}`),

    // Create category (Admin)
    create: (data: Partial<Category>) =>
        api.post<Category>("/admin/categories", data),

    // Update category (Admin)
    update: (id: number | string, data: Partial<Category>) =>
        api.put<Category>(`/admin/categories/${id}`, data),

    // Patch category (Admin - partial update)
    patch: (id: number | string, data: Partial<Category>) =>
        api.patch<Category>(`/admin/categories/${id}`, data),

    // Delete category (Admin)
    delete: (id: number | string) =>
        api.del<void>(`/admin/categories/${id}`),
};
