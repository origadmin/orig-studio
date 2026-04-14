// API 客户端 - 媒体模块
// 对应后端 /api/v1 路径
// 类型定义对齐后端 ent entity JSON 输出
import {api, getAccessToken, API_BASE_URL} from "../request";

// 统一响应格式接口
export interface ApiResponse<T> {
    code: number;
    message: string;
    data: T;
}

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
    encoding_status: string; // "pending" | "processing" | "success" | "partial" | "failed"
    state: string; // "draft" | "active" | "deleted"
    uuid?: string; // secure unique ID for public paths (HLS, thumbnails)
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
    create_time?: { seconds: number; nanos: number };
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
    items: Media[];
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

export interface EncodeProfile {
    id: number;
    name: string;
    description: string;
    extension: string;
    resolution: string;
    video_codec: string;
    video_bitrate: string;
    audio_codec: string;
    audio_bitrate: string;
    is_active: boolean;
}

export interface EncodingTask {
    id: number;
    media_id: number;
    profile_id: number;
    status: string; // "pending" | "processing" | "success" | "failed"
    progress: number;
    output_path: string;
    error_message: string;
    create_time: string;
    update_time: string;
}

export interface TranscodingStatusResponse {
    processing_count: number;
    pending_count: number;
    partial_count: number;
    failed_count: number;
    success_count: number;
    total: number;
    page: number;
    page_size: number;
    items: TranscodingMediaItem[];
}

export interface TranscodingMediaItem {
    media: Media;
    tasks: EncodingTask[];
}

export interface EncodingTaskListResponse {
    processing_count: number;
    pending_count: number;
    partial_count: number;
    failed_count: number;
    success_count: number;
    total: number;
    page: number;
    page_size: number;
    items: (EncodingTask & { profile_name?: string })[];
}

// 点赞/收藏响应
export interface LikeResponse {
    is_liked: boolean;
    is_disliked: boolean;
    like_count: number;
    dislike_count: number;
}

export interface FavoriteResponse {
    is_favorited: boolean;
    favorite_count: number;
}

export interface ShareResponse {
    url: string;
    title: string;
    twitter: string;
    facebook: string;
    linkedin: string;
    whatsapp: string;
    telegram: string;
}

// ==================== Encoding Module ====================
export const encodingApi = {
    // 获取转码事件流（SSE）
    getSSEUrl: (mediaId?: number) => {
        // 使用相对路径，让前端代理处理
        return `/api/v1/encoding/events${mediaId ? `?media_id=${mediaId}` : ""}`;
    },

    // 获取所有转码任务（扁平列表）
    // 当 status=all 时，返回完整统计
    // 当 status 为其他值时，返回过滤后的统计
    // 当 only_stats=true 时，只返回统计，不返回任务列表
    getTasks: (params?: {
        status?: string;
        page?: number;
        page_size?: number;
        media_id?: number;
        only_stats?: boolean;
    }) => api.get<EncodingTaskListResponse>("/encoding/tasks", params as Record<string, unknown>),

    // 重试单个任务
    retryTask: (taskId: number) => {
        return api.post<{ message: string; task: any }>('/encoding/retry', null, {
            params: {task_id: taskId}
        });
    },

    // 重试所有失败任务
    retryAllFailed: (mediaId?: number) => {
        return api.post<{ message: string; retried_count: number }>('/encoding/retry-all-failed', null, {
            params: {media_id: mediaId}
        });
    },

    // 编码配置管理
    profiles: {
        list: () => api.get<{ profiles: EncodeProfile[] }>('/encoding/profiles'),
        get: (id: number) => api.get<{ profile: EncodeProfile }>(`/encoding/profiles/${id}`),
        create: (data: Partial<EncodeProfile>) => api.post<{
            profile: EncodeProfile
        }>('/encoding/profiles', data),
        update: (id: number, data: Partial<EncodeProfile>) =>
            api.put<{ profile: EncodeProfile }>(`/encoding/profiles/${id}`, data),
        delete: (id: number) => api.del<void>(`/encoding/profiles/${id}`),
    },
};

// ==================== Media API ====================
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
    }) => api.get<MediaListResponse>("/medias", params as Record<string, unknown>),

    // 获取媒体详情（公开，自增播放量）
    get: (id: number | string) => {
        const cleanId = String(id).replace(/["']/g, '').trim();
        return api.get<Media>(`/medias/${cleanId}`);
    },

    // 管理端：获取所有媒体（包括未发布的）
    adminList: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        state?: string;
        keyword?: string;
    }) => api.get<MediaListResponse>("/medias", params as Record<string, unknown>),

    // 上传媒体文件（需要 JWT，支持进度回调）
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
                        const response = JSON.parse(xhr.responseText);
                        // 适配新的统一响应格式 {code, message, data}
                        const data = response.data || response;
                        resolve({data});
                    } catch {
                        reject(new Error("Invalid response"));
                    }
                } else {
                    try {
                        const err = JSON.parse(xhr.responseText);
                        reject(new Error(err.message || err.error || `Upload failed: ${xhr.status}`));
                    } catch {
                        reject(new Error(`Upload failed: ${xhr.status}`));
                    }
                }
            });

            xhr.addEventListener("error", () => reject(new Error("Network error")));
            xhr.addEventListener("abort", () => reject(new Error("Upload aborted")));

            xhr.open("POST", `${API_BASE_URL}/medias/upload`);
            if (token) {
                xhr.setRequestHeader("Authorization", `Bearer ${token}`);
            }
            xhr.send(formData);
        });
    },

    // 更新媒体（需要 JWT + owner 权限）
    update: (id: number | string, data: UpdateMediaRequest) =>
        api.put<Media>(`/medias/${id}`, data),

    // 删除媒体（需要 JWT + owner 权限）
    delete: (id: number | string) => api.del<void>(`/medias/${id}`),

    // 文件操作
    download: (id: number | string) => api.get<{ download_url: string }>(`/medias/${id}/download`),
    stream: (id: number | string) => api.get<{ stream_url: string }>(`/medias/${id}/stream`),
    getThumbnail: (id: number | string) => api.get<{ thumbnail_url: string }>(`/medias/${id}/thumbnail`),

    // 转码相关（单个媒体）
    encoding: {
        // 获取媒体转码任务
        getTasks: (mediaId: number | string) =>
            api.get<{ tasks: EncodingTask[] }>(`/medias/${mediaId}/tasks`),

        // 获取媒体转码变体
        getVariants: (mediaId: number | string) =>
            api.get<MediaVariantSummary>(`/medias/${mediaId}/variants`),

        // 重试媒体转码
        retry: (mediaId: number | string) =>
            api.post<{ message: string; media_id: number }>(`/medias/${mediaId}/tasks/:taskId/retry`),
    },

    // 旧版转码状态（兼容）
    getTranscodingStatus: (params?: {
        status?: string;
        page?: number;
        page_size?: number;
    }) => api.get<TranscodingStatusResponse>("/encoding/status", params as Record<string, unknown>),

    // 获取转码事件流（SSE）
    getSSEUrl: (mediaId?: number) => {
        // 使用相对路径，让前端代理处理
        return `/api/v1/encoding/events${mediaId ? `?media_id=${mediaId}` : ""}`;
    },

    // ==================== 点赞/点踩 API ====================
    likes: {
        // 获取点赞状态
        getStatus: (mediaId: string | number) =>
            api.get<LikeResponse>(`/medias/${mediaId}/likes`),
        // 点赞/取消点赞
        toggle: (mediaId: string | number) =>
            api.post<LikeResponse>(`/medias/${mediaId}/likes`),
        // 点踩/取消点踩
        toggleDislike: (mediaId: string | number) =>
            api.post<LikeResponse>(`/medias/${mediaId}/dislikes`),
    },

    // ==================== 收藏 API ====================
    favorites: {
        // 获取收藏状态
        getStatus: (mediaId: string | number) =>
            api.get<FavoriteResponse>(`/medias/${mediaId}/favorites`),
        // 收藏/取消收藏
        toggle: (mediaId: string | number) =>
            api.post<FavoriteResponse>(`/medias/${mediaId}/favorites`),
    },

    // ==================== 分享 API ====================
    shares: {
        // 获取分享链接
        getShareUrl: (mediaId: string | number, title?: string) =>
            api.get<ShareResponse>(`/medias/${mediaId}/shares`, title ? {title} : {}),
        // 分享视频（增加分享计数）
        share: (mediaId: string | number) =>
            api.post<{ success: boolean }>(`/medias/${mediaId}/shares`),
    },
};

// MediaVariantSummary is the aggregated transcoding status for a single media.
export interface MediaVariantSummary {
    media_id: number;
    uuid: string;
    encoding_status: string;
    hls_file?: string;
    thumbnail?: string;
    preview_file?: string;
    video_total_count: number;
    video_success_count: number;
    video_failed_count: number;
    video_pending_count?: number;
    video_processing_count?: number;
    variants: VariantInfo[];
}

export interface VariantInfo {
    task_id: number;
    profile_name: string;
    profile_id: number;
    resolution: string;
    codec: string;
    status: string;
    output_path: string;
    bandwidth: number;
    error_message?: string;
}

// 为了保持向后兼容，导出 encodingApi 的方法到 mediaApi
// 这些将在未来版本中移除
export const legacyMediaApi = {
    // 旧版路径 - 将在未来版本中移除
    listProfiles: () => api.get<{ profiles: EncodeProfile[] }>("/encoding/profiles"),
    getProfile: (id: number) => api.get<{ profile: EncodeProfile }>(`/encoding/profiles/${id}`),
    createProfile: (data: Partial<EncodeProfile>) => api.post<{ profile: EncodeProfile }>("/encoding/profiles", data),
    updateProfile: (id: number, data: Partial<EncodeProfile>) =>
        api.put<{ profile: EncodeProfile }>(`/encoding/profiles/${id}`, data),
    deleteProfile: (id: number) => api.del<void>(`/encoding/profiles/${id}`),
    getEncodingTasks: (params?: {
        status?: string;
        page?: number;
        page_size?: number;
        media_id?: number;
    }) => api.get<EncodingTaskListResponse>("/encoding/tasks", params as Record<string, unknown>),
    listTasks: (mediaId: number) => api.get<{ tasks: EncodingTask[] }>(`/medias/${mediaId}/tasks`),
    retryTranscode: (mediaId: number) =>
        api.post<{ message: string; media_id: number }>(`/medias/${mediaId}/tasks/:taskId/retry`),
    retryTask: (taskId: number) => {
        return fetch(`${API_BASE_URL}/encoding/retry?task_id=${taskId}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                ...(getAccessToken() ? {Authorization: `Bearer ${getAccessToken()}`} : {}),
            },
        }).then((r) => !r.ok ? r.json().then((e) => Promise.reject(e)) : r.json());
    },
    retryAllFailed: (mediaId: number) => {
        return fetch(`${API_BASE_URL}/encoding/retry-all-failed?media_id=${mediaId}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                ...(getAccessToken() ? {Authorization: `Bearer ${getAccessToken()}`} : {}),
            },
        }).then((r) => !r.ok ? r.json().then((e) => Promise.reject(e)) : r.json());
    },
    getVariants: (mediaId: number) =>
        api.get<MediaVariantSummary>(`/medias/${mediaId}/variants`),
    getSSEUrl: (mediaId?: number) => {
        return `${API_BASE_URL}/encoding/events${mediaId ? `?media_id=${mediaId}` : ""}`;
    },
};
