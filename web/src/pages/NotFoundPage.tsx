import React from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate, Link} from '@tanstack/react-router';
import {Home, Search} from 'lucide-react';

const NotFoundPage: React.FC = () => {
    const {t} = useTranslation();
    const navigate = useNavigate();

    return (
        <div className="min-h-[70vh] flex items-center justify-center bg-background px-4">
            <div className="text-center space-y-6 max-w-md">
                <div className="text-8xl font-bold text-primary/20 select-none">
                    404
                </div>

                <div className="w-16 h-1 bg-primary/30 rounded-full mx-auto"/>

                <div className="space-y-2">
                    <h1 className="text-2xl font-bold text-foreground flex items-center justify-center gap-2">
                        <Search className="w-6 h-6"/>
                        {t('notFound.title') || '页面未找到'}
                    </h1>
                    <p className="text-muted-foreground text-sm">
                        {t('notFound.description') || '抱歉，您访问的页面不存在或已被移除。'}
                    </p>
                </div>

                <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
                    <button
                        onClick={() => navigate({to: '/'})}
                        className="inline-flex items-center justify-center gap-2 px-5 py-2.5 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors font-medium"
                    >
                        <Home className="w-4 h-4"/>
                        {t('notFound.backToHome') || '返回首页'}
                    </button>
                    <Link
                        to="/"
                        className="inline-flex items-center justify-center gap-2 px-5 py-2.5 border border-input rounded-lg hover:bg-accent transition-colors font-medium"
                    >
                        <Search className="w-4 h-4"/>
                        {t('notFound.explore') || '探索内容'}
                    </Link>
                </div>

                <p className="text-xs text-muted-foreground pt-4">
                    Error ID: {window.location.pathname}
                </p>
            </div>
        </div>
    );
};

export default NotFoundPage;
