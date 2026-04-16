// Channel API
import {api} from "../request";
import {PaginatedResponse} from "./types";

export interface Channel {
    id: string;
    name: string;
    slug: string;
    description: string;
    user_id: string;
    media_count: number;
    subscriber_count: number;
    status: string;
    created_at: string;
}

export const channelApi = {
    getAll: () => api.get<PaginatedResponse<Channel>>('/channels'),
    get: (id: string) => api.get<Channel>(`/channels/${id}`),
    create: (data: Partial<Channel>) => api.post<Channel>('/channels', data),
    update: (id: string, data: Partial<Channel>) => api.put<Channel>(`/channels/${id}`, data),
    delete: (id: string) => api.del<void>(`/channels/${id}`),
};
