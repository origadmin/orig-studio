export interface MediaItem {
    id: number;
    title: string;
    description?: string;
    url: string;
    thumbnail: string;
    hls_file?: string;
    encoding_status?: 'pending' | 'processing' | 'success' | 'failed';
    duration: number;
    view_count: number;
    create_time: string;
    user_id: number;
    author_name: string;
    author_avatar?: string;
    category?: string;
    tags?: string[];
    likes?: number;
    dislikes?: number;
    is_premium?: boolean;
}
