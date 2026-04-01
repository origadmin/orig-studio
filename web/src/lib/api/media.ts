// API 客户端 - 媒体模块
import {api} from "../request";

export interface Media {
    id: string;
    title: string;
    description?: string;
    type: "video" | "audio" | "image" | "document";
    url: string;
    thumbnail?: string;
    size: number;
    duration?: number;
    status: string;
    user_id: string;
    category_id?: string;
    tags?: string[];
    views: number;
    likes: number;
    created_at: string;
    updated_at?: string;
}

export interface MediaListResponse {
    list: Media[];
    total: number;
    page: number;
    page_size: number;
}

export interface CreateMediaRequest {
    title: string;
    description?: string;
    type: Media["type"];
    url: string;
    thumbnail?: string;
    size: number;
    duration?: number;
    category_id?: string;
    tags?: string[];
}

export interface UpdateMediaRequest {
    title?: string;
    description?: string;
    thumbnail?: string;
    category_id?: string;
    tags?: string[];
    status?: string;
}

export const mediaApi = {
    // 获取所有媒体（管理端）
    getAll: () => api.get<MediaListResponse>("/media"),

    // 获取媒体列表（公开）
    list: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        category_id?: string;
        keyword?: string;
        status?: string;
    }) => api.get<MediaListResponse>("/media", {...params, status: params?.status || "published"}),

    // 获取媒体详情（公开）
    get: (id: string) => api.get<Media>(`/media/${id}`),

    // 获取推荐媒体
    featured: (limit?: number) => api.get<Media[]>("/media/featured", {limit}),

    // 获取最新媒体
    latest: (limit?: number) => api.get<Media[]>("/media/latest", {limit}),

    // 管理端：获取所有媒体（包括未发布的）
    adminList: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        status?: string;
        keyword?: string;
    }) => api.get<MediaListResponse>("/media", params),

    // 管理端：创建媒体
    create: (data: CreateMediaRequest) => api.post<Media>("/media", data),

    // 管理端：上传媒体文件
    upload: (file: File, metadata: { title?: string; description?: string; user_id?: string }) => {
        const formData = new FormData();
        formData.append("file", file);
        if (metadata.title) formData.append("title", metadata.title);
        if (metadata.description) formData.append("description", metadata.description);
        if (metadata.user_id) formData.append("user_id", metadata.user_id);

        return api.post<Media>("/media/upload", formData);
    },

    // 管理端：更新媒体
    update: (id: string, data: UpdateMediaRequest) => api.put<Media>(`/media/${id}`, data),

    // 管理端：删除媒体
    delete: (id: string) => api.del<void>(`/media/${id}`),

    // 管理端：更新媒体状态
    updateStatus: (id: string, status: string) =>
        api.patch<Media>(`/media/${id}/status`, {status}),
};
