/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 我的播放列表页
 */

import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {ListVideo, Plus, MoreVertical, Play, Video, Loader2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useQuery} from '@tanstack/react-query';
import {playlistApi} from '@/lib/api/playlist';
import {formatDate} from '@/lib/format';
import {getFullUrl} from '@/lib/utils';

const PlaylistsPage = () => {
    const {t} = useTranslation();
    const {user} = useAuth();

    const {data, isLoading, error} = useQuery({
        queryKey: ['playlists', user?.id],
        queryFn: async () => {
            if (!user) throw new Error('User not logged in');
            return await playlistApi.getAll(user.id);
        },
        enabled: !!user
    });

    const playlists = data || [];

    const visibilityLabel = (v: string) => {
        const map: Record<string, string> = {
            public: t('common.public'),
            private: t('common.private'),
            unlisted: t('common.unlisted')
        };
        return map[v] || v;
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error || !user) {
        return (
            <div className="text-center py-20 text-gray-400">
                <ListVideo size={48} className="mx-auto mb-3 opacity-30"/>
                <p className="text-lg mb-1">{t('playlists.empty')}</p>
                <p className="text-sm">{t('playlists.emptyDesc')}</p>
            </div>
        );
    }



    return (
        <div className="space-y-6">
            {/* 标题 + 新建按钮 */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <ListVideo size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('playlists.title')}</h1>
                    <span className="text-sm text-gray-500">{t('playlists.listCount', {count: playlists.length})}</span>
                </div>
                <Button
                    className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 transition-colors">
                    <Plus size={16}/> {t('playlists.newList')}
                </Button>
            </div>

            {/* 播放列表卡片 */}
            {playlists.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                    {playlists.map((pl) => (
                        <div
                            key={pl.id}
                            className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-100 dark:border-gray-700 hover:shadow-lg transition-all group"
                        >
                            {/* 封面 */}
                            <div className="relative aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700">
                                <div className="absolute inset-0 flex items-center justify-center">
                                    <ListVideo size={48} className="text-gray-300 dark:text-gray-600"/>
                                </div>
                                <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent"/>
                                <div className="absolute bottom-3 left-3 flex items-center gap-2">
                                    <Video size={14} className="text-white/80"/>
                                    <span
                                        className="text-white text-sm">{pl.media_ids?.length || 0} {t('common.videos_count')}</span>
                                </div>
                                <div className="absolute top-3 right-3">
                                    <span className={`text-xs px-2 py-0.5 rounded-full ${
                                        pl.is_public
                                            ? 'bg-emerald-500/80 text-white'
                                            : 'bg-gray-600/80 text-white'
                                    }`}>
                                        {visibilityLabel(pl.is_public ? 'public' : 'private')}
                                    </span>
                                </div>
                                {/* 播放全部按钮 */}
                                <div
                                    className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                    <div
                                        className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg">
                                        <Play className="w-5 h-5 text-gray-900 ml-0.5" fill="currentColor"/>
                                    </div>
                                </div>
                            </div>
                            {/* 信息 */}
                            <div className="p-4">
                                <h3 className="font-semibold text-gray-900 dark:text-white mb-1 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                    {pl.name}
                                </h3>
                                <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-1">{pl.description}</p>
                                <p className="text-xs text-gray-400 mt-2">{t('playlists.updated', {date: formatDate(pl.updated_at)})}</p>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 text-gray-400">
                    <ListVideo size={48} className="mx-auto mb-3 opacity-30"/>
                    <p className="text-lg mb-1">{t('playlists.empty')}</p>
                    <p className="text-sm">{t('playlists.emptyDesc')}</p>
                </div>
            )}
        </div>
    );
};

export default PlaylistsPage;
