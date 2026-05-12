import React, {useState, useEffect} from 'react';
import {Bell, Send, Settings, History, Loader2, Mail, MessageSquare, X, ChevronDown, Search} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Tabs, TabsContent, TabsList, TabsTrigger} from '@/components/ui/tabs';
import {notificationApi, type Notification} from '@/lib/api/notification';
import {adminUserApi, type User} from '@/lib/api/user';
import {formatDate} from '@/lib/format';
import {Spinner} from '@/components/ui/spinner';

const AdminNotifications: React.FC = () => {
    const {t} = useTranslation();
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);
    const [stats, setStats] = useState({total: 0, read: 0, unread: 0});
    const [form, setForm] = useState({
        action: 'system',
        title: '',
        body: '',
        method: 'in_app',
        notify: true,
        sendToAll: false,
    });
    const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
    const [users, setUsers] = useState<User[]>([]);
    const [userSearch, setUserSearch] = useState('');
    const [showUserPicker, setShowUserPicker] = useState(false);
    const [loadingUsers, setLoadingUsers] = useState(false);

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        try {
            setLoading(true);
            const response = await notificationApi.getAll({page_size: 50});
            const items = (response as any)?.items || response || [];
            setNotifications(items);
            const readCount = items.filter((n: Notification) => n.read).length;
            setStats({total: items.length, read: readCount, unread: items.length - readCount});
        } catch (err) {
            console.error('Failed to fetch notifications:', err);
        } finally {
            setLoading(false);
        }
    };

    const fetchUsers = async (keyword?: string) => {
        try {
            setLoadingUsers(true);
            const response = await adminUserApi.list({page_size: 50, keyword});
            const items = (response as any)?.items || response || [];
            setUsers(items);
        } catch (err) {
            console.error('Failed to fetch users:', err);
        } finally {
            setLoadingUsers(false);
        }
    };

    const handleOpenUserPicker = () => {
        setShowUserPicker(true);
        fetchUsers();
    };

    const toggleUser = (userId: string) => {
        setSelectedUserIds(prev =>
            prev.includes(userId) ? prev.filter(id => id !== userId) : [...prev, userId]
        );
    };

    const removeUser = (userId: string) => {
        setSelectedUserIds(prev => prev.filter(id => id !== userId));
    };

    const handleSend = async () => {
        if (!form.title || !form.body) return;
        try {
            setSending(true);
            if (form.sendToAll) {
                await notificationApi.create({
                    action: form.action,
                    title: form.title,
                    body: form.body,
                    method: form.method,
                    notify: form.notify,
                });
            } else if (selectedUserIds.length > 0) {
                for (const userId of selectedUserIds) {
                    await notificationApi.create({
                        action: form.action,
                        title: form.title,
                        body: form.body,
                        user_id: userId,
                        method: form.method,
                        notify: form.notify,
                    });
                }
            }
            setForm({action: 'system', title: '', body: '', method: 'in_app', notify: true, sendToAll: false});
            setSelectedUserIds([]);
            setShowUserPicker(false);
            fetchData();
        } catch (err) {
            console.error('Failed to send notification:', err);
        } finally {
            setSending(false);
        }
    };

    const handleDelete = async (id: number) => {
        try {
            await notificationApi.delete(id);
            setNotifications(prev => prev.filter(n => n.id !== id));
            setStats(prev => ({...prev, total: prev.total - 1}));
        } catch (err) {
            console.error('Failed to delete notification:', err);
        }
    };

    const selectedUsers = users.filter(u => selectedUserIds.includes(u.id));

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[200px]">
                <Spinner/>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                    <Bell className="w-6 h-6"/>
                    {t('admin.notifications')}
                </h1>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                                <Send className="w-5 h-5 text-blue-600 dark:text-blue-400"/>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">{t('notifications.totalSent')}</p>
                                <p className="text-2xl font-bold">{stats.total}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
                                <MessageSquare className="w-5 h-5 text-green-600 dark:text-green-400"/>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">{t('notifications.totalRead')}</p>
                                <p className="text-2xl font-bold">{stats.read}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-orange-100 dark:bg-orange-900 rounded-lg">
                                <Bell className="w-5 h-5 text-orange-600 dark:text-orange-400"/>
                            </div>
                            <div>
                                <p className="text-sm text-muted-foreground">{t('notifications.unreadCount', {count: stats.unread})}</p>
                                <p className="text-2xl font-bold">{stats.unread}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            <Tabs defaultValue="send" className="space-y-4">
                <TabsList>
                    <TabsTrigger value="send" className="gap-1.5">
                        <Send className="w-4 h-4"/>
                        {t('notifications.sendNotification')}
                    </TabsTrigger>
                    <TabsTrigger value="history" className="gap-1.5">
                        <History className="w-4 h-4"/>
                        {t('notifications.notificationHistory')}
                    </TabsTrigger>
                    <TabsTrigger value="config" className="gap-1.5">
                        <Settings className="w-4 h-4"/>
                        {t('notifications.config')}
                    </TabsTrigger>
                </TabsList>

                <TabsContent value="send">
                    <Card>
                        <CardHeader>
                            <CardTitle>{t('notifications.sendNotification')}</CardTitle>
                            <CardDescription>{t('notifications.batchSendDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-5">
                            <div>
                                <label className="block text-sm font-medium mb-1.5">{t('notifications.actionType')}</label>
                                <select
                                    value={form.action}
                                    onChange={e => setForm(f => ({...f, action: e.target.value}))}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600"
                                >
                                    <option value="system">System</option>
                                    <option value="comment">Comment</option>
                                    <option value="like">Like</option>
                                    <option value="follow">Follow</option>
                                    <option value="mention">Mention</option>
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1.5">{t('admin.title')} *</label>
                                <input
                                    type="text"
                                    value={form.title}
                                    onChange={e => setForm(f => ({...f, title: e.target.value}))}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600"
                                    placeholder={t('notifications.create')}
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium mb-1.5">{t('admin.description')} *</label>
                                <textarea
                                    value={form.body}
                                    onChange={e => setForm(f => ({...f, body: e.target.value}))}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600"
                                    rows={3}
                                    placeholder={t('notifications.create')}
                                />
                            </div>

                            <div className="space-y-3">
                                <label className="block text-sm font-medium">{t('notifications.targetUser')}</label>
                                <div className="flex items-center gap-2">
                                    <label className="flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="checkbox"
                                            checked={form.sendToAll}
                                            onChange={e => setForm(f => ({...f, sendToAll: e.target.checked}))}
                                            className="rounded"
                                        />
                                        <span className="text-sm">{t('notifications.sendToAll')}</span>
                                    </label>
                                </div>
                                {!form.sendToAll && (
                                    <div className="space-y-2">
                                        <div className="flex gap-2">
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={handleOpenUserPicker}
                                                className="gap-1"
                                            >
                                                <Search className="w-3.5 h-3.5"/>
                                                {t('notifications.selectUsers')}
                                            </Button>
                                            {selectedUserIds.length > 0 && (
                                                <span className="text-sm text-muted-foreground self-center">
                                                    {t('notifications.selectUsers')}: {selectedUserIds.length}
                                                </span>
                                            )}
                                        </div>
                                        {selectedUsers.length > 0 && (
                                            <div className="flex flex-wrap gap-1.5">
                                                {selectedUsers.map(user => (
                                                    <span
                                                        key={user.id}
                                                        className="inline-flex items-center gap-1 px-2 py-1 text-xs bg-primary/10 text-primary rounded-full"
                                                    >
                                                        {user.nickname || user.username}
                                                        <button
                                                            onClick={() => removeUser(user.id)}
                                                            className="hover:text-destructive"
                                                        >
                                                            <X className="w-3 h-3"/>
                                                        </button>
                                                    </span>
                                                ))}
                                            </div>
                                        )}
                                        {showUserPicker && (
                                            <div className="border rounded-lg p-3 space-y-2 dark:border-gray-600">
                                                <input
                                                    type="text"
                                                    value={userSearch}
                                                    onChange={e => {
                                                        setUserSearch(e.target.value);
                                                        fetchUsers(e.target.value);
                                                    }}
                                                    className="w-full px-3 py-1.5 text-sm border rounded dark:bg-gray-800 dark:border-gray-600"
                                                    placeholder={t('notifications.selectUsers') + '...'}
                                                />
                                                {loadingUsers ? (
                                                    <div className="text-center py-2 text-sm text-muted-foreground">
                                                        <Loader2 className="w-4 h-4 animate-spin inline"/>
                                                    </div>
                                                ) : (
                                                    <div className="max-h-40 overflow-y-auto space-y-1">
                                                        {users
                                                            .filter(u => !userSearch || u.username.includes(userSearch) || (u.nickname || '').includes(userSearch))
                                                            .map(user => (
                                                                <label
                                                                    key={user.id}
                                                                    className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-accent cursor-pointer"
                                                                >
                                                                    <input
                                                                        type="checkbox"
                                                                        checked={selectedUserIds.includes(user.id)}
                                                                        onChange={() => toggleUser(user.id)}
                                                                        className="rounded"
                                                                    />
                                                                    <span className="text-sm">{user.nickname || user.username}</span>
                                                                    <span className="text-xs text-muted-foreground">{user.email}</span>
                                                                </label>
                                                            ))}
                                                    </div>
                                                )}
                                                <Button variant="ghost" size="sm" onClick={() => setShowUserPicker(false)} className="w-full">
                                                    {t('common.cancel') || 'Close'}
                                                </Button>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>

                            <div>
                                <label className="block text-sm font-medium mb-1.5">{t('notifications.deliveryMethod')}</label>
                                <select
                                    value={form.method}
                                    onChange={e => setForm(f => ({...f, method: e.target.value}))}
                                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600"
                                >
                                    <option value="in_app">{t('notifications.inApp')}</option>
                                    <option value="email">{t('notifications.email')}</option>
                                </select>
                            </div>

                            <div className="flex items-center gap-2">
                                <input
                                    type="checkbox"
                                    id="notifyPush"
                                    checked={form.notify}
                                    onChange={e => setForm(f => ({...f, notify: e.target.checked}))}
                                    className="rounded"
                                />
                                <label htmlFor="notifyPush" className="text-sm">{t('notifications.notifyOption')}</label>
                            </div>
                            <div className="flex gap-2">
                                <Button onClick={handleSend} disabled={sending || !form.title || !form.body || (!form.sendToAll && selectedUserIds.length === 0)}>
                                    {sending ? <Loader2 className="w-4 h-4 mr-2 animate-spin"/> : <Send className="w-4 h-4 mr-2"/>}
                                    {t('notifications.sendNotification')}
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="history">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <History className="w-5 h-5"/>
                                {t('notifications.notificationHistory')} ({notifications.length})
                            </CardTitle>
                        </CardHeader>
                        <CardContent>
                            {notifications.length === 0 ? (
                                <div className="text-center py-12 text-gray-500 dark:text-muted-foreground">
                                    {t('notifications.noNotifications')}
                                </div>
                            ) : (
                                <div className="space-y-3">
                                    {notifications.map(notification => (
                                        <div
                                            key={notification.id}
                                            className={`p-4 rounded-lg border ${notification.read ? 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800' : 'border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/30'}`}
                                        >
                                            <div className="flex items-start justify-between gap-4">
                                                <div className="flex-1 space-y-1">
                                                    <div className="flex items-center gap-2">
                                                        <h4 className="font-medium text-gray-900 dark:text-white">
                                                            {notification.title}
                                                        </h4>
                                                        <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${notification.read ? 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200' : 'bg-blue-100 text-blue-800 dark:bg-blue-800 dark:text-blue-200'}`}>
                                                            {notification.read ? t('notifications.read') : t('notifications.unread')}
                                                        </span>
                                                    </div>
                                                    <p className="text-sm text-gray-600 dark:text-gray-300">
                                                        {notification.body}
                                                    </p>
                                                    <div className="flex items-center gap-3 text-xs text-gray-500 dark:text-muted-foreground">
                                                        <span>{t('notifications.actionType')}: {notification.action}</span>
                                                        <span>{t('notifications.targetUser')}: {notification.user_id}</span>
                                                        <span>{formatDate(notification.create_time)}</span>
                                                    </div>
                                                </div>
                                                <Button
                                                    variant="ghost"
                                                    size="sm"
                                                    className="h-8 w-8 p-0 text-destructive dark:text-red-400"
                                                    onClick={() => handleDelete(notification.id)}
                                                    title={t('common.delete')}
                                                >
                                                    <span className="text-lg leading-none">&times;</span>
                                                </Button>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="config">
                    <Card>
                        <CardHeader>
                            <CardTitle>{t('notifications.config')}</CardTitle>
                            <CardDescription>{t('notifications.configDesc')}</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            <div className="space-y-4">
                                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">{t('notifications.deliveryMethod')}</h3>
                                <div className="flex items-center justify-between py-2">
                                    <div className="flex items-center gap-3">
                                        <MessageSquare className="w-5 h-5 text-muted-foreground"/>
                                        <div>
                                            <p className="font-medium">{t('notifications.enableInApp')}</p>
                                            <p className="text-sm text-muted-foreground">{t('notifications.inApp')}</p>
                                        </div>
                                    </div>
                                    <input type="checkbox" defaultChecked className="rounded"/>
                                </div>
                                <div className="flex items-center justify-between py-2">
                                    <div className="flex items-center gap-3">
                                        <Mail className="w-5 h-5 text-muted-foreground"/>
                                        <div>
                                            <p className="font-medium">{t('notifications.enableEmail')}</p>
                                            <p className="text-sm text-muted-foreground">{t('notifications.email')}</p>
                                        </div>
                                    </div>
                                    <input type="checkbox" className="rounded"/>
                                </div>
                                <div className="flex items-center justify-between py-2">
                                    <div className="flex items-center gap-3">
                                        <Bell className="w-5 h-5 text-muted-foreground"/>
                                        <div>
                                            <p className="font-medium">{t('notifications.enablePush')}</p>
                                            <p className="text-sm text-muted-foreground">{t('notifications.notifyOption')}</p>
                                        </div>
                                    </div>
                                    <input type="checkbox" className="rounded"/>
                                </div>
                            </div>
                            <div className="space-y-4">
                                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">{t('admin.actions')}</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium mb-1">{t('notifications.defaultMethod')}</label>
                                        <select className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600">
                                            <option value="in_app">{t('notifications.inApp')}</option>
                                            <option value="email">{t('notifications.email')}</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium mb-1">{t('notifications.retentionDays')}</label>
                                        <input
                                            type="number"
                                            defaultValue={90}
                                            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600"
                                        />
                                    </div>
                                </div>
                            </div>
                            <div className="flex gap-2">
                                <Button>{t('common.save')}</Button>
                            </div>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
};

export default AdminNotifications;
