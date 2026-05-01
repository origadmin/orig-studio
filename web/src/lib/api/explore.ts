import {api} from '../request';

export interface TrendingItem {
    id: string;
    short_token: string;
    title: string;
    description?: string;
    thumbnail?: string;
    duration?: number;
    view_count: number;
    like_count: number;
    published_at?: string;
}

export interface TrendingResponse {
    items: TrendingItem[];
    total: number;
    page?: number;
    page_size?: number;
}

const exploreApi = {
    getTrending: async (params?: {limit?: number}): Promise<TrendingResponse> => {
        const searchParams = new URLSearchParams();
        if (params?.limit) searchParams.set('limit', String(params.limit));
        const query = searchParams.toString();
        return api.get<TrendingResponse>(`/explore/trending${query ? `?${query}` : ''}`);
    },
};

export {exploreApi};