// Like API
import {api} from "../request";

export interface LikeResponse {
    likes: number;
    dislikes: number;
    user_liked: string;
}

export interface ToggleLikeResponse {
    liked: boolean;
}

export const likeApi = {
    toggle: (mediaId: string, type: "like" | "dislike") =>
        api.post<ToggleLikeResponse>("/likes", {media_id: mediaId, type}),
    getMediaLikes: (mediaId: string) =>
        api.get<LikeResponse>(`/likes/media/${mediaId}`),
};
