import {api} from "../request";
import type {Media} from "./media";

export interface Playlist {
    id: string;
    title: string;
    description?: string;
    user_id: string;
    media_count: number;
    media?: Media[];
    created_at: string;
    create_time?: string;
    updated_at: string;
    update_time?: string;
}

export interface PlaylistListResponse {
    items: Playlist[];
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
    list: (params?: { page?: number; page_size?: number; user_id?: string }) =>
        api.get<PlaylistListResponse>("/playlists", params as Record<string, unknown>),

    getMyPlaylists: (params?: { page?: number; page_size?: number }) =>
        api.get<any>("/me/playlists", params),

    get: (id: string) => api.get<{ playlist: Playlist }>(`/playlists/${id}`),

    create: (data: CreatePlaylistRequest) =>
        api.post<{ playlist: Playlist }>("/me/playlists", data),

    update: (id: string, data: UpdatePlaylistRequest) =>
        api.put<{ playlist: Playlist }>(`/playlists/${id}`, data),

    delete: (id: string) => api.del<void>(`/playlists/${id}`),

    addMedia: (playlistId: string, mediaId: string) =>
        api.post<void>(`/me/playlists/${playlistId}/media`, {media_id: mediaId}),

    removeMedia: (playlistId: string, mediaId: string) =>
        api.del<void>(`/me/playlists/${playlistId}/media/${mediaId}`),

    reorderMedia: (playlistId: string, mediaOrders: Record<number, number>) =>
        api.put<void>(`/playlists/${playlistId}/reorder`, {media_orders: mediaOrders}),
};

// ==================== Admin Playlist API (UUID based, requires JWT + Admin) ====================
export const adminPlaylistApi = {
    // List all playlists (Admin, includes non-public)
    list: (params?: { page?: number; page_size?: number }) =>
        api.get<PlaylistListResponse>("/admin/playlists", params as Record<string, unknown>),

    // Get playlist detail by UUID (Admin)
    get: (id: string) =>
        api.get<{ playlist: Playlist }>(`/admin/playlists/${id}`),

    // Create playlist (Admin)
    create: (data: CreatePlaylistRequest & { user_id: string }) =>
        api.post<{ playlist: Playlist }>("/admin/playlists", data),

    // Update playlist by UUID (Admin)
    update: (id: string, data: UpdatePlaylistRequest) =>
        api.put<{ playlist: Playlist }>(`/admin/playlists/${id}`, data),

    // Delete playlist by UUID (Admin)
    delete: (id: string) =>
        api.del<void>(`/admin/playlists/${id}`),
};
