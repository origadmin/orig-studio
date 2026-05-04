// History service types - shared across local and remote implementations
// Re-exports ContentType and HistoryItem from the API layer for convenience

export type {ContentType} from '@/lib/api/history';
export type {HistoryItem} from '@/lib/api/history';

import type {ContentType, HistoryItem} from '@/lib/api/history';

/** Parameters for listing history */
export interface HistoryListParams {
    page?: number;
    page_size?: number;
    content_type?: ContentType;
}

/** Result of listing history */
export interface HistoryListResult {
    items: HistoryItem[];
    total: number;
}

/** Interface for history service implementations (local or remote) */
export interface IHistoryService {
    list(params?: HistoryListParams): Promise<HistoryListResult>;
    upsert(item: Omit<HistoryItem, 'id' | 'last_watched_at' | 'create_time' | 'update_time' | 'user_id'>): Promise<HistoryItem>;
    remove(id: string): Promise<void>;
    clear(): Promise<{deleted_count: number}>;
    sync(items: HistoryItem[]): Promise<{items: HistoryItem[]; merged_count: number}>;
}
