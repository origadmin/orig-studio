import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuLabel,
    DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, UserPlus, Users, Filter, Loader2, RotateCcw} from 'lucide-react';
import {formatDateTime} from '@/lib/format';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {adminApi, Channel} from '@/lib/api/admin';
import {TablePagination} from '@/components/common/TablePagination';
import {usePagination} from '@/hooks/usePagination';

const Channels: React.FC = () => {
    const {t} = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [channels, setChannels] = useState<Channel[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const {page, pageSize, total, setPage, setTotal, getParams} = usePagination();
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    const [currentChannel, setCurrentChannel] = useState<Channel | null>(null);
    const [formData, setFormData] = useState<Partial<Channel>>({
        name: '',
        handle: '',
        description: '',
        status: 'active',
    });

    // 加载频道数据
    useEffect(() => {
        loadChannels();
    }, [page]);

    const loadChannels = async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await adminApi.getChannels(getParams());
            const channelList = Array.isArray(response?.items) ? response.items : [];
            setChannels(channelList);
            if (response?.total !== undefined) {
                setTotal(response.total);
            }
        } catch (err) {
            setError('Failed to load channels');
            console.error('Error loading channels:', err);
        } finally {
            setLoading(false);
        }
    };

    const resetForm = () => {
        setFormData({
            name: '',
            handle: '',
            description: '',
            status: 'active',
        });
    };

    const handleCreate = async () => {
        try {
            await adminApi.createChannel(formData);
            await loadChannels();
            setShowCreateDialog(false);
            resetForm();
        } catch (err) {
            console.error('Failed to create channel:', err);
        }
    };

    const handleUpdate = async () => {
        if (!currentChannel) return;

        try {
            await adminApi.updateChannel(currentChannel.id, formData);
            await loadChannels();
            setShowEditDialog(false);
            resetForm();
            setCurrentChannel(null);
        } catch (err) {
            console.error('Failed to update channel:', err);
        }
    };

    const handleDelete = async () => {
        if (!currentChannel) return;

        try {
            await adminApi.deleteChannel(currentChannel.id);
            await loadChannels();
            setShowDeleteDialog(false);
            setCurrentChannel(null);
        } catch (err) {
            console.error('Failed to delete channel:', err);
        }
    };

    const openCreateDialog = () => {
        resetForm();
        setShowCreateDialog(true);
    };

    const openEditDialog = (channel: Channel) => {
        setCurrentChannel(channel);
        setFormData({
            name: channel.name,
            handle: channel.handle,
            description: channel.description,
            status: channel.status,
        });
        setShowEditDialog(true);
    };

    const openDeleteDialog = (channel: Channel) => {
        setCurrentChannel(channel);
        setShowDeleteDialog(true);
    };

    const filteredChannels = channels.filter(channel => {
        const matchesSearch = channel.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            channel.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || channel.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const totalSubscribers = channels.reduce((sum, c) => sum + (c.subscriber_count || 0), 0);
    const verifiedCount = channels.filter(c => c.status === 'verified').length;
    const pendingCount = channels.filter(c => c.status === 'pending').length;

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
            verified: 'default',
            active: 'secondary',
            pending: 'outline',
            banned: 'destructive',
        };
        const labels: Record<string, string> = {
            verified: t('common.verified'),
            active: t('admin.normal'),
            pending: t('admin.pending'),
            banned: t('admin.banned'),
        };
        return <Badge variant={variants[status] || 'outline'}>{labels[status] || status}</Badge>;
    };

    const formatNumber = (num: number | undefined | null) => {
        if (num === undefined || num === null) return '0';
        if (num >= 10000) return (num / 10000).toFixed(1) + t('common.wan');
        return num.toString();
    };

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.channels')}</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage your content channels
                                </p>
                            </div>
                        </div>

                        {/* 分隔线 */}
                        <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                        {/* 搜索和筛选 */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1 min-w-[120px] max-w-[400px]">
                                <div className="relative w-full">
                                    <Search
                                        className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                    <Input
                                        placeholder={t('admin.search') || t('admin.channels') + '...'}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={statusFilter} onValueChange={setStatusFilter}>
                                    <SelectTrigger className="w-[140px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {statusFilter === 'all' ? (
                                                <span className="text-muted-foreground">Status</span>
                                            ) : (
                                                <SelectValue placeholder="Status"/>
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                        <SelectItem value="verified">{t('common.verified')}</SelectItem>
                                        <SelectItem value="active">{t('admin.normal')}</SelectItem>
                                        <SelectItem value="pending">{t('admin.pending')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => {
                                            setSearchTerm('');
                                            setStatusFilter('all');
                                        }}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-2"/>
                                        Reset
                                    </Button>
                                    <Button
                                        variant="default"
                                        size="sm"
                                        onClick={() => {
                                        }}
                                    >
                                        <Search className="h-4 w-4 mr-2"/>
                                        Search
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 统计卡片 */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.channelTotal')}</p>
                                <p className="text-2xl font-bold text-info">{channels.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <Users className="w-6 h-6 text-info"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.totalSubscribers')}</p>
                                <p className="text-2xl font-bold text-purple-600">{formatNumber(totalSubscribers)}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <UserPlus className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-purple-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.verifiedChannels')}</p>
                                <p className="text-2xl font-bold text-success">{verifiedCount}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <Eye className="w-6 h-6 text-success"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.pending')}</p>
                                <p className="text-2xl font-bold text-yellow-600">{pendingCount}</p>
                            </div>
                            <div className="w-12 h-12 bg-yellow-100 rounded-xl flex items-center justify-center">
                                <Loader2 className="w-6 h-6 text-yellow-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-warning w-full opacity-10"/>
                    </CardContent>
                </Card>
            </div>

            {/* 频道表格 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.channelList')}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button size="sm" onClick={openCreateDialog}>
                                <Plus className="mr-2 h-4 w-4"/>
                                {t('admin.newChannel')}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>{t('admin.channel')}</TableHead>
                                <TableHead>{t('admin.owner')}</TableHead>
                                <TableHead className="text-right">{t('admin.subscriberCount')}</TableHead>
                                <TableHead className="text-right">{t('admin.videoCount')}</TableHead>
                                <TableHead>{t('admin.category')}</TableHead>
                                <TableHead>{t('admin.status')}</TableHead>
                                <TableHead>{t('admin.createdAt')}</TableHead>
                                <TableHead className="text-right">{t('admin.actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow key="loading">
                                    <TableCell colSpan={8} className="text-center py-8">
                                        <div className="animate-pulse">Loading channels...</div>
                                    </TableCell>
                                </TableRow>
                            ) : error ? (
                                <TableRow key="error">
                                    <TableCell colSpan={8} className="text-center py-8">
                                        <div className="text-destructive">{error}</div>
                                        <Button 
                                            variant="outline" 
                                            size="sm" 
                                            className="mt-2"
                                            onClick={() => window.location.reload()}
                                        >
                                            Retry
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ) : filteredChannels.length === 0 ? (
                                <TableRow key="empty">
                                    <TableCell colSpan={8} className="text-center py-8">
                                        No channels found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredChannels.map((channel) => (
                                    <TableRow key={channel.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-3">
                                                <Avatar className="h-10 w-10">
                                                    <AvatarFallback>{channel.name[0]}</AvatarFallback>
                                                </Avatar>
                                                <div>
                                                    <div className="font-medium">{channel.name}</div>
                                                    <div className="text-xs text-muted-foreground">@{channel.handle}</div>
                                                </div>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <Avatar className="h-6 w-6">
                                                    <AvatarFallback
                                                        className="text-xs">{channel.user_id.substring(0, 1).toUpperCase()}</AvatarFallback>
                                                </Avatar>
                                                <span className="text-muted-foreground">User ID: {channel.user_id}</span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-right font-medium">
                                            {formatNumber(channel.subscriber_count)}
                                        </TableCell>
                                        <TableCell className="text-right">{channel.media_count || 0}</TableCell>
                                        <TableCell>
                                            <Badge variant="outline">-</Badge>
                                        </TableCell>
                                        <TableCell>{getStatusBadge(channel.status)}</TableCell>
                                        <TableCell className="text-muted-foreground">
                                            {formatDateTime(channel.create_time)}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button 
                                                        variant="ghost" 
                                                        size="icon" 
                                                        className="h-6 w-6" 
                                                        title="More Actions"
                                                    >
                                                        <MoreHorizontal className="h-3 w-3"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem>
                                                        <Eye className="mr-2 h-4 w-4"/>
                                                        {t('admin.view')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => openEditDialog(channel)}>
                                                        <Edit className="mr-2 h-4 w-4"/>
                                                        {t('admin.edit')}
                                                    </DropdownMenuItem>
                                                    {channel.status === 'pending' && (
                                                        <DropdownMenuItem>
                                                            <UserPlus className="mr-2 h-4 w-4"/>
                                                            {t('admin.verify')}
                                                        </DropdownMenuItem>
                                                    )}
                                                    <DropdownMenuItem 
                                                        className="text-destructive focus:text-destructive" 
                                                        onClick={() => openDeleteDialog(channel)}
                                                    >
                                                        <Trash2 className="mr-2 h-4 w-4"/>
                                                        {t('admin.delete')}
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <TablePagination
                page={page}
                pageSize={pageSize}
                total={total}
                onPageChange={setPage}
            />

            {/* Create Channel Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.newChannel') || 'New Channel'}</DialogTitle>
                        <DialogDescription>
                            Create a new channel for your content
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Name *
                            </h4>
                            <Input
                                placeholder="Enter channel name"
                                value={formData.name || ''}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Handle *
                            </h4>
                            <Input
                                placeholder="Enter channel handle"
                                value={formData.handle || ''}
                                onChange={(e) => setFormData({...formData, handle: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Description
                            </h4>
                            <Textarea
                                placeholder="Enter channel description"
                                value={formData.description || ''}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Status
                            </h4>
                            <Select
                                value={formData.status || 'active'}
                                onValueChange={(value) => setFormData({...formData, status: value})}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select status"/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="active">Active</SelectItem>
                                    <SelectItem value="verified">Verified</SelectItem>
                                    <SelectItem value="pending">Pending</SelectItem>
                                    <SelectItem value="banned">Banned</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button onClick={handleCreate}>
                            {t('common.save') || 'Save'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Edit Channel Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.editChannel') || 'Edit Channel'}</DialogTitle>
                        <DialogDescription>
                            Update the channel information
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Name *
                            </h4>
                            <Input
                                placeholder="Enter channel name"
                                value={formData.name || ''}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Handle *
                            </h4>
                            <Input
                                placeholder="Enter channel handle"
                                value={formData.handle || ''}
                                onChange={(e) => setFormData({...formData, handle: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Description
                            </h4>
                            <Textarea
                                placeholder="Enter channel description"
                                value={formData.description || ''}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Status
                            </h4>
                            <Select
                                value={formData.status || 'active'}
                                onValueChange={(value) => setFormData({...formData, status: value})}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select status"/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="active">Active</SelectItem>
                                    <SelectItem value="verified">Verified</SelectItem>
                                    <SelectItem value="pending">Pending</SelectItem>
                                    <SelectItem value="banned">Banned</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button onClick={handleUpdate}>
                            {t('common.save') || 'Save'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Delete Channel Dialog */}
            <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t('admin.deleteChannel') || 'Delete Channel'}</AlertDialogTitle>
                        <AlertDialogDescription>
                            Are you sure you want to delete this channel? This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel onClick={() => setShowDeleteDialog(false)}>
                            {t('common.cancel') || 'Cancel'}
                        </AlertDialogCancel>
                        <AlertDialogAction 
                            onClick={handleDelete}
                            className="bg-red-600 hover:bg-red-700"
                        >
                            {t('common.delete') || 'Delete'}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
};

export default Channels;
