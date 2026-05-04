import React from 'react';
import {useModuleState} from '@/contexts/ModuleConfigContext';
import WelcomeLayout from './WelcomeLayout';
import ArticleHomeLayout from './ArticleHomeLayout';
import MixedHomeLayout from './MixedHomeLayout';
import DocHomeLayout from './DocHomeLayout';

const VideoHomeLayout = React.lazy(() => import('@/pages/home/index'));

const DynamicHomeLayout: React.FC = () => {
    const {layout} = useModuleState();

    switch (layout) {
        case 'doc':
            return <DocHomeLayout/>;
        case 'article':
            return <ArticleHomeLayout/>;
        case 'video':
            return (
                <React.Suspense fallback={
                    <div className="flex items-center justify-center min-h-[50vh]">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"/>
                    </div>
                }>
                    <VideoHomeLayout/>
                </React.Suspense>
            );
        case 'mixed':
            return <MixedHomeLayout/>;
        case 'welcome':
        default:
            return <WelcomeLayout/>;
    }
};

export default DynamicHomeLayout;
