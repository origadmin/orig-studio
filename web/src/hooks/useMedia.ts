/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// TODO(M2): Implement media state management hooks
// This is a placeholder that provides the media hook interface.

import {useState, useEffect, useCallback} from 'react';

export interface MediaItem {
    id: number;
    title: string;
    description: string;
    thumbnailUrl: string;
    viewCount: number;
    duration?: number;
    authorId: number;
    authorName?: string;
    createdAt: string;
    url?: string;
}

export interface UseMediaListReturn {
    items: MediaItem[];
    loading: boolean;
    error: string | null;
    refresh: () => void;
}

/**
 * useMediaList fetches a list of media items.
 * TODO(M2): Replace with real API call to /api/v1/feed or /api/v1/media
 */
export function useMediaList(params?: { page?: number; pageSize?: number }): UseMediaListReturn {
    const [items, setItems] = useState<MediaItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetch_ = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const qs = new URLSearchParams();
            if (params?.page) qs.set('page', String(params.page));
            if (params?.pageSize) qs.set('page_size', String(params.pageSize));
            const res = await fetch(`/api/v1/feed?${qs}`);
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const data = await res.json();
            const allItems: MediaItem[] = (data.sections ?? [])
                .flatMap((s: any) => s.items ?? [])
                .map((m: any) => ({
                    id: m.id,
                    title: m.title,
                    description: m.description,
                    thumbnailUrl: m.thumbnail_url ?? m.thumbnail ?? '',
                    viewCount: m.view_count ?? 0,
                    authorId: m.author_id ?? m.user_id ?? 0,
                    createdAt: m.created_at ?? m.create_time ?? '',
                }));
            setItems(allItems);
        } catch (e: any) {
            setError(e.message ?? 'Failed to load media');
        } finally {
            setLoading(false);
        }
    }, [params?.page, params?.pageSize]);

    useEffect(() => {
        fetch_();
    }, [fetch_]);

    return {items, loading, error, refresh: fetch_};
}
