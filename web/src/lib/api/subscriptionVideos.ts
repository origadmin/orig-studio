// API client - Subscription Videos
// GET /api/v1/subscriptions/videos
// Fetches videos from all subscribed channels
import {api} from '../request';
import {PaginatedResponse} from './types';

/**
 * Represents a video item from subscribed channels.
 * Extends the standard media fields with channel subscription context.
 */
export interface SubscriptionVideo {
    id: string;
    title: string;
    description?: string;
    short_token: string;
    thumbnail?: string;
    poster?: string;
    duration: number;
    type: string;
    url: string;
    hls_file?: string;
    view_count: number;
    like_count: number;
    comment_count: number;
    privacy: number;
    encoding_status: string;
    state: string;
    published_at?: string;
    create_time: string;
    update_time?: string;
    tags?: string[];
    channel_id: string;
    channel_name: string;
    channel_avatar?: string;
    channel_token?: string;
    user_id: string;
    username: string;
    user_avatar?: string;
}

/**
 * Query parameters for subscription videos endpoint
 */
export interface SubscriptionVideosParams {
    page?: number;
    page_size?: number;
    channel_id?: string;
    sort?: 'newest' | 'most_viewed' | 'trending';
}

export const subscriptionVideosApi = {
    /**
     * Get videos from subscribed channels
     * GET /api/v1/subscriptions/videos
     *
     * @param params - Query parameters including pagination, channel filter, and sort order
     */
    getVideos: (params?: SubscriptionVideosParams) =>
        api.get<PaginatedResponse<SubscriptionVideo>>('/subscriptions/videos', params as Record<string, unknown>),
};
