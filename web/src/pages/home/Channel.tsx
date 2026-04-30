import React, {useState, useEffect} from 'react';
import {useParams, Link} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews} from '@/lib/format';
import {channelApi} from '@/lib/api/channel';
import {mediaApi, normalizeMediaList} from '@/lib/api/media';
import type {Channel} from '@/lib/api/channel';

const ChannelPage = () => {
    const params = useParams({strict: false}) as {id: string};
    const id = params.id;
    const {t} = useTranslation();
    const [channel, setChannel] = useState<Channel | null>(null);
    const [videos, setVideos] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [activeTab, setActiveTab] = useState('videos');

    useEffect(() => {
        if (!id) return;
        setLoading(true);
        Promise.all([
            channelApi.get({id}),
            mediaApi.list({channel_id: id, page_size: 12}),
        ])
            .then(([channelRes, videosRes]: any[]) => {
                const ch = channelRes?.data || channelRes;
                if (ch) setChannel(ch);
                const vid = videosRes?.data?.items || videosRes?.items || [];
                setVideos(normalizeMediaList(vid));
            })
            .finally(() => setLoading(false));
    }, [id]);

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
                        {v: 'playlists', l: t('channel.tabPlaylists')},
                        {v: 'community', l: t('channel.tabCommunity')},
                        {v: 'about', l: t('channel.tabAbout')},
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
                <TabsContent value="playlists" className="mt-6">
                    <div className="text-center py-12 text-slate-500 dark:text-muted-foreground">{t('channel.noPlaylists')}</div>
                </TabsContent>
                <TabsContent value="community" className="mt-6">
                    <div className="text-center py-12 text-slate-500 dark:text-muted-foreground">{t('channel.noCommunity')}</div>
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