/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Notifications Page (TODO: i18n)
 */

import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {Bell, MessageSquare, Heart, UserPlus, Eye, CheckCheck} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';

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
        title: "New comment on your video", // TODO: i18n
        message: "React Master commented on 'Building Go Microservices'",
        avatar: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=100",
        is_read: false, created_at: "2024-03-15T14:30:00", link: "/v/1",
    },
    {
        id: "2", type: "like",
        title: "Someone liked your video", // TODO: i18n
        message: "DevOps Pro liked 'Docker Tutorial'",
        avatar: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&q=80&w=100",
        is_read: false, created_at: "2024-03-15T12:15:00", link: "/v/3",
    },
    {
        id: "3", type: "subscribe",
        title: "New subscriber", // TODO: i18n
        message: "TypeScript Guru subscribed to your channel",
        avatar: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=100",
        is_read: true, created_at: "2024-03-14T18:20:00", link: "/u/4",
    },
    {
        id: "4", type: "view",
        title: "Your video is trending", // TODO: i18n
        message: "'Go Microservices' reached 100K views",
        avatar: null,
        is_read: true, created_at: "2024-03-14T10:00:00", link: "/v/1",
    },
    {
        id: "5", type: "system",
        title: "System notification", // TODO: i18n
        message: "Your video has been approved and published",
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

// TODO: extract to lib/format.ts
function formatTime(dateStr: string): string {
    const diff = Date.now() - new Date(dateStr).getTime();
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);
    if (hours < 1) return 'Just now'; // TODO: i18n
    if (hours < 24) return `${hours}h ago`; // TODO: i18n
    if (days < 7) return `${days}d ago`; // TODO: i18n
    return new Date(dateStr).toLocaleDateString('en-US', {month: 'short', day: 'numeric'});
}

const NotificationsPage = () => {
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
                        {/* TODO: i18n */}通知
                    </h1>
                    <p className="text-gray-500 dark:text-gray-400 text-sm mt-1">
                        {unreadCount > 0
                            ? `${unreadCount} 条未读通知` // TODO: i18n
                            : '没有新通知'} {/* TODO: i18n */}
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button variant={filter === 'all' ? 'default' : 'outline'} size="sm"
                            onClick={() => setFilter('all')}>
                        {/* TODO: i18n */}全部
                    </Button>
                    <Button variant={filter === 'unread' ? 'default' : 'outline'} size="sm"
                            onClick={() => setFilter('unread')}>
                        {/* TODO: i18n */}未读
                    </Button>
                </div>
            </div>

            {/* Mark all as read */}
            {unreadCount > 0 && (
                <Button variant="ghost" size="sm" className="text-blue-600 dark:text-blue-400">
                    <CheckCheck className="w-4 h-4 mr-2"/>
                    {/* TODO: i18n */}全部已读
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
                    <p className="text-gray-500 dark:text-gray-400">{/* TODO: i18n */}暂无通知</p>
                </div>
            )}
        </div>
    );
};

export default NotificationsPage;
