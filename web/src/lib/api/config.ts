import {api} from "../request";

export type SettingCategory = "site" | "review" | "upload" | "encoding" | "comment" | "storage";

export interface SettingItem {
    key: string;
    value: string;
    value_type: "string" | "bool" | "int" | "json";
    description: string;
    is_secret: boolean;
    is_builtin: boolean;
    updated_at: string;
}

export interface SettingCategoryGroup {
    category: SettingCategory;
    settings: SettingItem[];
}

export interface SettingsResponse {
    categories: SettingCategoryGroup[];
}

export interface CategorySettingsResponse {
    category: SettingCategory;
    settings: SettingItem[];
}

export interface UpdateSettingRequest {
    value: string;
}

export interface BatchUpdateSettingItem {
    key: string;
    value: string;
}

export interface BatchUpdateSettingsRequest {
    settings: BatchUpdateSettingItem[];
}

export interface BatchUpdateResult {
    key: string;
    status: "success" | "failed";
    error?: string;
}

export interface BatchUpdateSettingsResponse {
    success_count: number;
    fail_count: number;
    results: BatchUpdateResult[];
}

export const configApi = {
    getAll: () =>
        api.get<SettingsResponse>('/admin/settings'),

    getByCategory: (category: SettingCategory) =>
        api.get<CategorySettingsResponse>(`/admin/settings/${category}`),

    updateOne: (key: string, data: UpdateSettingRequest) =>
        api.put<SettingItem>(`/admin/settings/${key}`, data),

    batchUpdate: (data: BatchUpdateSettingsRequest) =>
        api.put<BatchUpdateSettingsResponse>('/admin/settings', data),

    delete: (key: string) =>
        api.del<{ message: string; key: string }>(`/admin/settings/${key}`),
};
