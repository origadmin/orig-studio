// API 客户端 - 历史记录模块
import {api} from "../request";

export interface HistoryItem {
    id: number;
    media_id: number;
    user_id: number;
    progress: number;
    watched_at: string;
    media?: {
        id: number;
        title: string;
        thumbnail: string;
        duration: number;
        view_count: number;
        created_at: string;
        edges?: {
            user?: Array<{
                username: string;
                avatar: string;
            }>;
        };
    };
}

export interface HistoryListResponse {
    list: HistoryItem[];
    total: number;
    page: number;
    page_size: number;
}

export const historyApi = {
    // 获取历史记录列表
    list: (params?: { page?: number; page_size?: number }) =>
        api.get<HistoryListResponse>("/user/history", params),

    // 清除历史记录
    clear: () => api.del<void>("/user/history"),

    // 删除单个历史记录
    remove: (id: number) => api.del<void>(`/user/history/${id}`),
};
