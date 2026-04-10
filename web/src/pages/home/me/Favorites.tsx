import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Heart, Trash2, Loader2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useQuery, useMutation, useQueryClient} from '@tanstack/react-query';
import {favoriteApi} from '@/lib/api/favorite';
import {getFullUrl} from '@/lib/utils';

const FavoritesPage = () => {
    const {t} = useTranslation();
    const {user} = useAuth();
    const queryClient = useQueryClient();

    const {data, isLoading, error} = useQuery({
        queryKey: ['favorites', user?.id],
        queryFn: async () => {
            if (!user) throw new Error('User not logged in');
            return await favoriteApi.list();
        },
        enabled: !!user
    });

    const deleteMutation = useMutation({
        mutationFn: (mediaId: number) => favoriteApi.remove({media_id: mediaId?.toString() || ''}),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['favorites', user?.id]});
        }
    });

    const favorites = data?.list || [];

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error || !user) {
        return (
            <div className="text-center py-20">
                <Heart className="w-16 h-16 text-slate-200 dark:text-gray-700 mx-auto mb-4"/>
                <p className="text-slate-500 dark:text-gray-400">{t('favorites.empty')}</p>
                <p className="text-sm text-slate-400 dark:text-gray-500 mt-1">{t('favorites.emptyDesc')}</p>
            </div>
        );
    }



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
                    {favorites.map(favorite => {
                        const video = favorite.media;
                        return (
                            <Link key={video.id} to="/watch" search={{v: String(video.id)}} className="group">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                    <div className="relative aspect-video">
                                        <img src={video.thumbnail ? getFullUrl(video.thumbnail) : undefined}
                                             alt={video.title}
                                             className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                        <div className="absolute top-2 right-2">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-8 w-8 p-0 hover:bg-white/80"
                                                onClick={(e) => {
                                                    e.preventDefault();
                                                    deleteMutation.mutate(video.id);
                                                }}
                                            >
                                                <Trash2 className="w-4 h-4 text-slate-500 hover:text-rose-500"/>
                                            </Button>
                                        </div>
                                    </div>
                                    <div className="p-3">
                                        <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                        <div className="flex items-center gap-2 mt-2">
                                            <span className="text-xs text-slate-500 dark:text-gray-400">
                                                {video.edges?.user?.[0]?.username || 'Unknown'}
                                            </span>
                                        </div>
                                        <p className="text-xs text-slate-400 dark:text-gray-500 mt-1">
                                            {formatViews(video.view_count)} {t('common.views')} · {formatDate(video.created_at)}
                                        </p>
                                    </div>
                                </div>
                            </Link>
                        );
                    })}
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
