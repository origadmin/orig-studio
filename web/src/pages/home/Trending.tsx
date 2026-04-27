import {useState, useEffect} from 'react';
import {Link} from '@tanstack/react-router';
import {TrendingUp, Play, Eye, Heart, Clock} from 'lucide-react';
import {exploreApi} from '../../lib/api/explore';
import type {TrendingItem} from '../../lib/api/explore';

export default function Trending() {
    const [items, setItems] = useState<TrendingItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        exploreApi.getTrending({limit: 50})
            .then((res) => {
                if (res.code === 0) {
                    setItems(res.data.items || []);
                } else {
                    setError(res.message || 'Failed to load trending');
                }
            })
            .catch((err: Error) => setError(err.message))
            .finally(() => setLoading(false));
    }, []);

    const formatCount = (n: number) => {
        if (n >= 1000000) return `${(n / 1000000).toFixed(1)}M`;
        if (n >= 1000) return `${(n / 1000).toFixed(1)}K`;
        return String(n);
    };

    const formatDuration = (seconds?: number) => {
        if (!seconds) return '--:--';
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
        return `${m}:${String(s).padStart(2, '0')}`;
    };

    if (loading) {
        return (
            <div className="max-w-7xl mx-auto px-4 py-8">
                <div className="flex items-center gap-3 mb-8">
                    <TrendingUp className="w-8 h-8 text-emerald-500"/>
                    <h1 className="text-2xl font-bold">热门内容</h1>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                    {Array.from({length: 12}).map((_, i) => (
                        <div key={i} className="animate-pulse">
                            <div className="aspect-video bg-muted dark:bg-gray-800 rounded-lg mb-3"/>
                            <div className="h-4 bg-muted dark:bg-gray-800 rounded w-3/4 mb-2"/>
                            <div className="h-3 bg-muted dark:bg-gray-800 rounded w-1/2"/>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="max-w-7xl mx-auto px-4 py-16 text-center">
                <p className="text-destructive">{error}</p>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto px-4 py-8">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-3">
                    <TrendingUp className="w-8 h-8 text-emerald-500"/>
                    <h1 className="text-2xl font-bold">热门内容</h1>
                </div>
                <span className="text-sm text-gray-500">共 {items.length} 个结果</span>
            </div>

            {items.length === 0 ? (
                <div className="text-center py-16 text-gray-500">
                    暂无热门内容
                </div>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                    {items.map((item) => (
                        <Link
                            key={item.id}
                            to="/watch"
                            search={{v: item.short_token}}
                            className="group block"
                        >
                            <div className="relative aspect-video rounded-lg overflow-hidden bg-muted dark:bg-gray-800 mb-3">
                                {item.thumbnail ? (
                                    <img
                                        src={item.thumbnail}
                                        alt={item.title}
                                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
                                        loading="lazy"
                                    />
                                ) : null}
                                {item.duration ? (
                                    <span className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-0.5 rounded">
                                        {formatDuration(item.duration)}
                                    </span>
                                ) : null}
                                <Play className="absolute inset-0 m-auto w-12 h-12 text-white opacity-0 group-hover:opacity-80 transition-opacity pointer-events-none"/>
                            </div>
                            <h3 className="font-medium text-sm line-clamp-2 group-hover:text-emerald-600 transition-colors">
                                {item.title}
                            </h3>
                            <div className="flex items-center gap-3 mt-1.5 text-xs text-gray-500">
                                <span className="flex items-center gap-1">
                                    <Eye size={12}/>
                                    {formatCount(item.view_count)}
                                </span>
                                <span className="flex items-center gap-1">
                                    <Heart size={12}/>
                                    {formatCount(item.like_count)}
                                </span>
                            </div>
                        </Link>
                    ))}
                </div>
            )}
        </div>
    );
}