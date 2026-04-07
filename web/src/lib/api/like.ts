// Like API
import {api} from "../request";

export interface LikeResponse {
    is_liked: boolean;
    count: number;
}

export interface ToggleLikeResponse {
    success: boolean;
    is_liked: boolean;
    count: number;
}

export const likeApi = {
    toggle: (mediaId: string) =>
        api.post<LikeResponse>(`/media/${mediaId}/like`),
    getStatus: (mediaId: string) =>
        api.get<LikeResponse>(`/media/${mediaId}/like`),
};