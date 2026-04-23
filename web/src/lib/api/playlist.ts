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
    getMyPlaylists: (params?: { page?: number; page_size?: number }) =>
        api.get<any>("/me/playlists", params),

    create: (data: CreatePlaylistRequest) =>
        api.post<{ playlist: Playlist }>("/me/playlists", data),

    addMedia: (playlistId: string, mediaId: string) =>
        api.post<void>(`/me/playlists/${playlistId}/media`, {media_id: mediaId}),

    removeMedia: (playlistId: string, mediaId: string) =>
        api.del<void>(`/me/playlists/${playlistId}/media/${mediaId}`),
};
