/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Categories Page
 */

import React, {useState, useEffect} from 'react';
import {Link} from '@tanstack/react-router';
import {Folder, Play, Eye, Loader2} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {categoryApi} from '@/lib/api/category';
import {useMediaList} from '@/hooks/queries';
import {getFullUrl} from '@/lib/utils';

const CategoriesPage = () => {
    const {t} = useTranslation();
    const [categories, setCategories] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // 获取分类列表
    useEffect(() => {
        const fetchCategories = async () => {
            try {
                setLoading(true);
                const response = await categoryApi.getAll();
                setCategories(response || []);
            } catch (err) {
                setError(t('common.error'));
                console.error('Failed to fetch categories:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchCategories();
    }, [t]);



    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error || categories.length === 0) {
        return (
            <div className="text-center py-16 text-gray-400">
                <Folder size={48} className="mx-auto mb-3 opacity-30"/>
                <p>{error || t('categories.noCategories')}</p>
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {/* Header */}
            <div className="flex items-center gap-3">
                <Folder size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('categories.title')}</h1>
            </div>

            {/* Categories */}
            {categories.map((category) => (
                <div key={category.id} className="space-y-4">
                    <h2 className="text-xl font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                        <Folder size={18} className="text-emerald-500"/>
                        {category.name}
                    </h2>

                    {/* Category videos */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
                        {category.media?.slice(0, 4).map((media: any) => (
                            <Link key={media.id} to="/watch" search={{v: media.friendly_token || String(media.id)}} className="group">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5"
                                >
                                    <div className="relative aspect-video overflow-hidden">
                                        <img
                                            src={media.thumbnail ? getFullUrl(media.thumbnail) : undefined}
                                            alt={media.title}
                                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                                        />
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/80 text-white text-xs font-medium px-1.5 py-0.5 rounded"
                                        >
                                            {formatDuration(media.duration)}
                                        </div>
                                        <div
                                            className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                                        >
                                            <div
                                                className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg"
                                            >
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
                                                src={media.edges?.user?.[0]?.avatar ? getFullUrl(media.edges.user[0].avatar) : undefined}
                                                alt={media.edges?.user?.[0]?.username}
                                                className="w-5 h-5 rounded-full object-cover"
                                            />
                                            <span className="text-xs text-gray-500 dark:text-gray-400">
                                                {media.edges?.user?.[0]?.username || 'Unknown'}
                                            </span>
                                        </div>
                                        <div
                                            className="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
                                            <span className="flex items-center gap-1">
                                                <Eye size={12}/>
                                                {formatViews(media.view_count)}
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </Link>
                        )) || (
                            <div className="col-span-full text-center py-8 text-gray-400">
                                <p>{t('categories.noVideos')}</p>
                            </div>
                        )}
                    </div>

                    <div className="flex justify-end">
                        <Button variant="outline" asChild>
                            <Link to={`/category/${category.slug}`}>
                                {t('categories.viewMore')}
                            </Link>
                        </Button>
                    </div>
                </div>
            ))}
        </div>
    );
};

export default CategoriesPage;