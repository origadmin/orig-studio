export interface User {
    id: number;
    username: string;
    nickname?: string;
    avatar?: string;
}

export interface Media {
    id: number;
    title: string;
    description?: string;
    thumbnail?: string;
    duration?: number;
    view_count?: number;
    likes?: number;
    created_at?: string;
    type?: string;
    size?: string;
    edges?: {
        user?: Array<{
            id: number;
            username: string;
            nickname?: string;
        }>;
        category?: {
            name: string;
        };
    };
    encoding_status?: string;
    friendly_token?: string;
    tags?: string[];
    state?: string;
}

export interface MediaItem extends Media {
    url: string;
    author_name: string;
    author_avatar: string;
    create_time: string;
}

export interface Comment {
    id: number;
    content: string;
    created_at: string;
    user: User;
}

export interface LikeResponse {
    is_liked: boolean;
    count: number;
}

export interface ToggleLikeResponse {
    success: boolean;
    is_liked: boolean;
    count: number;
}

export interface Favorite {
    id: number;
    media_id: number;
    media: Media;
    created_at: string;
}

export interface ToggleFavoriteResponse {
    success: boolean;
    is_favorited: boolean;
}

export interface UploadTask {
    id: string;
    file: File;
    progress: number;
    status: 'pending' | 'uploading' | 'success' | 'error';
    error?: string;
    uploadId?: string;
    parts?: Array<{
        ETag: string;
        PartNumber: number;
    }>;
    title: string;
    description: string;
    categoryId: number;
    tags: string[];
    privacy: number;
    speed?: number;
}
