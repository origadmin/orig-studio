// Comment API
import {api} from "../request";

export interface Comment {
    id: string;
    content_id?: string;
    media_id?: string;
    user_id: string;
    username: string;
    parent_id?: string;
    body: string;
    status: string;
    created_at: string;
    updated_at: string;
}

export const commentApi = {
    getAll: (params?: { media_id?: string; content_id?: string }) =>
        api.get<Comment[]>("/comments", params),
    get: (id: string) => api.get<Comment>(`/comments/${id}`),
    create: (data: { media_id?: string; content_id?: string; parent_id?: string; body: string }) =>
        api.post<Comment>("/comments", data),
    update: (id: string, data: { body: string }) =>
        api.put<Comment>(`/comments/${id}`, data),
    delete: (id: string) => api.del<void>(`/comments/${id}`),
};
