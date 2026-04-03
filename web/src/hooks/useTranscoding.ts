/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useEffect, useState, useCallback, useRef} from "react";
import {mediaApi} from "../lib/api/media";

interface TranscodingEvent {
    media_id: number;
    task_id: number;
    status: "pending" | "processing" | "success" | "failed";
    progress: number;
}

interface TranscodingSSEStatus {
    connected: boolean;
    reconnecting: boolean;
}

export function useTranscoding(mediaId?: number): {
    lastEvent: TranscodingEvent | null;
    sseStatus: TranscodingSSEStatus;
    connect: () => void;
    disconnect: () => void;
} {
    const [lastEvent, setLastEvent] = useState<TranscodingEvent | null>(null);
    const [sseStatus, setSseStatus] = useState<TranscodingSSEStatus>({
        connected: false,
        reconnecting: false,
    });
    const eventSourceRef = useRef<EventSource | null>(null);
    const mountedRef = useRef(true);

    const disconnect = useCallback(() => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
            eventSourceRef.current = null;
        }
        if (mountedRef.current) {
            setSseStatus({connected: false, reconnecting: false});
        }
    }, []);

    const connect = useCallback(() => {
        if (!mountedRef.current) return;

        // Close any existing connection first.
        disconnect();

        const sseUrl = mediaApi.getSSEUrl(mediaId);
        const eventSource = new EventSource(sseUrl);
        eventSourceRef.current = eventSource;

        eventSource.onopen = () => {
            if (mountedRef.current) {
                setSseStatus({connected: true, reconnecting: false});
            }
        };

        eventSource.onerror = () => {
            if (mountedRef.current) {
                setSseStatus({connected: false, reconnecting: true});
            }
            // EventSource will auto-reconnect, but we close and retry
            // to get a clean state.
            eventSource.close();
            eventSourceRef.current = null;
        };

        eventSource.addEventListener("transcoding_progress", (event) => {
            try {
                const data: TranscodingEvent = JSON.parse(event.data);
                if (mountedRef.current) {
                    setLastEvent(data);
                }
            } catch (err) {
                console.error("Failed to parse transcoding event:", err);
            }
        });
    }, [mediaId, disconnect]);

    // Auto-reconnect with backoff when disconnected.
    useEffect(() => {
        if (!sseStatus.reconnecting) return;

        let delay = 1000;
        const maxDelay = 10000;
        let timerId: ReturnType<typeof setTimeout>;

        const attemptReconnect = () => {
            if (!mountedRef.current || !sseStatus.reconnecting) return;
            connect();
            delay = Math.min(delay * 2, maxDelay);
            timerId = setTimeout(attemptReconnect, delay);
        };

        timerId = setTimeout(attemptReconnect, delay);
        return () => clearTimeout(timerId);
    }, [sseStatus.reconnecting, connect]);

    // Connect on mount, disconnect on unmount.
    useEffect(() => {
        mountedRef.current = true;
        connect();
        return () => {
            mountedRef.current = false;
            disconnect();
        };
    }, [connect, disconnect]);

    return {lastEvent, sseStatus, connect, disconnect};
}
