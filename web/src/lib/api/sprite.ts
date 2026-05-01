import {api, API_BASE_URL} from "@/lib/request";

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
    /** Get the WebVTT sprite sheet URL for a media item (by short_token) */
    getVttUrl: (mediaId: string) =>
        `${API_BASE_URL}/medias/${mediaId}/sprite.vtt`,

    /** Get the sprite sheet JPEG URL for a media item (by short_token) */
    getSpriteUrl: (mediaId: string) =>
        `${API_BASE_URL}/medias/${mediaId}/sprite.jpg`,

    /** Trigger asynchronous sprite sheet regeneration (admin only) */
    regenerateSprite: (mediaId: string) =>
        api.post<RegenerateSpriteResponse>(`/admin/medias/${mediaId}/regenerate-sprite`),

    /** Trigger thumbnail regeneration at an optional timestamp (admin only) */
    regenerateThumbnail: (mediaId: string, data?: RegenerateThumbnailRequest) =>
        api.post<RegenerateThumbnailResponse>(`/admin/medias/${mediaId}/regenerate-thumbnail`, data),
};
