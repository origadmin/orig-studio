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
    // TODO: 后端尚未注册 /medias/:id/sprite.vtt 路由
    getVttUrl: (mediaId: string) =>
        `${API_BASE_URL}/medias/${mediaId}/sprite.vtt`,

    // TODO: 后端尚未注册 /medias/:id/sprite.jpg 路由
    getSpriteUrl: (mediaId: string) =>
        `${API_BASE_URL}/medias/${mediaId}/sprite.jpg`,

    // TODO: 后端尚未注册 POST /admin/medias/:id/regenerate-sprite 路由
    regenerateSprite: (mediaId: string) =>
        api.post<RegenerateSpriteResponse>(`/admin/medias/${mediaId}/regenerate-sprite`),

    // TODO: 后端尚未注册 POST /admin/medias/:id/regenerate-thumbnail 路由
    regenerateThumbnail: (mediaId: string, data?: RegenerateThumbnailRequest) =>
        api.post<RegenerateThumbnailResponse>(`/admin/medias/${mediaId}/regenerate-thumbnail`, data),
};
