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
        api.post<Comment>("/comments", {
            text: data.body,
            media_id: data.media_id,
            content_id: data.content_id,
            parent_id: data.parent_id
        }),
    update: (id: string, data: { body: string }) =>
        api.put<Comment>(`/comments/${id}`, {text: data.body}),
    delete: (id: string) => api.del<void>(`/comments/${id}`),
};
