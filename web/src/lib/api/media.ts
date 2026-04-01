// API 客户端 - 媒体模块
// 类型定义对齐后端 ent entity JSON 输出
import {api, getAccessToken} from "../request";

// Media 对齐后端 entity.Media JSON 序列化字段
export interface Media {
    id: number;
    title: string;
    description?: string;
    friendly_token?: string;
    type: string; // "video" | "image" | "audio"
    url: string;
    hls_file?: string;
    thumbnail?: string;
    poster?: string;
    preview_file_path?: string;
    duration: number;
    size?: string; // 后端存的是 string
    width: number;
    height: number;
    mime_type?: string;
    md5sum?: string;
    extension?: string;
    privacy: number; // 1: public, 2: private, 3: unlisted
    encoding_status: string; // "pending" | "processing" | "success" | "failed"
    state: string; // "draft" | "active" | "deleted"
    view_count: number;
    like_count: number;
    dislike_count: number;
    comment_count: number;
    favorite_count: number;
    download_count: number;
    allow_download: boolean;
    enable_comments: boolean;
    featured: boolean;
    is_reviewed: boolean;
    reported_times: number;
    tags?: string[];
    user_id: number;
    published_at?: string;
    created_at: string;
    updated_at: string;
    // edges
    edges?: {
        user?: UserSummary[];
        category?: CategorySummary;
        comments?: unknown[];
        channels?: unknown[];
        playlists?: unknown[];
        tags_rel?: unknown[];
        favorites?: unknown[];
        likes?: unknown[];
    };
}

// UserSummary 是 edges.user 中返回的用户摘要
export interface UserSummary {
    id: number;
    username: string;
    nickname?: string;
    avatar?: string;
}

// CategorySummary 是 edges.category 中返回的分类摘要
export interface CategorySummary {
    id: number;
    name: string;
    slug?: string;
    icon?: string;
    color?: string;
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
    type: string;
    url: string;
    thumbnail?: string;
    size?: string;
    duration?: number;
    category_id?: number;
    tags?: string[];
    privacy?: number;
}

export interface UpdateMediaRequest {
    title?: string;
    description?: string;
    thumbnail?: string;
    category_id?: number | null;
    tags?: string[];
    state?: string;
    privacy?: number;
    featured?: boolean;
}

export const mediaApi = {
    // 获取媒体列表（公开，默认只返回 active 状态）
    list: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        category_id?: number;
        keyword?: string;
        user_id?: number;
        state?: string;
        featured?: string;
        order_by?: string;
        descending?: boolean;
    }) => api.get<MediaListResponse>("/media", params as Record<string, unknown>),

    // 获取媒体详情（公开，自增播放量）
    get: (id: number | string) => api.get<Media>(`/media/${id}`),

    // 管理端：获取所有媒体（包括未发布的）
    adminList: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        state?: string;
        keyword?: string;
    }) => api.get<MediaListResponse>("/media", params as Record<string, unknown>),

    // 上传媒体文件（需要 JWT，支持进度回调，使用 XHR 支持 onUploadProgress）
    upload: (
        file: File,
        metadata: {
            title?: string;
            description?: string;
            category_id?: number;
            tags?: string[];
            privacy?: number;
        },
        onProgress?: (percent: number) => void,
    ) => {
        const formData = new FormData();
        formData.append("file", file);
        if (metadata.title) formData.append("title", metadata.title);
        if (metadata.description) formData.append("description", metadata.description);
        if (metadata.category_id) formData.append("category_id", String(metadata.category_id));
        if (metadata.tags?.length) formData.append("tags", metadata.tags.join(","));
        if (metadata.privacy) formData.append("privacy", String(metadata.privacy));

        const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:9090";
        const token = getAccessToken();

        return new Promise<{ data: Media }>((resolve, reject) => {
            const xhr = new XMLHttpRequest();

            if (onProgress) {
                xhr.upload.addEventListener("progress", (e) => {
                    if (e.lengthComputable) {
                        onProgress(Math.round((e.loaded / e.total) * 100));
                    }
                });
            }

            xhr.addEventListener("load", () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    try {
                        const data = JSON.parse(xhr.responseText);
                        resolve({data});
                    } catch {
                        reject(new Error("Invalid response"));
                    }
                } else {
                    try {
                        const err = JSON.parse(xhr.responseText);
                        reject(new Error(err.error || `Upload failed: ${xhr.status}`));
                    } catch {
                        reject(new Error(`Upload failed: ${xhr.status}`));
                    }
                }
            });

            xhr.addEventListener("error", () => reject(new Error("Network error")));
            xhr.addEventListener("abort", () => reject(new Error("Upload aborted")));

            xhr.open("POST", `${API_BASE_URL}/api/v1/media/upload`);
            if (token) {
                xhr.setRequestHeader("Authorization", `Bearer ${token}`);
            }
            xhr.send(formData);
        });
    },

    // 更新媒体（需要 JWT + owner 权限）
    update: (id: number | string, data: UpdateMediaRequest) =>
        api.put<Media>(`/media/${id}`, data),

    // 删除媒体（需要 JWT + owner 权限）
    delete: (id: number | string) => api.del<void>(`/media/${id}`),
};
