// Notification API
import {api} from "../request";

export interface Notification {
    id: number;
    user_id: string;
    action: string;
    title: string;
    body: string;
    read: boolean;
    create_time: string;
    update_time: string;
}

export interface NotificationListResponse {
    items: Notification[];
    total: number;
    page: number;
    page_size: number;
    unread_count: number;
}

export const notificationApi = {
    getAll: (params?: { page?: number; page_size?: number; read?: boolean }) =>
        api.get<NotificationListResponse>('/notifications', params),

    getUnreadCount: () =>
        api.get<{ unread_count: number }>('/notifications/unread-count'),

    create: (data: { action: string; title: string; body: string; user_id?: string; method?: string; notify?: boolean }) =>
        api.post<Notification>('/notifications', data),

    markAsRead: (id: number) =>
        api.post<Notification>(`/notifications/${id}/read`),

    markAllAsRead: () =>
        api.post<{ success: boolean }>('/notifications/read-all'),

    delete: (id: number) =>
        api.del<void>(`/notifications/${id}`),
};
