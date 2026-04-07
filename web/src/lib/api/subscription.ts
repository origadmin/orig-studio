// API 客户端 - 订阅模块
import {api} from "../request";

export interface SubscriptionStatus {
    is_subscribed: boolean;
    subscriber_count: number;
}

export interface SubscriptionListResponse {
    list: {
        id: string;
        user_id: string;
        username: string;
        avatar?: string;
        subscribed_at: string;
    }[];
    total: number;
    page: number;
    page_size: number;
}

export const subscriptionApi = {
    // 获取订阅状态
    getStatus: (userId: string) => api.get<SubscriptionStatus>(`/users/${userId}/subscription`),

    // 订阅用户
    subscribe: (userId: string) => api.post<void>(`/users/${userId}/subscribe`),

    // 取消订阅
    unsubscribe: (userId: string) => api.delete<void>(`/users/${userId}/subscribe`),

    // 获取订阅列表
    getSubscriptions: (params?: { page?: number; page_size?: number; keyword?: string }) =>
        api.get<SubscriptionListResponse>("/subscriptions", params),

    // 获取粉丝列表
    getFollowers: (params?: { page?: number; page_size?: number; keyword?: string }) =>
        api.get<SubscriptionListResponse>("/followers", params),
};
