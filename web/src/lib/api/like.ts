// Like API
import {api} from "../request";

export interface LikeResponse {
    is_liked: boolean;
    is_disliked: boolean;
    like_count: number;
    dislike_count: number;
}

export const likeApi = {
    toggle: (mediaId: string) =>
        api.post<LikeResponse>(`/media/${mediaId}/like`),
    toggleDislike: (mediaId: string) =>
        api.post<LikeResponse>(`/media/${mediaId}/dislike`),
    getStatus: (mediaId: string) =>
        api.get<LikeResponse>(`/media/${mediaId}/like`),
};