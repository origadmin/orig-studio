/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 历史记录页
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {History, Play, Eye, Trash2, X} from 'lucide-react';
import {useTranslation} from 'react-i18next';

interface HistoryItem {
    id: number;
    title: string;
    thumbnail: string;
    duration: number;
    view_count: number;
    author_name: string;
    watchedAt: string;
    progress: number; // 观看进度 0-100
}

const mockHistory: HistoryItem[] = Array.from({length: 20}, (_, i) => ({
    id: 200 + i,
    title: `已观看的视频 ${i + 1} - ${['Go 并发', 'React Hooks', 'Docker 部署', 'K8s 编排', 'TypeScript 类型', 'Python ML'][i % 6]}`,
    thumbnail: `https://images.unsplash.com/photo-${1517694712202 + (i % 3)}141592653?auto=format&fit=crop&q=80&w=400&h=225`,
    duration: 1800 + Math.floor(Math.random() * 5400),
    view_count: Math.floor(Math.random() * 100000) + 1000,
    author_name: ['Gopher Expert', 'React Master', 'DevOps Pro', 'TS Guru'][i % 4],
    watchedAt: new Date(Date.now() - i * 3600000 * Math.random() * 48).toISOString(),
    progress: Math.floor(Math.random() * 100),
}));

const formatDuration = (s: number) => {
    const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), sec = s % 60;
    return h > 0 ? `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}` : `${m}:${String(sec).padStart(2, '0')}`;
};

const formatTimeAgo = (iso: string) => {
    const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 60000);
    if (diff < 60) return `${diff} 分钟前`;
    if (diff < 1440) return `${Math.floor(diff / 60)} 小时前`;
    return `${Math.floor(diff / 1440)} 天前`;
};

const HistoryPage = () => {
    const {t} = useTranslation();
    const [items, setItems] = useState(mockHistory);
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(false);
    const sentinelRef = useRef<HTMLDivElement>(null);

    const clearHistory = () => {
        setItems([]);
    };

    const loadMore = useCallback(() => {
        if (loading || !hasMore) return;
        setLoading(true);
        setTimeout(() => setLoading(false), 500);
    }, [loading, hasMore]);

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

    if (items.length === 0) {
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
                <button
                    onClick={clearHistory}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                >
                    <Trash2 size={14}/> {t('history.clear')}
                </button>
            </div>

            {/* 历史列表 */}
            <div className="space-y-2">
                {items.map((item) => (
                    <Link
                        key={item.id}
                        to="/watch" search={{v: String(item.id)}}
                        className="group flex items-center gap-4 p-3 bg-white dark:bg-gray-800 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
                    >
                        {/* 缩略图 */}
                        <div className="relative w-40 shrink-0 aspect-video rounded-lg overflow-hidden">
                            <img src={item.thumbnail} alt={item.title}
                                 className="w-full h-full object-cover"/>
                            <div
                                className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1 py-0.5 rounded">
                                {formatDuration(item.duration)}
                            </div>
                            {/* 进度条 */}
                            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-gray-600/30">
                                <div
                                    className="h-full bg-emerald-500"
                                    style={{width: `${item.progress}%`}}
                                />
                            </div>
                        </div>

                        {/* 信息 */}
                        <div className="flex-1 min-w-0">
                            <h3 className="text-sm font-medium text-gray-900 dark:text-white line-clamp-1 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                {item.title}
                            </h3>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{item.author_name}</p>
                            <p className="text-xs text-gray-400 mt-0.5">{formatTimeAgo(item.watchedAt)}</p>
                        </div>

                        {/* 移除按钮 */}
                        <button
                            onClick={(e) => {
                                e.preventDefault();
                                setItems(prev => prev.filter(h => h.id !== item.id));
                            }}
                            className="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg opacity-0 group-hover:opacity-100 transition-all"
                        >
                            <X size={14}/>
                        </button>
                    </Link>
                ))}
            </div>

            {/* 无限滚动哨兵 */}
            <div ref={sentinelRef} className="flex items-center justify-center py-6">
                {loading && (
                    <div
                        className="animate-spin w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full"/>
                )}
            </div>
        </div>
    );
};

export default HistoryPage;
