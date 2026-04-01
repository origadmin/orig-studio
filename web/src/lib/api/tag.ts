// Tag API
import {api} from "../request";

export interface Tag {
    id: string;
    name: string;
    slug: string;
    count: number;
    created_at: string;
}

export const tagApi = {
    getAll: () => api.get<Tag[]>("/tags"),
    get: (id: string) => api.get<Tag>(`/tags/${id}`),
    create: (data: Partial<Tag>) => api.post<Tag>("/tags", data),
    update: (id: string, data: Partial<Tag>) => api.put<Tag>(`/tags/${id}`, data),
    delete: (id: string) => api.del<void>(`/tags/${id}`),
};
