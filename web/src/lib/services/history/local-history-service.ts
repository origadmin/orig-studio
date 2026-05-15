// Local history service - uses localStorage for anonymous/offline users
// Max 500 items, ~200KB storage budget

import type {HistoryItem, ContentType} from '@/lib/api/history';
import type {HistoryListParams, HistoryListResult, IHistoryService} from './types';

const STORAGE_KEY = 'origstudio_watch_history';
const MAX_ITEMS = 500;
const CAPACITY_BYTES = 4 * 1024 * 1024; // 4MB safety limit

export class LocalHistoryService implements IHistoryService {
    private getAll(): HistoryItem[] {
        try {
            const raw = localStorage.getItem(STORAGE_KEY);
            if (!raw) return [];
            const items = JSON.parse(raw) as HistoryItem[];
            // Deduplicate: same short_token + content_type = same video
            const seen = new Map<string, number>();
            for (let i = items.length - 1; i >= 0; i--) {
                const key = (items[i].short_token || items[i].content_id) + ':' + items[i].content_type;
                if (seen.has(key)) {
                    const prevIdx = seen.get(key)!;
                    // Keep the one with more progress
                    if (items[i].progress_seconds > items[prevIdx].progress_seconds) {
                        items.splice(prevIdx, 1);
                        seen.set(key, i - 1);
                    } else {
                        items.splice(i, 1);
                    }
                } else {
                    seen.set(key, i);
                }
            }
            return items;
        } catch {
            return [];
        }
    }

    private saveAll(items: HistoryItem[]): void {
        try {
            const json = JSON.stringify(items);
            // Capacity check: remove oldest items until under capacity
            if (json.length > CAPACITY_BYTES) {
                items.sort((a, b) =>
                    new Date(b.last_watched_at).getTime() - new Date(a.last_watched_at).getTime()
                );
                while (JSON.stringify(items).length > CAPACITY_BYTES && items.length > 0) {
                    items.pop();
                }
            }
            localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
        } catch {
            // localStorage full or unavailable - silently fail
        }
    }

    async list(params?: HistoryListParams): Promise<HistoryListResult> {
        let items = this.getAll();

        // Sort by last_watched_at DESC
        items.sort((a, b) =>
            new Date(b.last_watched_at).getTime() - new Date(a.last_watched_at).getTime()
        );

        // Filter by content_type
        if (params?.content_type) {
            items = items.filter(i => i.content_type === params.content_type);
        }

        // Paginate
        const page = params?.page ?? 1;
        const pageSize = params?.page_size ?? 20;
        const start = (page - 1) * pageSize;
        const paged = items.slice(start, start + pageSize);

        return {items: paged, total: items.length};
    }

    async upsert(item: Omit<HistoryItem, 'id' | 'last_watched_at' | 'create_time' | 'update_time' | 'user_id'>): Promise<HistoryItem> {
        const items = this.getAll();
        let existingIdx = items.findIndex(
            i => i.content_id === item.content_id && i.content_type === item.content_type
        );

        // If not found by content_id, try matching by short_token (migration compat)
        if (existingIdx < 0 && item.short_token) {
            existingIdx = items.findIndex(
                i => i.short_token === item.short_token && i.content_type === item.content_type
            );
            if (existingIdx >= 0) {
                // Migrate: update content_id from short_token to UUID
                items[existingIdx].content_id = item.content_id;
            }
        }

        if (existingIdx >= 0) {
            const existing = items[existingIdx];
            existing.progress_seconds = Math.max(existing.progress_seconds, item.progress_seconds);
            existing.duration_seconds = item.duration_seconds || existing.duration_seconds;
            existing.is_finished = item.is_finished || existing.is_finished;
            existing.last_watched_at = new Date().toISOString();
            if (item.title) existing.title = item.title;
            if (item.thumbnail) existing.thumbnail = item.thumbnail;
            if (item.short_token) existing.short_token = item.short_token;
            this.saveAll(items);
            return existing;
        }

        // Insert
        const newItem: HistoryItem = {
            ...item,
            id: crypto.randomUUID(),
            last_watched_at: new Date().toISOString(),
            create_time: new Date().toISOString(),
            user_id: '',
        };
        items.unshift(newItem);

        // Enforce max items
        if (items.length > MAX_ITEMS) {
            items.sort((a, b) =>
                new Date(b.last_watched_at).getTime() - new Date(a.last_watched_at).getTime()
            );
            items.length = MAX_ITEMS;
        }

        this.saveAll(items);
        return newItem;
    }

    async remove(id: string): Promise<void> {
        const items = this.getAll();
        const filtered = items.filter(i => i.id !== id);
        this.saveAll(filtered);
    }

    async clear(): Promise<{deleted_count: number}> {
        const items = this.getAll();
        const count = items.length;
        localStorage.removeItem(STORAGE_KEY);
        return {deleted_count: count};
    }

    async sync(items: HistoryItem[]): Promise<{items: HistoryItem[]; merged_count: number}> {
        const existing = this.getAll();
        let mergedCount = 0;

        for (const item of items) {
            const existingIdx = existing.findIndex(
                e => e.content_id === item.content_id && e.content_type === item.content_type
            );

            if (existingIdx >= 0) {
                // Merge: max progress, latest time
                const e = existing[existingIdx];
                e.progress_seconds = Math.max(e.progress_seconds, item.progress_seconds);
                e.duration_seconds = item.duration_seconds || e.duration_seconds;
                e.is_finished = e.is_finished || item.is_finished;
                if (new Date(item.last_watched_at).getTime() > new Date(e.last_watched_at).getTime()) {
                    e.last_watched_at = item.last_watched_at;
                }
                mergedCount++;
            } else {
                existing.push({...item});
                mergedCount++;
            }
        }

        this.saveAll(existing);
        return {items: existing, merged_count: mergedCount};
    }

    // Extra methods for sync manager

    /** Replace all local items with the given list */
    replaceAll(items: HistoryItem[]): void {
        this.saveAll(items);
    }

    /** Get items that are pending sync to the server */
    getPendingItems(): HistoryItem[] {
        try {
            const raw = localStorage.getItem(STORAGE_KEY + '_pending');
            if (!raw) return [];
            return JSON.parse(raw) as HistoryItem[];
        } catch {
            return [];
        }
    }

    /** Mark a content item as pending sync */
    markPendingSync(contentId: string, contentType: ContentType): void {
        const items = this.getAll();
        const item = items.find(i => i.content_id === contentId && i.content_type === contentType);
        if (item) {
            const pending = this.getPendingItems();
            if (!pending.find(p => p.content_id === contentId && p.content_type === contentType)) {
                pending.push(item);
                localStorage.setItem(STORAGE_KEY + '_pending', JSON.stringify(pending));
            }
        }
    }

    /** Clear all pending sync flags */
    clearPendingFlags(): void {
        localStorage.removeItem(STORAGE_KEY + '_pending');
    }
}
