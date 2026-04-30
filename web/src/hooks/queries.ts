import {useQuery, useMutation, useQueryClient, useInfiniteQuery} from '@tanstack/react-query';
import {mediaApi, publicMediaApi, adminMediaApi, type Media, type UpdateMediaRequest, normalizeMedia, normalizeMediaList} from '@/lib/api/media';
import {categoryApi, type Category} from '@/lib/api/category';
import {channelApi, type ChannelDetail} from '@/lib/api/channel';
import {playlistApi, type Playlist, type PlaylistListResponse} from '@/lib/api/playlist';
import {portalApi, adminPortalApi} from '@/lib/api/portal';
import {reviewApi} from '@/lib/api/review';
import {adminCommentApi} from '@/lib/api/comment';
import {configApi, type SettingCategory} from '@/lib/api/config';
import {adminPermissionApi} from '@/lib/api/permission';
import {spriteApi} from '@/lib/api/sprite';
import {PAGINATION} from '@/config/pagination';

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
    featured?: boolean | string;
}) {
    return useQuery({
        queryKey: mediaKeys.list(params),
        queryFn: async () => {
            const res = await mediaApi.list({
                ...params,
                keyword: params.search || params.keyword,
                category_id: params.category_id || undefined,
                user_id: params.user_id ? Number(params.user_id) : undefined,
                featured: params.featured != null ? String(params.featured) : undefined,
            });
            // Normalize flat edge fields (user, category, channel) to edges structure
            if (res?.items) {
                res.items = normalizeMediaList(res.items);
            }
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
            // Normalize flat edge fields (user, category, channel) to edges structure
            if (res?.items) {
                res.items = normalizeMediaList(res.items);
            }
            return res;
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage, allPages) => {
            const size = params.page_size || PAGINATION.DEFAULT_PAGE_SIZE;
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
            // Normalize flat edge fields (user, category, channel) to edges structure
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
 * Returns public fields only, auto-increments view count
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
