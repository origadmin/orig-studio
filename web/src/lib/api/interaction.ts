// @deprecated 此文件已废弃，请直接使用以下替代方案：
// - 点赞: 使用 mediaApi.likes (from './media') 或 likeApi (from './like')
// - 收藏: 使用 mediaApi.favorites (from './media') 或 favoriteApi (from './favorite')
// - 订阅: 使用 channelApi (from './channel') 或 subscriptionApi (from './subscription')
// - 关注: 使用 userApi (from './user')
// - 分享: 使用 mediaApi.shares (from './media') 或 shareApi (from './share')
// 保留此文件仅为向后兼容，新代码请勿使用

// API 客户端 - 交互模块（点赞、订阅、分享）
// 对应后端 /api/v1/interactions 路径
import {api} from "../request";

// ==================== Like Types ====================
export interface LikeResponse {
    is_liked: boolean;
    is_disliked: boolean;
    like_count: number;
    dislike_count: number;
}

export interface LikeStatusBatchResponse {
    status: Record<string, boolean>;
}

// ==================== Subscription Types ====================
export interface SubscriptionStatus {
    is_subscribed: boolean;
    subscriber_count: number;
}

export interface SubscriptionListResponse {
    items: {
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

// ==================== Share Types ====================
export interface ShareResponse {
    success: boolean;
    url: string;
}

// ==================== API ====================
export const interactionApi = {
    // ==================== Likes ====================
    likes: {
        // Get like list
        list: () => api.get<{ items: unknown[]; total: number; page: number; page_size: number }>("/interactions/likes"),

        // Toggle like/dislike
        toggle: (mediaId: string, type: 'like' | 'dislike' = 'like') =>
            api.post<LikeResponse>("/interactions/likes", {media_id: mediaId, type}),

        // Batch get like status
        getStatusBatch: (mediaIds: string[]) =>
            api.get<LikeStatusBatchResponse>("/interactions/likes/status", {ids: mediaIds.join(",")}),
    },

    // ==================== Subscriptions ====================
    subscriptions: {
        // Get my subscriptions list
        list: (params?: { page?: number; page_size?: number }) =>
            api.get<SubscriptionListResponse>("/interactions/subscriptions", params),

        // Get subscription count
        count: () => api.get<{ count: number }>("/interactions/subscriptions/count"),

        // Subscribe to channel
        subscribe: (channelId: string) =>
            api.post<void>(`/channels/${channelId}/subscription`),

        // Unsubscribe from channel
        unsubscribe: (channelId: string) =>
            api.del<void>(`/channels/${channelId}/subscription`),

        // Get channel subscription status
        getStatus: (channelId: string) =>
            api.get<SubscriptionStatus>(`/channels/${channelId}/subscription`),
    },

    // ==================== Followers ====================
    followers: {
        // Get my followers list
        list: (params?: { page?: number; page_size?: number }) =>
            api.get<SubscriptionListResponse>("/interactions/followers", params),

        // Get follower count
        count: () => api.get<{ count: number }>("/interactions/followers/count"),
    },

    // ==================== Shares ====================
    shares: {
        // Create share
        create: (mediaId: string) =>
            api.post<ShareResponse>("/interactions/shares", {media_id: mediaId}),
    },
};

// Backward-compatible re-exports
// NOTE: favoriteApi has been removed from this file.
// Use `import { favoriteApi } from './favorite'` instead.
export const likeApi = {
    toggle: (mediaId: string) => interactionApi.likes.toggle(mediaId, 'like'),
    toggleDislike: (mediaId: string) => interactionApi.likes.toggle(mediaId, 'dislike'),
    getStatus: (mediaId: string) => interactionApi.likes.getStatusBatch([mediaId]).then(r => ({
        is_liked: r.status[mediaId] || false,
        is_disliked: false,
        like_count: 0,
        dislike_count: 0,
    })),
};

export const subscriptionApi = {
    getStatus: (channelId: string) => interactionApi.subscriptions.getStatus(channelId),
    subscribe: (channelId: string) => interactionApi.subscriptions.subscribe(channelId),
    unsubscribe: (channelId: string) => interactionApi.subscriptions.unsubscribe(channelId),
    getSubscriptions: (params?: { page?: number; page_size?: number }) =>
        interactionApi.subscriptions.list(params),
    getFollowers: (params?: { page?: number; page_size?: number }) =>
        interactionApi.followers.list(params),
};

export const shareApi = {
    share: (mediaId: string) => interactionApi.shares.create(mediaId),
    getShareUrl: (mediaId: string) => interactionApi.shares.create(mediaId),
};
