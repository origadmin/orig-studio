/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 首页 - 信息流 + 无限滚动
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Clock, TrendingUp, Plus} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {formatDuration, formatViews, formatDate} from '@/lib/format';

/* ── Mock 数据生成 ─────────────────────────────────────────────────────── */

const titles = [
    'Go 微服务架构实战', 'React 18 新特性详解', 'Kubernetes 入门到精通',
    'TypeScript 高级类型编程', 'Docker 容器化部署', 'Python 数据分析',
    'AWS 云服务实践', 'Vue 3 组合式 API', 'Redis 缓存策略',
    'GraphQL API 设计', 'Nginx 高性能配置', 'CI/CD 流水线搭建',
    'Rust 系统编程入门', 'MongoDB 文档数据库', 'Linux 运维指南',
    'Electron 桌面应用开发', 'WebAssembly 前沿探索', '微前端架构实践',
    'Prometheus 监控体系', 'gRPC 服务通信',
];
const authors = ['Gopher 专家', 'React 大师', '运维专家', 'TS 达人', '数据科学家', '云架构师', 'Vue 大师', '全栈工程师'];
const categories = ['技术', '编程', '运维', '数据科学', '云计算', '前端', '职业'];
const tagPool = ['Go', 'React', 'Docker', 'K8s', 'TypeScript', 'Python', 'AWS', 'Vue', 'Redis', 'GraphQL'];

const generateMockData = (startId: number, count: number) =>
    Array.from({length: count}, (_, i) => {
        const idx = (startId + i - 1) % titles.length;
        const daysAgo = startId + i - 1;
        const d = new Date();
        d.setDate(d.getDate() - daysAgo);
        return {
            id: startId + i,
            title: titles[idx],
            description: `深入了解 ${titles[idx]} 的核心概念和最佳实践。`,
            thumbnail: `https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400&h=225&sig=${startId + i}`,
            duration: 1800 + Math.floor(Math.random() * 5400),
            view_count: Math.floor(Math.random() * 500000) + 1000,
            create_time: d.toISOString().split('T')[0],
            user_id: (idx % 8) + 1,
            author_name: authors[idx % authors.length],
            author_avatar: `https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=100&sig=${idx}`,
            category: categories[idx % categories.length],
            tags: [tagPool[idx % tagPool.length], tagPool[(idx + 3) % tagPool.length]],
        };
    });

const PAGE_SIZE = 12;

/* ── Component ─────────────────────────────────────────────────────────── */

const HomePage = () => {
    const [items, setItems] = useState(() => generateMockData(0, PAGE_SIZE));
    const [activeCategory, setActiveCategory] = useState('全部');
    const [loading, setLoading] = useState(false);
    const [hasMore, setHasMore] = useState(true);
    const sentinelRef = useRef<HTMLDivElement>(null);

    const allCategories = ['全部', ...categories];

    const filteredItems = activeCategory === '全部'
        ? items
        : items.filter(m => m.category === activeCategory);

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
        const obs = new IntersectionObserver(([e]) => {
            if (e.isIntersecting) loadMore();
        }, {rootMargin: '200px'});
        obs.observe(el);
        return () => obs.disconnect();
    }, [loadMore]);

    return (
        <div className="space-y-8">
            {/* Hero Banner */}
            <section
                className="relative rounded-2xl overflow-hidden bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 text-white">
                <div
                    className="absolute inset-0 bg-[url('https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&q=80')] bg-cover bg-center opacity-20"/>
                <div className="relative px-6 py-8 flex items-center">
                    <div className="max-w-xl">
                        <Badge className="bg-emerald-500/20 text-emerald-300 hover:bg-emerald-500/30 mb-4">
                            <TrendingUp className="w-3 h-3 mr-1"/> 精选推荐
                        </Badge>
                        <h1 className="text-4xl font-black mb-4 leading-tight">发现优质视频内容</h1>
                        <p className="text-lg text-slate-300 mb-6">浏览来自全球创作者的原创视频，探索感兴趣的领域。</p>
                        <div className="flex gap-4">
                            <Link to="/featured">
                                <Button size="lg" className="bg-emerald-600 hover:bg-emerald-700">探索内容</Button>
                            </Link>
                            <Link to="/me/upload">
                                <Button size="lg" variant="outline"
                                        className="border-slate-600 text-white hover:bg-slate-800">
                                    开始创作
                                </Button>
                            </Link>
                        </div>
                    </div>
                </div>
            </section>

            {/* 分类标签 */}
            <section className="flex flex-wrap gap-2">
                {allCategories.map(cat => (
                    <Button
                        key={cat}
                        variant={activeCategory === cat ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setActiveCategory(cat)}
                        className={activeCategory === cat ? 'bg-emerald-600 hover:bg-emerald-700' : ''}
                    >{cat}</Button>
                ))}
            </section>

            {/* 列表标题 */}
            <div className="flex items-center justify-between">
                <h2 className="text-2xl font-bold text-slate-900 dark:text-white">
                    {activeCategory === '全部' ? '最新视频' : `${activeCategory}视频`}
                </h2>
                <Link to="/latest"
                      className="text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 font-medium">
                    查看全部 →
                </Link>
            </div>

            {/* 视频网格 */}
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                {filteredItems.map(media => (
                    <Link key={media.id} to="/v/$id" params={{id: String(media.id)}} className="group">
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
                                <div className="flex flex-wrap gap-1 mt-2">
                                    {media.tags?.slice(0, 2).map(tag => (
                                        <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </Link>
                ))}
            </div>

            {/* 无限滚动哨兵 */}
            <div ref={sentinelRef} className="flex flex-col items-center py-8">
                {loading && (
                    <div className="flex items-center gap-3 text-gray-400">
                        <div
                            className="animate-spin w-5 h-5 border-2 border-emerald-600 border-t-transparent rounded-full"/>
                        <span className="text-sm">加载中...</span>
                    </div>
                )}
                {!hasMore && items.length > 0 && (
                    <p className="text-sm text-gray-400 py-4">— 已加载全部内容 —</p>
                )}
            </div>
        </div>
    );
};

export default HomePage;
