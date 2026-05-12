import React, {createContext, useContext, useState, useEffect, useCallback} from 'react';
import {notificationApi, type Notification} from '@/lib/api/notification';
import {useAuth} from '@/hooks/useAuth';

interface NotificationState {
    unreadCount: number;
    recentNotifications: Notification[];
    refresh: () => Promise<void>;
    markAsRead: (id: number) => Promise<void>;
    markAllAsRead: () => Promise<void>;
    deleteNotification: (id: number) => Promise<void>;
}

const NotificationContext = createContext<NotificationState>({
    unreadCount: 0,
    recentNotifications: [],
    refresh: async () => {},
    markAsRead: async () => {},
    markAllAsRead: async () => {},
    deleteNotification: async () => {},
});

export const useNotificationState = () => useContext(NotificationContext);

export const NotificationProvider: React.FC<{ children: React.ReactNode }> = ({children}) => {
    const {user} = useAuth();
    const [unreadCount, setUnreadCount] = useState(0);
    const [recentNotifications, setRecentNotifications] = useState<Notification[]>([]);

    const refresh = useCallback(async () => {
        if (!user) return;
        try {
            const [countRes, notifsRes] = await Promise.all([
                notificationApi.getUnreadCount(),
                notificationApi.getAll({page_size: 5}),
            ]);
            setUnreadCount((countRes as any)?.unread_count || (countRes as any)?.count || 0);
            setRecentNotifications((notifsRes as any)?.items || notifsRes || []);
        } catch (err) {
            console.error('Failed to refresh notification state:', err);
        }
    }, [user]);

    const markAsRead = useCallback(async (id: number) => {
        try {
            await notificationApi.markAsRead(id);
            setRecentNotifications(prev => prev.map(n => n.id === id ? {...n, read: true} : n));
            setUnreadCount(prev => Math.max(0, prev - 1));
        } catch (err) {
            console.error('Failed to mark as read:', err);
        }
    }, []);

    const markAllAsRead = useCallback(async () => {
        try {
            await notificationApi.markAllAsRead();
            setRecentNotifications(prev => prev.map(n => ({...n, read: true})));
            setUnreadCount(0);
        } catch (err) {
            console.error('Failed to mark all as read:', err);
        }
    }, []);

    const deleteNotification = useCallback(async (id: number) => {
        try {
            await notificationApi.delete(id);
            setRecentNotifications(prev => prev.filter(n => n.id !== id));
            setUnreadCount(prev => {
                const deleted = recentNotifications.find(n => n.id === id);
                return deleted && !deleted.read ? Math.max(0, prev - 1) : prev;
            });
        } catch (err) {
            console.error('Failed to delete notification:', err);
        }
    }, [recentNotifications]);

    useEffect(() => {
        if (user) {
            refresh();
            const interval = setInterval(refresh, 30000);
            return () => clearInterval(interval);
        }
    }, [user, refresh]);

    return (
        <NotificationContext.Provider value={{
            unreadCount,
            recentNotifications,
            refresh,
            markAsRead,
            markAllAsRead,
            deleteNotification,
        }}>
            {children}
        </NotificationContext.Provider>
    );
};
