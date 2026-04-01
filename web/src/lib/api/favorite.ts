// Favorite API
import {api} from "../request";
import {Media} from "./media";

export interface Favorite {
    id: string;
    media_id: string;
    media?: Media;
    created_at: string;
}

export interface ToggleFavoriteResponse {
    favorited: boolean;
}

export const favoriteApi = {
    getAll: () => api.get<Favorite[]>("/favorites"),
    toggle: (mediaId: string) =>
        api.post<ToggleFavoriteResponse>("/favorites", {media_id: mediaId}),
    remove: (id: string) => api.del<void>(`/favorites/${id}`),
};
