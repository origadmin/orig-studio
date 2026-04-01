import React from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {Filter} from 'lucide-react';

const SearchPage = () => {
    const location = useLocation();
    const searchParams = new URLSearchParams(location.search);
    const q = searchParams.get('q') || '';

    return (
        <div className="space-y-12">
            <div className="pb-8 border-b border-slate-100 dark:border-gray-700 flex justify-between items-center">
                <h1 className="text-3xl font-black text-slate-900 dark:text-white">
                    {q ? (
                        <>搜索 <span className="text-emerald-600">"{q}"</span> 的结果</>
                    ) : '搜索'}
                </h1>
                <button
                    className="flex items-center space-x-2 px-6 py-2.5 bg-slate-900 dark:bg-gray-800 text-white rounded-2xl text-xs font-black hover:bg-emerald-600 transition-all">
                    <Filter size={16}/><span>筛选</span>
                </button>
            </div>
            <div className="space-y-8">
                {[1, 2, 3].map(i => (
                    <Link key={i} to="/v/$id" params={{id: String(i)}}
                          className="flex flex-col md:flex-row gap-8 group p-4 rounded-3xl hover:bg-slate-50 dark:hover:bg-gray-800 transition-all">
                        <div
                            className="relative w-full md:w-80 aspect-video bg-slate-200 rounded-2xl overflow-hidden shrink-0 border border-slate-100 dark:border-gray-700 shadow-lg">
                            <img src={`https://picsum.photos/seed/${i + 500}/600/400`}
                                 className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                                 alt="搜索结果"/>
                        </div>
                        <div className="flex-1 space-y-3 min-w-0">
                            <h3 className="text-xl font-black text-slate-900 dark:text-white group-hover:text-emerald-600 transition-colors line-clamp-2 leading-tight">
                                Go 与 React 构建现代平台
                            </h3>
                            <div
                                className="flex items-center space-x-3 text-xs font-black text-slate-400 uppercase tracking-widest">
                                <span>12.4万次观看</span><span>•</span><span>2 周前</span>
                            </div>
                            <p className="text-sm font-medium text-slate-500 dark:text-gray-400 line-clamp-2 leading-relaxed">
                                深入解析我们最新 Go 框架背后的架构设计决策...
                            </p>
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    );
};
export default SearchPage;
