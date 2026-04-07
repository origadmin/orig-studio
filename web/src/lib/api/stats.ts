// Stats API
import {api} from "../request";

export interface DashboardStats {
    total_users: number;
    total_media: number;
    total_views: number;
    total_comments: number;
    total_subscribers: number;
    total_revenue: number;
    active_users: number;
    new_users_today: number;
    new_media_today: number;
    new_views_today: number;
    new_comments_today: number;
    new_subscribers_today: number;
    media_by_type: {
        video: number;
        image: number;
        audio: number;
        other: number;
    };
    users_by_role: {
        admin: number;
        editor: number;
        user: number;
    };
    views_by_date: Array<{
        date: string;
        views: number;
    }>;
    media_by_date: Array<{
        date: string;
        count: number;
    }>;
    top_categories: Array<{
        id: string;
        name: string;
        count: number;
    }>;
    top_creators: Array<{
        id: string;
        name: string;
        media_count: number;
        views: number;
    }>;
    top_media: Array<{
        id: string;
        title: string;
        views: number;
        created_at: string;
    }>;
}

export interface MediaStats {
    total: number;
    by_type: {
        video: number;
        image: number;
        audio: number;
        other: number;
    };
    by_status: {
        pending: number;
        approved: number;
        rejected: number;
    };
    by_date: Array<{
        date: string;
        count: number;
    }>;
}

export interface UserStats {
    total: number;
    by_role: {
        admin: number;
        editor: number;
        user: number;
    };
    by_status: {
        active: number;
        inactive: number;
        banned: number;
    };
    by_date: Array<{
        date: string;
        count: number;
    }>;
}

export const statsApi = {
    // 获取仪表盘统计数据
    getDashboard: () =>
        api.get<DashboardStats>('/stats/dashboard'),

    // 获取媒体统计数据
    getMedia: () =>
        api.get<MediaStats>('/stats/media'),

    // 获取用户统计数据
    getUsers: () =>
        api.get<UserStats>('/stats/users'),

    // 获取流量统计数据
    getTraffic: (params?: { days?: number }) =>
        api.get<{
            views: Array<{
                date: string;
                views: number;
            }>;
            unique_visitors: Array<{
                date: string;
                visitors: number;
            }>;
        }>('/stats/traffic', params),

    // 获取收入统计数据
    getRevenue: (params?: { days?: number; type?: 'daily' | 'weekly' | 'monthly' }) =>
        api.get<{
            revenue: Array<{
                date: string;
                amount: number;
            }>;
        }>('/stats/revenue', params),
};
