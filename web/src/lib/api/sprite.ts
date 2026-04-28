import {api, API_BASE_URL} from "../request";

export interface RegenerateSpriteResponse {
    media_id: string;
    sprite_status: string;
    message: string;
}

export interface RegenerateThumbnailRequest {
    timestamp?: number;
}

export interface RegenerateThumbnailResponse {
    media_id: string;
    thumbnail: string;
    thumbnail_time: number;
    message: string;
}

export const spriteApi = {
    getVttUrl: (mediaId: string) =>
        `${API_BASE_URL}/medias/${mediaId}/sprite.vtt`,

    getSpriteUrl: (mediaId: string) =>
        `${API_BASE_URL}/medias/${mediaId}/sprite.jpg`,

    regenerateSprite: (mediaId: string) =>
        api.post<RegenerateSpriteResponse>(`/admin/medias/${mediaId}/regenerate-sprite`),

    regenerateThumbnail: (mediaId: string, data?: RegenerateThumbnailRequest) =>
        api.post<RegenerateThumbnailResponse>(`/admin/medias/${mediaId}/regenerate-thumbnail`, data),
};
