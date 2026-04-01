/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Notifications Page
 */

import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {Bell, MessageSquare, Heart, UserPlus, Eye, CheckCheck} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatRelativeTime} from '@/lib/format';
import {useTranslation} from 'react-i18next';

interface Notification {
    id: string;
    type: 'comment' | 'like' | 'subscribe' | 'view' | 'system';
    title: string;
    message: string;
    avatar: string | null;
    is_read: boolean;
    created_at: string;
    link: string;
}

// TODO: replace with API call
const mockNotifications: Notification[] = [
    {
        id: "1", type: "comment",
        title: "视频收到了新评论",
        message: "React 大师 评论了「从零构建 Go 微服务」",
        avatar: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=100",
        is_read: false, created_at: "2024-03-15T14:30:00", link: "/v/1",
    },
    {
        id: "2", type: "like",
        title: "有人赞了你的视频",
        message: "运维专家 赞了「Docker 入门教程」",
        avatar: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&q=80&w=100",
        is_read: false, created_at: "2024-03-15T12:15:00", link: "/v/3",
    },
    {
        id: "3", type: "subscribe",
        title: "新订阅者",
        message: "TypeScript 达人 订阅了你的频道",
        avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=100",
        is_read: true, created_at: "2024-03-14T18:20:00", link: "/u/4",
    },
    {
        id: "4", type: "view",
        title: "你的视频正在流行",
        message: "「Go 微服务」播放量突破 10 万",
        avatar: null,
        is_read: true, created_at: "2024-03-14T10:00:00", link: "/v/1",
    },
    {
        id: "5", type: "system",
        title: "系统通知",
        message: "你的视频已审核通过并发布",
        avatar: null,
        is_read: true, created_at: "2024-03-13T09:00:00", link: "/v/1",
    },
];

function getIcon(type: Notification['type']) {
    switch (type) {
        case 'comment':
            return <MessageSquare className="w-5 h-5 text-blue-500"/>;
        case 'like':
            return <Heart className="w-5 h-5 text-rose-500"/>;
        case 'subscribe':
            return <UserPlus className="w-5 h-5 text-green-500"/>;
        case 'view':
            return <Eye className="w-5 h-5 text-purple-500"/>;
        default:
            return <Bell className="w-5 h-5 text-gray-500"/>;
    }
}

const NotificationsPage = () => {
    const {t} = useTranslation();
    const [notifications] = useState(mockNotifications);
    const [filter, setFilter] = useState<'all' | 'unread'>('all');

    const filtered = filter === 'all'
        ? notifications
        : notifications.filter(n => !n.is_read);

    const unreadCount = notifications.filter(n => !n.is_read).length;

    return (
        <div className="max-w-2xl mx-auto space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                        <Bell className="w-6 h-6"/>
                        {t('notifications.title')}
                    </h1>
                    <p className="text-gray-500 dark:text-gray-400 text-sm mt-1">
                        {unreadCount > 0
                            ? t('notifications.unreadCount', {count: unreadCount})
                            : t('notifications.noNew')}
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button variant={filter === 'all' ? 'default' : 'outline'} size="sm"
                            onClick={() => setFilter('all')}>
                        {t('notifications.all')}
                    </Button>
                    <Button variant={filter === 'unread' ? 'default' : 'outline'} size="sm"
                            onClick={() => setFilter('unread')}>
                        {t('notifications.unread')}
                    </Button>
                </div>
            </div>

            {/* Mark all as read */}
            {unreadCount > 0 && (
                <Button variant="ghost" size="sm" className="text-blue-600 dark:text-blue-400">
                    <CheckCheck className="w-4 h-4 mr-2"/>
                    {t('notifications.markAllRead')}
                </Button>
            )}

            {/* List */}
            <div className="space-y-3">
                {filtered.map(n => (
                    <Link key={n.id} to={n.link}
                        className={`block p-4 rounded-xl transition-colors ${
                            n.is_read
                                ? 'bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700/50'
                                : 'bg-blue-50 dark:bg-blue-950/30 hover:bg-blue-100 dark:hover:bg-blue-950/50 border-l-4 border-blue-500'
                        }`}>
                        <div className="flex gap-4">
                            {n.avatar ? (
                                <Avatar className="w-10 h-10 shrink-0">
                                    <AvatarImage src={n.avatar}/>
                                    <AvatarFallback>U</AvatarFallback>
                                </Avatar>
                            ) : (
                                <div
                                    className="w-10 h-10 bg-gray-200 dark:bg-gray-600 rounded-full flex items-center justify-center shrink-0">
                                    {getIcon(n.type)}
                                </div>
                            )}
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 mb-1">
                                    <p className="font-semibold text-gray-900 dark:text-white">{n.title}</p>
                                    {!n.is_read && <div className="w-2 h-2 bg-blue-500 rounded-full"/>}
                                </div>
                                <p className="text-sm text-gray-500 dark:text-gray-400 line-clamp-1">{n.message}</p>
                                <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{formatTime(n.created_at)}</p>
                            </div>
                            <div className="shrink-0 hidden sm:block">{getIcon(n.type)}</div>
                        </div>
                    </Link>
                ))}
            </div>

            {filtered.length === 0 && (
                <div className="text-center py-20">
                    <Bell className="w-16 h-16 text-gray-200 dark:text-gray-700 mx-auto mb-4"/>
                    <p className="text-gray-500 dark:text-gray-400">{t('notifications.empty')}</p>
                </div>
            )}
        </div>
    );
};

export default NotificationsPage;
