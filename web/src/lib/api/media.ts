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
    id: string;
    title: string;
    description?: string;
    short_token?: string;
    type: string;
    url: string;
    hls_file?: string;
    thumbnail?: string;
    poster?: string;
    preview_file_path?: string;
    preview_file?: string;
    duration: number;
    size?: string;
    width: number;
    height: number;
    mime_type?: string;
    md5sum?: string;
    extension?: string;
    privacy: number;
    encoding_status: string;
    state: string;
    view_count: number;
    like_count: number;
    dislike_count: number;
    comment_count: number;
    favorite_count: number;
    download_count?: number;
    allow_download?: boolean;
    enable_comments?: boolean;
    featured?: boolean;
    review_status?: string;
    listable?: boolean;
    reported_times?: number;
    tags?: string[];
    user_id: string;
    channel_id?: string;
    category_id?: string;
    published_at?: string;
    created_at: string;
    updated_at?: string;
    edges?: {
        user?: UserSummary[];
        category?: CategorySummary;
        channels?: ChannelSummary[];
        comments?: unknown[];
        playlists?: unknown[];
        tags_rel?: unknown[];
        favorites?: unknown[];
        likes?: unknown[];
    };
}

export interface UserSummary {
    id: string;
    username: string;
    nickname?: string;
    avatar?: string;
    subscriber_count?: number;
}

export interface ChannelSummary {
    id: string;
    name: string;
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
    enable_comments?: boolean;
    allow_download?: boolean;
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
    id: string;
    media_id: string;
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
// NOTE: 后端路径 /api/v1/admin/encoding/* (有 admin 前缀)
export const encodingApi = {
    // 获取转码事件流（SSE）
    getSSEUrl: (mediaId?: string) => {
        return `/api/v1/admin/encoding/events${mediaId ? `?media_id=${mediaId}` : ""}`;
    },

    // 获取所有转码任务（扁平列表）
    getTasks: (params?: {
        status?: string;
        page?: number;
        page_size?: number;
        media_id?: string;
        only_stats?: boolean;
    }) => api.get<EncodingTaskListResponse>('/admin/encoding/tasks', params as Record<string, unknown>),

    // 重试单个任务
    retryTask: (taskId: string) => {
        return api.post<{ message: string; task: any }>(`/admin/encoding/tasks/${taskId}/retry`);
    },

    // 重试所有失败任务
    retryAllFailed: (mediaId?: string) => {
        return api.post<{ message: string; retried_count: number }>('/admin/encoding/retry-failed', null, {
            params: {media_id: mediaId}
        });
    },

    // 编码配置管理
    profiles: {
        list: () => api.get<{ profiles: EncodeProfile[] }>('/admin/encoding/profiles'),
        get: (id: number) => api.get<{ profile: EncodeProfile }>(`/admin/encoding/profiles/${id}`),
        create: (data: Partial<EncodeProfile>) => api.post<{
            profile: EncodeProfile
        }>('/admin/encoding/profiles', data),
        update: (id: number, data: Partial<EncodeProfile>) =>
            api.put<{ profile: EncodeProfile }>(`/admin/encoding/profiles/${id}`, data),
        delete: (id: number) => api.del<void>(`/admin/encoding/profiles/${id}`),
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
    }) => api.get<MediaListResponse>("/admin/medias", params as Record<string, unknown>),

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
    getSSEUrl: (mediaId?: string) => {
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
    media_id: string;
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
    task_id: string;
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
    // 旧版路径 - 已在 encodingApi 中统一
    listProfiles: () => api.get<{ profiles: EncodeProfile[] }>("/admin/encoding/profiles"),
    getProfile: (id: number) => api.get<{ profile: EncodeProfile }>(`/admin/encoding/profiles/${id}`),
    createProfile: (data: Partial<EncodeProfile>) => api.post<{ profile: EncodeProfile }>("/admin/encoding/profiles", data),
    updateProfile: (id: number, data: Partial<EncodeProfile>) =>
        api.put<{ profile: EncodeProfile }>(`/admin/encoding/profiles/${id}`, data),
    deleteProfile: (id: number) => api.del<void>(`/admin/encoding/profiles/${id}`),
    getEncodingTasks: (params?: {
        status?: string;
        page?: number;
        page_size?: number;
        media_id?: number;
    }) => api.get<EncodingTaskListResponse>("/admin/encoding/tasks", params as Record<string, unknown>),
    listTasks: (mediaId: number) => api.get<{ tasks: EncodingTask[] }>(`/medias/${mediaId}/tasks`),
    retryTranscode: (mediaId: number) =>
        api.post<{ message: string; media_id: number }>(`/medias/${mediaId}/tasks/:taskId/retry`),
    retryTask: (taskId: string) => {
        return api.post<{ message: string; task: any }>(`/admin/encoding/tasks/${taskId}/retry`);
    },
    retryAllFailed: (mediaId: number) => {
        return api.post<{ message: string; retried_count: number }>('/admin/encoding/retry-failed', null, {
            params: {media_id: mediaId}
        });
    },
    getVariants: (mediaId: number) =>
        api.get<MediaVariantSummary>(`/medias/${mediaId}/variants`),
    getSSEUrl: (mediaId?: number) => {
        return `/api/v1/admin/encoding/events${mediaId ? `?media_id=${mediaId}` : ""}`;
    },
};

// ==================== Public Media API (short_token based) ====================
// MediaCMS style: /api/v1/medias/{short_token}
// 用于公开页面、Watch 页面、用户交互操作
// 无需认证或可选 JWT 认证
export const publicMediaApi = {
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

    // 获取媒体公开详情（使用 short_token）
    // 返回公开字段，不包含敏感信息
    // 自动增加观看计数
    get: (shortToken: string) => {
        const cleanToken = String(shortToken).replace(/["']/g, '').trim();
        return api.get<Media>(`/medias/${cleanToken}`);
    },

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

    // ==================== 点赞/点踩 API (使用 short_token) ====================
    likes: {
        // 获取点赞状态（无需认证）
        getStatus: (shortToken: string) =>
            api.get<LikeResponse>(`/medias/${shortToken}/likes`),
        // 点赞/取消点赞（需要 JWT）
        toggle: (shortToken: string) =>
            api.post<LikeResponse>(`/medias/${shortToken}/likes`),
        // 点踩/取消点踩（需要 JWT）
        toggleDislike: (shortToken: string) =>
            api.post<LikeResponse>(`/medias/${shortToken}/dislikes`),
    },

    // ==================== 收藏 API (使用 short_token) ====================
    favorites: {
        // 获取收藏状态（无需认证）
        getStatus: (shortToken: string) =>
            api.get<FavoriteResponse>(`/medias/${shortToken}/favorites`),
        // 收藏/取消收藏（需要 JWT）
        toggle: (shortToken: string) =>
            api.post<FavoriteResponse>(`/medias/${shortToken}/favorites`),
    },

    // ==================== 分享 API (使用 short_token) ====================
    shares: {
        // 获取分享链接（返回 /watch?v={short_token} 格式）
        getShareUrl: (shortToken: string) =>
            api.get<ShareResponse>(`/medias/${shortToken}/shares`),
        // 分享视频（增加分享计数，需要 JWT）
        share: (shortToken: string) =>
            api.post<{ success: boolean }>(`/medias/${shortToken}/shares`),
    },
};

// ==================== Admin Media API (ID based, requires JWT + Admin) ====================
// MediaCMS style: /api/v1/admin/medias/:id
// 用于管理后台、CRUD 操作、完整数据访问
// 需要 JWT + Admin 角色权限
export const adminMediaApi = {
    // 管理端：获取所有媒体（包括未发布的，支持更多过滤条件）
    list: (params?: {
        page?: number;
        page_size?: number;
        type?: string;
        state?: string;
        keyword?: string;
        user_id?: number | string;
        category_id?: number | string;
        featured?: boolean;
        tags?: string[];
        order_by?: string;
        descending?: boolean;
    }) => api.get<MediaListResponse>("/admin/medias", params as Record<string, unknown>),

    // 获取媒体完整详情（返回所有字段，包括私有视频）
    // 使用 UUID ID，不接受 short_token
    getById: (id: string) => api.get<Media>(`/admin/medias/${id}`),

    // 更新媒体（Admin 可以编辑任何媒体）
    update: (id: string, data: UpdateMediaRequest) =>
        api.put<Media>(`/admin/medias/${id}`, data),

    // 删除媒体（Admin 可以删除任何媒体）
    delete: (id: string) => api.del<void>(`/admin/medias/${id}`),

    // 获取统计数据
    getStats: (id: string) =>
        api.get<{
            id: string;
            view_count: number;
            like_count: number;
            dislike_count: number;
            comment_count: number;
            favorite_count: number;
            encoding_status: string;
        }>(`/admin/medias/${id}/stats`),

    // 获取转码变体信息（Admin 版本，返回详细数据）
    getVariants: (id: string) =>
        api.get<MediaVariantSummary>(`/admin/medias/${id}/variants`),

    // 变更媒体状态（用于审核流程）
    changeState: (id: string, state: string, comment?: string) =>
        api.put<{
            id: string;
            state: string;
            updated_at: string;
            changed_by: string;
        }>(`/admin/medias/${id}/state`, {state, comment}),

    // 获取编码任务列表
    getTasks: (id: string) =>
        api.get<{ tasks: EncodingTask[] }>(`/admin/medias/${id}/tasks`),

    // 重试编码任务
    retryTask: (id: string, taskId: string) =>
        api.post<{ message: string }>(`/admin/medias/${id}/tasks/${taskId}/retry`),
};
