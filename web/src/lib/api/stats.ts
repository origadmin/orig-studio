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
    // Get dashboard stats (Admin)
    getDashboard: () =>
        api.get<DashboardStats>('/admin/stats/dashboard'),

    // Get media stats (Admin)
    getMedia: () =>
        api.get<MediaStats>('/admin/stats/medias'),

    // Get user stats (Admin)
    getUsers: () =>
        api.get<UserStats>('/admin/stats/users'),

    // Get traffic stats (Admin)
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
        }>('/admin/stats/traffic', params),

    // Get revenue stats (Admin)
    getRevenue: (params?: { days?: number; type?: 'daily' | 'weekly' | 'monthly' }) =>
        api.get<{
            revenue: Array<{
                date: string;
                amount: number;
            }>;
        }>('/admin/stats/revenue', params),
};
