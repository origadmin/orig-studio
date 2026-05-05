import React from 'react';
import {Link} from '@tanstack/react-router';
import {articleApi} from '@/lib/api/article';
import {useMediaList} from '@/hooks/queries';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Spinner} from '@/components/ui/spinner';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {FileText, Video, Clock, User, Eye} from 'lucide-react';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {useQuery} from '@tanstack/react-query';
import {useTranslation} from 'react-i18next';

const MixedHomeLayout: React.FC = () => {
    const {t} = useTranslation();
    const articlesQuery = useQuery({
        queryKey: ['articles', 'latest'],
        queryFn: () => articleApi.list({page: 1, page_size: 10, state: 'published'}),
        staleTime: 30_000,
    });

    // Use useMediaList hook which applies normalizeMedia for edge fields
    const videosQuery = useMediaList({
        page: 1,
        page_size: 12,
    });

    const isLoading = articlesQuery.isLoading || videosQuery.isLoading;

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[50vh]">
                <Spinner className="h-8 w-8"/>
            </div>
        );
    }

    const articles = articlesQuery.data?.items ?? [];
    const videos = videosQuery.data?.items ?? [];

    return (
        <div className="max-w-5xl mx-auto">
            <Tabs defaultValue="all" className="w-full">
                <div className="flex items-center justify-between mb-6">
                    <h1 className="text-2xl font-bold">{t('nav.home')}</h1>
                    <TabsList>
                        <TabsTrigger value="all">{t('home.all')}</TabsTrigger>
                        <TabsTrigger value="articles">{t('nav.articles')}</TabsTrigger>
                        <TabsTrigger value="videos">{t('common.videos')}</TabsTrigger>
                    </TabsList>
                </div>

                <TabsContent value="all" className="space-y-8">
                    {videos.length > 0 && (
                        <section>
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-lg font-semibold flex items-center gap-2">
                                    <Video className="h-5 w-5"/>
                                    {t('home.latestVideos')}
                                </h2>
                                <Link to="/latest" className="text-sm text-primary hover:underline">
                                    {t('home.viewAll')}
                                </Link>
                            </div>
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                {videos.slice(0, 6).map((video) => {
                                    const user = video?.edges?.user?.[0];
                                    const thumbUrl = getImageUrl(video?.thumbnail, 'thumbnail');
                                    return (
                                        <Link
                                            key={video.id}
                                            to="/watch"
                                            search={{v: video.short_token || video.id}}
                                            className="group"
                                        >
                                            <div className="aspect-video bg-muted rounded-lg overflow-hidden mb-2">
                                                {video.thumbnail ? (
                                                    <img
                                                        src={thumbUrl}
                                                        alt={video.title}
                                                        onError={(e) => handleImageError(e, 'thumbnail')}
                                                        className="w-full h-full object-cover group-hover:scale-105 transition-transform"
                                                    />
                                                ) : (
                                                    <div className="w-full h-full flex items-center justify-center">
                                                        <Video className="h-8 w-8 text-muted-foreground"/>
                                                    </div>
                                                )}
                                            </div>
                                            <h3 className="font-medium text-sm line-clamp-2">{video.title}</h3>
                                            <div className="flex items-center gap-2 mt-1">
                                                <span className="text-xs text-muted-foreground">
                                                    {user?.nickname || user?.username || 'Unknown'}
                                                </span>
                                                <span className="flex items-center gap-1 text-xs text-muted-foreground">
                                                    <Eye size={12}/>{formatViews(video.view_count || 0)}
                                                </span>
                                            </div>
                                        </Link>
                                    );
                                })}
                            </div>
                        </section>
                    )}

                    {articles.length > 0 && (
                        <section>
                            <div className="flex items-center justify-between mb-4">
                                <h2 className="text-lg font-semibold flex items-center gap-2">
                                    <FileText className="h-5 w-5"/>
                                    {t('home.latestArticles')}
                                </h2>
                                <Link to="/articles" className="text-sm text-primary hover:underline">
                                    {t('home.viewAll')}
                                </Link>
                            </div>
                            <div className="divide-y divide-border">
                                {articles.slice(0, 5).map((article) => (
                                    <Link
                                        key={article.id}
                                        to="/articles/$slug"
                                        params={{slug: article.slug}}
                                        className="block py-3 hover:bg-muted/50 -mx-2 px-2 rounded-lg transition-colors"
                                    >
                                        <div className="flex items-center gap-2 mb-1">
                                            <Avatar className="h-5 w-5">
                                                <AvatarFallback className="text-[10px]">
                                                    <User className="h-3 w-3"/>
                                                </AvatarFallback>
                                            </Avatar>
                                            <span className="text-sm font-medium">
                                                {article.user_id ? `User` : 'Anonymous'}
                                            </span>
                                            <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                <Clock className="h-3 w-3"/>
                                                {formatDate(article.create_time)}
                                            </span>
                                        </div>
                                        <h3 className="font-semibold text-sm line-clamp-1">{article.title}</h3>
                                    </Link>
                                ))}
                            </div>
                        </section>
                    )}
                </TabsContent>

                <TabsContent value="articles">
                    <div className="divide-y divide-border">
                        {articles.map((article) => (
                            <Link
                                key={article.id}
                                to="/articles/$slug"
                                params={{slug: article.slug}}
                                className="block py-4 hover:bg-muted/50 -mx-2 px-2 rounded-lg transition-colors"
                            >
                                <div className="flex gap-3">
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center gap-2 mb-1">
                                            <span className="text-sm font-medium">
                                                {article.user_id ? `User` : 'Anonymous'}
                                            </span>
                                            <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                <Clock className="h-3 w-3"/>
                                                {formatDate(article.create_time)}
                                            </span>
                                        </div>
                                        <h3 className="font-semibold text-base leading-snug mb-1 line-clamp-2">
                                            {article.title}
                                        </h3>
                                        {article.summary && (
                                            <p className="text-sm text-muted-foreground line-clamp-2">
                                                {article.summary}
                                            </p>
                                        )}
                                    </div>
                                    {article.thumbnail && (
                                        <div className="flex-shrink-0">
                                            <img
                                                src={getImageUrl(article.thumbnail, 'cover')}
                                                alt=""
                                                onError={(e) => handleImageError(e, 'cover')}
                                                className="w-24 h-24 object-cover rounded-lg"
                                            />
                                        </div>
                                    )}
                                </div>
                            </Link>
                        ))}
                    </div>
                </TabsContent>

                <TabsContent value="videos">
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                        {videos.map((video) => {
                            const user = video?.edges?.user?.[0];
                            const thumbUrl = getImageUrl(video?.thumbnail, 'thumbnail');
                            return (
                                <Link
                                    key={video.id}
                                    to="/watch"
                                    search={{v: video.short_token || video.id}}
                                    className="group"
                                >
                                    <div className="aspect-video bg-muted rounded-lg overflow-hidden mb-2">
                                        {video.thumbnail ? (
                                            <img
                                                src={thumbUrl}
                                                alt={video.title}
                                                onError={(e) => handleImageError(e, 'thumbnail')}
                                                className="w-full h-full object-cover group-hover:scale-105 transition-transform"
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <Video className="h-8 w-8 text-muted-foreground"/>
                                            </div>
                                        )}
                                    </div>
                                    <h3 className="font-medium text-sm line-clamp-2">{video.title}</h3>
                                    <div className="flex items-center gap-2 mt-1">
                                        <span className="text-xs text-muted-foreground">
                                            {user?.nickname || user?.username || 'Unknown'}
                                        </span>
                                        <span className="flex items-center gap-1 text-xs text-muted-foreground">
                                            <Eye size={12}/>{formatViews(video.view_count || 0)}
                                        </span>
                                    </div>
                                </Link>
                            );
                        })}
                    </div>
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default MixedHomeLayout;
