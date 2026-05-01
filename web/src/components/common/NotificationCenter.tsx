﻿import {Spinner} from "@/components/ui/spinner"
import React, {useState, useEffect} from 'react';
import {Bell, Check, Trash2, Loader2, X} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDate} from '@/lib/format';
import {notificationApi, type Notification} from '@/lib/api/notification';
import ErrorPage from '@/components/common/ErrorPage';

const NotificationCenter: React.FC = () => {
    const {t} = useTranslation();
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [unreadCount, setUnreadCount] = useState(0);
    const [isMarkingAllRead, setIsMarkingAllRead] = useState(false);

    useEffect(() => {
        fetchNotifications();
        fetchUnreadCount();
    }, []);

    const fetchNotifications = async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await notificationApi.getAll({page_size: 20});
            setNotifications((response as any)?.items || response || []);
        } catch (err) {
            setError('Failed to fetch notifications');
            console.error('Failed to fetch notifications:', err);
        } finally {
            setLoading(false);
        }
    };

    const fetchUnreadCount = async () => {
        try {
            const response = await notificationApi.getUnreadCount();
            setUnreadCount((response as any)?.unread_count || (response as any)?.count || 0);
        } catch (err) {
            console.error('Failed to fetch unread count:', err);
        }
    };

    const handleMarkAsRead = async (id: string) => {
        try {
            await notificationApi.markAsRead(id);
            setNotifications(prev => prev.map(notification =>
                notification.id === id ? {...notification, read: true} : notification
            ));
            fetchUnreadCount();
        } catch (err) {
            console.error('Failed to mark notification as read:', err);
        }
    };

    const handleMarkAllAsRead = async () => {
        try {
            setIsMarkingAllRead(true);
            await notificationApi.markAllAsRead();
            setNotifications(prev => prev.map(notification => ({...notification, read: true})));
            setUnreadCount(0);
        } catch (err) {
            console.error('Failed to mark all notifications as read:', err);
        } finally {
            setIsMarkingAllRead(false);
        }
    };

    const handleDelete = async (id: string) => {
        try {
            await notificationApi.delete(id);
            setNotifications(prev => prev.filter(notification => notification.id !== id));
        } catch (err) {
            console.error('Failed to delete notification:', err);
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[200px]">
                <Spinner />
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    return (
        <div className="space-y-4">
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <CardTitle className="flex items-center gap-2">
                            <Bell className="w-5 h-5"/>
                            {t('notifications.title')}
                            {unreadCount > 0 && (
                                <span
                                    className="text-xs font-medium px-2 py-0.5 bg-blue-100 text-blue-800 rounded-full">
                  {unreadCount}
                </span>
                            )}
                        </CardTitle>
                        {notifications.length > 0 && (
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={handleMarkAllAsRead}
                                disabled={isMarkingAllRead}
                            >
                                {isMarkingAllRead ? (
                                    <>
                                        <Loader2 className="w-4 h-4 mr-2 animate-spin"/>
                                        {t('notifications.markingAllRead')}
                                    </>
                                ) : (
                                    <>
                                        <Check className="w-4 h-4 mr-2"/>
                                        {t('notifications.markAllAsRead')}
                                    </>
                                )}
                            </Button>
                        )}
                    </div>
                </CardHeader>
                <CardContent>
                    {notifications.length === 0 ? (
                        <div className="text-center py-12 text-gray-500 dark:text-muted-foreground">
                            {t('notifications.noNotifications')}
                        </div>
                    ) : (
                        <div className="space-y-4">
                            {notifications.map(notification => (
                                <div
                                    key={notification.id}
                                    className={`p-4 rounded-lg border ${notification.read ? 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800' : 'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/30'}`}
                                >
                                    <div className="flex items-start justify-between gap-4">
                                        <div className="flex-1 space-y-2">
                                            <h4 className="font-medium text-gray-900 dark:text-white">
                                                {notification.title}
                                            </h4>
                                            <p className="text-sm text-gray-600 dark:text-gray-300">
                                                {notification.body}
                                            </p>
                                            <div className="flex items-center gap-4">
                        <span className="text-xs text-gray-500 dark:text-muted-foreground">
                          {formatDate(notification.create_time)}
                        </span>
                                                <span
                                                    className={`text-xs font-medium px-2 py-0.5 rounded-full ${notification.read ? 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200' : 'bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-200'}`}>
                          {notification.read ? t('notifications.read') : t('notifications.unread')}
                        </span>
                                            </div>
                                        </div>
                                        <div className="flex flex-col gap-2">
                                            {!notification.read && (
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="h-8 w-8 p-0"
                                                    onClick={() => handleMarkAsRead(notification.id)}
                                                    title={t('notifications.markAsRead')}
                                                >
                                                    <Check className="w-4 h-4"/>
                                                </Button>
                                            )}
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                className="h-8 w-8 p-0 text-destructive dark:text-red-400"
                                                onClick={() => handleDelete(notification.id)}
                                                title={t('common.delete')}
                                            >
                                                <Trash2 className="w-4 h-4"/>
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};

export default NotificationCenter;
