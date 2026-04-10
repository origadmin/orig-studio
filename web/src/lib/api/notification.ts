// Notification API
import {api} from "../request";

export interface Notification {
    id: string;
    user_id: string;
    type: string;
    title: string;
    body: string;
    read: boolean;
    created_at: string;
    updated_at: string;
    data?: Record<string, any>;
}

export interface NotificationListResponse {
    items: Notification[];
    total: number;
    page: number;
    page_size: number;
    unread_count: number;
}

export const notificationApi = {
    // 获取通知列表（包含未读数量）
    getAll: (params?: { page?: number; page_size?: number; read?: boolean }) =>
        api.get<NotificationListResponse>('/notifications', params),

    // 获取未读通知数量
    getUnreadCount: () =>
        api.get<{ unread_count: number }>('/notifications/unread-count'),

    // 标记通知为已读 (使用 POST 方法)
    markAsRead: (id: string) =>
        api.post<Notification>(`/notifications/${id}/read`),

    // 标记所有通知为已读 (使用 POST 方法)
    markAllAsRead: () =>
        api.post<{ success: boolean }>('/notifications/read-all'),

    // 删除通知
    delete: (id: string) =>
        api.del<void>(`/notifications/${id}`),
};
