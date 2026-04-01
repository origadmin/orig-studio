/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Latest Page — infinite scroll
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {Clock, Play, Eye} from 'lucide-react';
import {MediaItem} from '@/types/media';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';

// TODO: replace with API call
function generateMockData(startId: number, count: number): MediaItem[] {
    const titles = [
        'Go 微服务架构实战', 'React 18 新特性详解', 'Kubernetes 入门到精通',
        'TypeScript 高级类型编程', 'Docker 容器化部署', 'Python 数据分析',
        'AWS 云服务实践', 'Vue 3 组合式 API', 'Redis 缓存策略',
        'GraphQL API 设计', 'Nginx 高性能配置', 'CI/CD 流水线搭建',
        'Rust 系统编程入门', 'MongoDB 文档数据库', 'Linux 运维指南',
        'Electron 桌面应用开发', 'WebAssembly 前沿探索', '微前端架构实践',
        'Prometheus 监控体系', 'gRPC 服务通信',
    ];
    const authors = ['Gopher Expert', 'React Master', 'DevOps Pro', 'TS Guru', 'Data Scientist', 'Cloud Expert', 'Vue Master', 'Full Stack'];
    const tagPool = ['Go', 'React', 'Docker', 'K8s', 'TypeScript', 'Python', 'AWS', 'Vue', 'Redis', 'GraphQL'];

    return Array.from({length: count}, (_, i) => {
        const idx = (startId + i - 1) % titles.length;
        const daysAgo = startId + i - 1;
        const date = new Date();
        date.setDate(date.getDate() - daysAgo);
        return {
            id: startId + i,
            title: titles[idx],
            description: `深入了解 ${titles[idx]} 的核心概念和最佳实践。`,
            thumbnail: `https://images.unsplash.com/photo-${1517694712202 + idx % 3}141592653?auto=format&fit=crop&q=80&w=400&h=225`,
            duration: 1800 + Math.floor(Math.random() * 5400),
            view_count: Math.floor(Math.random() * 500000) + 1000,
            create_time: date.toISOString().split('T')[0],
            user_id: (idx % 8) + 1,
            author_name: authors[idx % authors.length],
            author_avatar: `https://images.unsplash.com/photo-${1535713875002 + (idx % 5) * 7}0000?auto=format&fit=crop&q=80&w=100`,
            tags: [tagPool[idx % tagPool.length], tagPool[(idx + 3) % tagPool.length]],
        };
    });
}

const PAGE_SIZE = 12;

const LatestPage = () => {
    const {t} = useTranslation();
    const [items, setItems] = useState<MediaItem[]>(() => generateMockData(0, PAGE_SIZE));
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const sentinelRef = useRef<HTMLDivElement>(null);

    const loadMore = useCallback(() => {
        if (loading || !hasMore) return;
        setLoading(true);
        setTimeout(() => {
            const newItems = generateMockData(items.length, PAGE_SIZE);
            setItems(prev => [...prev, ...newItems]);
            setLoading(false);
            if (items.length + PAGE_SIZE >= 60) setHasMore(false);
        }, 600);
    }, [loading, hasMore, items.length]);

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

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Clock size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('latest.title')}</h1>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                {items.map((media) => (
                    <Link key={media.id} to="/watch" search={{v: String(media.id)}} className="group">
                        <div
                            className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                            <div className="relative aspect-video overflow-hidden">
                                <img src={media.thumbnail} alt={media.title}
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
                                    <img src={media.author_avatar} alt={media.author_name}
                                         className="w-5 h-5 rounded-full object-cover"/>
                                    <span
                                        className="text-xs text-gray-500 dark:text-gray-400">{media.author_name}</span>
                                </div>
                                <div className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                    <span className="flex items-center gap-1"><Eye
                                        size={12}/>{formatViews(media.view_count)}</span>
                                    <span>{formatDate(media.create_time)}</span>
                                </div>
                            </div>
                        </div>
                    </Link>
                ))}
            </div>

            <div ref={sentinelRef} className="flex flex-col items-center py-8">
                {loading && (
                    <div className="flex items-center gap-3 text-gray-400">
                        <div
                            className="animate-spin w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full"/>
                        <span className="text-sm">{t('common.loading')}</span>
                    </div>
                )}
                {!hasMore && items.length > 0 && (
                    <p className="text-sm text-gray-400 py-4">— {t('common.allLoaded')} —</p>
                )}
            </div>
        </div>
    );
};

export default LatestPage;
