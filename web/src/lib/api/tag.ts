// Tag API
import {api} from "../request";
import {PaginatedResponse} from "./types";

export interface Tag {
    id: string;
    name: string;
    slug: string;
    description?: string;
    color?: string;
    status?: string;
    count?: number;
    create_time: string;
    update_time?: string;
}

export const tagApi = {
    getAll: () => api.get<PaginatedResponse<Tag>>("/tags"),
    get: (id: string) => api.get<Tag>(`/tags/${id}`),
    create: (data: Partial<Tag>) => api.post<Tag>("/tags", data),
    update: (id: string, data: Partial<Tag>) => api.put<Tag>(`/tags/${id}`, data),
    delete: (id: string) => api.del<void>(`/tags/${id}`),
};
