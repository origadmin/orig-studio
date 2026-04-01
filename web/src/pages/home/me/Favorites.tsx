import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Heart, Trash2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';

const mockFavorites = [
    {
        id: 1,
        title: '从零构建 Go 微服务',
        thumbnail: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400',
        duration: 3600,
        view_count: 125400,
        create_time: '2024-03-15',
        author_name: 'Gopher 专家'
    },
    {
        id: 2,
        title: 'React 高级模式与最佳实践',
        thumbnail: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?auto=format&fit=crop&q=80&w=400',
        duration: 2400,
        view_count: 89400,
        create_time: '2024-03-14',
        author_name: 'React 大师'
    },
    {
        id: 3,
        title: 'Docker & Kubernetes 深度解析',
        thumbnail: 'https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&q=80&w=400',
        duration: 5400,
        view_count: 234500,
        create_time: '2024-03-12',
        author_name: '运维专家'
    },
    {
        id: 4,
        title: 'TypeScript 高级类型大师课',
        thumbnail: 'https://images.unsplash.com/photo-1516116216624-53e697fedbea?auto=format&fit=crop&q=80&w=400',
        duration: 1800,
        view_count: 67800,
        create_time: '2024-03-10',
        author_name: 'TS 达人'
    },
];

const FavoritesPage = () => {
    const {t} = useTranslation();
    const [favorites] = useState(mockFavorites);

    return (
        <div className="max-w-6xl mx-auto space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <Heart className="w-6 h-6 text-rose-500 fill-current"/>{t('favorites.title')}
                </h1>
                <p className="text-slate-500 dark:text-gray-400 text-sm mt-1">{t('favorites.savedCount', {count: favorites.length})}</p>
            </div>

            {favorites.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                    {favorites.map(video => (
                        <Link key={video.id} to="/v/$id" params={{id: String(video.id)}} className="group">
                            <div
                                className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                <div className="relative aspect-video">
                                    <img src={video.thumbnail} alt={video.title}
                                         className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                    <div
                                        className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                    <div className="absolute top-2 right-2">
                                        <Button variant="ghost" size="sm"
                                                className="h-8 w-8 p-0 hover:bg-white/80"><Trash2
                                            className="w-4 h-4 text-slate-500 hover:text-rose-500"/></Button>
                                    </div>
                                </div>
                                <div className="p-3">
                                    <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                    <div className="flex items-center gap-2 mt-2"><span
                                        className="text-xs text-slate-500 dark:text-gray-400">{video.author_name}</span>
                                    </div>
                                    <p className="text-xs text-slate-400 dark:text-gray-500 mt-1">{formatViews(video.view_count)} {t('common.views')}
                                        · {formatDate(video.create_time)}</p>
                                </div>
                            </div>
                        </Link>
                    ))}
                </div>
            ) : (
                <div className="text-center py-20">
                    <Heart className="w-16 h-16 text-slate-200 dark:text-gray-700 mx-auto mb-4"/>
                    <p className="text-slate-500 dark:text-gray-400">{t('favorites.empty')}</p>
                    <p className="text-sm text-slate-400 dark:text-gray-500 mt-1">{t('favorites.emptyDesc')}</p>
                </div>
            )}
        </div>
    );
};
export default FavoritesPage;
