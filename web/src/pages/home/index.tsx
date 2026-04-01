/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 首页 - 信息流 + 无限滚动 (对接真实 API)
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, TrendingUp} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {mediaApi, type Media} from '@/lib/api/media';

const categories = [
    {id: 1, name: '技术'},
    {id: 2, name: '编程'},
    {id: 3, name: '运维'},
    {id: 4, name: '数据科学'},
    {id: 5, name: '云计算'},
    {id: 6, name: '前端'},
    {id: 7, name: '职业'}
];

const PAGE_SIZE = 12;

const HomePage = () => {
    const {t} = useTranslation();
    const [items, setItems] = useState<Media[]>([]);
    const [activeCategoryId, setActiveCategoryId] = useState<number | null>(null);
    const [loading, setLoading] = useState(false);
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);
    const sentinelRef = useRef<HTMLDivElement>(null);

    // 获取数据的方法
    const fetchMedia = useCallback(async (pageNum: number, categoryId: number | null, append: boolean = true) => {
        if (loading) return;
        setLoading(true);
        try {
            const res = await mediaApi.list({
                page: pageNum,
                page_size: PAGE_SIZE,
                category_id: categoryId || undefined,
                state: 'active'
            });

            const newItems = res.list || [];
            setItems(prev => append ? [...prev, ...newItems] : newItems);
            setHasMore(newItems.length === PAGE_SIZE);
            setPage(pageNum);
        } catch (err) {
            console.error("Failed to fetch media:", err);
        } finally {
            setLoading(false);
        }
    }, [loading]);

    // 初始加载
    useEffect(() => {
        fetchMedia(1, activeCategoryId, false);
    }, [activeCategoryId]);

    // 加载更多
    const loadMore = useCallback(() => {
        if (loading || !hasMore) return;
        fetchMedia(page + 1, activeCategoryId, true);
    }, [loading, hasMore, page, activeCategoryId, fetchMedia]);

    // 滚动监听
    useEffect(() => {
        const el = sentinelRef.current;
        if (!el) return;
        const obs = new IntersectionObserver(([e]) => {
            if (e.isIntersecting) loadMore();
        }, {rootMargin: '200px'});
        obs.observe(el);
        return () => obs.disconnect();
    }, [loadMore]);

    const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:9090";

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

            {/* 分类标签 */}
            <section className="flex flex-wrap gap-2">
                <Button
                    variant={activeCategoryId === null ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setActiveCategoryId(null)}
                    className={activeCategoryId === null ? 'bg-emerald-600 hover:bg-emerald-700' : ''}
                >{t('home.all')}</Button>
                {categories.map((cat) => (
                    <Button
                        key={cat.id}
                        variant={activeCategoryId === cat.id ? 'default' : 'outline'}
                        size="sm"
                        onClick={() => setActiveCategoryId(cat.id)}
                        className={activeCategoryId === cat.id ? 'bg-emerald-600 hover:bg-emerald-700' : ''}
                    >{cat.name}</Button>
                ))}
            </section>

            {/* 列表标题 */}
            <div className="flex items-center justify-between">
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

            {/* 视频网格 */}
            {items.length === 0 && !loading ? (
                <div className="py-20 text-center text-gray-500">
                    <p>{t('common.noData')}</p>
                </div>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                    {items.map(media => {
                        const user = media.edges?.user?.[0];
                        // 处理缩略图路径，如果不是绝对路径则拼接 BaseURL
                        const thumbUrl = media.thumbnail
                            ? (media.thumbnail.startsWith('http') ? media.thumbnail : `${API_BASE_URL}${media.thumbnail.startsWith('/') ? '' : '/'}${media.thumbnail}`)
                            : 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400&h=225';

                        return (
                            <Link key={media.id} to="/watch" search={{v: String(media.id)}} className="group">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
                                    <div className="relative aspect-video overflow-hidden">
                                        <img src={thumbUrl} alt={media.title}
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
                                                src={user?.avatar || `https://ui-avatars.com/api/?name=${user?.username || 'U'}`}
                                                alt={user?.username}
                                                className="w-5 h-5 rounded-full object-cover"/>
                                            <span
                                                className="text-xs text-gray-500 dark:text-gray-400">{user?.nickname || user?.username || 'Unknown'}</span>
                                        </div>
                                        <div
                                            className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                            <span className="flex items-center gap-1"><Eye
                                                size={12}/>{formatViews(media.view_count)}</span>
                                            <span>{formatDate(media.created_at)}</span>
                                        </div>
                                        <div className="flex flex-wrap gap-1 mt-2">
                                            {media.tags?.slice(0, 2).map((tag, tIdx) => (
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

            {/* 无限滚动哨兵 */}
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

export default HomePage;
