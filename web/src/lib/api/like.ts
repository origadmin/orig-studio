// Like API - 已整合到 media.ts，此文件保留用于向后兼容
// 推荐使用 mediaApi.likes 替代
import {api} from "../request";
import {mediaApi, LikeResponse} from "./media";

export {LikeResponse};

export const likeApi = {
    // 获取点赞状态 - 使用新的 /medias/:id/likes 路径
    getStatus: (mediaId: string) =>
        mediaApi.likes.getStatus(mediaId),

    // 点赞/取消点赞 - 使用新的 /medias/:id/likes 路径
    toggle: (mediaId: string) =>
        mediaApi.likes.toggle(mediaId),

    // 点踩/取消点踩 - 使用新的 /medias/:id/dislikes 路径
    toggleDislike: (mediaId: string) =>
        mediaApi.likes.toggleDislike(mediaId),
};

export default likeApi;
