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
    list: ReviewItem[];
    total: number;
    page: number;
    page_size: number;
}

export const reviewApi = {
    // 获取待审核内容列表
    getPending: (params?: { page?: number; page_size?: number; type?: string }) =>
        api.get<ReviewListResponse>('/review/pending', params),

    // 获取审核历史
    getHistory: (params?: { page?: number; page_size?: number; type?: string; status?: string }) =>
        api.get<ReviewListResponse>('/review/history', params),

    // 审核内容
    review: (id: string, data: { status: string; reason?: string }) =>
        api.put<{ success: boolean; message: string }>(`/review/${id}`, data),

    // 获取审核详情
    getDetail: (id: string) =>
        api.get<ReviewItem>(`/review/${id}`),

    // 批量审核
    batchReview: (data: { ids: string[]; status: string; reason?: string }) =>
        api.put<{ success: boolean; message: string }>('/review/batch', data),
};
