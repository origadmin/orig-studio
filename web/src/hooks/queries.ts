import {useQuery, useMutation, useQueryClient, useInfiniteQuery} from '@tanstack/react-query';
import {mediaApi, publicMediaApi, adminMediaApi, type Media} from '@/lib/api/media';
import {categoryApi, type Category} from '@/lib/api/category';
import {channelApi, type ChannelDetail} from '@/lib/api/channel';
import {playlistApi, type Playlist, type PlaylistListResponse} from '@/lib/api/playlist';

/**
 * keys factory
 */
export const mediaKeys = {
    all: ['media'] as const,
    lists: () => [...mediaKeys.all, 'list'] as const,
    list: (params: Record<string, any>) => [...mediaKeys.lists(), params] as const,
    adminLists: () => [...mediaKeys.all, 'adminList'] as const,
    adminList: (params: Record<string, any>) => [...mediaKeys.adminLists(), params] as const,
    details: () => [...mediaKeys.all, 'detail'] as const,
    detail: (id: string) => [...mediaKeys.details(), id] as const,
};

/**
 * useMediaList: Fetch paginated media list for user
 */
export function useMediaList(params: {
    page?: number;
    page_size?: number;
    status?: string;
    type?: string;
    category_id?: number | null;
    user_id?: string | number;
    keyword?: string;
    search?: string;
}) {
    return useQuery({
        queryKey: mediaKeys.list(params),
        queryFn: async () => {
            const res = await mediaApi.list({
                ...params,
                keyword: params.search || params.keyword,
                category_id: params.category_id || undefined,
                user_id: params.user_id ? Number(params.user_id) : undefined
            });
            return res;
        },
    });
}

/**
 * useInfiniteMediaList: Fetch paginated media list with infinite scroll
 */
export function useInfiniteMediaList(params: {
    page_size?: number;
    status?: string;
    type?: string;
    category_id?: number | null;
    user_id?: string | number;
}) {
    return useInfiniteQuery({
        queryKey: mediaKeys.list(params),
        queryFn: async ({pageParam = 1}) => {
            const res = await mediaApi.list({
                ...params,
                page: pageParam,
                category_id: params.category_id || undefined,
                user_id: params.user_id ? Number(params.user_id) : undefined
            });
            return res;
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            const size = params.page_size || 20;
            const items = lastPage.items || [];
            return items.length === size ? allPages.length + 1 : undefined;
        },
    });
}

/**
 * useAdminMediaList: Fetch paginated media list for admin
 */
export function useAdminMediaList(params: {
    page?: number;
    page_size?: number;
}) {
    return useQuery({
        queryKey: mediaKeys.adminList(params),
        queryFn: async () => {
            const res = await mediaApi.adminList(params);
            return res;
        },
    });
}

/**
 * useMediaDetail: Fetch single media details (Legacy - uses ID or short_token)
 */
export function useMediaDetail(id: string | null) {
    // 彻底清理 id：移除任何引号、空格，并确保是纯数字字符串
    const cleanId = id ? String(id).replace(/["']/g, '').trim() : null;
    return useQuery({
        queryKey: mediaKeys.detail(cleanId!),
        queryFn: async () => {
            const res = await mediaApi.get(cleanId!);
            return res;
        },
        enabled: !!cleanId,
    });
}

/**
 * usePublicMediaDetail: Fetch public media details using short_token (Recommended)
 * MediaCMS style: /api/v1/medias/{short_token}
 * Returns public fields only, auto-increments view count
 */
export function usePublicMediaDetail(shortToken: string | null) {
    // 清理 short_token
    const cleanToken = shortToken ? String(shortToken).replace(/["']/g, '').trim() : null;
    return useQuery({
        queryKey: ['publicMedia', 'detail', cleanToken!],
        queryFn: async () => {
            const res = await publicMediaApi.get(cleanToken!);
            return res;
        },
        enabled: !!cleanToken && cleanToken.length > 0,
    });
}

/**
 * useAdminMediaDetail: Fetch full media details using ID (Admin only)
 * MediaCMS style: /api/v1/admin/medias/:id
 * Requires JWT + Admin role, returns all fields including private media
 */
export function useAdminMediaDetail(id: string | null) {
    const cleanId = id ? String(id).replace(/["']/g, '').trim() : null;
    return useQuery({
        queryKey: ['adminMedia', 'detail', cleanId!],
        queryFn: async () => {
            const res = await adminMediaApi.getById(cleanId!);
            return res;
        },
        enabled: !!cleanId,
    });
}

/**
 * useDeleteMedia: Handle media deletion
 */
export function useDeleteMedia() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => mediaApi.delete(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: mediaKeys.all});
        },
    });
}

/**
 * useUpdateMedia: Handle media update
 */
export function useUpdateMedia() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({id, data}: { id: string; data: Partial<Media> }) =>
            mediaApi.update(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: mediaKeys.all});
        },
    });
}

/**
 * useCategoryList: Fetch all categories
 */
export function useCategoryList() {
    return useQuery({
        queryKey: ['categories'],
        queryFn: async () => {
            const res = await categoryApi.getAll();
            return res;
        },
    });
}

export function useChannelByToken(token: string | null) {
    return useQuery({
        queryKey: ['channel', 'token', token],
        queryFn: async () => {
            const res = await channelApi.getByToken(token!);
            return res as ChannelDetail;
        },
        enabled: !!token,
    });
}

export function useChannelByHandle(handle: string | null) {
    return useQuery({
        queryKey: ['channel', 'handle', handle],
        queryFn: async () => {
            const res = await channelApi.get({username: handle!});
            return (res as any).data || res as ChannelDetail;
        },
        enabled: !!handle,
    });
}

export function useMyChannel(enabled: boolean) {
    return useQuery({
        queryKey: ['channel', 'me'],
        queryFn: async () => {
            const res = await channelApi.getMyChannel();
            return res as ChannelDetail;
        },
        enabled,
    });
}

export function useSubscriptionStatus(channelToken: string | null) {
    return useQuery({
        queryKey: ['subscription', channelToken],
        queryFn: async () => {
            const res = await channelApi.getSubscriptionStatus(channelToken!);
            return res as {is_subscribed: boolean};
        },
        enabled: !!channelToken,
    });
}

export function useSubscribe() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (channelToken: string) => channelApi.subscribe(channelToken),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['subscription']});
            queryClient.invalidateQueries({queryKey: ['channel']});
        },
    });
}

export function useUnsubscribe() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (channelToken: string) => channelApi.unsubscribe(channelToken),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['subscription']});
            queryClient.invalidateQueries({queryKey: ['channel']});
        },
    });
}

export function useUpdateNotificationSetting() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({channelToken, setting}: {channelToken: string; setting: string}) =>
            channelApi.updateNotificationSetting(channelToken, setting),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['channel']});
        },
    });
}

export function useChannelVideos(channelToken: string | null, params?: {sort?: string; keyword?: string; page_size?: number; page?: number}) {
    return useQuery({
        queryKey: ['channelVideos', channelToken, params],
        queryFn: async () => {
            const res = await channelApi.listAll({page: params?.page, limit: params?.page_size});
            return res;
        },
        enabled: !!channelToken,
    });
}

export function useChannelPlaylists(channelToken: string | null) {
    return useQuery({
        queryKey: ['channelPlaylists', channelToken],
        queryFn: async () => {
            const res = await playlistApi.getMyPlaylists();
            return res as PlaylistListResponse;
        },
        enabled: !!channelToken,
    });
}
