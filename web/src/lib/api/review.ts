// Review API
import {api} from "../request";

export interface ReviewItem {
    id: string;
    media_id: string;
    media_title: string;
    media_type: string;
    user_id: string;
    username: string;
    status: string; // "pending" | "approved" | "rejected"
    reason?: string;
    created_at: string;
    updated_at: string;
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
    // Get pending review items (Admin)
    getPending: (params?: { page?: number; page_size?: number; type?: string }) =>
        api.get<ReviewListResponse>('/admin/medias/review/pending', params),

    // Get review history (Admin)
    getHistory: (params?: { page?: number; page_size?: number; type?: string; status?: string }) =>
        api.get<ReviewListResponse>('/admin/medias/review/history', params),

    // Review a media item (Admin)
    review: (mediaId: string, data: { status: string; reason?: string }) =>
        api.put<{ success: boolean; message: string }>(`/admin/medias/${mediaId}/review`, data),

    // Get review detail for a media item (Admin)
    getDetail: (mediaId: string) =>
        api.get<ReviewItem>(`/admin/medias/${mediaId}/review-logs`),

    // Batch review (Admin)
    batchReview: (data: { ids: string[]; status: string; reason?: string }) =>
        api.post<{ success: boolean; message: string }>('/admin/medias/review/batch', data),
};
