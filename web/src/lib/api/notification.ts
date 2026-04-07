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

export const notificationApi = {
    // 获取通知列表
    getAll: (params?: { page?: number; page_size?: number; read?: boolean }) =>
        api.get<Notification[]>('/notifications', params),

    // 获取未读通知数量
    getUnreadCount: () =>
        api.get<{ count: number }>('/notifications/unread/count'),

    // 标记通知为已读
    markAsRead: (id: string) =>
        api.put<Notification>(`/notifications/${id}/read`),

    // 标记所有通知为已读
    markAllAsRead: () =>
        api.put<{ success: boolean }>('/notifications/read-all'),

    // 删除通知
    delete: (id: string) =>
        api.del<void>(`/notifications/${id}`),
};
