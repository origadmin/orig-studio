// History sync manager - handles login merge, progress reporting, and offline recovery

import type {HistoryItem, ContentType} from '@/lib/api/history';
import {isTokenExpired} from '@/lib/request';
import {LocalHistoryService} from './local-history-service';
import {RemoteHistoryService} from './remote-history-service';

export class HistorySyncManager {
    private localService = new LocalHistoryService();
    private remoteService = new RemoteHistoryService();

    /**
     * Called on login: merge local history with server.
     * 1. Read all local items
     * 2. Send to server via SyncHistory API
     * 3. Overwrite local with server response
     */
    async mergeOnLogin(): Promise<void> {
        const localResult = await this.localService.list({page: 1, page_size: 500});
        if (localResult.items.length === 0) return;

        try {
            const result = await this.remoteService.sync(localResult.items);
            // Overwrite local with server's merged result
            this.localService.replaceAll(result.items);
        } catch (error) {
            // Sync failure: keep local data, retry later
            console.warn('History sync on login failed:', error);
        }
    }

    /**
     * Called during playback: report progress to both local and remote.
     * Local is always updated first (low latency guarantee).
     * Remote is updated asynchronously (fire-and-forget with error marking).
     */
    async reportProgress(
        contentId: string,
        contentType: ContentType,
        progressSeconds: number,
        durationSeconds: number,
        meta?: {title?: string; thumbnail?: string; shortToken?: string},
    ): Promise<void> {
        const isFinished = durationSeconds > 0 && progressSeconds >= durationSeconds * 0.9;

        // Always update local first
        await this.localService.upsert({
            content_id: contentId,
            content_type: contentType,
            progress_seconds: progressSeconds,
            duration_seconds: durationSeconds,
            is_finished: isFinished,
            title: meta?.title || '',
            thumbnail: meta?.thumbnail || '',
            short_token: meta?.shortToken || '',
        });

        // Try remote if authenticated
        if (this.isAuthenticated()) {
            try {
                await this.remoteService.upsert({
                    content_id: contentId,
                    content_type: contentType,
                    progress_seconds: progressSeconds,
                    duration_seconds: durationSeconds,
                    is_finished: isFinished,
                    title: meta?.title || '',
                    thumbnail: meta?.thumbnail || '',
                    short_token: meta?.shortToken || '',
                });
            } catch {
                // Remote failure: mark for later sync
                this.localService.markPendingSync(contentId, contentType);
            }
        }
    }

    /**
     * Called on 'online' event: sync pending items to server.
     */
    async syncPendingItems(): Promise<void> {
        const pending = this.localService.getPendingItems();
        if (pending.length === 0) return;

        try {
            await this.remoteService.sync(pending);
            this.localService.clearPendingFlags();
        } catch {
            // Still offline or server error - keep pending items
        }
    }

    private isAuthenticated(): boolean {
        // Check if user is logged in by looking for a valid (non-expired) auth token.
        // Must match the same logic as AuthProvider to avoid sending requests
        // with expired tokens (which would 401 and fail silently).
        const token = localStorage.getItem('origcms_token') ||
                      sessionStorage.getItem('origcms_token');
        if (!token) return false;
        // isTokenExpired() parses the JWT exp claim and returns true if expired
        return !isTokenExpired();
    }
}

// Singleton instance
export const historySyncManager = new HistorySyncManager();
