import {useQuery, useMutation, useQueryClient, useInfiniteQuery} from '@tanstack/react-query';
import {useMemo} from 'react';
import {mediaApi, publicMediaApi, adminMediaApi, type Media, type UpdateMediaRequest, normalizeMedia, normalizeMediaList} from '@/lib/api/media';
import {categoryApi, type Category} from '@/lib/api/category';
import {channelApi, type Channel, type ChannelDetail, type ChannelLimits} from '@/lib/api/channel';
import {userApi, type PublicProfile} from '@/lib/api/user';
import {playlistApi, type Playlist, type PlaylistListResponse} from '@/lib/api/playlist';
import {portalApi, adminPortalApi} from '@/lib/api/portal';
import {reviewApi} from '@/lib/api/review';
import {adminCommentApi} from '@/lib/api/comment';
import {configApi, type SettingCategory} from '@/lib/api/config';
import {adminPermissionApi} from '@/lib/api/permission';
import {spriteApi} from '@/lib/api/sprite';
import {favoriteApi} from '@/lib/api/favorite';
import {PAGINATION_CONFIG} from '@/config/pagination';
import {useAuth} from '@/hooks/useAuth';

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
 *
 * Parameter mapping (hook params → API params):
 *   status    → state       (backend uses "state", not "status")
 *   search    → keyword     (consolidated into "keyword")
 *   featured  → featured    (boolean → string conversion)
 *   sort      → order_by    (backend uses "order_by", not "sort")
 *   order     → descending  ("desc" → true, "asc" → false)
 */
export function useMediaList(params: {
    page?: number;
    page_size?: number;
    status?: string;
    type?: string;
    category_id?: number | null;
    category_ids?: number[];
    user_id?: string | number;
    keyword?: string;
    search?: string;
    featured?: boolean | string;
    order_by?: string;
    descending?: boolean;
    /** @deprecated Use order_by instead */
    sort?: string;
    /** @deprecated Use descending instead */
    order?: string;
}) {
    return useQuery({
        queryKey: mediaKeys.list(params),
        queryFn: async () => {
            // Explicitly construct API params to avoid leaking unrecognized fields
            // (e.g. "status" → backend expects "state"; "search" → mapped to "keyword")
            const apiParams: Record<string, unknown> = {
                page: params.page,
                page_size: params.page_size,
                type: params.type,
                category_id: params.category_id != null && params.category_id > 0 ? params.category_id : undefined,
                category_ids: params.category_ids && params.category_ids.length > 0 ? params.category_ids.join(',') : undefined,
                user_id: params.user_id || undefined,
                keyword: params.search || params.keyword,
                // Map status → state (backend field name)
                state: params.status,
                featured: params.featured != null ? String(params.featured) : undefined,
                // Map sort/order → order_by/descending (backend field names)
                order_by: params.order_by || params.sort,
                descending: params.descending != null
                    ? params.descending
                    : params.order === 'desc' ? true : params.order === 'asc' ? false : undefined,
            };
            // Remove undefined values to keep URL clean
            Object.keys(apiParams).forEach(key => {
                if (apiParams[key] === undefined || apiParams[key] === null) {
                    delete apiParams[key];
                }
            });
            const res = await mediaApi.list(apiParams as Parameters<typeof mediaApi.list>[0]);
            // Normalize edge fields for each media item
            if (res?.items) {
                res.items = normalizeMediaList(res.items);
            }
            return res;
        },
    });
}

/**
 * useInfiniteMediaList: Fetch paginated media list with infinite scroll
 *
 * Parameter mapping (hook params → API params):
 *   status → state (backend uses "state", not "status")
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
            // Explicitly construct API params to avoid leaking unrecognized fields
            const apiParams: Record<string, unknown> = {
                page: pageParam,
                page_size: params.page_size,
                type: params.type,
                category_id: params.category_id != null && params.category_id > 0 ? params.category_id : undefined,
                user_id: params.user_id ? Number(params.user_id) : undefined,
                // Map status → state (backend field name)
                state: params.status,
            };
            // Remove undefined values to keep URL clean
            Object.keys(apiParams).forEach(key => {
                if (apiParams[key] === undefined || apiParams[key] === null) {
                    delete apiParams[key];
                }
            });
            const res = await mediaApi.list(apiParams as Parameters<typeof mediaApi.list>[0]);
            // Normalize edge fields for each media item
            if (res?.items) {
                res.items = normalizeMediaList(res.items);
            }
            return res;
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            const size = params.page_size || PAGINATION_CONFIG.DEFAULT_PAGE_SIZE;
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
    keyword?: string;
    state?: string;
    type?: string;
}) {
    return useQuery({
        queryKey: mediaKeys.adminList(params),
        queryFn: async () => {
            const res = await mediaApi.adminList(params);
            // Normalize edge fields for each media item
            if (res?.items) {
                res.items = normalizeMediaList(res.items);
            }
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
            return normalizeMedia(res);
        },
        enabled: !!cleanId,
    });
}

/**
 * usePublicMediaDetail: Fetch public media details using short_token (Recommended)
 * MediaCMS style: /api/v1/medias/{short_token}
 * Returns public fields only
 */
export function usePublicMediaDetail(shortToken: string | null) {
    // 清理 short_token
    const cleanToken = shortToken ? String(shortToken).replace(/["']/g, '').trim() : null;
    return useQuery({
        queryKey: ['publicMedia', 'detail', cleanToken!],
        queryFn: async () => {
            const res = await publicMediaApi.get(cleanToken!);
            return normalizeMedia(res);
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
            return normalizeMedia(res);
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
        mutationFn: (id: string) => adminMediaApi.delete(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: mediaKeys.all});
        },
    });
}

/**
 * useUpdateMedia: Handle media update (admin)
 */
export function useUpdateMedia() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({id, data}: { id: string; data: Partial<Media> }) =>
            adminMediaApi.update(id, data as any),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: mediaKeys.all});
            queryClient.invalidateQueries({queryKey: ['adminMedia']});
        },
    });
}

/**
 * useUpdatePublicMedia: Handle media update via short_token (owner/admin)
 */
export function useUpdatePublicMedia() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({shortToken, data}: { shortToken: string; data: UpdateMediaRequest }) =>
            publicMediaApi.update(shortToken, data),
        onSuccess: (_data, variables) => {
            queryClient.invalidateQueries({queryKey: ['publicMedia', 'detail', variables.shortToken]});
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
            const res = await channelApi.resolveHandle(handle!);
            if (res.type === 'channel' && res.channel) {
                return res.channel as ChannelDetail;
            }
            return null;
        },
        enabled: !!handle,
    });
}

export function usePublicProfile(username: string | null) {
    const {user: currentUser, isAuthenticated} = useAuth();
    const query = useQuery({
        queryKey: ['profile', username],
        queryFn: async () => {
            const res = await userApi.getPublicProfile(username!);
            const raw = (res as any)?.user ?? res;
            return {
                id: raw.id,
                username: raw.username,
                nickname: raw.nickname || undefined,
                avatar: raw.avatar || undefined,
                slug: raw.slug || undefined,
                bio: raw.description || raw.bio || undefined,
                location: raw.location || undefined,
                website: raw.website || undefined,
                title: raw.title || undefined,
                is_featured: raw.is_verified || false,
                media_count: raw.media_count || 0,
                subscriber_count: raw.subscriber_count || 0,
                created_at: raw.create_time || raw.created_at,
                default_channel_token: raw.default_channel_token || undefined,
                // is_owner computed outside queryFn via useMemo below,
                // so it stays in sync with auth state changes.
                is_subscribed: raw.is_subscribed || false,
            } as Omit<PublicProfile, 'is_owner'>;
        },
        enabled: !!username,
        staleTime: 60_000,
    });

    // Derive is_owner from current auth state and profile data.
    // Using useMemo ensures is_owner is recomputed whenever auth state changes,
    // without requiring a re-fetch of the profile data.
    const data = useMemo(() => {
        if (!query.data) return undefined;
        return {
            ...query.data,
            is_owner: isAuthenticated && !!currentUser && currentUser.username === query.data.username,
        } as PublicProfile;
    }, [query.data, isAuthenticated, currentUser]);

    return {...query, data};
}

export function useMyChannel(enabled: boolean) {
    return useQuery({
        queryKey: ['channel', 'me'],
        queryFn: async () => {
            const res = await channelApi.getMyChannels();
            // Backend returns { items: [], total: 0 } when user has no channels
            const data = (res as any)?.items ?? [];
            return (data.length > 0 ? data[0] : null) as ChannelDetail | null;
        },
        enabled,
    });
}

export function useMyChannels(enabled: boolean) {
    return useQuery({
        queryKey: ['channels', 'me'],
        queryFn: async () => {
            const res = await channelApi.getMyChannels();
            const data = (res as any)?.items ?? [];
            return data as Channel[];
        },
        enabled,
    });
}

export function useChannelLimits(enabled: boolean) {
    return useQuery({
        queryKey: ['channel', 'limits'],
        queryFn: async () => {
            const res = await channelApi.getChannelLimits();
            return res as ChannelLimits;
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
            // Build params conditionally to avoid passing page: undefined
            const listParams: { page?: number; limit?: number } = {};
            if (params?.page !== undefined) listParams.page = params.page;
            if (params?.page_size !== undefined) listParams.limit = params.page_size;
            const res = await channelApi.listAll(listParams);
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

// ==================== Portal Hooks ====================

/**
 * usePortalConfig: Fetch portal configuration (navigation, banners, featured users, site info)
 */
export function usePortalConfig() {
    return useQuery({
        queryKey: ['portal', 'config'],
        queryFn: async () => {
            const res = await portalApi.getConfig();
            return res;
        },
    });
}

/**
 * useAdminNavItems: Fetch admin navigation items list
 */
export function useAdminNavItems() {
    return useQuery({
        queryKey: ['admin', 'navItems'],
        queryFn: async () => {
            const res = await adminPortalApi.listNavItems();
            return res;
        },
    });
}

/**
 * useAdminBanners: Fetch admin banners list
 */
export function useAdminBanners() {
    return useQuery({
        queryKey: ['admin', 'banners'],
        queryFn: async () => {
            const res = await adminPortalApi.listBanners();
            return res;
        },
    });
}

/**
 * useCreateNavItem: Create a new navigation item
 */
export function useCreateNavItem() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: Parameters<typeof adminPortalApi.createNavItem>[0]) =>
            adminPortalApi.createNavItem(data),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'navItems']});
            queryClient.invalidateQueries({queryKey: ['portal', 'config']});
        },
    });
}

/**
 * useUpdateNavItem: Update an existing navigation item
 */
export function useUpdateNavItem() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({id, data}: {id: string; data: Parameters<typeof adminPortalApi.updateNavItem>[1]}) =>
            adminPortalApi.updateNavItem(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'navItems']});
            queryClient.invalidateQueries({queryKey: ['portal', 'config']});
        },
    });
}

/**
 * useDeleteNavItem: Delete a navigation item
 */
export function useDeleteNavItem() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => adminPortalApi.deleteNavItem(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'navItems']});
            queryClient.invalidateQueries({queryKey: ['portal', 'config']});
        },
    });
}

/**
 * useCreateBanner: Create a new banner
 */
export function useCreateBanner() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: Parameters<typeof adminPortalApi.createBanner>[0]) =>
            adminPortalApi.createBanner(data),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'banners']});
            queryClient.invalidateQueries({queryKey: ['portal', 'config']});
        },
    });
}

/**
 * useUpdateBanner: Update an existing banner
 */
export function useUpdateBanner() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({id, data}: {id: string; data: Parameters<typeof adminPortalApi.updateBanner>[1]}) =>
            adminPortalApi.updateBanner(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'banners']});
            queryClient.invalidateQueries({queryKey: ['portal', 'config']});
        },
    });
}

/**
 * useToggleBanner: Toggle banner active status
 */
export function useToggleBanner() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => adminPortalApi.toggleBanner(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'banners']});
            queryClient.invalidateQueries({queryKey: ['portal', 'config']});
        },
    });
}

// ==================== Review Hooks ====================

/**
 * useReviewList: Fetch review list (pending or history)
 */
export function useReviewList(params?: { page?: number; page_size?: number; type?: string; status?: string }) {
    const isHistory = !!params?.status;
    return useQuery({
        queryKey: ['reviews', params],
        queryFn: async () => {
            if (isHistory) {
                const res = await reviewApi.getHistory(params);
                return res;
            }
            const res = await reviewApi.getPending(params);
            return res;
        },
    });
}

/**
 * useApproveReview: Approve a review item
 */
export function useApproveReview() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            reviewApi.review(id, {action: 'approve'}),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['reviews']});
        },
    });
}

/**
 * useRejectReview: Reject a review item
 */
export function useRejectReview() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({id, reason}: {id: string; reason?: string}) =>
            reviewApi.review(id, {action: 'reject', comment: reason}),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['reviews']});
        },
    });
}

// ==================== Admin Comment Hooks ====================

/**
 * useAdminCommentList: Fetch admin comment list with optional filters
 */
export function useAdminCommentList(params?: { page?: number; page_size?: number; media_id?: string; status?: string }) {
    return useQuery({
        queryKey: ['admin', 'comments', params],
        queryFn: async () => {
            const res = await adminCommentApi.list(params);
            return res;
        },
    });
}

/**
 * useDeleteComment: Delete a comment (admin)
 */
export function useDeleteComment() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (id: string) => adminCommentApi.delete(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['admin', 'comments']});
        },
    });
}

// ==================== Config Hooks ====================

/**
 * useSettingCategories: Fetch all setting categories
 */
export function useSettingCategories() {
    return useQuery({
        queryKey: ['settings', 'categories'],
        queryFn: async () => {
            const res = await configApi.getAll();
            return res;
        },
    });
}

/**
 * useUpdateSetting: Update a single setting by key
 */
export function useUpdateSetting() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({key, value}: {key: string; value: string}) =>
            configApi.updateOne(key, {value}),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['settings']});
        },
    });
}

// ==================== Permission Hooks ====================

/**
 * usePermissionList: Fetch permission groups list
 */
export function usePermissionList(params?: { page?: number; page_size?: number; is_active?: boolean }) {
    return useQuery({
        queryKey: ['permissions', params],
        queryFn: async () => {
            const res = await adminPermissionApi.list(params);
            return res;
        },
    });
}

export function usePermissionGroups(params?: { page?: number; page_size?: number; is_active?: boolean }) {
    return useQuery({
        queryKey: ['permissionGroups', params],
        queryFn: async () => {
            const res = await adminPermissionApi.list(params);
            return res;
        },
    });
}

export function usePermissionGroup(id: string | null) {
    return useQuery({
        queryKey: ['permissionGroup', id],
        queryFn: async () => {
            const res = await adminPermissionApi.get(id!);
            return res;
        },
        enabled: !!id,
    });
}

export function useGroupMembers(groupId: string | null, params?: { page?: number; page_size?: number }) {
    return useQuery({
        queryKey: ['groupMembers', groupId, params],
        queryFn: async () => {
            const res = await adminPermissionApi.getMembers(groupId!, params);
            return res;
        },
        enabled: !!groupId,
    });
}

export function useUserPermissions(userId: string | null) {
    return useQuery({
        queryKey: ['userPermissions', userId],
        queryFn: async () => {
            const res = await adminPermissionApi.getUserPermissions(userId!);
            return res;
        },
        enabled: !!userId,
    });
}

/**
 * useUpdatePermission: Update a permission group
 */
export function useUpdatePermission() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({id, data}: {id: string; data: Parameters<typeof adminPermissionApi.update>[1]}) =>
            adminPermissionApi.update(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['permissions']});
        },
    });
}

// ==================== Sprite Hooks ====================

/**
 * useSpriteList: Get sprite URLs for a media item
 */
export function useSpriteList(mediaId: string | null) {
    return useQuery({
        queryKey: ['sprite', mediaId],
        queryFn: async () => {
            const vttUrl = spriteApi.getVttUrl(mediaId!);
            const spriteUrl = spriteApi.getSpriteUrl(mediaId!);
            return {vttUrl, spriteUrl};
        },
        enabled: !!mediaId,
    });
}

/**
 * useGenerateSprite: Regenerate sprite for a media item
 */
export function useGenerateSprite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (mediaId: string) => spriteApi.regenerateSprite(mediaId),
        onSuccess: (_data, mediaId) => {
            queryClient.invalidateQueries({queryKey: ['sprite', mediaId]});
        },
    });
}

// ==================== Favorite Hooks ====================

/**
 * useFavoriteStatus: Get favorite status for a media item with TanStack Query caching
 */
export function useFavoriteStatus(mediaId: string | null | undefined, shortToken?: string) {
    const identifier = shortToken || mediaId;
    const usePublicApi = !!shortToken && shortToken.trim().length > 0;
    return useQuery({
        queryKey: ['favoriteStatus', identifier],
        queryFn: async () => {
            if (!identifier) return null;
            if (usePublicApi) {
                return await publicMediaApi.favorites.getStatus(identifier);
            }
            return await mediaApi.favorites.getStatus(identifier!);
        },
        enabled: !!identifier,
    });
}

/**
 * useToggleFavorite: Toggle favorite status with optimistic update
 */
export function useToggleFavorite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: async ({mediaId, shortToken}: { mediaId: string; shortToken?: string }) => {
            const identifier = shortToken || mediaId;
            const usePublicApi = !!shortToken && shortToken.trim().length > 0;
            if (usePublicApi) {
                return await publicMediaApi.favorites.toggle(identifier);
            }
            return await mediaApi.favorites.toggle(identifier);
        },
        onSuccess: (_data, variables) => {
            const identifier = variables.shortToken || variables.mediaId;
            // Invalidate favorite status cache
            queryClient.invalidateQueries({queryKey: ['favoriteStatus', identifier]});
            // Invalidate favorites list cache
            queryClient.invalidateQueries({queryKey: ['favorites']});
            // Invalidate media detail cache (favorite_count may have changed)
            queryClient.invalidateQueries({queryKey: ['publicMedia', 'detail', identifier]});
            queryClient.invalidateQueries({queryKey: mediaKeys.detail(identifier)});
        },
    });
}

/**
 * useFavoriteList: Get user's favorite list with pagination
 */
export function useFavoriteList(params?: { page?: number; page_size?: number }, userId?: string) {
    return useQuery({
        queryKey: ['favorites', userId, params],
        queryFn: async () => {
            return await favoriteApi.list(params);
        },
        enabled: !!userId,
    });
}

/**
 * useRemoveFavorite: Remove a favorite by its ID
 */
export function useRemoveFavorite() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (favoriteId: string) => favoriteApi.remove(favoriteId),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['favorites']});
        },
    });
}

// ==================== History Hooks ====================

import {historyApi, type ContentType} from '@/lib/api/history';
import {createHistoryService} from '@/lib/services/history';

/**
 * useHistoryList: Get watch history list with pagination
 * Supports both authenticated (remote) and anonymous (local) users
 */
export function useHistoryList(params: {
    page?: number;
    page_size?: number;
    content_type?: ContentType;
    isAuthenticated?: boolean;
    userId?: number | string;
}) {
    const service = createHistoryService(!!params.isAuthenticated);
    return useQuery({
        queryKey: ['history', params.userId, params.page, params.content_type],
        queryFn: async () => {
            // Build params conditionally to avoid passing page: undefined
            const listParams: { page?: number; page_size?: number; content_type?: ContentType } = {};
            if (params.page !== undefined) listParams.page = params.page;
            if (params.page_size !== undefined) listParams.page_size = params.page_size;
            if (params.content_type !== undefined) listParams.content_type = params.content_type;
            return await service.list(listParams);
        },
        staleTime: 0,
        refetchOnMount: 'always',
    });
}

/**
 * useUpsertHistory: Report watch progress (upsert a history record)
 */
export function useUpsertHistory() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (data: {
            content_id: string;
            content_type: ContentType;
            progress_seconds: number;
            duration_seconds: number;
        }) => historyApi.upsert(data),
        onSuccess: () => {
            // Don't invalidate on every upsert (too frequent during playback)
            // Progress is reported via useWatchProgress which handles its own cadence
        },
    });
}

/**
 * useClearHistory: Clear all watch history
 */
export function useClearHistory(isAuthenticated?: boolean) {
    const queryClient = useQueryClient();
    const service = createHistoryService(!!isAuthenticated);
    return useMutation({
        mutationFn: () => service.clear(),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['history']});
        },
    });
}

/**
 * useRemoveHistoryItem: Remove a single history record by ID
 */
export function useRemoveHistoryItem(isAuthenticated?: boolean) {
    const queryClient = useQueryClient();
    const service = createHistoryService(!!isAuthenticated);
    return useMutation({
        mutationFn: (id: string) => service.remove(id),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['history']});
        },
    });
}
