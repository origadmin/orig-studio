// Favorite API - 已整合到 media.ts，此文件保留用于向后兼容
// 推荐使用 mediaApi.favorites 替代
import {api} from "../request";
import type {Media} from "../../types";
import {mediaApi} from "./media";

export interface Favorite {
    id: number;
    media_id: number;
    media: Media;
    created_at: string;
}

export interface ToggleFavoriteResponse {
    is_favorited: boolean;
    favorite_count: number;
}

export interface FavoriteListResponse {
    items: Favorite[];
    total: number;
    page: number;
    page_size: number;
}

export const favoriteApi = {
    // 获取收藏状态 - 使用新的 /medias/:id/favorites 路径
    getStatus: (mediaId: string) =>
        mediaApi.favorites.getStatus(mediaId),

    // 收藏/取消收藏 - 使用新的 /medias/:id/favorites 路径
    toggle: (mediaId: string) =>
        mediaApi.favorites.toggle(mediaId),

    // 获取收藏列表
    list: (params?: { page?: number; page_size?: number }) =>
        api.get<FavoriteListResponse>('/favorites', params),

    // 移除收藏
    remove: (favoriteId: string) =>
        api.del<void>(`/favorites/${favoriteId}`),
};

export default favoriteApi;
