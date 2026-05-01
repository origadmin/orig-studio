export interface User {
    id: string;
    username: string;
    nickname?: string;
    avatar?: string;
    channel_id?: string;
}

export interface Media {
    id: string;
    title: string;
    description?: string;
    thumbnail?: string;
    duration?: number;
    view_count?: number;
    likes?: number;
    create_time?: string;
    type?: string;
    size?: string;
    user_id?: string;
    channel_id?: string;
    // Flat edge fields (proto-based API returns these at top level)
    user?: {
        id: string;
        username: string;
        nickname?: string;
    };
    category?: {
        id: number;
        name: string;
    };
    channel?: {
        id: string;
        name: string;
    };
    // Nested edges structure (used by frontend components)
    edges?: {
        user?: Array<{
            id: string;
            username: string;
            nickname?: string;
        }>;
        category?: {
            id: number;
            name: string;
        };
        channels?: Array<{
            id: string;
            name: string;
        }>;
    };
    encoding_status?: string;
    short_token?: string;
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
    id: string;
    content: string;
    create_time: string;
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
    id: string;
    media_id: string;
    media: Media;
    create_time: string;
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
