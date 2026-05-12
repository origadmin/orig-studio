import React from 'react';
import {Bell, Check, ExternalLink, Mail, MessageSquare, AtSign, Heart, UserPlus} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Popover, PopoverContent, PopoverTrigger} from '@/components/ui/popover';
import {ScrollArea} from '@/components/ui/scroll-area';
import {Separator} from '@/components/ui/separator';
import {useNotificationState} from '@/contexts/NotificationContext';
import {useAuth} from '@/hooks/useAuth';
import {useTranslation} from 'react-i18next';
import {formatDate} from '@/lib/format';

const actionIcons: Record<string, React.ReactNode> = {
    system: <MessageSquare className="w-4 h-4"/>,
    comment: <MessageSquare className="w-4 h-4"/>,
    like: <Heart className="w-4 h-4"/>,
    follow: <UserPlus className="w-4 h-4"/>,
    mention: <AtSign className="w-4 h-4"/>,
    email: <Mail className="w-4 h-4"/>,
};

const actionColors: Record<string, string> = {
    system: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400',
    comment: 'bg-blue-100 text-blue-600 dark:bg-blue-900 dark:text-blue-400',
    like: 'bg-rose-100 text-rose-600 dark:bg-rose-900 dark:text-rose-400',
    follow: 'bg-emerald-100 text-emerald-600 dark:bg-emerald-900 dark:text-emerald-400',
    mention: 'bg-amber-100 text-amber-600 dark:bg-amber-900 dark:text-amber-400',
    email: 'bg-violet-100 text-violet-600 dark:bg-violet-900 dark:text-violet-400',
};

const NotificationBadge: React.FC = () => {
    const {user} = useAuth();
    const {t} = useTranslation();
    const {unreadCount, recentNotifications, markAsRead} = useNotificationState();

    if (!user) return null;

    return (
        <Popover>
            <PopoverTrigger asChild>
                <Button variant="ghost" className="relative h-9 w-9 p-0 rounded-full hover:bg-accent">
                    <Bell className="h-[18px] w-[18px]"/>
                    {unreadCount > 0 && (
                        <span className="absolute -top-0.5 -right-0.5 min-h-[16px] min-w-[16px] px-1 rounded-full bg-red-500 text-white text-[10px] font-medium flex items-center justify-center">
                            {unreadCount > 99 ? '99+' : unreadCount}
                        </span>
                    )}
                </Button>
            </PopoverTrigger>
            <PopoverContent align="end" className="w-96 p-0 rounded-xl shadow-lg border dark:border-gray-700" sideOffset={8}>
                <div className="px-4 py-3 flex items-center justify-between bg-gradient-to-r from-slate-50 to-slate-100 dark:from-gray-800 dark:to-gray-900 rounded-t-xl">
                    <div className="flex items-center gap-2">
                        <Bell className="w-4 h-4 text-muted-foreground"/>
                        <span className="font-semibold text-sm">{t('notifications.title')}</span>
                    </div>
                    {unreadCount > 0 && (
                        <span className="text-xs font-medium bg-red-500 text-white px-2 py-0.5 rounded-full">
                            {unreadCount}
                        </span>
                    )}
                </div>
                <Separator/>
                {recentNotifications.length === 0 ? (
                    <div className="px-4 py-8 text-center">
                        <Bell className="w-8 h-8 mx-auto text-muted-foreground/40 mb-2"/>
                        <p className="text-sm text-muted-foreground">{t('notifications.noNotifications')}</p>
                    </div>
                ) : (
                    <ScrollArea className="max-h-[400px]">
                        <div className="divide-y divide-border">
                            {recentNotifications.map(notification => (
                                <div
                                    key={notification.id}
                                    className={`px-4 py-3 flex items-start gap-3 transition-colors hover:bg-accent/50 ${!notification.read ? 'bg-blue-50/50 dark:bg-blue-950/30' : ''}`}
                                >
                                    <div className={`mt-0.5 p-1.5 rounded-lg flex-shrink-0 ${actionColors[notification.action] || actionColors.system}`}>
                                        {actionIcons[notification.action] || actionIcons.system}
                                    </div>
                                    <div className="flex-1 min-w-0 space-y-1">
                                        <div className="flex items-center gap-2">
                                            {!notification.read && (
                                                <span className="h-1.5 w-1.5 rounded-full bg-blue-500 flex-shrink-0"/>
                                            )}
                                            <p className={`text-sm leading-tight ${notification.read ? 'text-muted-foreground' : 'font-medium text-foreground'}`}>
                                                {notification.title}
                                            </p>
                                        </div>
                                        <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                                            {notification.body}
                                        </p>
                                        <p className="text-[10px] text-muted-foreground/70">
                                            {formatDate(notification.create_time)}
                                        </p>
                                    </div>
                                    {!notification.read && (
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            className="h-6 px-2 text-[10px] flex-shrink-0 mt-1 rounded-full border-blue-200 text-blue-600 hover:bg-blue-50 dark:border-blue-800 dark:text-blue-400 dark:hover:bg-blue-950"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                markAsRead(notification.id);
                                            }}
                                        >
                                            <Check className="w-3 h-3 mr-0.5"/>
                                            {t('notifications.markAsRead')}
                                        </Button>
                                    )}
                                </div>
                            ))}
                        </div>
                    </ScrollArea>
                )}
                <Separator/>
                <div className="p-2">
                    <a
                        href="/me/notifications"
                        className="flex items-center justify-center gap-1.5 text-sm font-medium text-primary hover:text-primary/80 py-1.5 rounded-lg hover:bg-accent/50 transition-colors"
                    >
                        {t('notifications.viewAll')}
                        <ExternalLink className="w-3.5 h-3.5"/>
                    </a>
                </div>
            </PopoverContent>
        </Popover>
    );
};

export default NotificationBadge;
