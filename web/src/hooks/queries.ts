import {useQuery, useMutation, useQueryClient, useInfiniteQuery} from '@tanstack/react-query';
import {mediaApi, type Media} from '@/lib/api/media';

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
}) {
    return useQuery({
        queryKey: mediaKeys.list(params),
        queryFn: async () => {
            const res = await mediaApi.list({
                ...params,
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
            const items = lastPage.list || (Array.isArray(lastPage) ? lastPage : []);
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
 * useMediaDetail: Fetch single media details
 */
export function useMediaDetail(id: string | null) {
    return useQuery({
        queryKey: mediaKeys.detail(id!),
        queryFn: async () => {
            const res = await mediaApi.get(id!);
            return res;
        },
        enabled: !!id,
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
