import React from 'react';
import { createFileRoute } from '@tanstack/react-router';
import PortalLayout from '@/layout/PortalLayout';
import DocPortalLayout from '@/layout/DocPortalLayout';
import { ModuleConfigProvider, useModuleState } from '@/contexts/ModuleConfigContext';

const LayoutSwitcher: React.FC = () => {
    const { layout } = useModuleState();

    if (layout === 'doc') {
        return <DocPortalLayout />;
    }

    return <PortalLayout />;
};

export const Route = createFileRoute('/_portal')({
    component: () => (
        <ModuleConfigProvider>
            <LayoutSwitcher />
        </ModuleConfigProvider>
    ),
});
