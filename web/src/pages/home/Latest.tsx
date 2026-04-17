/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Latest Page — infinite scroll
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {Clock, Play, Eye} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {useMediaList, useCategoryList} from '@/hooks/queries';
import {getFullUrl} from '@/lib/utils';
import ErrorPage from '@/components/common/ErrorPage';

const PAGE_SIZE = 12;

const LatestPage = () => {
    const {t} = useTranslation();
    const [page, setPage] = useState(1);
    const [items, setItems] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const [activeCategoryId, setActiveCategoryId] = useState<number | null>(null);
    const sentinelRef = useRef<HTMLDivElement>(null);

    // 获取分类列表
    const {data: categories} = useCategoryList();

    const {data, isLoading, error} = useMediaList({
        page,
        page_size: PAGE_SIZE,
        status: 'active',
        sort: 'created_at',
        order: 'desc',
        category_id: activeCategoryId || undefined
    });

    // 当分类切换时重置页面和数据
    useEffect(() => {
        setPage(1);
        setItems([]);
        setHasMore(true);
    }, [activeCategoryId]);

    // Load initial data and append when page changes
    useEffect(() => {
        if (data?.items && data.items.length > 0) {
            if (page === 1) {
                setItems(data.items);
            } else {
                setItems(prev => [...prev, ...data.items]);
            }
            setHasMore(data.items.length === PAGE_SIZE);
        } else if (page > 1) {
            setHasMore(false);
        }
    }, [data, page]);

    const loadMore = useCallback(() => {
        if (isLoading || !hasMore) return;
        setPage(prev => prev + 1);
    }, [isLoading, hasMore]);

    useEffect(() => {
        const el = sentinelRef.current;
        if (!el) return;
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting) loadMore();
            },
            {rootMargin: '200px'},
        );
        observer.observe(el);
        return () => observer.disconnect();
    }, [loadMore]);


    if (error && items.length === 0) {
        return <ErrorPage message={error.message || t('common.error')}/>;
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Clock size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('latest.title')}</h1>
            </div>

            {/* 分类标签 */}
            <section className="flex flex-wrap gap-2 mb-6">
                <Button
                    variant={activeCategoryId === null ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setActiveCategoryId(null)}
                    className={activeCategoryId === null ? 'bg-emerald-600 hover:bg-emerald-700' : ''}
                >{t('home.all')}</Button>
                {categories?.items?.map((cat) => (
                    <Button
                        key={cat.id}
                        variant={activeCategoryId === cat.id ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setActiveCategoryId(cat.id)}
                        className={activeCategoryId === cat.id ? 'bg-emerald-600 hover:bg-emerald-700' : ''}
                    >{cat.name}</Button>
                ))}
            </section>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                {items.map((media) => (
                    <Link key={media.id} to="/watch" search={{v: media.friendly_token || String(media.id)}} className="group">
                        <div
                            className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                            <div className="relative aspect-video overflow-hidden">
                                <img src={media.thumbnail ? getFullUrl(media.thumbnail) : undefined} alt={media.title}
                                     className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                                <div
                                    className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
                                    {formatDuration(media.duration)}
                                </div>
                                <div
                                    className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                    <div
                                        className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg">
                                        <Play className="w-5 h-5 text-gray-900 ml-0.5" fill="currentColor"/>
                                    </div>
                                </div>
                            </div>
                            <div className="p-3">
                                <h3 className="font-medium text-gray-900 dark:text-white text-sm line-clamp-2 mb-1.5 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                    {media.title}
                                </h3>
                                <div className="flex items-center gap-2 mb-1">
                                    <img
                                        src={media.edges?.user?.[0]?.avatar ? getFullUrl(media.edges.user[0].avatar) : undefined}
                                         alt={media.edges?.user?.[0]?.username}
                                         className="w-5 h-5 rounded-full object-cover"/>
                                    <span
                                        className="text-xs text-gray-500 dark:text-gray-400">{media.edges?.user?.[0]?.username || 'Unknown'}</span>
                                </div>
                                <div className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                    <span className="flex items-center gap-1"><Eye
                                        size={12}/>{formatViews(media.view_count)}</span>
                                    <span>{formatDate(media.created_at)}</span>
                                </div>
                            </div>
                        </div>
                    </Link>
                ))}
            </div>

            <div ref={sentinelRef} className="flex flex-col items-center py-8">
                {isLoading && (
                    <div className="flex items-center gap-3 text-gray-400">
                        <div
                            className="animate-spin w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full"/>
                        <span className="text-sm">{t('common.loading')}</span>
                    </div>
                )}
                {!hasMore && items.length > 0 && (
                    <p className="text-sm text-gray-400 py-4">— {t('common.allLoaded')} —</p>
                )}
                {error && items.length > 0 && (
                    <p className="text-sm text-red-500 py-4">{t('common.error')}</p>
                )}
            </div>
        </div>
    );
};

export default LatestPage;
