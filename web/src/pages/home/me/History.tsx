/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Watch History page - supports both authenticated (remote) and anonymous (local) users
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {History, Trash2, X, AlertCircle} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useQuery, useMutation, useQueryClient} from '@tanstack/react-query';
import {formatDuration, formatDate} from '@/lib/format';
import {createHistoryService} from '@/lib/services/history';
import {historyApi} from '@/lib/api/history';
import type {HistoryItem} from '@/lib/api/history';
import {getFullUrl} from '@/lib/utils';
import {DeleteConfirmDialog} from '@/components/common/DeleteConfirmDialog';
import {Spinner} from '@/components/ui/spinner';

const HistoryPage = () => {
    const {t} = useTranslation();
    const {user, isAuthenticated} = useAuth();
    const queryClient = useQueryClient();
    const [showClearConfirm, setShowClearConfirm] = useState(false);
    const sentinelRef = useRef<HTMLDivElement>(null);

    const [page, setPage] = useState(1);
    const pageSize = 20;

    // Select service based on auth state
    const historyService = createHistoryService(isAuthenticated);

    const {data, isLoading, error, refetch} = useQuery({
        queryKey: ['history', user?.id, page],
        queryFn: async () => {
            return await historyService.list({page, page_size: pageSize});
        },
        staleTime: 0,
        refetchOnMount: 'always',
    });

    const clearHistory = useMutation({
        mutationFn: () => historyService.clear(),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['history', user?.id]});
            setShowClearConfirm(false);
        },
    });

    const removeHistory = useMutation({
        mutationFn: (historyId: string) => historyService.remove(historyId),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['history', user?.id]});
        },
    });

    const [hasMore, setHasMore] = useState(false);

    const loadMore = useCallback(() => {
        if (hasMore) {
            setPage(prev => prev + 1);
        }
    }, [hasMore]);

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
                <Spinner/>
            </div>
        );
    }

    if (error || items.length === 0) {
        return (
            <div className="space-y-6">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <History size={24} className="text-emerald-600"/>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('history.title')}</h1>
                    </div>
                </div>
                <div className="text-center py-20 text-muted-foreground">
                    <History size={48} className="mx-auto mb-3 opacity-30"/>
                    <p className="text-lg mb-1">{t('history.empty')}</p>
                    <p className="text-sm">{t('history.emptyDesc')}</p>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <History size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('history.title')}</h1>
                    <span className="text-sm text-gray-500">{t('history.recordCount', {count: data?.total || items.length})}</span>
                </div>
                <Button
                    onClick={() => setShowClearConfirm(true)}
                    variant="ghost"
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-destructive hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                >
                    <Trash2 size={14}/> {t('history.clear')}
                </Button>
            </div>

            {/* History list */}
            <div className="space-y-2">
                {items.map((item: HistoryItem) => {
                    const progressPercent = item.duration_seconds > 0
                        ? Math.min(100, (item.progress_seconds / item.duration_seconds) * 100)
                        : 0;

                    const isDeleted = !item.title && !item.short_token;
                    const contentLink = isDeleted ? '' : (item.short_token
                        ? `/watch?v=${item.short_token}`
                        : item.content_type === 'article'
                            ? `/articles/${item.content_id}`
                            : `/watch?v=${item.content_id}`);

                    const content = (
                        <>
                            {/* Thumbnail */}
                            <div className="relative w-40 shrink-0 aspect-video rounded-lg overflow-hidden">
                                {isDeleted ? (
                                    <div className="w-full h-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center">
                                        <AlertCircle size={24} className="text-gray-400"/>
                                    </div>
                                ) : item.thumbnail ? (
                                    <img
                                        src={getFullUrl(item.thumbnail)}
                                        alt={item.title || ''}
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <div className="w-full h-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
                                        <History size={24} className="text-gray-400"/>
                                    </div>
                                )}
                                {item.duration_seconds > 0 && !isDeleted && (
                                    <div className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1 py-0.5 rounded">
                                        {formatDuration(item.duration_seconds)}
                                    </div>
                                )}
                                {!isDeleted && (
                                    <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-gray-600/30">
                                        <div
                                            className="h-full bg-emerald-500"
                                            style={{width: `${progressPercent}%`}}
                                        />
                                    </div>
                                )}
                            </div>

                            {/* Info */}
                            <div className="flex-1 min-w-0">
                                <h3 className={`text-sm font-medium line-clamp-1 transition-colors ${
                                    isDeleted
                                        ? 'text-gray-400 dark:text-gray-500 line-through'
                                        : 'text-gray-900 dark:text-white group-hover:text-emerald-600 dark:group-hover:text-emerald-400'
                                }`}>
                                    {isDeleted
                                        ? (t('history.contentUnavailable') || 'Content no longer available')
                                        : (item.title || item.content_id)
                                    }
                                </h3>
                                <div className="flex items-center gap-2 mt-1">
                                    {isDeleted && (
                                        <span className="text-[10px] bg-gray-100 dark:bg-gray-700 text-gray-500 px-1.5 py-0.5 rounded">
                                            {t('history.removed') || 'Removed'}
                                        </span>
                                    )}
                                    {item.is_finished && !isDeleted && (
                                        <span className="text-[10px] bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 px-1.5 py-0.5 rounded">
                                            {t('history.watched') || 'Watched'}
                                        </span>
                                    )}
                                    {!isDeleted && item.progress_seconds > 0 && item.duration_seconds > 0 && (
                                        <span className="text-xs text-muted-foreground">
                                            {formatDuration(item.progress_seconds)} / {formatDuration(item.duration_seconds)}
                                        </span>
                                    )}
                                    <span className="text-xs text-muted-foreground">{formatDate(item.last_watched_at)}</span>
                                </div>
                            </div>

                            {/* Remove button */}
                            <Button
                                onClick={(e) => {
                                    e.preventDefault();
                                    e.stopPropagation();
                                    removeHistory.mutate(item.id);
                                }}
                                variant="ghost"
                                size="sm"
                                className="p-1.5 text-muted-foreground hover:text-destructive hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg opacity-0 group-hover:opacity-100 transition-all"
                            >
                                <X size={14}/>
                            </Button>
                        </>
                    );

                    return isDeleted ? (
                        <div
                            key={item.id}
                            className="group flex items-center gap-4 p-3 bg-white dark:bg-gray-800 rounded-xl opacity-60"
                        >
                            {content}
                        </div>
                    ) : (
                        <Link
                            key={item.id}
                            to={contentLink}
                            className="group flex items-center gap-4 p-3 bg-white dark:bg-gray-800 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
                        >
                            {content}
                        </Link>
                    );
                })}
            </div>

            {/* Infinite scroll sentinel */}
            <div ref={sentinelRef} className="flex items-center justify-center py-6">
                {hasMore && (
                    <Spinner size="sm"/>
                )}
            </div>

            {/* Clear history confirmation dialog */}
            <DeleteConfirmDialog
                open={showClearConfirm}
                onOpenChange={setShowClearConfirm}
                title={t('history.clearConfirmTitle') || 'Clear Watch History'}
                description={t('history.clearConfirmDesc') || 'Are you sure you want to clear all watch history? This action cannot be undone.'}
                confirmLabel={t('history.clearConfirm') || 'Clear All'}
                confirmVariant="destructive"
                loadingLabel={t('history.clearing') || 'Clearing...'}
                isDeleting={clearHistory.isPending}
                onConfirm={() => clearHistory.mutate()}
            />
        </div>
    );
};

export default HistoryPage;
