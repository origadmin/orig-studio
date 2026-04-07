// Share API
import {api} from "../request";

export interface ShareResponse {
    success: boolean;
    url: string;
}

export const shareApi = {
    // 分享视频
    share: (mediaId: string) => api.post<ShareResponse>(`/media/${mediaId}/share`),

    // 获取分享链接
    getShareUrl: (mediaId: string) => api.get<ShareResponse>(`/media/${mediaId}/share`),
};
