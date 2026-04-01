/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 我的播放列表页
 */

import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {ListVideo, Plus, MoreVertical, Play, Video} from 'lucide-react';

interface PlaylistInfo {
    id: number;
    title: string;
    description: string;
    cover: string;
    count: number;
    updatedAt: string;
    visibility: 'public' | 'private' | 'unlisted';
}

const mockPlaylists: PlaylistInfo[] = [
    {
        id: 1, title: 'Go 学习路线', description: '从入门到精通的 Go 语言学习资源',
        cover: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400&h=225',
        count: 8, updatedAt: '2024-03-15', visibility: 'public',
    },
    {
        id: 2, title: 'DevOps 工具链', description: 'Docker, K8s, CI/CD 相关视频合集',
        cover: 'https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&q=80&w=400&h=225',
        count: 12, updatedAt: '2024-03-12', visibility: 'public',
    },
    {
        id: 3, title: '周末充电', description: '周末想看的视频收藏',
        cover: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?auto=format&fit=crop&q=80&w=400&h=225',
        count: 5, updatedAt: '2024-03-10', visibility: 'private',
    },
];

const visibilityLabel = (v: string) => {
    const map: Record<string, string> = {public: '公开', private: '私密', unlisted: '不公开'};
    return map[v] || v;
};

const PlaylistsPage = () => {
    const [playlists] = useState(mockPlaylists);

    return (
        <div className="space-y-6">
            {/* 标题 + 新建按钮 */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <ListVideo size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">我的播放列表</h1>
                    <span className="text-sm text-gray-500">{playlists.length} 个列表</span>
                </div>
                <button
                    className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 transition-colors">
                    <Plus size={16}/> 新建列表
                </button>
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
                            <div className="relative aspect-video overflow-hidden">
                                <img src={pl.cover} alt={pl.title}
                                     className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"/>
                                <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent"/>
                                <div className="absolute bottom-3 left-3 flex items-center gap-2">
                                    <Video size={14} className="text-white/80"/>
                                    <span className="text-white text-sm">{pl.count} 个视频</span>
                                </div>
                                <div className="absolute top-3 right-3">
                                    <span className={`text-xs px-2 py-0.5 rounded-full ${
                                        pl.visibility === 'public'
                                            ? 'bg-emerald-500/80 text-white'
                                            : pl.visibility === 'private'
                                                ? 'bg-gray-600/80 text-white'
                                                : 'bg-yellow-500/80 text-white'
                                    }`}>
                                        {visibilityLabel(pl.visibility)}
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
                                    {pl.title}
                                </h3>
                                <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-1">{pl.description}</p>
                                <p className="text-xs text-gray-400 mt-2">更新于 {pl.updatedAt}</p>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 text-gray-400">
                    <ListVideo size={48} className="mx-auto mb-3 opacity-30"/>
                    <p className="text-lg mb-1">还没有播放列表</p>
                    <p className="text-sm">创建你的第一个播放列表，收藏喜欢的视频</p>
                </div>
            )}
        </div>
    );
};

export default PlaylistsPage;
