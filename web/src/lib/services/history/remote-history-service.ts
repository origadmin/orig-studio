// Remote history service - calls backend API for authenticated users

import type {HistoryItem} from '@/lib/api/history';
import type {HistoryListParams, HistoryListResult, IHistoryService} from './types';
import {historyApi} from '@/lib/api/history';

export class RemoteHistoryService implements IHistoryService {
    async list(params?: HistoryListParams): Promise<HistoryListResult> {
        const response = await historyApi.list({
            page: params?.page,
            page_size: params?.page_size,
            content_type: params?.content_type,
        });
        return {
            items: response.items,
            total: response.total,
        };
    }

    async upsert(item: Omit<HistoryItem, 'id' | 'last_watched_at' | 'create_time' | 'update_time' | 'user_id'>): Promise<HistoryItem> {
        const response = await historyApi.upsert({
            content_id: item.content_id,
            content_type: item.content_type,
            progress_seconds: item.progress_seconds,
            duration_seconds: item.duration_seconds,
        });
        return response.item;
    }

    async remove(id: string): Promise<void> {
        await historyApi.remove(id);
    }

    async clear(): Promise<{deleted_count: number}> {
        const response = await historyApi.clear();
        return {deleted_count: response.deleted_count};
    }

    async sync(items: HistoryItem[]): Promise<{items: HistoryItem[]; merged_count: number}> {
        const response = await historyApi.sync({
            items: items.map(item => ({
                content_id: item.content_id,
                content_type: item.content_type,
                progress_seconds: item.progress_seconds,
                duration_seconds: item.duration_seconds,
                is_finished: item.is_finished,
            })),
        });
        return {
            items: response.items,
            merged_count: response.merged_count,
        };
    }
}
