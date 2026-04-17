/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 历史记录页
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {History, Play, Eye, Trash2, X, Loader2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useQuery, useMutation, useQueryClient} from '@tanstack/react-query';
import {formatDuration, formatDate} from '@/lib/format';
import {historyApi} from '@/lib/api/history';
import {getFullUrl} from '@/lib/utils';

const HistoryPage = () => {
    const {t} = useTranslation();
    const {user} = useAuth();
    const queryClient = useQueryClient();
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(false);
    const sentinelRef = useRef<HTMLDivElement>(null);

    const [page, setPage] = useState(1);
    const pageSize = 10;

    const {data, isLoading, error} = useQuery({
        queryKey: ['history', user?.id, page],
        queryFn: async () => {
            if (!user) throw new Error('User not logged in');
            return await historyApi.list({page, page_size: pageSize});
        },
        enabled: !!user
    });

    const clearHistory = useMutation({
        mutationFn: () => historyApi.clear(),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['history', user?.id]});
        }
    });

    const removeHistory = useMutation({
        mutationFn: (historyId: number) => historyApi.remove(historyId),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['history', user?.id]});
        }
    });

    const loadMore = useCallback(() => {
        if (loading || !hasMore) return;
        setPage(prev => prev + 1);
    }, [loading, hasMore]);

    // Update hasMore when data changes
    useEffect(() => {
        if (data) {
            setHasMore(data.items.length === pageSize);
        }
    }, [data, pageSize]);

    useEffect(() => {
        const el = sentinelRef.current;
        if (!el) return;
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting) loadMore();
            },
            {rootMargin: '200px'}
        );
        observer.observe(el);
        return () => observer.disconnect();
    }, [loadMore]);

    const items = data?.items || [];

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error || !user || items.length === 0) {
        return (
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <History size={24} className="text-emerald-600"/>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('history.title')}</h1>
                    </div>
                </div>
                <div className="text-center py-20 text-gray-400">
                    <History size={48} className="mx-auto mb-3 opacity-30"/>
                    <p className="text-lg mb-1">{t('history.empty')}</p>
                    <p className="text-sm">{t('history.emptyDesc')}</p>
                </div>
            </div>
        );
    }



    return (
        <div className="space-y-6">
            {/* 标题 */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <History size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('history.title')}</h1>
                    <span className="text-sm text-gray-500">{t('history.recordCount', {count: items.length})}</span>
                </div>
                <Button
                    onClick={clearHistory}
                    variant="ghost"
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                >
                    <Trash2 size={14}/> {t('history.clear')}
                </Button>
            </div>

            {/* 历史列表 */}
            <div className="space-y-2">
                {items.map((item) => (
                    <Link
                        key={item.id}
                        to="/watch" search={{v: item.friendly_token || String(item.media_id)}}
                        className="group flex items-center gap-4 p-3 bg-white dark:bg-gray-800 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
                    >
                        {/* 缩略图 */}
                        <div className="relative w-40 shrink-0 aspect-video rounded-lg overflow-hidden">
                            <img src={item.media?.thumbnail ? getFullUrl(item.media.thumbnail) : undefined}
                                 alt={item.media?.title}
                                 className="w-full h-full object-cover"/>
                            <div
                                className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1 py-0.5 rounded">
                                {formatDuration(item.media?.duration || 0)}
                            </div>
                            {/* 进度条 */}
                            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-gray-600/30">
                                <div
                                    className="h-full bg-emerald-500"
                                    style={{width: `${item.progress || 0}%`}}
                                />
                            </div>
                        </div>

                        {/* 信息 */}
                        <div className="flex-1 min-w-0">
                            <h3 className="text-sm font-medium text-gray-900 dark:text-white line-clamp-1 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                {item.media?.title}
                            </h3>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{item.media?.edges?.user?.[0]?.username || 'Unknown'}</p>
                            <p className="text-xs text-gray-400 mt-0.5">{formatDate(item.watched_at)}</p>
                        </div>

                        {/* 移除按钮 */}
                        <Button
                            onClick={(e) => {
                                e.preventDefault();
                                removeHistory.mutate(item.id);
                            }}
                            variant="ghost"
                            size="sm"
                            className="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg opacity-0 group-hover:opacity-100 transition-all"
                        >
                            <X size={14}/>
                        </Button>
                    </Link>
                ))}
            </div>

            {/* 无限滚动哨兵 */}
            <div ref={sentinelRef} className="flex items-center justify-center py-6">
                {loading && (
                    <div
                        className="animate-spin w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full"
                    />
                )}
            </div>
        </div>
    );
};

export default HistoryPage;
