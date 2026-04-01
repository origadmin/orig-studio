/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 视频播放页 - 对接真实数据
 */

import React, {useState, useEffect} from 'react';
import {useSearch, Link} from '@tanstack/react-router';
import {
    ThumbsUp, ThumbsDown, Share2, MessageCircle,
    MoreHorizontal, UserPlus, Eye
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent} from '@/components/ui/card';
import {Skeleton} from '@/components/ui/skeleton';
import {formatViews, formatDate, formatDuration} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {type Media} from '@/lib/api/media';
import {useMediaDetail, useMediaList} from '@/hooks/queries';

const WatchPage = () => {
    const {t} = useTranslation();
    const {v: id} = useSearch({strict: false});
    const {data: media, isLoading: isMediaLoading, error: mediaError} = useMediaDetail(id as string);

    const {data: recData} = useMediaList({
        page_size: 10,
        category_id: media?.edges?.category?.id || undefined,
        status: 'active'
    });

    const recommendations = recData?.list?.filter(m => m.id !== Number(id)) || [];
    const loading = isMediaLoading;
    const error = mediaError ? t('watch.failedToLoad') : null;

    if (loading) {
        return (
            <div className="flex flex-col lg:flex-row gap-6 animate-pulse">
                <div className="flex-1 space-y-4">
                    <Skeleton className="aspect-video w-full rounded-2xl"/>
                    <Skeleton className="h-8 w-3/4"/>
                    <div className="flex justify-between items-center">
                        <div className="flex items-center gap-3">
                            <Skeleton className="h-12 w-12 rounded-full"/>
                            <div className="space-y-2">
                                <Skeleton className="h-4 w-24"/>
                                <Skeleton className="h-3 w-16"/>
                            </div>
                        </div>
                        <Skeleton className="h-10 w-32 rounded-full"/>
                    </div>
                </div>
                <div className="lg:w-80 xl:w-96 space-y-4">
                    {Array.from({length: 5}).map((_, i) => (
                        <div key={i} className="flex gap-3">
                            <Skeleton className="w-40 aspect-video rounded-lg shrink-0"/>
                            <div className="flex-1 space-y-2 py-1">
                                <Skeleton className="h-4 w-full"/>
                                <Skeleton className="h-3 w-2/3"/>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    if (error || !media) {
        return (
            <div className="py-20 text-center space-y-4">
                <div className="text-red-500 text-lg">{error || "Video not found"}</div>
                <Link to="/">
                    <Button variant="outline">{t('common.backToHome')}</Button>
                </Link>
            </div>
        );
    }

    const API_BASE_URL = (import.meta as any).env.VITE_API_BASE_URL || "http://localhost:9090";

    const getFullUrl = (path?: string) => {
        if (!path) return '';
        if (path.startsWith('http')) return path;
        const base = API_BASE_URL.replace(/\/$/, '');
        const sep = path.startsWith('/') ? '' : '/';
        return `${base}${sep}${path}`;
    };

    const videoUrl = getFullUrl(media.url);
    const user = media.edges?.user?.[0];

    return (
        <div className="flex flex-col lg:flex-row gap-6">
            {/* Main Content: Player & Details */}
            <div className="flex-1 min-w-0">
                {/* Player Container */}
                <div className="bg-black rounded-2xl overflow-hidden aspect-video shadow-2xl relative group">
                    <video
                        key={videoUrl}
                        src={videoUrl}
                        controls
                        autoPlay
                        className="w-full h-full"
                        poster={media.poster ? getFullUrl(media.poster) : (media.thumbnail ? getFullUrl(media.thumbnail) : undefined)}
                    >
                        Your browser does not support the video tag.
                    </video>
                </div>

                {/* Video Info */}
                <div className="mt-6 space-y-4">
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white line-clamp-2">
                        {media.title}
                    </h1>

                    <div
                        className="flex flex-wrap items-center justify-between gap-4 py-2 border-b dark:border-gray-800">
                        <div className="flex items-center gap-4">
                            <Link to="/u/$id" params={{id: String(media.user_id)}}>
                                <Avatar className="h-12 w-12 ring-2 ring-gray-100 dark:ring-gray-800">
                                    <AvatarImage src={user?.avatar}/>
                                    <AvatarFallback>{user?.username?.[0] || 'U'}</AvatarFallback>
                                </Avatar>
                            </Link>
                            <div>
                                <Link to="/u/$id" params={{id: String(media.user_id)}}
                                      className="font-bold text-gray-900 dark:text-white hover:text-blue-600 transition-colors">
                                    {user?.nickname || user?.username || 'Unknown Gopher'}
                                </Link>
                                <p className="text-xs text-gray-500 dark:text-gray-400">1.2M {t('watch.subscribers')}</p>
                            </div>
                            <Button
                                className="ml-4 rounded-full bg-gray-900 dark:bg-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-200">
                                {t('watch.subscribe')}
                            </Button>
                        </div>

                        <div className="flex items-center bg-gray-100 dark:bg-gray-800 rounded-full p-1">
                            <Button variant="ghost"
                                    className="rounded-l-full gap-2 px-4 hover:bg-gray-200 dark:hover:bg-gray-700">
                                <ThumbsUp className="w-5 h-5"/>
                                <span className="text-sm font-medium">{formatViews(media.like_count)}</span>
                            </Button>
                            <div className="w-[1px] h-6 bg-gray-300 dark:bg-gray-600"/>
                            <Button variant="ghost"
                                    className="rounded-r-full px-4 hover:bg-gray-200 dark:hover:bg-gray-700">
                                <ThumbsDown className="w-5 h-5"/>
                            </Button>
                        </div>
                    </div>

                    {/* Meta & Description */}
                    <Card
                        className="bg-gray-100 dark:bg-gray-800 border-none shadow-none rounded-xl overflow-hidden mt-4">
                        <CardContent className="p-4 space-y-2">
                            <div className="flex gap-3 text-sm font-bold text-gray-900 dark:text-white">
                                <span>{formatViews(media.view_count)} {t('watch.views')}</span>
                                <span>{formatDate(media.created_at)}</span>
                                {media.tags?.map(tag => (
                                    <span key={tag}
                                          className="text-blue-600 dark:text-blue-400 cursor-pointer hover:underline">#{tag}</span>
                                ))}
                            </div>
                            <p className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed">
                                {media.description || t('watch.noDescription')}
                            </p>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Sidebar: Recommendations */}
            <div className="lg:w-80 xl:w-96 shrink-0 space-y-4">
                <h3 className="font-bold text-lg text-gray-900 dark:text-white flex items-center gap-2 mb-4">
                    {t('watch.nextVideos')}
                </h3>

                <div className="space-y-4">
                    {recommendations.length === 0 ? (
                        <p className="text-sm text-gray-500 py-4 italic">{t('watch.noRecommendations')}</p>
                    ) : (
                        recommendations.map((item) => {
                            const recUser = item.edges?.user?.[0];
                            const recThumb = item.thumbnail
                                ? getFullUrl(item.thumbnail)
                                : 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400&h=225';

                            return (
                                <Link
                                    key={item.id}
                                    to="/watch"
                                    search={{v: String(item.id)}}
                                    className="flex gap-3 group"
                                >
                                    <div className="relative w-40 aspect-video rounded-lg overflow-hidden shrink-0">
                                        <img
                                            src={recThumb}
                                            alt={item.title}
                                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                                        />
                                        <div
                                            className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1 rounded">
                                            {formatDuration(item.duration)}
                                        </div>
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <h4 className="text-sm font-bold text-gray-900 dark:text-white line-clamp-2 leading-snug group-hover:text-blue-600 transition-colors">
                                            {item.title}
                                        </h4>
                                        <p className="text-xs text-gray-500 mt-1">{recUser?.nickname || recUser?.username || 'Unknown'}</p>
                                        <div className="flex items-center gap-2 text-xs text-gray-400">
                                            <span>{formatViews(item.view_count)} views</span>
                                            <span>·</span>
                                            <span>{formatDate(item.created_at)}</span>
                                        </div>
                                    </div>
                                </Link>
                            );
                        })
                    )}
                </div>
            </div>
        </div>
    );
};

export default WatchPage;
