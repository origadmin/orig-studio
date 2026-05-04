// Admin API v4.0 (管理端 - UUID only)
import {api} from "../request";
import {Channel, ChannelList} from "./channel";

// Re-export Channel type for admin pages
export type {Channel} from "./channel";
export type {ChannelList} from "./channel";

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
     * Get channel detail by UUID (Admin)
     *
     * @param id - Channel UUID (not short_token!)
     */
    getChannelById: (id: string) =>
        api.get<AdminChannelDetail>(`/admin/channels/${id}`),

    /**
     * Create a new channel (Admin)
     */
    createChannel: (data: Partial<Channel>) =>
        api.post<Channel>('/admin/channels', data),

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
