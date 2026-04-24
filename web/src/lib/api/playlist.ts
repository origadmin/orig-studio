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
    updated_at: string;
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
