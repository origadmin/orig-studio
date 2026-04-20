// Comment API
import {api} from "../request";

export interface Comment {
    id: string;
    content?: string;
    media_id?: string;
    user_id?: string;
    username?: string;
    parent_id?: string;
    like_count?: number;
    status?: string;
    create_time?: string;
    update_time?: string;
}

export interface CommentListResponse {
    total: number;
    comments: Comment[];
    page: number;
    page_size: number;
}

export const commentApi = {
    getAll: (params?: { media_id?: string; content_id?: string }) => {
        return api.get<CommentListResponse>('/comments', params || {});
    },
    get: (id: string) => api.get<Comment>(`/comments/${id}`),
    create: (data: { media_id?: string; content_id?: string; parent_id?: string; content: string }) => {
        return api.post<Comment>("/comments", {
            comment: {
                content: data.content,
                ...(data.media_id && { media_id: data.media_id }),
                ...(data.content_id && { content_id: data.content_id }),
                ...(data.parent_id && { parent_id: data.parent_id }),
            }
        });
    },
    update: (id: string, data: { content: string }) =>
        api.put<Comment>(`/comments/${id}`, {
            comment: { content: data.content }
        }),
    delete: (id: string) => api.del<void>(`/comments/${id}`),
};
