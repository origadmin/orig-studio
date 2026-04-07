import React, {useState} from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {Filter, Eye, Play, Loader2} from 'lucide-react';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {useMediaList} from '@/hooks/queries';

const SearchPage = () => {
    const {t} = useTranslation();
    const location = useLocation();
    const searchParams = new URLSearchParams(location.search);
    const q = searchParams.get('q') || '';
    const categoryId = searchParams.get('category_id') || undefined;
    const [page, setPage] = useState(1);
    const pageSize = 10;

    const {data, isLoading, error} = useMediaList({
        page,
        page_size: pageSize,
        status: 'active',
        search: q,
        category_id: categoryId ? Number(categoryId) : undefined
    });

    const searchResults = data?.list || [];

    const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:9090";

    const getFullUrl = (path?: string) => {
        if (!path) return '';
        if (path.startsWith('http')) return path;
        const base = API_BASE_URL.replace(/\/$/, '');
        const sep = path.startsWith('/') ? '' : '/';
        return `${base}${sep}${path}`;
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error || searchResults.length === 0) {
        return (
            <div className="py-20 text-center space-y-4">
                <div className="text-gray-500 text-lg">{error ? t('common.loading') : t('common.noData')}</div>
                <Link to="/">
                    <button
                        className="flex items-center space-x-2 px-6 py-2.5 bg-slate-900 dark:bg-gray-800 text-white rounded-2xl text-xs font-black hover:bg-emerald-600 transition-all">
                        <span>{t('common.backToHome')}</span>
                    </button>
                </Link>
            </div>
        );
    }

    return (
        <div className="space-y-12">
            <div className="pb-8 border-b border-slate-100 dark:border-gray-700 flex justify-between items-center">
                <h1 className="text-3xl font-black text-slate-900 dark:text-white">
                    {q ? (
                        <>{t('search.resultsFor', {query: q})}</>
                    ) : t('search.title')}
                </h1>
                <button
                    className="flex items-center space-x-2 px-6 py-2.5 bg-slate-900 dark:bg-gray-800 text-white rounded-2xl text-xs font-black hover:bg-emerald-600 transition-all">
                    <Filter size={16}/><span>{t('common.filter')}</span>
                </button>
            </div>
            <div className="space-y-8">
                {searchResults.map(item => (
                    <Link key={item.id} to="/watch" search={{v: String(item.id)}}
                          className="flex flex-col md:flex-row gap-8 group p-4 rounded-3xl hover:bg-slate-50 dark:hover:bg-gray-800 transition-all">
                        <div
                            className="relative w-full md:w-80 aspect-video bg-slate-200 rounded-2xl overflow-hidden shrink-0 border border-slate-100 dark:border-gray-700 shadow-lg">
                            <img src={getFullUrl(item.thumbnail)}
                                 className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                                 alt={item.title}/>
                            <div
                                className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-1.5 py-0.5 rounded">{formatDuration(item.duration)}</div>
                            <div
                                className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                <div
                                    className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg">
                                    <Play className="w-5 h-5 text-gray-900 ml-0.5" fill="currentColor"/>
                                </div>
                            </div>
                        </div>
                        <div className="flex-1 space-y-3 min-w-0">
                            <h3 className="text-xl font-black text-slate-900 dark:text-white group-hover:text-emerald-600 transition-colors line-clamp-2 leading-tight">
                                {item.title}
                            </h3>
                            <div
                                className="flex items-center space-x-3 text-xs font-black text-slate-400 uppercase tracking-widest">
                                <span>{item.edges?.user?.[0]?.username || 'Unknown'}</span>
                                <span>·</span>
                                <span className="flex items-center gap-1"><Eye
                                    size={12}/>{formatViews(item.view_count)} {t('common.views')}</span>
                                <span>·</span>
                                <span>{formatDate(item.created_at)}</span>
                            </div>
                            <p className="text-sm font-medium text-slate-500 dark:text-gray-400 line-clamp-2 leading-relaxed">
                                {item.description || t('watch.noDescription')}
                            </p>
                        </div>
                    </Link>
                ))}
            </div>
            {data && data.total > pageSize && (
                <div className="flex justify-center pt-8">
                    <div className="flex gap-2">
                        {Array.from({length: Math.ceil(data.total / pageSize)}).map((_, i) => (
                            <button
                                key={i}
                                className={`px-4 py-2 rounded-md ${page === i + 1 ? 'bg-emerald-600 text-white' : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'}`}
                                onClick={() => setPage(i + 1)}
                            >
                                {i + 1}
                            </button>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};
export default SearchPage;
