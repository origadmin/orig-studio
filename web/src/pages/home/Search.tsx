import React from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {Filter, Eye, Play} from 'lucide-react';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';

const searchResults = [
    {
        id: 1,
        title: 'Go 与 React 构建现代平台',
        thumbnail: 'https://picsum.photos/seed/501/600/400',
        duration: 3600,
        view_count: 124000,
        create_time: '2024-03-10',
        description: '深入解析我们最新 Go 框架背后的架构设计决策，从前端到后端的完整技术栈方案。',
    },
    {
        id: 2,
        title: 'Kubernetes 集群管理实战',
        thumbnail: 'https://picsum.photos/seed/502/600/400',
        duration: 5400,
        view_count: 89000,
        create_time: '2024-03-08',
        description: '从零搭建生产级 K8s 集群，涵盖监控、日志、自动扩缩容等核心话题。',
    },
    {
        id: 3,
        title: 'TypeScript 高级类型编程',
        thumbnail: 'https://picsum.photos/seed/503/600/400',
        duration: 2400,
        view_count: 67500,
        create_time: '2024-03-05',
        description: '掌握条件类型、映射类型、模板字面量类型等高级技巧，写出类型安全的代码。',
    },
];

const SearchPage = () => {
    const {t} = useTranslation();
    const location = useLocation();
    const searchParams = new URLSearchParams(location.search);
    const q = searchParams.get('q') || '';

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
                    <Link key={item.id} to="/v/$id" params={{id: String(item.id)}}
                          className="flex flex-col md:flex-row gap-8 group p-4 rounded-3xl hover:bg-slate-50 dark:hover:bg-gray-800 transition-all">
                        <div
                            className="relative w-full md:w-80 aspect-video bg-slate-200 rounded-2xl overflow-hidden shrink-0 border border-slate-100 dark:border-gray-700 shadow-lg">
                            <img src={item.thumbnail}
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
                                <span className="flex items-center gap-1"><Eye
                                    size={12}/>{formatViews(item.view_count)} {t('common.views')}</span>
                                <span>·</span>
                                <span>{formatDate(item.create_time)}</span>
                            </div>
                            <p className="text-sm font-medium text-slate-500 dark:text-gray-400 line-clamp-2 leading-relaxed">
                                {item.description}
                            </p>
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    );
};
export default SearchPage;
