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

export interface AdminComment {
    id: string;
    content?: string;
    status?: string;
    media_id?: string;
    user_id?: string;
    username?: string;
    avatar?: string;
    like_count?: number;
    reply_count?: number;
    report_count?: number;
    is_spam?: boolean;
    create_time?: string;
    media?: { id: string; title: string };
    moderated_by?: string;
    moderated_at?: string;
    parent_id?: string;
    depth?: number;
    has_replies?: boolean;
    children?: AdminComment[];
    has_pending_reports?: boolean;
}

export interface AdminCommentListResponse {
    items: AdminComment[];
    total: number;
    page: number;
    page_size: number;
}

export interface CommentReport {
    id: string;
    comment_id: string;
    reporter_id: string;
    reason: string;
    description?: string;
    status: string;
    create_time: string;
    username?: string;
}

export interface CommentStats {
    pending: number;
    approved: number;
    rejected: number;
    blocked: number;
    total: number;
    reported_pending: number;
}

export interface CommentListResponse {
    items: Comment[];
    total: number;
    page: number;
    page_size: number;
}

export interface CommentLikeResponse {
    like_count: number;
    is_liked: boolean;
    is_disliked: boolean;
}

export type CommentSortBy = 'newest' | 'oldest' | 'popular';

export const commentApi = {
    getAll: (params?: { media_id?: string; content_id?: string; page?: number; page_size?: number; sort_by?: string; order?: string }) => {
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

    // Comment Likes API
    likes: {
        getStatus: (commentId: string) =>
            api.get<CommentLikeResponse>(`/comments/${commentId}/likes`),
        toggle: (commentId: string) =>
            api.post<CommentLikeResponse>(`/comments/${commentId}/likes`),
        toggleDislike: (commentId: string) =>
            api.post<CommentLikeResponse>(`/comments/${commentId}/dislikes`),
    },

    report: (commentId: string, data: { reason: string; description?: string }) =>
        api.post<{ message: string; report_count: number; status: string }>(`/comments/${commentId}/report`, data),
};

export const adminCommentApi = {
    list: (params?: { page?: number; page_size?: number; media_id?: string; status?: string; report_status?: string; tree?: boolean }) =>
        api.get<AdminCommentListResponse>('/admin/comments', params),

    get: (id: string) =>
        api.get<AdminComment>(`/admin/comments/${id}`),

    delete: (id: string) =>
        api.del<void>(`/admin/comments/${id}`),

    approve: (id: string) =>
        api.post(`/admin/comments/${id}/approve`),

    reject: (id: string) =>
        api.post(`/admin/comments/${id}/reject`),

    block: (id: string) =>
        api.post(`/admin/comments/${id}/block`),

    unblock: (id: string) =>
        api.post(`/admin/comments/${id}/unblock`),

    dismissReports: (id: string) =>
        api.post(`/admin/comments/${id}/dismiss-reports`),

    getReports: (id: string) =>
        api.get(`/admin/comments/${id}/reports`),

    getStats: (params?: { media_id?: string }) =>
        api.get<CommentStats>('/admin/comments/stats', params),

    batchApprove: (ids: string[]) =>
        api.post('/admin/comments/batch-approve', { ids }),

    batchReject: (ids: string[]) =>
        api.post('/admin/comments/batch-reject', { ids }),
};
