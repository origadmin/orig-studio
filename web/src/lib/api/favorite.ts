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
    toggle: (params: { media_id: string }) =>
        api.post<ToggleFavoriteResponse>("/api/v1/media/:mediaId/favorite", {}, {params}),
    getStatus: (params: { media_id: string }) =>
        api.get<ToggleFavoriteResponse>("/api/v1/media/:mediaId/favorite", {params}),
    list: () =>
        api.get<FavoriteListResponse>("/api/v1/favorites"),
    remove: (params: { media_id: string }) =>
        api.del("/api/v1/favorites/:mediaId", {params}),
};