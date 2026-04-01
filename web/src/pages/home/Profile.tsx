import React, {useState} from 'react';
import {useParams} from '@tanstack/react-router';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Calendar, Settings, Bell, Heart, UserPlus, MessageSquare} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews, formatDate} from '@/lib/format';

const mockUser = {
    id: 1, username: 'gopher_expert', nickname: 'Gopher 专家',
    avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=200',
    cover: 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&q=80&w=1200',
    bio: '全栈开发者，热爱 Go、React 和微服务架构。通过教程和课程分享知识。',
    subscriber_count: 125400, video_count: 48, total_views: 2345000,
    joined_date: '2023-06-15', is_verified: true, is_subscribed: false, is_me: false,
};

const mockVideos = [
    {
        id: 1,
        title: '从零构建 Go 微服务',
        thumbnail: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400',
        duration: 3600,
        view_count: 125400,
        create_time: '2024-03-15'
    },
    {
        id: 2,
        title: 'Go 并发模式进阶',
        thumbnail: 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?auto=format&fit=crop&q=80&w=400',
        duration: 2700,
        view_count: 89400,
        create_time: '2024-03-10'
    },
    {
        id: 3,
        title: 'Go 错误处理最佳实践',
        thumbnail: 'https://images.unsplash.com/photo-1542831371-29b0f74f9713?auto=format&fit=crop&q=80&w=400',
        duration: 1800,
        view_count: 67800,
        create_time: '2024-03-05'
    },
    {
        id: 4,
        title: '构建 REST API',
        thumbnail: 'https://images.unsplash.com/photo-1516116216624-53e697fedbea?auto=format&fit=crop&q=80&w=400',
        duration: 4200,
        view_count: 156000,
        create_time: '2024-02-28'
    },
    {
        id: 5,
        title: 'Go 数据库集成教程',
        thumbnail: 'https://images.unsplash.com/photo-1504639725590-34d0984388bd?auto=format&fit=crop&q=80&w=400',
        duration: 3000,
        view_count: 112000,
        create_time: '2024-02-20'
    },
    {
        id: 6,
        title: 'Go 测试策略',
        thumbnail: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400',
        duration: 2400,
        view_count: 89000,
        create_time: '2024-02-15'
    },
];

const ProfilePage = () => {
    const {id} = useParams({from: '/u/$id'});
    const {t} = useTranslation();
    const [user] = useState(mockUser);
    const [videos] = useState(mockVideos);

    return (
        <div className="space-y-8">
            <div className="relative">
                <div className="h-48 md:h-64 rounded-2xl bg-cover bg-center"
                     style={{backgroundImage: `url(${user.cover})`}}/>
                <div className="absolute -bottom-16 left-6 flex items-end gap-6">
                    <Avatar className="w-32 h-32 border-4 border-white dark:border-gray-900 shadow-lg">
                        <AvatarImage src={user.avatar}/>
                        <AvatarFallback className="text-3xl">{user.nickname.charAt(0)}</AvatarFallback>
                    </Avatar>
                </div>
                <div className="absolute top-4 right-4">
                    {user.is_me ? (
                        <Button variant="outline" className="bg-white dark:bg-gray-800"><Settings
                            className="w-4 h-4 mr-2"/>{t('common.editProfile')}</Button>
                    ) : user.is_subscribed ? (
                        <Button variant="outline" className="bg-white dark:bg-gray-800"><UserPlus
                            className="w-4 h-4 mr-2"/>{t('common.subscribed')}</Button>
                    ) : (
                        <Button className="bg-red-600 hover:bg-red-700">{t('common.subscribe')}</Button>
                    )}
                </div>
            </div>

            <div className="pt-20 px-6 space-y-4">
                <div className="flex items-center gap-3">
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white">{user.nickname}</h1>
                    {user.is_verified &&
                        <Badge variant="default" className="bg-emerald-500">{t('common.verified')}</Badge>}
                </div>
                <p className="text-slate-500 dark:text-gray-400">@{user.username}</p>
                <p className="text-slate-600 dark:text-gray-300 max-w-2xl">{user.bio}</p>
                <div className="flex flex-wrap gap-6 text-sm">
                    <div className="flex items-center gap-2"><UserPlus className="w-4 h-4 text-slate-400"/><span
                        className="font-semibold text-slate-900 dark:text-white">{formatViews(user.subscriber_count)}</span><span
                        className="text-slate-500 dark:text-gray-400">{t('common.subscribers')}</span></div>
                    <div className="flex items-center gap-2"><Play className="w-4 h-4 text-slate-400"/><span
                        className="font-semibold text-slate-900 dark:text-white">{user.video_count}</span><span
                        className="text-slate-500 dark:text-gray-400">{t('common.videos_count')}</span></div>
                    <div className="flex items-center gap-2"><Eye className="w-4 h-4 text-slate-400"/><span
                        className="font-semibold text-slate-900 dark:text-white">{formatViews(user.total_views)}</span><span
                        className="text-slate-500 dark:text-gray-400">{t('common.views')}</span></div>
                    <div className="flex items-center gap-2"><Calendar className="w-4 h-4 text-slate-400"/><span
                        className="text-slate-500 dark:text-gray-400">{t('common.joinedAt', {date: formatDate(user.joined_date)})}</span>
                    </div>
                </div>
            </div>

            <Tabs defaultValue="videos" className="w-full">
                <TabsList className="w-full justify-start border-b dark:border-gray-700 bg-transparent h-auto p-0">
                    {[
                        {v: 'videos', icon: <Play className="w-4 h-4 mr-2"/>, l: t('channel.tabVideos')},
                        {v: 'playlists', l: t('profile.tabPlaylists')},
                        {v: 'favorites', icon: <Heart className="w-4 h-4 mr-2"/>, l: t('profile.tabFavorites')},
                        {v: 'community', icon: <MessageSquare className="w-4 h-4 mr-2"/>, l: t('profile.tabCommunity')},
                    ].map(t => (
                        <TabsTrigger key={t.v} value={t.v}
                                     className="data-[state=active]:bg-transparent data-[state=active]:border-b-2 data-[state=active]:border-emerald-600 rounded-none px-4 py-3">
                            {t.icon}{t.l}
                        </TabsTrigger>
                    ))}
                </TabsList>
                <TabsContent value="videos" className="mt-6">
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                        {videos.map(video => (
                            <Link key={video.id} to="/watch" search={{v: String(video.id)}} className="group">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                    <div className="relative aspect-video">
                                        <img src={video.thumbnail} alt={video.title}
                                             className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                    </div>
                                    <div className="p-3">
                                        <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                        <p className="text-xs text-slate-500 dark:text-gray-400 mt-2">{formatViews(video.view_count)} {t('common.views')}
                                            · {formatDate(video.create_time)}</p>
                                    </div>
                                </div>
                            </Link>
                        ))}
                    </div>
                </TabsContent>
                <TabsContent value="playlists" className="mt-6">
                    <div
                        className="text-center py-12 text-slate-500 dark:text-gray-400">{t('profile.noPlaylists')}</div>
                </TabsContent>
                <TabsContent value="favorites" className="mt-6">
                    <div
                        className="text-center py-12 text-slate-500 dark:text-gray-400">{t('profile.noFavorites')}</div>
                </TabsContent>
                <TabsContent value="community" className="mt-6">
                    <div
                        className="text-center py-12 text-slate-500 dark:text-gray-400">{t('profile.noCommunity')}</div>
                </TabsContent>
            </Tabs>
        </div>
    );
};
export default ProfilePage;
