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
export type SettingType = 'string' | 'int' | 'bool' | 'json';
export type SettingCategory = 'general' | 'upload' | 'review' | 'email' | 'module';

export interface SettingItem {
    id: string;
    key: string;
    value: string;
    type: SettingType;
    category: SettingCategory;
    description?: string;
    is_sensitive: boolean;
    fallback_value?: string;
    is_builtin: boolean;
    create_time: string;
    update_time: string;
}

export type GroupedSettings = Record<SettingCategory, SettingItem[]>;

export interface UpdateSettingItem {
    key: string;
    value: string;
}

export interface UpdateSettingsRequest {
    settings: UpdateSettingItem[];
}

// ==================== Stats API ====================
export const statsApi = {
    getDashboard: () => api.get<DashboardStats>("/admin/stats/dashboard"),
    getMedia: () => api.get<MediaStats>("/admin/stats/medias"),
    getUsers: () => api.get<UserStats>("/admin/stats/users"),
    getTraffic: (params?: { page?: number; page_size?: number }) =>
        api.get<TrafficStatsResponse>("/admin/stats/traffic", params),
};

// ==================== Settings API ====================
export const settingsApi = {
    get: () => api.get<GroupedSettings>("/admin/settings"),
    update: (data: UpdateSettingsRequest) =>
        api.put<GroupedSettings>("/admin/settings", data),
    getByKey: (key: string) =>
        api.get<SettingItem>(`/system/settings/${key}`),
    resetKey: (key: string) =>
        api.post<SettingItem>(`/system/settings/${key}/reset`),
};

// ==================== System API ====================
export const systemApi = {
    stats: statsApi,
    settings: settingsApi,
};
