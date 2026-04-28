// API 客户端 - 系统模块（统计、配置）
// 对应后端 /api/v1/system 路径
import {api} from "../request";

// ==================== Stats Types ====================
export interface DashboardStats {
    total_users: number;
    total_media: number;
    total_views: number;
    new_users_today: number;
    new_media_today: number;
    views_today: number;
    encoding_pending: number;
    encoding_failed: number;
}

export interface MediaStats {
    total: number;
    video_count: number;
    audio_count: number;
    image_count: number;
    public_count: number;
    private_count: number;
    encoding_pending: number;
    encoding_failed: number;
}

export interface UserStats {
    total: number;
    active_today: number;
    new_today: number;
    admin_count: number;
    editor_count: number;
    regular_count: number;
}

export interface TrafficStatsItem {
    date: string;
    views: number;
    unique_visitors: number;
    bandwidth: number;
}

export interface TrafficStatsResponse {
    list: TrafficStatsItem[];
    total: number;
    page: number;
    page_size: number;
}

// ==================== Settings Types ====================
export interface SystemSettings {
    site_name: string;
    site_description: string;
    allow_register: boolean;
    allow_upload: boolean;
    max_upload_size: number; // bytes
    // 可以添加更多配置项
}

export interface UpdateSettingsRequest {
    site_name?: string;
    site_description?: string;
    allow_register?: boolean;
    allow_upload?: boolean;
    max_upload_size?: number;
}

// ==================== Stats API ====================
export const statsApi = {
    // Get Dashboard stats (Admin)
    getDashboard: () => api.get<DashboardStats>("/admin/stats/dashboard"),

    // Get media stats (Admin)
    getMedia: () => api.get<MediaStats>("/admin/stats/medias"),

    // Get user stats (Admin)
    getUsers: () => api.get<UserStats>("/admin/stats/users"),

    // Get traffic stats (Admin)
    getTraffic: (params?: { page?: number; page_size?: number }) =>
        api.get<TrafficStatsResponse>("/admin/stats/traffic", params),
};

// ==================== Settings API ====================
export const settingsApi = {
    // Get system settings (Admin)
    get: () => api.get<SystemSettings>("/admin/settings"),

    // Update system settings (Admin)
    update: (data: UpdateSettingsRequest) =>
        api.put<SystemSettings>("/admin/settings", data),
};

// ==================== System API ====================
export const systemApi = {
    stats: statsApi,
    settings: settingsApi,
};
