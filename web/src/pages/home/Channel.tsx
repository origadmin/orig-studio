import React, {useState} from 'react';
import {useParams} from '@tanstack/react-router';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Settings, Bell, Crown} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews, formatDate} from '@/lib/format';

const mockChannel = {
    id: 1, name: '技术教程', slug: 'tech-tutorials',
    avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=200',
    banner: 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&q=80&w=1200',
    description: '高质量编程、云计算和软件开发教程。每周更新 Go、React、Kubernetes 等内容。',
    subscriber_count: 125400, video_count: 156, total_views: 4560000,
    is_verified: true, is_subscribed: false, is_owner: false,
    links: {website: 'https://example.com', github: 'tech-tutorials'},
};

const mockVideos = [
    {
        id: 1,
        title: '从零构建 Go 微服务 - 完整指南',
        thumbnail: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400',
        duration: 3600,
        view_count: 125400,
        create_time: '2024-03-15',
        is_premium: false
    },
    {
        id: 2,
        title: 'Kubernetes 入门教程',
        thumbnail: 'https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&q=80&w=400',
        duration: 5400,
        view_count: 234500,
        create_time: '2024-03-12',
        is_premium: true
    },
    {
        id: 3,
        title: 'React 18 新特性完整指南',
        thumbnail: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?auto=format&fit=crop&q=80&w=400',
        duration: 2400,
        view_count: 89400,
        create_time: '2024-03-10',
        is_premium: false
    },
    {
        id: 4,
        title: 'Docker 深度解析',
        thumbnail: 'https://images.unsplash.com/photo-1605745341112-85968b19335b?auto=format&fit=crop&q=80&w=400',
        duration: 7200,
        view_count: 456000,
        create_time: '2024-03-05',
        is_premium: true
    },
    {
        id: 5,
        title: 'AWS 解决方案架构师课程',
        thumbnail: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?auto=format&fit=crop&q=80&w=400',
        duration: 18000,
        view_count: 890000,
        create_time: '2024-02-28',
        is_premium: true
    },
    {
        id: 6,
        title: 'TypeScript 高级模式',
        thumbnail: 'https://images.unsplash.com/photo-1516116216624-53e697fedbea?auto=format&fit=crop&q=80&w=400',
        duration: 2700,
        view_count: 67800,
        create_time: '2024-02-20',
        is_premium: false
    },
];

const ChannelPage = () => {
    const {id} = useParams({from: '/c/$id'});
    const [channel] = useState(mockChannel);
    const [videos] = useState(mockVideos);
    const [activeTab, setActiveTab] = useState('videos');

    return (
        <div className="space-y-8">
            <div className="relative">
                <div className="h-48 md:h-80 rounded-2xl bg-cover bg-center"
                     style={{backgroundImage: `url(${channel.banner})`}}/>
                <div className="absolute -bottom-16 left-6 flex items-end gap-6">
                    <Avatar className="w-32 h-32 border-4 border-white dark:border-gray-900 shadow-lg">
                        <AvatarImage src={channel.avatar}/>
                        <AvatarFallback className="text-3xl">{channel.name.charAt(0)}</AvatarFallback>
                    </Avatar>
                </div>
                <div className="absolute top-4 right-4 flex gap-2">
                    {channel.is_owner ? (
                        <Button variant="outline" className="bg-white dark:bg-gray-800"><Settings
                            className="w-4 h-4 mr-2"/>管理频道</Button>
                    ) : channel.is_subscribed ? (
                        <Button variant="outline" className="bg-white dark:bg-gray-800"><Bell className="w-4 h-4 mr-2"/>已通知</Button>
                    ) : (
                        <Button className="bg-red-600 hover:bg-red-700"><Bell className="w-4 h-4 mr-2"/>订阅</Button>
                    )}
                </div>
            </div>

            <div className="pt-20 px-6 space-y-4">
                <div className="flex items-center gap-3">
                    <h1 className="text-3xl font-bold text-slate-900 dark:text-white">{channel.name}</h1>
                    {channel.is_verified && <Badge variant="default" className="bg-emerald-500">已认证</Badge>}
                </div>
                <div className="flex flex-wrap gap-6 text-sm">
                    <div><span
                        className="font-semibold text-slate-900 dark:text-white">{formatViews(channel.subscriber_count)}</span><span
                        className="text-slate-500 dark:text-gray-400"> 订阅者</span></div>
                    <div><span
                        className="font-semibold text-slate-900 dark:text-white">{channel.video_count}</span><span
                        className="text-slate-500 dark:text-gray-400"> 个视频</span></div>
                    <div><span
                        className="font-semibold text-slate-900 dark:text-white">{formatViews(channel.total_views)}</span><span
                        className="text-slate-500 dark:text-gray-400"> 次观看</span></div>
                </div>
                <p className="text-slate-600 dark:text-gray-300 max-w-2xl">{channel.description}</p>
                {channel.links && (
                    <div className="flex gap-4">
                        {channel.links.website && <a href={channel.links.website}
                                                     className="text-emerald-600 hover:underline text-sm">网站</a>}
                        {channel.links.github &&
                            <a href="#" className="text-emerald-600 hover:underline text-sm">{channel.links.github}</a>}
                    </div>
                )}
            </div>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                <TabsList className="w-full justify-start border-b dark:border-gray-700 bg-transparent h-auto p-0">
                    {[
                        {v: 'videos', icon: <Play className="w-4 h-4 mr-2"/>, l: '视频'},
                        {v: 'playlists', l: '播放列表'},
                        {v: 'community', l: '社区'},
                        {v: 'about', l: '关于'},
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
                            <Link key={video.id} to="/v/$id" params={{id: String(video.id)}} className="group">
                                <div
                                    className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                    <div className="relative aspect-video">
                                        <img src={video.thumbnail} alt={video.title}
                                             className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                        <div
                                            className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                        {video.is_premium && <div className="absolute top-2 left-2"><Badge
                                            className="bg-amber-500 hover:bg-amber-600"><Crown
                                            className="w-3 h-3 mr-1"/>付费</Badge></div>}
                                    </div>
                                    <div className="p-3">
                                        <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                        <p className="text-xs text-slate-500 dark:text-gray-400 mt-2">{formatViews(video.view_count)} 次观看
                                            · {formatDate(video.create_time)}</p>
                                    </div>
                                </div>
                            </Link>
                        ))}
                    </div>
                </TabsContent>
                <TabsContent value="playlists" className="mt-6">
                    <div className="text-center py-12 text-slate-500 dark:text-gray-400">暂无播放列表</div>
                </TabsContent>
                <TabsContent value="community" className="mt-6">
                    <div className="text-center py-12 text-slate-500 dark:text-gray-400">暂无社区帖子</div>
                </TabsContent>
                <TabsContent value="about" className="mt-6">
                    <div className="max-w-2xl"><h3
                        className="font-semibold text-slate-900 dark:text-white mb-4">简介</h3><p
                        className="text-slate-600 dark:text-gray-300">{channel.description}</p></div>
                </TabsContent>
            </Tabs>
        </div>
    );
};
export default ChannelPage;
