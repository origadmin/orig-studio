import React, {createContext, useContext, useMemo} from 'react';
import {useModuleConfig, type ModulePortalConfig} from '@/hooks/useModuleConfig';

const DEFAULT_CONFIG: ModulePortalConfig = {
    modules: {articles: true, videos: true, music: false},
    layout: 'video',
    site: {
        site_name: 'OrigCMS',
        site_description: '',
        allow_registration: true,
        allow_upload: true,
    },
};

const ModuleConfigContext = createContext<ModulePortalConfig>(DEFAULT_CONFIG);

export function ModuleConfigProvider({children}: { children: React.ReactNode }) {
    const {data} = useModuleConfig();
    const value = useMemo(() => data ?? DEFAULT_CONFIG, [data]);
    return (
        <ModuleConfigContext.Provider value={value}>
            {children}
        </ModuleConfigContext.Provider>
    );
}

export function useModuleState(): ModulePortalConfig {
    return useContext(ModuleConfigContext);
}
