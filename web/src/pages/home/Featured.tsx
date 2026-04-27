import {Spinner} from "@/components/ui/spinner"
/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Featured Page
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Star, Loader2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {useMediaList} from '@/hooks/queries';
import {getFullUrl} from '@/lib/utils';
import ErrorPage from '@/components/common/ErrorPage';

const FeaturedPage = () => {
    const {t} = useTranslation();

    const {data, isLoading, error} = useMediaList({
        featured: 'true',
        page_size: 10,
        status: 'active'
    });

    const featuredMedia = data?.items || [];

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error.message || t('common.error')}/>;
    }

    if (featuredMedia.length === 0) {
        return (
            <div className="py-20 text-center space-y-4">
                <div className="text-gray-500 text-lg">{t('common.noData')}</div>
                <Link to="/">
                    <Button variant="outline">{t('common.backToHome')}</Button>
                </Link>
            </div>
        );
    }



    return (
        <div className="space-y-8">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <Star size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('featured.title')}</h1>
                </div>
                <span className="text-sm text-gray-500 dark:text-muted-foreground">
                    {t('featured.featuredCount', {count: featuredMedia.length})}
                </span>
            </div>

            {/* Hero cards */}
            {featuredMedia.slice(0, 2).map((item) => (
                <Link key={item.id} to="/watch" search={{v: item.short_token}} className="group block">
                    <div
                        className="relative rounded-2xl overflow-hidden bg-gradient-to-r from-gray-900 to-gray-800 dark:from-gray-800 dark:to-gray-900">
                        <div className="flex flex-col md:flex-row">
                            <div className="relative aspect-video md:w-[480px] shrink-0">
                                <img src={item.thumbnail ? getFullUrl(item.thumbnail) : undefined} alt={item.title}
                                     className="w-full h-full object-cover"/>
                                <div className="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent"/>
                                <div
                                    className="absolute bottom-3 right-3 bg-black/80 text-white text-xs px-2 py-1 rounded">
                                    {formatDuration(item.duration)}
                                </div>
                                <div
                                    className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                    <div
                                        className="w-16 h-16 bg-white/90 rounded-full flex items-center justify-center shadow-xl">
                                        <Play className="w-7 h-7 text-gray-900 ml-1" fill="currentColor"/>
                                    </div>
                                </div>
                            </div>
                            <div className="p-6 md:p-8 flex flex-col justify-center">
                            <span className="inline-flex items-center gap-1 text-emerald-400 text-xs font-medium mb-3">
                                <Star size={12} fill="currentColor"/> {' '}
                                {t('featured.editorPick')}
                            </span>
                                <h2 className="text-2xl font-bold text-white mb-3 group-hover:text-emerald-300 transition-colors">
                                    {item.title}
                                </h2>
                                <p className="text-gray-300 text-sm mb-4 line-clamp-2">{item.description || t('watch.noDescription')}</p>
                                <div className="flex items-center gap-3">
                                    <img
                                        src={item.edges?.user?.[0]?.avatar ? getFullUrl(item.edges.user[0].avatar) : undefined}
                                         alt={item.edges?.user?.[0]?.username}
                                         className="w-8 h-8 rounded-full"/>
                                    <span
                                        className="text-muted-foreground text-sm">{item.edges?.user?.[0]?.username || 'Unknown'}</span>
                                    <span className="text-gray-500 text-sm flex items-center gap-1">
                                        <Eye size={14}/>{formatViews(item.view_count)} {t('common.views')}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>
                </Link>
            ))}

            {/* Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {featuredMedia.slice(2).map((item) => (
                    <Link key={item.id} to="/watch" search={{v: item.short_token}} className="group">
                        <div
                            className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                            <div className="relative aspect-video overflow-hidden">
                                <img src={item.thumbnail ? getFullUrl(item.thumbnail) : undefined} alt={item.title}
                                     className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                                <div
                                    className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">
                                    {formatDuration(item.duration)}
                                </div>
                            </div>
                            <div className="p-4">
                                <h3 className="font-semibold text-gray-900 dark:text-white line-clamp-2 mb-2 group-hover:text-emerald-600 transition-colors">
                                    {item.title}
                                </h3>
                                <p className="text-sm text-gray-500 dark:text-muted-foreground line-clamp-1">{item.description || t('watch.noDescription')}</p>
                            </div>
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    );
};

export default FeaturedPage;
