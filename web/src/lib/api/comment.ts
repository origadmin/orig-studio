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
    getAll: (params?: { media_id?: string; content_id?: string }) => {
        const queryParams: any = {};
        if (params?.media_id) {
            queryParams.media_id = parseInt(params.media_id, 10);
        }
        if (params?.content_id) {
            queryParams.content_id = parseInt(params.content_id, 10);
        }
        // Only pass params if there are any
        const hasParams = Object.keys(queryParams).length > 0;
        return api.get<Comment[]>('/comments', hasParams ? queryParams : {});
    },
    get: (id: string) => api.get<Comment>(`/comments/${id}`),
    create: (data: { media_id?: string; content_id?: string; parent_id?: string; body: string }) => {
        const requestData: any = {
            text: data.body
        };
        if (data.media_id) {
            requestData.media_id = parseInt(data.media_id, 10);
        }
        if (data.content_id) {
            requestData.content_id = parseInt(data.content_id, 10);
        }
        if (data.parent_id) {
            requestData.parent_id = parseInt(data.parent_id, 10);
        }
        return api.post<Comment>("/comments", requestData);
    },
    update: (id: string, data: { body: string }) =>
        api.put<Comment>(`/comments/${id}`, {text: data.body}),
    delete: (id: string) => api.del<void>(`/comments/${id}`),
};
