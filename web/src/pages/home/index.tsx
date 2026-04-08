/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Home Page - Feed + Infinite Scroll (Connected to Real API)
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, TrendingUp, Star} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {mediaApi, type Media} from '@/lib/api/media';
import {API_BASE_URL} from '@/lib/request';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {useInfiniteMediaList, useMediaList} from '@/hooks/queries';
import HorizontalScroll from '@/components/common/HorizontalScroll';

const categories = [
    {id: 1, name: 'Technology'},
    {id: 2, name: 'Programming'},
    {id: 3, name: 'Design'},
    {id: 4, name: 'Lifestyle'},
    {id: 5, name: 'Entertainment'},
];

const HomePage = () => {
    const {t} = useTranslation();
    const [activeCategoryId, setActiveCategoryId] = useState<number | null>(null);

    // Featured videos
    const {data: featuredData} = useMediaList({
        page: 1,
        page_size: 6,
        featured: true,
    });
    const featuredVideos = featuredData?.list || [];

    // Recommended videos
    const {data: recommendedData} = useMediaList({
        page: 1,
        page_size: 8,
    });
    const recommendedVideos = recommendedData?.list || [];

    // Infinite scroll video list
    const {
        data,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useInfiniteMediaList({
        page_size: 12,
    });

    const items = data?.pages.flatMap(page => page.list).filter(Boolean) || [];

    // Infinite scroll
    const sentinelRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const observer = new IntersectionObserver(
            (entries) => {
                if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
                    fetchNextPage();
                }
            },
            {threshold: 0.1}
        );

        if (sentinelRef.current) {
            observer.observe(sentinelRef.current);
        }

        return () => observer.disconnect();
    }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

    return (
        <div className="space-y-8">
            {/* Hero Banner */}
            <section
                className="relative rounded-2xl overflow-hidden bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 text-white">
                <div
                    className="absolute inset-0 bg-cover bg-center opacity-20"
                    style={{backgroundImage: 'url(/assets/images/cover-placeholder.svg)'}}/>
                <div className="relative px-6 py-6 flex items-center">
                    <div className="max-w-xl">
                        <Badge className="bg-emerald-500/20 text-emerald-300 hover:bg-emerald-500/30 mb-4">
                            <TrendingUp className="w-3 h-3 mr-1"/> {t('home.heroBadge')}
                        </Badge>
                        <h1 className="text-4xl font-black mb-4 leading-tight">{t('home.heroTitle')}</h1>
                        <p className="text-lg text-slate-300 mb-6">{t('home.heroDesc')}</p>
                        <div className="flex gap-4">
                            <Link to="/featured">
                                <Button size="lg"
                                        className="bg-emerald-600 hover:bg-emerald-700">{t('home.exploreContent')}</Button>
                            </Link>
                            <Link to="/me/upload">
                                <Button size="lg" variant="outline"
                                        className="border-slate-600 text-white hover:bg-slate-800">
                                    {t('home.startCreating')}
                                </Button>
                            </Link>
                        </div>
                    </div>
                </div>
            </section>

            {/* Featured Videos */}
            <section className="mb-12">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                        <Star className="w-6 h-6 text-yellow-500"/>
                        {t('home.featuredVideos')}
                    </h2>
                    <Link to="/featured"
                          className="text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 font-medium">
                        {t('home.viewAll')}
                    </Link>
                </div>
                <HorizontalScroll>
                    {featuredVideos.map(media => {
                        const user = media?.edges?.user?.[0];
                        // Handle thumbnail path
                        const thumbUrl = getImageUrl(media?.thumbnail, 'thumbnail');

                        return (
                            <Link key={media?.id} to="/watch" search={{v: String(media?.id)}}
                                  className="group w-64 flex-shrink-0">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                                    <div className="relative aspect-video overflow-hidden">
                                        <img src={thumbUrl} alt={media?.title}
                                             onError={(e) => handleImageError(e, 'thumbnail')}
                                             className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
                                            {formatDuration(media?.duration || 0)}
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
                                            {media?.title || 'Untitled'}
                                        </h3>
                                        <div className="flex items-center gap-2 mb-1">
                                            <img
                                                src={getImageUrl(user?.avatar, 'avatar')}
                                                alt={user?.username}
                                                onError={(e) => handleImageError(e, 'avatar')}
                                                className="w-5 h-5 rounded-full object-cover"/>
                                            <span
                                                className="text-xs text-gray-500 dark:text-gray-400">{user?.nickname || user?.username || 'Unknown'}</span>
                                        </div>
                                        <div
                                            className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                                <span className="flex items-center gap-1"><Eye
                                                    size={12}/>{formatViews(media?.view_count || 0)}</span>
                                        </div>
                                    </div>
                                </div>
                            </Link>
                        );
                    })}
                </HorizontalScroll>
            </section>

            {/* Recommended Videos */}
            <section className="mb-12">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-2xl font-bold text-slate-900 dark:text-white">
                        {t('home.recommendedVideos')}
                    </h2>
                    <Link to="/latest"
                          className="text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 font-medium">
                        {t('home.viewAll')}
                    </Link>
                </div>
                <HorizontalScroll>
                    {recommendedVideos.map(media => {
                        const user = media?.edges?.user?.[0];
                        // Handle thumbnail path
                        const thumbUrl = getImageUrl(media?.thumbnail, 'thumbnail');

                        return (
                            <Link key={media?.id} to="/watch" search={{v: String(media?.id)}}
                                  className="group w-64 flex-shrink-0">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                                    <div className="relative aspect-video overflow-hidden">
                                        <img src={thumbUrl} alt={media?.title}
                                             onError={(e) => handleImageError(e, 'thumbnail')}
                                             className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
                                            {formatDuration(media?.duration || 0)}
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
                                            {media?.title || 'Untitled'}
                                        </h3>
                                        <div className="flex items-center gap-2 mb-1">
                                            <img
                                                src={getImageUrl(user?.avatar, 'avatar')}
                                                alt={user?.username}
                                                onError={(e) => handleImageError(e, 'avatar')}
                                                className="w-5 h-5 rounded-full object-cover"/>
                                            <span
                                                className="text-xs text-gray-500 dark:text-gray-400">{user?.nickname || user?.username || 'Unknown'}</span>
                                        </div>
                                        <div
                                            className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                                <span className="flex items-center gap-1"><Eye
                                                    size={12}/>{formatViews(media?.view_count || 0)}</span>
                                        </div>
                                    </div>
                                </div>
                            </Link>
                        );
                    })}
                </HorizontalScroll>
            </section>

            {/* Latest Videos */}
            <section className="mb-12">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-2xl font-bold text-slate-900 dark:text-white">
                        {activeCategoryId === null
                            ? t('home.latestVideos')
                            : t('home.categoryVideos', {category: categories.find(c => c.id === activeCategoryId)?.name})}
                    </h2>
                    <Link to="/latest"
                          className="text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 font-medium">
                        {t('home.viewAll')}
                    </Link>
                </div>

                {/* Video Grid */}
                {items.length === 0 && !isFetchingNextPage ? (
                    <div className="py-20 text-center text-gray-500">
                        <p>{t('common.noData')}</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                        {items.map(media => {
                            const user = media?.edges?.user?.[0];
                            // Handle thumbnail path
                            const thumbUrl = getImageUrl(media?.thumbnail, 'thumbnail');

                            return (
                                <Link key={media?.id} to="/watch" search={{v: String(media?.id)}} className="group">
                                    <div
                                        className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                                        <div className="relative aspect-video overflow-hidden">
                                            <img src={thumbUrl} alt={media?.title}
                                                 onError={(e) => handleImageError(e, 'thumbnail')}
                                                 className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                                            <div
                                                className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
                                                {formatDuration(media?.duration || 0)}
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
                                                {media?.title || 'Untitled'}
                                            </h3>
                                            <div className="flex items-center gap-2 mb-1">
                                                <img
                                                    src={getImageUrl(user?.avatar, 'avatar')}
                                                    alt={user?.username}
                                                    onError={(e) => handleImageError(e, 'avatar')}
                                                    className="w-5 h-5 rounded-full object-cover"/>
                                                <span
                                                    className="text-xs text-gray-500 dark:text-gray-400">{user?.nickname || user?.username || 'Unknown'}</span>
                                            </div>
                                            <div
                                                className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                                <span className="flex items-center gap-1"><Eye
                                                    size={12}/>{formatViews(media?.view_count || 0)}</span>
                                                <span>{formatDate(media?.created_at || new Date().toISOString())}</span>
                                            </div>
                                            <div className="flex flex-wrap gap-1 mt-2">
                                                {media?.tags?.slice(0, 2).map((tag: string, tIdx: number) => (
                                                    <Badge key={`${tag}-${tIdx}`} variant="secondary"
                                                           className="text-xs">{tag}</Badge>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                </Link>
                            );
                        })}
                    </div>
                )}
            </section>

            {/* Infinite Scroll Sentinel */}
            <div ref={sentinelRef} className="flex flex-col items-center py-8">
                {isFetchingNextPage && (
                    <div className="flex items-center gap-3 text-gray-400 py-2">
                        <div
                            className="animate-spin w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full"/>
                        <span className="text-sm">{t('common.loading')}</span>
                    </div>
                )}
                {!hasNextPage && items.length > 0 && (
                    <p className="text-sm text-gray-400 py-4">— {t('common.allLoaded')} —</p>
                )}
            </div>
        </div>
    );
};

export default HomePage;