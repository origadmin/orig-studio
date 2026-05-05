// Channel API v4.0 (F019: Channel creation + handle resolution)
import {api} from "../request";
import {PaginatedResponse} from "./types";

export interface ChannelLink {
    type: 'website' | 'social' | 'custom';
    platform?: string;
    url: string;
    title: string;
}

export interface Channel {
    id: string;
    name: string;
    title?: string;
    slug?: string;
    handle?: string;
    short_token?: string;
    description: string;
    user_id: string;
    avatar?: string;
    banner?: string;
    banner_logo?: string;
    status?: string;
    privacy?: string;
    is_verified?: boolean;
    tags?: string[];
    category_id?: number;
    subscriber_count: number;
    media_count: number;
    article_count?: number;
    total_views?: number;
    links?: ChannelLink[];
    create_time?: string;
    update_time?: string;
    // View context
    is_owner?: boolean;
    is_subscribed?: boolean;
}

export interface ChannelDetail extends Channel {}

export interface ChannelPlaylist {
    id: string;
    short_token?: string;
    title: string;
    name?: string;
    description?: string;
    media_count: number;
    video_count?: number;
    thumbnail?: string;
    cover_images?: string[];
    update_time?: string;
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

export interface HandleResolution {
    type: 'channel' | 'user' | 'not_found';
    channel?: Channel;
    user?: {
        id: string;
        username: string;
        name: string;
        logo?: string;
    };
}

export interface ChannelLimits {
    max_channels: number;
    current_count: number;
    can_create: boolean;
}

export interface HandleValidation {
    available: boolean;
    reason?: string;
}

export interface CreateChannelInput {
    name: string;
    handle: string;
    description?: string;
    avatar?: string;
    banner?: string;
    privacy?: string;
    tags?: string[];
    category_id?: number;
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
     * Get single channel by short_token (路径参数方式)
     * RESTful, MediaCMS风格, 推荐!
     */
    getByToken: (token: string) =>
        api.get<ChannelDetail>(`/channels/${token}`),

    /**
     * Get channels with query parameters (查询参数方式)
     */
    get: (params?: ChannelQueryParams) =>
        api.get<Channel | ChannelList>('/channels', params as Record<string, unknown>),

    /**
     * List all public channels (公开频道列表)
     */
    listAll: (params?: {page?: number; limit?: number}) =>
        api.get<ChannelList>('/channels', params),

    /**
     * Get current user's channels (F019: now returns list, not single)
     */
    getMyChannels: (params?: {page?: number; page_size?: number}) =>
        api.get<ChannelList>('/channels/me', params),

    /**
     * Create a new channel (F019: with handle, limits check)
     */
    create: (data: CreateChannelInput) => api.post<Channel>('/channels', data),

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

    /**
     * Resolve a @handle to a channel or user (F019)
     * GET /api/v1/resolve/@{handle}
     */
    resolveHandle: (handle: string) =>
        api.get<HandleResolution>(`/resolve/@${handle}`),

    /**
     * Validate if a handle is available (F019)
     * GET /api/v1/channels/validate-handle?handle=xxx
     */
    validateHandle: (handle: string) =>
        api.get<HandleValidation>('/channels/validate-handle', {handle}),

    /**
     * Get channel creation limits for current user (F019)
     * GET /api/v1/system/config/channel-limits
     */
    getChannelLimits: () =>
        api.get<ChannelLimits>('/system/config/channel-limits'),

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
            params
        ),

    getSubscriberCount: (channelToken: string) =>
        api.get<{count: number}>(`/channels/${channelToken}/subscribers`, {count: 'true'}),

    getAll: (params?: {page?: number; page_size?: number}) => api.get<PaginatedResponse<Channel>>('/channels', params),
};
