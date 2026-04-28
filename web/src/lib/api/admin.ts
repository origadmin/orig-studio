// Admin API v3.2 (管理端 - UUID only)
import {api} from "../request";
import {Channel, ChannelList} from "./channel";

export interface AdminChannelDetail extends Omit<Channel, 'media_count'> {
    media_count?: number;
}

export interface AdminChannelFilters {
    page?: number;
    page_size?: number;
    status?: string;
    user_id?: string;
    search?: string;
}

export const adminApi = {
    // ================================
    // Channel Management (Admin - UUID only)
    // ================================

    /**
     * List all channels (Admin)
     * Returns paginated list of all channels including non-public
     */
    getChannels: (filters?: AdminChannelFilters) =>
        api.get<ChannelList>('/admin/channels', {params: filters}),

    /**
     * Get channel detail by UUID (Admin) ⭐
     *
     * @param id - Channel UUID (not short_token!)
     * @example getChannelById('550e8400-e29b-41d4-a716-446655440000')
     */
    getChannelById: (id: string) =>
        api.get<AdminChannelDetail>(`/admin/channels/${id}`),

    /**
     * Update any channel by UUID (Admin)
     */
    updateChannel: (id: string, data: Partial<Channel>) =>
        api.put<Channel>(`/admin/channels/${id}`, data),

    /**
     * Delete any channel by UUID (Admin)
     */
    deleteChannel: (id: string) =>
        api.del<void>(`/admin/channels/${id}`),
};
