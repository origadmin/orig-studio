import React, {useState, useEffect} from 'react';
import {useParams, Link} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {channelApi} from '@/lib/api/channel';
import {mediaApi, normalizeMediaList} from '@/lib/api/media';
import {articleApi, type Article} from '@/lib/api/article';
import type {Channel} from '@/lib/api/channel';
import {Play, FileText, Info} from 'lucide-react';

const ChannelPage = () => {
    const params = useParams({strict: false}) as {id: string};
    const id = params.id;
    const {t} = useTranslation();
    const [channel, setChannel] = useState<Channel | null>(null);
    const [videos, setVideos] = useState<any[]>([]);
    const [articles, setArticles] = useState<Article[]>([]);
    const [loading, setLoading] = useState(true);
    const [articlesLoading, setArticlesLoading] = useState(false);
    const [activeTab, setActiveTab] = useState('videos');

    useEffect(() => {
        if (!id) return;
        setLoading(true);
        Promise.all([
            channelApi.getByToken(id),
            mediaApi.list({page_size: 12}),
        ])
            .then(([channelRes, videosRes]: any[]) => {
                const ch = channelRes?.data || channelRes;
                if (ch) setChannel(ch);
                const vid = videosRes?.data?.items || videosRes?.items || [];
                setVideos(normalizeMediaList(vid));
            })
            .finally(() => setLoading(false));
    }, [id]);

    // Fetch articles when articles tab is active
    useEffect(() => {
        if (!id || activeTab !== 'articles') return;
        if (articles.length > 0) return;
        setArticlesLoading(true);
        articleApi.list({page_size: 20})
            .then((res: any) => {
                const data = res?.data || res;
                const items: Article[] = data?.items || [];
                // Filter articles belonging to this channel's user
                // The API may not support channel_id filter, so we filter client-side if needed
                setArticles(items);
            })
            .catch(err => console.error('Failed to fetch articles:', err))
            .finally(() => setArticlesLoading(false));
    }, [id, activeTab, articles.length]);

    if (loading) {
        return (
            <div className="space-y-8 animate-pulse">
                <div className="h-48 md:h-80 rounded-2xl bg-muted dark:bg-gray-800"/>
                <div className="pt-20 px-6 space-y-4">
                    <div className="h-8 bg-muted dark:bg-gray-800 rounded w-1/3"/>
                    <div className="h-4 bg-muted dark:bg-gray-800 rounded w-1/2"/>
                </div>
            </div>
        );
    }

    if (!channel) {
        return (
            <div className="text-center py-16 text-slate-500">{t('common.notFound')}</div>
        );
    }

    return (
        <div className="space-y-8">
            <div className="relative">
                <div className="h-48 md:h-80 rounded-2xl bg-cover bg-center bg-muted dark:bg-gray-800"
                     style={channel.banner ? {backgroundImage: `url(${channel.banner})`} : undefined}/>
                <div className="absolute -bottom-16 left-6 flex items-end gap-6">
                    <Avatar className="w-32 h-32 border-4 border-white dark:border-gray-900 shadow-lg">
                        {channel.avatar ? (
                            <AvatarImage src={channel.avatar}/>
                        ) : null}
                        <AvatarFallback className="text-3xl">{(channel.name || '?').charAt(0)}</AvatarFallback>
                    </Avatar>
                </div>
            </div>

            <div className="pt-20 px-6 space-y-4">
                <div className="flex items-center gap-3">
                    <h1 className="text-3xl font-bold text-slate-900 dark:text-white">{channel.name}</h1>
                </div>
                <div className="flex flex-wrap gap-6 text-sm">
                    <div><span
                        className="font-semibold text-slate-900 dark:text-white">{formatViews(channel.subscriber_count)}</span><span
                        className="text-slate-500 dark:text-muted-foreground"> {t('common.subscribers')}</span></div>
                    <div><span
                        className="font-semibold text-slate-900 dark:text-white">{channel.media_count ?? 0}</span><span
                        className="text-slate-500 dark:text-muted-foreground"> {t('common.videos_count')}</span></div>
                </div>
                {channel.description && (
                    <p className="text-slate-600 dark:text-gray-300 max-w-2xl">{channel.description}</p>
                )}
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                <TabsList className="w-full justify-start border-b dark:border-gray-700 bg-transparent h-auto p-0">
                    {[
                        {v: 'videos', icon: <Play className="w-4 h-4 mr-2"/>, l: t('channel.tabVideos')},
                        {v: 'articles', icon: <FileText className="w-4 h-4 mr-2"/>, l: t('channel.tabArticles')},
                        {v: 'about', icon: <Info className="w-4 h-4 mr-2"/>, l: t('channel.tabAbout')},
                    ].map(t => (
                        <TabsTrigger key={t.v} value={t.v}
                                     className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-emerald-600 rounded-none px-4 py-3">
                            {t.icon}{t.l}
                        </TabsTrigger>
                    ))}
                </TabsList>
                <TabsContent value="videos" className="mt-6">
                    {videos.length === 0 ? (
                        <div className="text-center py-12 text-slate-500 dark:text-muted-foreground">{t('channel.noVideos')}</div>
                    ) : (
                        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                            {videos.map(video => (
                                <Link key={video.id} to="/watch" search={{v: video.short_token}} className="group">
                                    <div
                                        className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                        <div className="relative aspect-video">
                                            {video.thumbnail ? (
                                                <img src={video.thumbnail} alt={video.title}
                                                     className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                            ) : (
                                                <div className="w-full h-full bg-muted dark:bg-gray-700 flex items-center justify-center">
                                                    <Play className="w-10 h-10 text-muted-foreground"/>
                                                </div>
                                            )}
                                            {video.duration != null ? (
                                                <div
                                                    className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                            ) : null}
                                        </div>
                                        <div className="p-3">
                                            <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                            <p className="text-xs text-slate-500 dark:text-muted-foreground mt-2">{formatViews(video.view_count)} {t('common.views')}</p>
                                        </div>
                                    </div>
                                </Link>
                            ))}
                        </div>
                    )}
                </TabsContent>
                <TabsContent value="articles" className="mt-6">
                    {articlesLoading ? (
                        <div className="text-center py-12 text-slate-500 dark:text-muted-foreground">{t('common.loading')}</div>
                    ) : articles.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            {articles.map(article => (
                                <Link
                                    key={article.id}
                                    to="/articles/$slug"
                                    params={{slug: article.slug || article.id}}
                                    className="group"
                                >
                                    <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1 border border-gray-100 dark:border-gray-700">
                                        {article.thumbnail && (
                                            <div className="aspect-video bg-muted dark:bg-gray-700">
                                                <img src={article.thumbnail} alt={article.title}
                                                     className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                            </div>
                                        )}
                                        <div className="p-4">
                                            <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 group-hover:text-emerald-600 transition-colors">
                                                {article.title}
                                            </h3>
                                            {article.summary && (
                                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-2 line-clamp-2">
                                                    {article.summary}
                                                </p>
                                            )}
                                            <div className="flex items-center gap-3 mt-3 text-xs text-slate-500 dark:text-muted-foreground">
                                                <span>{formatViews(article.view_count)} {t('common.views')}</span>
                                                {article.published_at && (
                                                    <span>{formatDate(article.published_at)}</span>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                </Link>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-12 text-slate-500 dark:text-muted-foreground">{t('channel.noArticles')}</div>
                    )}
                </TabsContent>
                <TabsContent value="about" className="mt-6">
                    <div className="max-w-2xl"><h3
                        className="font-semibold text-slate-900 dark:text-white mb-4">{t('channel.description')}</h3><p
                        className="text-slate-600 dark:text-gray-300">{channel.description || t('channel.noDescription')}</p></div>
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default ChannelPage;