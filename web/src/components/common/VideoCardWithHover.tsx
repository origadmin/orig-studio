/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * VideoCardWithHover - Video card with hover action overlay
 * Used in subscription content page and similar video grid layouts
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Clock, Eye, Share2, ListPlus} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import type {SubscriptionVideo} from '@/lib/api/subscriptionVideos';

interface VideoCardWithHoverProps {
    video: SubscriptionVideo;
}

const VideoCardWithHover: React.FC<VideoCardWithHoverProps> = ({video}) => {
    const {t} = useTranslation();

    return (
        <div
            className="group cursor-pointer rounded-xl bg-white dark:bg-gray-800 overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5 border border-gray-100 dark:border-gray-700">
            {/* Thumbnail area */}
            <Link to="/watch" search={{v: video.short_token || String(video.id)}}
                  className="block relative aspect-video overflow-hidden">
                <img
                    src={getImageUrl(video.thumbnail, 'thumbnail')}
                    alt={video.title}
                    loading="lazy"
                    onError={(e) => handleImageError(e, 'thumbnail')}
                    className="object-cover w-full h-full group-hover:scale-105 transition-transform duration-500"
                />

                {/* Duration badge */}
                <div
                    className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded">
                    {formatDuration(video.duration)}
                </div>

                {/* Play overlay on hover */}
                <div
                    className="absolute inset-0 bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center">
                    <div
                        className="w-12 h-12 bg-white/90 dark:bg-gray-900/90 rounded-full flex items-center justify-center shadow-lg transform scale-75 group-hover:scale-100 transition-transform duration-300">
                        <Play className="w-5 h-5 text-gray-900 dark:text-white ml-0.5" fill="currentColor"/>
                    </div>
                </div>

                {/* Quick action buttons on hover */}
                <div
                    className="absolute top-2 right-2 flex flex-col gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity duration-300">
                    <Button
                        variant="secondary"
                        size="icon"
                        className="h-8 w-8 rounded-full bg-white/90 dark:bg-gray-800/90 hover:bg-white dark:hover:bg-gray-800 shadow-md"
                        title={t('subscriptions.play')}
                        onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                        }}
                    >
                        <Play className="h-4 w-4"/>
                    </Button>
                    <Button
                        variant="secondary"
                        size="icon"
                        className="h-8 w-8 rounded-full bg-white/90 dark:bg-gray-800/90 hover:bg-white dark:hover:bg-gray-800 shadow-md"
                        title={t('subscriptions.watchLater')}
                        onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                        }}
                    >
                        <ListPlus className="h-4 w-4"/>
                    </Button>
                    <Button
                        variant="secondary"
                        size="icon"
                        className="h-8 w-8 rounded-full bg-white/90 dark:bg-gray-800/90 hover:bg-white dark:hover:bg-gray-800 shadow-md"
                        title={t('subscriptions.share')}
                        onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                        }}
                    >
                        <Share2 className="h-4 w-4"/>
                    </Button>
                </div>
            </Link>

            {/* Info area */}
            <div className="p-3">
                <h3 className="font-medium text-gray-900 dark:text-white text-sm line-clamp-2 mb-2 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                    <Link to="/watch" search={{v: video.short_token || String(video.id)}}>{video.title}</Link>
                </h3>

                <div className="flex items-center gap-2 mb-1.5">
                    <Avatar className="h-5 w-5">
                        <AvatarImage
                            src={getImageUrl(video.channel_avatar || video.user_avatar, 'avatar')}
                            alt={video.channel_name || video.username}
                        />
                        <AvatarFallback className="text-[10px]">
                            {(video.channel_name || video.username)?.[0] || 'U'}
                        </AvatarFallback>
                    </Avatar>
                    <span
                        className="text-xs text-gray-500 dark:text-muted-foreground truncate">{video.channel_name || video.username}</span>
                </div>

                <div className="flex items-center gap-3 text-xs text-muted-foreground dark:text-gray-500">
                    <span className="flex items-center gap-1">
                        <Eye size={12}/>
                        {formatViews(video.view_count)}
                    </span>
                    <span className="flex items-center gap-1">
                        <Clock size={12}/>
                        {formatDate(video.published_at || video.created_at)}
                    </span>
                </div>
            </div>
        </div>
    );
};

export default VideoCardWithHover;
