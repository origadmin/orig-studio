import {getFullUrl} from './utils';

// 本地默认图片路径 - 使用更语义化的命名和目录结构
const PLACEHOLDER_IMAGES = {
    videoThumbnail: '/assets/images/video.svg',
    userAvatar: '/assets/images/avatar.svg',
    coverImage: '/assets/images/cover.svg',
};

/**
 * 获取图片URL，支持远程->本地切换和失败回退
 * @param path 图片路径
 * @param type 图片类型
 * @returns 处理后的图片URL
 */
export const getImageUrl = (path?: string, type: 'thumbnail' | 'avatar' | 'cover' = 'thumbnail'): string => {
    // 如果有路径，使用getFullUrl处理
    if (path) {
        return getFullUrl(path) || PLACEHOLDER_IMAGES.videoThumbnail;
    }

    // 根据类型返回本地占位图片
    switch (type) {
        case 'avatar':
            return PLACEHOLDER_IMAGES.userAvatar;
        case 'cover':
            return PLACEHOLDER_IMAGES.coverImage;
        case 'thumbnail':
        default:
            return PLACEHOLDER_IMAGES.videoThumbnail;
    }
};

/**
 * 处理图片加载失败的事件
 * @param event 图片加载失败事件
 * @param type 图片类型
 */
export const handleImageError = (event: React.SyntheticEvent<HTMLImageElement>, type: 'thumbnail' | 'avatar' | 'cover' = 'thumbnail') => {
    const img = event.currentTarget;
    // 切换到本地占位图片
    switch (type) {
        case 'avatar':
            img.src = PLACEHOLDER_IMAGES.userAvatar;
            break;
        case 'cover':
            img.src = PLACEHOLDER_IMAGES.coverImage;
            break;
        case 'thumbnail':
        default:
            img.src = PLACEHOLDER_IMAGES.videoThumbnail;
            break;
    }
};
