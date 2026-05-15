import {useRef, useCallback, useEffect} from 'react';
import {historySyncManager} from '@/lib/services/history';
import type {ContentType} from '@/lib/services/history';

interface UseWatchProgressOptions {
    contentId: string;
    contentType: ContentType;
    duration: number;
    enabled?: boolean;
    title?: string;
    thumbnail?: string;
    shortToken?: string;
}

export function useWatchProgress({
    contentId,
    contentType,
    duration,
    enabled = true,
    title,
    thumbnail,
    shortToken,
}: UseWatchProgressOptions) {
    const lastReportRef = useRef(0);
    const currentTimeRef = useRef(0);
    const THROTTLE_MS = 10000;

    const handleTimeUpdate = useCallback((time: number) => {
        currentTimeRef.current = time;
        if (!enabled) return;

        const now = Date.now();
        if (now - lastReportRef.current < THROTTLE_MS) return;
        lastReportRef.current = now;

        historySyncManager.reportProgress(
            contentId,
            contentType,
            Math.floor(time),
            duration,
            {title, thumbnail, shortToken},
        );
    }, [contentId, contentType, duration, enabled, title, thumbnail, shortToken]);

    const handlePause = useCallback(() => {
        if (!enabled) return;
        historySyncManager.reportProgress(
            contentId,
            contentType,
            Math.floor(currentTimeRef.current),
            duration,
            {title, thumbnail, shortToken},
        );
    }, [contentId, contentType, duration, enabled, title, thumbnail, shortToken]);

    const handleEnded = useCallback(() => {
        if (!enabled) return;
        historySyncManager.reportProgress(
            contentId,
            contentType,
            duration,
            duration,
            {title, thumbnail, shortToken},
        );
    }, [contentId, contentType, duration, enabled, title, thumbnail, shortToken]);

    // beforeunload: fetch + keepalive (supports Authorization header)
    useEffect(() => {
        if (!enabled) return;

        const handleBeforeUnload = () => {
            const progress = Math.floor(currentTimeRef.current);
            if (progress <= 0) return;

            const data = {
                content_id: contentId,
                content_type: contentType,
                progress_seconds: progress,
                duration_seconds: duration,
                title: title || '',
                thumbnail: thumbnail || '',
                short_token: shortToken || '',
            };

            const token = localStorage.getItem('origstudio_token');
            const headers: Record<string, string> = {
                'Content-Type': 'application/json',
            };
            if (token) {
                headers['Authorization'] = `Bearer ${token}`;
            }

            fetch('/api/v1/me/history', {
                method: 'POST',
                headers,
                body: JSON.stringify(data),
                keepalive: true,
            }).catch(() => {});
        };

        window.addEventListener('beforeunload', handleBeforeUnload);
        return () => window.removeEventListener('beforeunload', handleBeforeUnload);
    }, [contentId, contentType, duration, enabled, title, thumbnail, shortToken]);

    // visibilitychange: save progress when tab becomes hidden
    useEffect(() => {
        if (!enabled) return;

        const handleVisibilityChange = () => {
            if (document.hidden) {
                historySyncManager.reportProgress(
                    contentId,
                    contentType,
                    Math.floor(currentTimeRef.current),
                    duration,
                    {title, thumbnail, shortToken},
                );
            }
        };

        document.addEventListener('visibilitychange', handleVisibilityChange);
        return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
    }, [contentId, contentType, duration, enabled, title, thumbnail, shortToken]);

    // online event: sync pending items
    useEffect(() => {
        if (!enabled) return;

        const handleOnline = () => {
            historySyncManager.syncPendingItems();
        };

        window.addEventListener('online', handleOnline);
        return () => window.removeEventListener('online', handleOnline);
    }, [enabled]);

    return {handleTimeUpdate, handlePause, handleEnded};
}
