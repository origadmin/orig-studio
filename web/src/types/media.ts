export interface MediaItem {
    id: number;
    title: string;
    description?: string;
    thumbnail: string;
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
