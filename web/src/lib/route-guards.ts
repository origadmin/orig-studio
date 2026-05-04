import {redirect} from '@tanstack/react-router';
import {useModuleState} from '@/contexts/ModuleConfigContext';
import type {ModulePortalConfig} from '@/hooks/useModuleConfig';

type ModuleKey = keyof ModulePortalConfig['modules'];

export function useModuleGuard(module: ModuleKey): boolean {
    const {modules} = useModuleState();
    return modules[module] === true;
}

export const ROUTE_MODULE_MAP: Record<string, ModuleKey> = {
    '/articles': 'articles',
    '/me/articles': 'articles',
    '/me/upload': 'videos',
    '/me/videos': 'videos',
    '/watch': 'videos',
    '/featured': 'videos',
    '/latest': 'videos',
    '/me/channels': 'videos',
    '/me/playlists': 'videos',
    '/subscriptions': 'videos',
};
