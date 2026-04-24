// @deprecated 此文件已废弃，请直接使用以下替代方案：
// - 点赞/收藏/分享: 使用 mediaApi (from './media')
// - 订阅: 使用 channelApi (from './channel')
// - 关注: 使用 userApi (from './user')
// 保留此文件仅为向后兼容，新代码请勿使用

// API 客户端 - 交互模块（点赞、收藏、订阅、分享）
// 对应后端 /api/v1/interactions 路径
import {api} from "../request";
import type {Media} from "../../types";

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

// ==================== Favorite Types ====================
export interface Favorite {
    id: number;
    media_id: number;
    media: Media;
    created_at: string;
}

export interface ToggleFavoriteResponse {
    success: boolean;
    is_favorited: boolean;
}

export interface FavoriteListResponse {
    items: Favorite[];
    total: number;
    page: number;
    page_size: number;
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
        // 获取点赞列表
        list: () => api.get<{ items: unknown[]; total: number; page: number; page_size: number }>("/interactions/likes"),

        // 点赞/取消点赞/点踩
        toggle: (mediaId: string, type: 'like' | 'dislike' = 'like') =>
            api.post<LikeResponse>("/interactions/likes", {media_id: mediaId, type}),

        // 批量获取点赞状态
        getStatusBatch: (mediaIds: string[]) =>
            api.get<LikeStatusBatchResponse>("/interactions/likes/status", {ids: mediaIds.join(",")}),
    },

    // ==================== Favorites ====================
    favorites: {
        // 获取收藏列表
        list: () => api.get<FavoriteListResponse>("/interactions/favorites"),

        // 收藏/取消收藏
        toggle: (mediaId: string) =>
            api.post<ToggleFavoriteResponse>("/interactions/favorites", {media_id: mediaId}),

        // 检查是否已收藏
        check: (mediaId: string) =>
            api.get<{ is_favorited: boolean }>("/interactions/favorites/check", {media_id: mediaId}),
    },

    // ==================== Subscriptions ====================
    subscriptions: {
        // 获取我的订阅列表
        list: (params?: { page?: number; page_size?: number }) =>
            api.get<SubscriptionListResponse>("/interactions/subscriptions", params),

        // 获取订阅数量
        count: () => api.get<{ count: number }>("/interactions/subscriptions/count"),

        // 订阅频道（通过频道ID）
        subscribe: (channelId: string) =>
            api.post<void>(`/channels/${channelId}/subscription`),

        // 取消订阅频道
        unsubscribe: (channelId: string) =>
            api.del<void>(`/channels/${channelId}/subscription`),

        // 获取频道订阅状态
        getStatus: (channelId: string) =>
            api.get<SubscriptionStatus>(`/channels/${channelId}/subscription`),
    },

    // ==================== Followers ====================
    followers: {
        // 获取我的粉丝列表
        list: (params?: { page?: number; page_size?: number }) =>
            api.get<SubscriptionListResponse>("/interactions/followers", params),

        // 获取粉丝数量
        count: () => api.get<{ count: number }>("/interactions/followers/count"),
    },

    // ==================== Shares ====================
    shares: {
        // 创建分享
        create: (mediaId: string) =>
            api.post<ShareResponse>("/interactions/shares", {media_id: mediaId}),
    },
};

// 为了保持向后兼容，导出单独的API对象
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

export const favoriteApi = {
    toggle: (mediaId: string) => interactionApi.favorites.toggle(mediaId),
    getStatus: (mediaId: string) => interactionApi.favorites.check(mediaId),
    list: () => interactionApi.favorites.list(),
    remove: (mediaId: string) => interactionApi.favorites.toggle(mediaId),
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
