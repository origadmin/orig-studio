// API 客户端 - 播放列表模块
// 对应后端 /api/v1/playlists 路径
import {api} from "../request";
import type {Media} from "./media";

export interface Playlist {
    id: number;
    title: string;
    description?: string;
    user_id: number;
    media_count: number;
    media?: Media[];
    created_at: string;
    updated_at: string;
}

export interface PlaylistListResponse {
    list: Playlist[];
    total: number;
    page: number;
    page_size: number;
}

export interface CreatePlaylistRequest {
    title: string;
    description?: string;
}

export interface UpdatePlaylistRequest {
    title?: string;
    description?: string;
}

export const playlistApi = {
    // 获取播放列表（公开）
    list: (params?: { page?: number; page_size?: number; user_id?: number }) =>
        api.get<PlaylistListResponse>("/playlists", params as Record<string, unknown>),

    // 获取我的播放列表
    getMyPlaylists: (params?: { page?: number; page_size?: number }) =>
        api.get<PlaylistListResponse>("/playlists/me", params),

    // 获取播放列表详情
    get: (id: string) => api.get<{ playlist: Playlist }>(`/playlists/${id}`),

    // 创建播放列表
    create: (data: CreatePlaylistRequest) =>
        api.post<{ playlist: Playlist }>("/playlists", data),

    // 更新播放列表
    update: (id: string, data: UpdatePlaylistRequest) =>
        api.put<{ playlist: Playlist }>(`/playlists/${id}`, data),

    // 删除播放列表
    delete: (id: string) => api.del<void>(`/playlists/${id}`),

    // 添加媒体到播放列表
    addMedia: (playlistId: string, mediaId: string) =>
        api.post<void>(`/playlists/${playlistId}/media`, {media_id: mediaId}),

    // 从播放列表移除媒体
    removeMedia: (playlistId: string, mediaId: string) =>
        api.del<void>(`/playlists/${playlistId}/media/${mediaId}`),

    // 重新排序播放列表中的媒体
    reorderMedia: (playlistId: string, mediaOrders: Record<number, number>) =>
        api.put<void>(`/playlists/${playlistId}/reorder`, {media_orders: mediaOrders}),
};
