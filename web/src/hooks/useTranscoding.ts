/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useEffect, useState} from 'react';
import {mediaApi} from '../lib/api/media';

interface TranscodingEvent {
    media_id: number;
    task_id: number;
    status: 'pending' | 'processing' | 'success' | 'failed';
    progress: number;
}

export function useTranscoding(mediaId?: number) {
    const [lastEvent, setLastEvent] = useState<TranscodingEvent | null>(null);
    const [status, setStatus] = useState<'connected' | 'disconnected' | 'error'>('disconnected');

    useEffect(() => {
        if (!mediaId) return;

        const sseUrl = mediaApi.getSSEUrl(mediaId);
        const eventSource = new EventSource(sseUrl);

        eventSource.onopen = () => {
            setStatus('connected');
        };

        eventSource.onerror = () => {
            setStatus('error');
            eventSource.close();
        };

        eventSource.addEventListener('transcoding_progress', (event) => {
            try {
                const data: TranscodingEvent = JSON.parse(event.data);
                setLastEvent(data);
            } catch (err) {
                console.error('Failed to parse transcoding event:', err);
            }
        });

        return () => {
            eventSource.close();
            setStatus('disconnected');
        };
    }, [mediaId]);

    return {lastEvent, status};
}
