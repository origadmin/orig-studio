import {api} from "../request";

// PlaylistMediaItem represents a media item within a playlist (simplified for display).
// Matches backend biz.PlaylistMediaItem struct.
export interface PlaylistMediaItem {
    id: string;
    short_token: string;
    title: string;
    thumbnail: string;
    duration: number;
    type: string;
    view_count: number;
    encoding_status: string;
    create_time: string;
}

// Playlist interface matching the backend biz.Playlist struct.
// Backend fields: id, title, description, short_token, user_id, is_public, create_time, update_time, media_items, media_details
export interface Playlist {
    id: string;
    title: string;
    description?: string;
    short_token?: string;
    user_id: string;
    is_public: boolean;
    media_items?: string[];  // Array of media IDs in the playlist
    media_details?: PlaylistMediaItem[];  // Full media details for display
    create_time: string;
    update_time: string;
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
    is_public?: boolean;
}

export interface UpdatePlaylistRequest {
    title?: string;
    description?: string;
    is_public?: boolean;
}

// ==================== User Playlist API (/me/playlists - requires JWT) ====================
export const playlistApi = {
    // List current user's playlists
    getMyPlaylists: (params?: { page?: number; page_size?: number }) =>
        api.get<PlaylistListResponse>("/me/playlists", params as Record<string, unknown>),

    // Get a public playlist by short_token (portal view)
    // Portal routes use short_token, not database id (A005 design principle)
    get: (shortToken: string) =>
        api.get<{ playlist: Playlist }>(`/playlists/${shortToken}`),

    // Create a new playlist for the current user
    create: (data: CreatePlaylistRequest) =>
        api.post<{ playlist: Playlist }>("/me/playlists", data),

    // Update a playlist owned by the current user (PATCH, not PUT)
    update: (id: string, data: UpdatePlaylistRequest) =>
        api.patch<{ playlist: Playlist }>(`/me/playlists/${id}`, data),

    // Delete a playlist owned by the current user
    delete: (id: string) =>
        api.del<void>(`/me/playlists/${id}`),

    // Add a media item to a playlist
    addMedia: (playlistId: string, mediaId: string) =>
        api.post<void>(`/me/playlists/${playlistId}/media`, {media_id: mediaId}),

    // Remove a media item from a playlist
    removeMedia: (playlistId: string, mediaId: string) =>
        api.del<void>(`/me/playlists/${playlistId}/media/${mediaId}`),

    // Reorder media items in a playlist
    reorderMedia: (playlistId: string, mediaOrders: Record<string, number>) =>
        api.patch<void>(`/me/playlists/${playlistId}/media/reorder`, {media_orders: mediaOrders}),
};

// ==================== Admin Playlist API (/admin/playlists - requires JWT + Admin) ====================
export const adminPlaylistApi = {
    // List all playlists (Admin, includes non-public)
    list: (params?: { page?: number; page_size?: number }) =>
        api.get<PlaylistListResponse>("/admin/playlists", params as Record<string, unknown>),

    // Get playlist detail by UUID (Admin)
    get: (id: string) =>
        api.get<{ playlist: Playlist }>(`/admin/playlists/${id}`),

    // Create playlist (Admin)
    create: (data: CreatePlaylistRequest & { user_id: string; is_public?: boolean }) =>
        api.post<{ playlist: Playlist }>("/admin/playlists", data),

    // Update playlist by UUID (Admin)
    update: (id: string, data: UpdatePlaylistRequest) =>
        api.put<{ playlist: Playlist }>(`/admin/playlists/${id}`, data),

    // Delete playlist by UUID (Admin)
    delete: (id: string) =>
        api.del<void>(`/admin/playlists/${id}`),
};
