import {useQuery} from '@tanstack/react-query';
import {portalApi, type ModulePortalConfig} from '@/lib/api/portal';

const FALLBACK_CONFIG: ModulePortalConfig = {
    modules: {articles: true, videos: true, music: false},
    layout: 'video',
    site: {
        site_name: 'OrigCMS',
        site_description: '',
        allow_registration: true,
        allow_upload: true,
    },
};

export function useModuleConfig() {
    return useQuery({
        queryKey: ['portal-config'],
        queryFn: () => portalApi.getModuleConfig(),
        staleTime: 5 * 60_000,
        gcTime: 30 * 60_000,
        retry: 1,
        refetchOnWindowFocus: false,
        placeholderData: FALLBACK_CONFIG,
    });
}

export type {ModulePortalConfig};
