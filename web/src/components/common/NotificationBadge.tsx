import React, {useState, useEffect} from 'react';
import {Bell} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger} from '@/components/ui/dropdown-menu';
import {notificationApi} from '@/lib/api/notification';
import {useAuth} from '@/hooks/useAuth';

const NotificationBadge: React.FC = () => {
    const {user} = useAuth();
    const [unreadCount, setUnreadCount] = useState(0);

    useEffect(() => {
        let isMounted = true;
        
        const fetchUnreadCount = async () => {
            if (!user) return;
            try {
                const response = await notificationApi.getUnreadCount();
                if (isMounted) {
                    setUnreadCount(response.unread_count || 0);
                }
            } catch (err) {
                console.error('Failed to fetch unread count:', err);
            }
        };

        if (user) {
            fetchUnreadCount();
            // 30秒刷新一次
            const interval = setInterval(fetchUnreadCount, 30000);
            return () => {
                isMounted = false;
                clearInterval(interval);
            };
        }
    }, [user]);

    if (!user) return null;

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-10 w-10 p-0">
                    <Bell className="h-5 w-5"/>
                    {unreadCount > 0 && (
                        <span
                            className="absolute top-0 right-0 h-4 w-4 rounded-full bg-blue-600 text-white text-[10px] flex items-center justify-center">
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
                    )}
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-80">
                <DropdownMenuItem>
                    <a href="/me/notifications" className="w-full">
                        View all notifications
                    </a>
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
};

export default NotificationBadge;
