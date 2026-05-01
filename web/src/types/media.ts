export interface UserSummary {
    id: string;
    username: string;
    nickname?: string;
    avatar?: string;
    subscriber_count?: number;
}

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
    create_time: string | { seconds: number; nanos: number };
    user_id: number;
    /** @deprecated Use edges.user instead. Kept for backward compatibility. */
    author_name?: string;
    /** @deprecated Use edges.user instead. Kept for backward compatibility. */
    author_avatar?: string;
    category?: string;
    tags?: string[];
    likes?: number;
    dislikes?: number;
    is_premium?: boolean;
    short_token?: string;
    /** Sprite generation status: pending | processing | success | failed */
    sprite_status?: string;
    /** Sprite sheet image path */
    sprite_path?: string;
    /** WebVTT file path */
    vtt_path?: string;
    // Flat edge fields from proto-based API
    user?: UserSummary;
    // Nested edges structure (populated by normalizeMedia)
    edges?: {
        user?: UserSummary[];
        category?: unknown;
        channels?: unknown[];
    };
}
