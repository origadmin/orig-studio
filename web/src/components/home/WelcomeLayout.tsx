import React from 'react';
import {Link} from '@tanstack/react-router';
import {useModuleState} from '@/contexts/ModuleConfigContext';
import {FileText, Video, Music, Settings} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useAuth} from '@/hooks/useAuth';

const WelcomeLayout: React.FC = () => {
    const {modules, site} = useModuleState();
    const {isAdmin} = useAuth();

    const availableModules = [
        {key: 'articles' as const, label: 'Articles', icon: FileText, enabled: modules.articles},
        {key: 'videos' as const, label: 'Videos', icon: Video, enabled: modules.videos},
        {key: 'music' as const, label: 'Music', icon: Music, enabled: modules.music},
    ];

    return (
        <div className="flex flex-col items-center justify-center min-h-[70vh] text-center px-4">
            <h1 className="text-4xl font-bold mb-4">
                {site.site_name || 'Welcome'}
            </h1>
            {site.site_description && (
                <p className="text-muted-foreground text-lg mb-8 max-w-md">
                    {site.site_description}
                </p>
            )}

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8 max-w-lg w-full">
                {availableModules.map((mod) => (
                    <div
                        key={mod.key}
                        className={`flex flex-col items-center p-4 rounded-lg border ${
                            mod.enabled
                                ? 'border-primary/50 bg-primary/5'
                                : 'border-muted bg-muted/30 opacity-50'
                        }`}
                    >
                        <mod.icon className="h-8 w-8 mb-2"/>
                        <span className="text-sm font-medium">{mod.label}</span>
                        <span className="text-xs text-muted-foreground">
                            {mod.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </div>
                ))}
            </div>

            {isAdmin && (
                <Link to="/admin/settings">
                    <Button variant="outline" className="gap-2">
                        <Settings className="h-4 w-4"/>
                        Configure Modules
                    </Button>
                </Link>
            )}
        </div>
    );
};

export default WelcomeLayout;
