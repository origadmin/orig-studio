// Channel API v3.2 (路径参数版)
import {api} from "../request";
import {PaginatedResponse} from "./types";

export interface Channel {
    id: string;
    name: string;
    title?: string;
    slug: string;
    handle?: string;
    friendly_token?: string;
    description: string;
    user_id: string;
    owner_id?: string;
    media_count: number;
    subscriber_count: number;
    status: string;
    is_public?: boolean;
    is_default?: boolean;
    banner_logo?: string;
    avatar?: string;
    banner?: string;
    total_views?: number;
    video_count?: number;
    is_verified?: boolean;
    created_at?: string;
    links?: Array<{
        type: 'website' | 'social' | 'custom';
        platform?: string;
        url: string;
        title: string;
    }>;
    tags?: string[];
}

export interface ChannelDetail extends Channel {}

export interface ChannelPlaylist {
    id: string;
    title: string;
    name?: string;
    description?: string;
    media_count: number;
    video_count?: number;
    thumbnail?: string;
    cover_images?: string[];
    updated_at?: string;
}

export interface ChannelList {
    items: Channel[];
    total: number;
    page: number;
    page_size: number;
}

export interface SubscribeResponse {
    success: boolean;
    message: string;
}

export interface NotificationSettingResponse {
    success: boolean;
    setting: string;
    channel_id: string;
    message: string;
}

/**
 * Channel Query Parameters for GET /channels (查询参数方式)
 */
export interface ChannelQueryParams {
    /** @username 查询模式 - 按 username 查询默认频道 (两步方案) */
    username?: string;
    /** 用户ID查询模式 - 查询某用户的所有频道 */
    user_id?: string;
    /** 分页参数 */
    page?: number;
    limit?: number;
}

export const channelApi = {
    /**
     * Get single channel by short_token (路径参数方式) ⭐
     * RESTful, MediaCMS风格, 推荐!
     *
     * @param token - short_token [a-zA-Z0-9]{6,12}
     * @example getByToken('adm001') → GET /channels/adm001
     */
    getByToken: (token: string) =>
        api.get<ChannelDetail>(`/channels/${token}`),

    /**
     * Get channels with query parameters (查询参数方式)
     * 支持3种模式:
     *   1. ?username=xxx  → @username 两步方案
     *   2. ?user_id=xxx   → 用户频道列表
     *   3. (无参数)        → 公开频道列表
     *
     * @example get({ username: 'admin' }) → GET /channels?username=admin
     * @example get({ user_id: 'uuid-xxx', page: 1 }) → GET /channels?user_id=uuid-xxx&page=1
     * @example get({ page: 1, limit: 20 }) → GET /channels?page=1&limit=20
     */
    get: (params?: ChannelQueryParams) =>
        api.get<Channel | ChannelList>('/channels', {params}),

    /**
     * List all public channels (公开频道列表)
     * Alias for get() without params
     */
    listAll: (params?: {page?: number; limit?: number}) =>
        api.get<ChannelList>('/channels', {params}),

    /**
     * Get current user's channel(s)
     */
    getMyChannel: () =>
        api.get<ChannelDetail>('/channels/me'),

    /**
     * Create a new channel
     */
    create: (data: Partial<Channel>) => api.post<Channel>('/channels', data),

    /**
     * Update a channel (by short_token)
     */
    update: (token: string, data: Partial<Channel>) => api.put<Channel>(`/channels/${token}`, data),

    /**
     * Delete a channel (by short_token)
     */
    delete: (token: string) => api.del<void>(`/channels/${token}`),

    /**
     * Update current user's channel handle/slug
     */
    updateHandle: (handle: string) =>
        api.put<ChannelDetail>('/channels/me/handle', {handle}),

    // ================================
    // Subscription APIs (基于 short_token)
    // ================================

    subscribe: (channelToken: string) =>
        api.post<SubscribeResponse>(`/channels/${channelToken}/subscription`),

    unsubscribe: (channelToken: string) =>
        api.del<SubscribeResponse>(`/channels/${channelToken}/subscription`),

    getSubscriptionStatus: (channelToken: string) =>
        api.get<{is_subscribed: boolean}>(`/channels/${channelToken}/subscription`),

    updateNotificationSetting: (channelToken: string, setting: string) =>
        api.put<NotificationSettingResponse>(`/channels/${channelToken}/notification`, {setting}),

    getSubscribers: (channelToken: string, params?: {page?: number; page_size?: number}) =>
        api.get<{items: string[]; total: number; page: number; page_size: number}>(
            `/channels/${channelToken}/subscribers`,
            {params}
        ),

    getSubscriberCount: (channelToken: string) =>
        api.get<{count: number}>(`/channels/${channelToken}/subscribers`, {params: {count: 'true'}}),

    getAll: (params?: {page?: number; page_size?: number}) => api.get<PaginatedResponse<Channel>>('/channels', {params}),
};
