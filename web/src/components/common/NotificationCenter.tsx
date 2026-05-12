import {Spinner} from "@/components/ui/spinner"
import React, {useState, useEffect} from 'react';
import {Bell, Check, Trash2, Loader2, CheckSquare, Square, X} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Checkbox} from '@/components/ui/checkbox';
import {formatDate} from '@/lib/format';
import {notificationApi, type Notification} from '@/lib/api/notification';
import {useNotificationState} from '@/contexts/NotificationContext';
import ErrorPage from '@/components/common/ErrorPage';

const NotificationCenter: React.FC = () => {
    const {t} = useTranslation();
    const {unreadCount, markAsRead, markAllAsRead, refresh} = useNotificationState();
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [isMarkingAllRead, setIsMarkingAllRead] = useState(false);
    const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
    const [batchMode, setBatchMode] = useState(false);
    const [batchLoading, setBatchLoading] = useState(false);

    useEffect(() => {
        fetchNotifications();
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

    const handleMarkAsRead = async (id: number) => {
        try {
            await markAsRead(id);
            setNotifications(prev => prev.map(n => n.id === id ? {...n, read: true} : n));
        } catch (err) {
            console.error('Failed to mark notification as read:', err);
        }
    };

    const handleMarkAllAsRead = async () => {
        try {
            setIsMarkingAllRead(true);
            await markAllAsRead();
            setNotifications(prev => prev.map(n => ({...n, read: true})));
        } catch (err) {
            console.error('Failed to mark all notifications as read:', err);
        } finally {
            setIsMarkingAllRead(false);
        }
    };

    const handleDelete = async (id: number) => {
        try {
            await notificationApi.delete(id);
            setNotifications(prev => prev.filter(n => n.id !== id));
            setSelectedIds(prev => {
                const next = new Set(prev);
                next.delete(id);
                return next;
            });
            refresh();
        } catch (err) {
            console.error('Failed to delete notification:', err);
        }
    };

    const toggleSelect = (id: number) => {
        setSelectedIds(prev => {
            const next = new Set(prev);
            if (next.has(id)) next.delete(id);
            else next.add(id);
            return next;
        });
    };

    const toggleSelectAll = () => {
        if (selectedIds.size === notifications.length) {
            setSelectedIds(new Set());
        } else {
            setSelectedIds(new Set(notifications.map(n => n.id)));
        }
    };

    const handleBatchMarkRead = async () => {
        if (selectedIds.size === 0) return;
        try {
            setBatchLoading(true);
            await Promise.all([...selectedIds].map(id => notificationApi.markAsRead(id)));
            setNotifications(prev => prev.map(n => selectedIds.has(n.id) ? {...n, read: true} : n));
            setSelectedIds(new Set());
            refresh();
        } catch (err) {
            console.error('Failed to batch mark as read:', err);
        } finally {
            setBatchLoading(false);
        }
    };

    const handleBatchDelete = async () => {
        if (selectedIds.size === 0) return;
        try {
            setBatchLoading(true);
            await Promise.all([...selectedIds].map(id => notificationApi.delete(id)));
            setNotifications(prev => prev.filter(n => !selectedIds.has(n.id)));
            setSelectedIds(new Set());
            refresh();
        } catch (err) {
            console.error('Failed to batch delete:', err);
        } finally {
            setBatchLoading(false);
        }
    };

    const exitBatchMode = () => {
        setBatchMode(false);
        setSelectedIds(new Set());
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[200px]">
                <Spinner/>
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
                                <span className="text-xs font-medium px-2 py-0.5 bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-200 rounded-full">
                                    {unreadCount}
                                </span>
                            )}
                        </CardTitle>
                        <div className="flex items-center gap-2">
                            {batchMode ? (
                                <>
                                    {selectedIds.size > 0 && (
                                        <>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={handleBatchMarkRead}
                                                disabled={batchLoading}
                                            >
                                                <Check className="w-4 h-4 mr-1"/>
                                                {t('notifications.markAsRead')} ({selectedIds.size})
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={handleBatchDelete}
                                                disabled={batchLoading}
                                                className="text-destructive hover:text-destructive"
                                            >
                                                <Trash2 className="w-4 h-4 mr-1"/>
                                                {t('common.delete')} ({selectedIds.size})
                                            </Button>
                                        </>
                                    )}
                                    <Button variant="ghost" size="sm" onClick={exitBatchMode}>
                                        <X className="w-4 h-4 mr-1"/>
                                        {t('common.cancel') || 'Cancel'}
                                    </Button>
                                </>
                            ) : (
                                <>
                                    {notifications.length > 0 && (
                                        <>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => setBatchMode(true)}
                                            >
                                                <CheckSquare className="w-4 h-4 mr-1"/>
                                                {t('notifications.selectUsers') || 'Select'}
                                            </Button>
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
                                        </>
                                    )}
                                </>
                            )}
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    {notifications.length === 0 ? (
                        <div className="text-center py-12 text-gray-500 dark:text-muted-foreground">
                            {t('notifications.noNotifications')}
                        </div>
                    ) : (
                        <>
                            {batchMode && (
                                <div className="flex items-center gap-2 pb-3 border-b mb-3">
                                    <Checkbox
                                        checked={selectedIds.size === notifications.length}
                                        onCheckedChange={toggleSelectAll}
                                    />
                                    <span className="text-sm text-muted-foreground">
                                        {selectedIds.size > 0
                                            ? t('notifications.selectUsers') + ` (${selectedIds.size}/${notifications.length})`
                                            : t('notifications.allUsers') || 'Select all'}
                                    </span>
                                </div>
                            )}
                            <div className="space-y-2">
                                {notifications.map(notification => (
                                    <div
                                        key={notification.id}
                                        className={`p-4 rounded-lg border transition-colors ${
                                            selectedIds.has(notification.id) ? 'border-primary bg-primary/5' :
                                            notification.read ? 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800' :
                                            'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/30'
                                        }`}
                                    >
                                        <div className="flex items-start gap-3">
                                            {batchMode && (
                                                <Checkbox
                                                    checked={selectedIds.has(notification.id)}
                                                    onCheckedChange={() => toggleSelect(notification.id)}
                                                    className="mt-1"
                                                />
                                            )}
                                            <div className="flex-1 min-w-0 space-y-2">
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
                                                    <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${notification.read ? 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200' : 'bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-200'}`}>
                                                        {notification.read ? t('notifications.read') : t('notifications.unread')}
                                                    </span>
                                                </div>
                                            </div>
                                            {!batchMode && (
                                                <div className="flex flex-col gap-1">
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
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};

export default NotificationCenter;
