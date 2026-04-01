// Playlist API
import {api} from "../request";
import {Media} from "./media";

export interface Playlist {
    id: string;
    name: string;
    description?: string;
    user_id: string;
    media_ids: string[];
    is_public: boolean;
    created_at: string;
    updated_at: string;
}

export interface PlaylistDetail {
    playlist: Playlist;
    media: Media[];
}

export const playlistApi = {
    getAll: (userId?: string) =>
        api.get<Playlist[]>("/playlists", {user_id: userId}),
    get: (id: string) => api.get<PlaylistDetail>(`/playlists/${id}`),
    create: (data: Partial<Playlist>) => api.post<Playlist>("/playlists", data),
    update: (id: string, data: Partial<Playlist>) =>
        api.put<Playlist>(`/playlists/${id}`, data),
    delete: (id: string) => api.del<void>(`/playlists/${id}`),
    addMedia: (playlistId: string, mediaId: string) =>
        api.post<Playlist>(`/playlists/${playlistId}/media`, {media_id: mediaId}),
    removeMedia: (playlistId: string, mediaId: string) =>
        api.del<void>(`/playlists/${playlistId}/media/${mediaId}`),
};
