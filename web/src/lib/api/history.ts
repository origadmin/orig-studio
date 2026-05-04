// API client - Watch History module
// Updated to match new Proto schema:
//   - media_id -> content_id (supports multi-content-type)
//   - progress -> progress_seconds (clarified unit)
//   - duration -> duration_seconds (clarified unit)
//   - added content_type field
//   - added is_finished field
import {api} from "../request";

/** Content type for history tracking (matches backend lowercase values) */
export type ContentType = 'video' | 'article' | 'audio';

/** History item returned by the server (matches Proto HistoryItem) */
export interface HistoryItem {
    id: string;
    user_id: string;
    content_id: string;
    content_type: ContentType;
    progress_seconds: number;
    duration_seconds: number;
    is_finished: boolean;
    last_watched_at: string;
    create_time: string;
    update_time?: string;

    // Display fields (populated from content details, not stored in DB)
    title?: string;
    thumbnail?: string;
    short_token?: string;
    deleted?: boolean;
}

/** Request body for upserting a history record */
export interface UpsertHistoryRequest {
    content_id: string;
    content_type: ContentType;
    progress_seconds: number;
    duration_seconds: number;
    title?: string;
    thumbnail?: string;
    short_token?: string;
}

/** Response for upserting a history record */
export interface UpsertHistoryResponse {
    item: HistoryItem;
}

/** Request body for syncing history records */
export interface SyncHistoryRequest {
    items: Array<{
        content_id: string;
        content_type: ContentType;
        progress_seconds: number;
        duration_seconds: number;
        is_finished: boolean;
    }>;
}

/** Response for syncing history records */
export interface SyncHistoryResponse {
    items: HistoryItem[];
    merged_count: number;
}

/** Response for listing history */
export interface HistoryListResponse {
    items: HistoryItem[];
    total: number;
    page: number;
    page_size: number;
}

/** Response for clearing history */
export interface ClearHistoryResponse {
    deleted_count: number;
}

export const historyApi = {
    // Get watch history list with optional filters
    list: (params?: { page?: number; page_size?: number; content_type?: ContentType }) =>
        api.get<HistoryListResponse>("/me/history", params),

    // Upsert (create or update) a history record
    upsert: (data: UpsertHistoryRequest) =>
        api.post<UpsertHistoryResponse>("/me/history", data),

    // Batch sync history records (login merge)
    sync: (data: SyncHistoryRequest) =>
        api.post<SyncHistoryResponse>("/me/history/sync", data),

    // Clear all watch history
    clear: () =>
        api.del<ClearHistoryResponse>("/me/history"),

    // Remove a single history record by ID
    remove: (id: string) =>
        api.del<void>(`/me/history/${id}`),
};
