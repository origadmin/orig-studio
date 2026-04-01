/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 视频播放页面
 */

import React, {useState} from 'react';
import {useParams, Link} from '@tanstack/react-router';
import {Play, ThumbsUp, ThumbsDown, Share2, Flag, Eye, Calendar, Star} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDuration, formatViews, formatDate} from '@/lib/format';

const mockVideo = {
    id: 1,
    title: '从零构建 Go 微服务',
    description: '学习如何使用 Go 和 Kratos 框架构建生产级微服务。\n\n本教程涵盖：\n• Go 项目结构搭建\n• RESTful API 实现\n• 数据库集成\n• 认证与授权\n• 错误处理与日志\n• 部署策略',
    thumbnail: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=800',
    duration: 3600,
    view_count: 125400,
    create_time: '2024-03-15',
    user_id: 1,
    author_name: 'Gopher 专家',
    author_avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=100',
    category: '技术',
    tags: ['Go', '微服务', '后端'],
    likes: 8400,
    dislikes: 120,
};

const mockRelatedVideos = [
    {
        id: 2,
        title: 'React 18 高级模式与最佳实践',
        thumbnail: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?auto=format&fit=crop&q=80&w=300',
        duration: 2400,
        view_count: 89400,
        author_name: 'React 大师'
    },
    {
        id: 3,
        title: 'Docker & Kubernetes 深度解析',
        thumbnail: 'https://images.unsplash.com/photo-1667372393119-3d4c48d07fc9?auto=format&fit=crop&q=80&w=300',
        duration: 5400,
        view_count: 234500,
        author_name: '运维专家'
    },
    {
        id: 4,
        title: 'TypeScript 高级类型大师课',
        thumbnail: 'https://images.unsplash.com/photo-1516116216624-53e697fedbea?auto=format&fit=crop&q=80&w=300',
        duration: 1800,
        view_count: 67800,
        author_name: 'TS 达人'
    },
    {
        id: 5,
        title: '系统设计面试完全指南',
        thumbnail: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?auto=format&fit=crop&q=80&w=300',
        duration: 7200,
        view_count: 456000,
        author_name: '面试教练'
    },
    {
        id: 6,
        title: 'Python 数据科学实战',
        thumbnail: 'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&q=80&w=300',
        duration: 4800,
        view_count: 189000,
        author_name: '数据科学家'
    },
];

const WatchPage = () => {
    const {id} = useParams({from: '/v/$id'});
    const [video] = useState(mockVideo);
    const [relatedVideos] = useState(mockRelatedVideos);
    const [isLiked, setIsLiked] = useState(false);
    const [isDisliked, setIsDisliked] = useState(false);
    const [isFavorited, setIsFavorited] = useState(false);
    const [likeCount, setLikeCount] = useState(mockVideo.likes);

    const handleLike = () => {
        if (isLiked) {
            setIsLiked(false);
            setLikeCount(likeCount - 1);
        } else {
            setIsLiked(true);
            setLikeCount(likeCount + 1);
            if (isDisliked) setIsDisliked(false);
        }
    };

    const handleDislike = () => {
        if (isDisliked) {
            setIsDisliked(false);
        } else {
            setIsDisliked(true);
            if (isLiked) {
                setIsLiked(false);
                setLikeCount(likeCount - 1);
            }
        }
    };

    return (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <div className="lg:col-span-2 space-y-6">
                {/* 播放器 */}
                <div className="relative aspect-video bg-slate-900 rounded-2xl overflow-hidden group">
                    <img src={video.thumbnail} alt={video.title} className="w-full h-full object-cover opacity-80"/>
                    <div className="absolute inset-0 flex items-center justify-center">
                        <button
                            className="w-20 h-20 bg-white/20 backdrop-blur-sm rounded-full flex items-center justify-center transform group-hover:scale-110 transition-transform">
                            <Play className="w-10 h-10 text-white ml-1" fill="currentColor"/>
                        </button>
                    </div>
                    <div className="absolute bottom-4 right-4 bg-black/80 text-white text-sm px-3 py-1 rounded">
                        {formatDuration(video.duration)}
                    </div>
                </div>

                {/* 视频信息 */}
                <div className="space-y-4">
                    <h1 className="text-2xl font-bold text-slate-900 dark:text-white">{video.title}</h1>

                    <div
                        className="flex flex-wrap items-center justify-between gap-4 pb-4 border-b border-gray-200 dark:border-gray-700">
                        <div className="flex items-center gap-4 text-sm text-slate-500 dark:text-gray-400">
                            <span className="flex items-center gap-1"><Eye
                                className="w-4 h-4"/>{formatViews(video.view_count)} 次观看</span>
                            <span className="flex items-center gap-1"><Calendar
                                className="w-4 h-4"/>{formatDate(video.create_time)}</span>
                        </div>

                        <div className="flex items-center gap-2">
                            <Button variant={isLiked ? 'default' : 'outline'} size="sm" onClick={handleLike}
                                    className={isLiked ? 'bg-blue-600' : ''}>
                                <ThumbsUp className="w-4 h-4 mr-1"/>{formatViews(likeCount)}
                            </Button>
                            <Button variant={isDisliked ? 'default' : 'outline'} size="sm" onClick={handleDislike}
                                    className={isDisliked ? 'bg-slate-600' : ''}>
                                <ThumbsDown className="w-4 h-4"/>
                            </Button>
                            <Button variant={isFavorited ? 'default' : 'outline'} size="sm"
                                    onClick={() => setIsFavorited(!isFavorited)}
                                    className={isFavorited ? 'bg-rose-600' : ''}>
                                <Star
                                    className={`w-4 h-4 mr-1 ${isFavorited ? 'fill-current' : ''}`}/>{isFavorited ? '已收藏' : '收藏'}
                            </Button>
                            <Button variant="outline" size="sm"><Share2 className="w-4 h-4 mr-1"/>分享</Button>
                            <Button variant="ghost" size="sm"><Flag className="w-4 h-4"/></Button>
                        </div>
                    </div>

                    {/* 作者 */}
                    <div className="flex items-center justify-between p-4 bg-slate-50 dark:bg-gray-800 rounded-xl">
                        <Link to={`/u/${video.user_id}`} className="flex items-center gap-3">
                            <Avatar className="w-12 h-12">
                                <AvatarImage src={video.author_avatar}/>
                                <AvatarFallback>{video.author_name.charAt(0)}</AvatarFallback>
                            </Avatar>
                            <div>
                                <p className="font-semibold text-slate-900 dark:text-white">{video.author_name}</p>
                                <p className="text-sm text-slate-500 dark:text-gray-400">12.5 万订阅者</p>
                            </div>
                        </Link>
                        <Button className="bg-red-600 hover:bg-red-700">订阅</Button>
                    </div>

                    {/* 描述 */}
                    <div className="p-4 bg-slate-50 dark:bg-gray-800 rounded-xl">
                        <div className="flex flex-wrap gap-2 mb-4">
                            {video.tags.map(tag => <Badge key={tag} variant="secondary">{tag}</Badge>)}
                        </div>
                        <p className="text-sm text-slate-600 dark:text-gray-300 whitespace-pre-line">{video.description}</p>
                    </div>
                </div>
            </div>

            {/* 相关推荐 */}
            <div className="space-y-4">
                <h3 className="font-semibold text-slate-900 dark:text-white">接下来播放</h3>
                {relatedVideos.map(item => (
                    <Link key={item.id} to="/v/$id" params={{id: String(item.id)}} className="flex gap-3 group">
                        <div className="relative w-40 h-24 bg-slate-200 rounded-lg overflow-hidden shrink-0">
                            <img src={item.thumbnail} alt={item.title}
                                 className="w-full h-full object-cover group-hover:scale-105 transition-transform"/>
                            <div
                                className="absolute bottom-1 right-1 bg-black/80 text-white text-xs px-1.5 py-0.5 rounded">{formatDuration(item.duration)}</div>
                        </div>
                        <div className="min-w-0 flex-1">
                            <h4 className="text-sm font-medium text-slate-900 dark:text-white line-clamp-2 group-hover:text-blue-600 transition-colors">{item.title}</h4>
                            <p className="text-xs text-slate-500 dark:text-gray-400 mt-1">{item.author_name}</p>
                            <p className="text-xs text-slate-400 dark:text-gray-500">{formatViews(item.view_count)} 次观看</p>
                        </div>
                    </Link>
                ))}
            </div>
        </div>
    );
};

export default WatchPage;
