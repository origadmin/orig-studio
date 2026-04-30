import {api} from "../request";

export interface ReviewItem {
    id: string;
    media_id: string;
    media_title: string;
    media_type: string;
    user_id: string;
    username: string;
    review_status: string;
    reason?: string;
    created_at: string;
    create_time?: string;
    updated_at: string;
    update_time?: string;
    reviewer_id?: string;
    reviewer_name?: string;
}

export interface ReviewListResponse {
    items: ReviewItem[];
    total: number;
    page: number;
    page_size: number;
}

export const reviewApi = {
    getPending: (params?: { page?: number; page_size?: number; type?: string }) =>
        api.get<ReviewListResponse>('/admin/medias/review/pending', params),

    getHistory: (params?: { page?: number; page_size?: number; type?: string; status?: string }) =>
        api.get<ReviewListResponse>('/admin/medias/review/history', params),

    review: (mediaId: string, data: { action: 'approve' | 'reject'; comment?: string }) =>
        api.put<{ id: string; review_status: string; listable: boolean; updated_at: string }>(`/admin/medias/${mediaId}/review`, data),

    getDetail: (mediaId: string) =>
        api.get<{ items: ReviewItem[] }>(`/admin/medias/${mediaId}/review-logs`),

    batchReview: (data: { media_ids: string[]; action: 'approve' | 'reject'; comment?: string }) =>
        api.post<{ succeeded: string[]; failed: { media_id: string; error: string }[] }>('/admin/medias/review/batch', data),
};
