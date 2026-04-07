// Favorite API
import {api} from "../request";
import type {Media} from "../../types";

export interface Favorite {
    id: number;
    media_id: number;
    media: Media;
    created_at: string;
}

export interface ToggleFavoriteResponse {
    success: boolean;
    is_favorited: boolean;
}

export interface FavoriteListResponse {
    list: Favorite[];
    total: number;
}

export const favoriteApi = {
    toggle: (mediaId: string) =>
        api.post<ToggleFavoriteResponse>(`/media/${mediaId}/favorite`),
    getStatus: (mediaId: string) =>
        api.get<ToggleFavoriteResponse>(`/media/${mediaId}/favorite`),
    list: () =>
        api.get<FavoriteListResponse>('/favorites'),
    remove: (mediaId: string) =>
        api.del(`/favorites/${mediaId}`),
};