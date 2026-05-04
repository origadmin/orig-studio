import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
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
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from '@/components/ui/dialog';
import {MoreHorizontal, Search, Edit, Trash2, Eye, PlayCircle, Lock, Globe, User, Filter, RotateCcw, Plus, Loader2} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {adminPlaylistApi, playlistApi, Playlist} from '@/lib/api/playlist';
import {formatDateTime} from '@/lib/format';
import {extractList} from '@/lib/extract';
import {TablePagination} from '@/components/common/TablePagination';
import {usePagination} from '@/hooks/usePagination';
import {Link} from '@tanstack/react-router';

const Playlists: React.FC = () => {
    const {t} = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [visibilityFilter, setVisibilityFilter] = useState('all');
    const [playlists, setPlaylists] = useState<Playlist[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const {page, pageSize, total, setPage, setTotal, getParams} = usePagination();

    // Create dialog state
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [createTitle, setCreateTitle] = useState('');
    const [createDescription, setCreateDescription] = useState('');
    const [createUserId, setCreateUserId] = useState('');
    const [createIsPublic, setCreateIsPublic] = useState(true);
    const [isCreating, setIsCreating] = useState(false);

    // Edit dialog state
    const [editTarget, setEditTarget] = useState<Playlist | null>(null);
    const [editTitle, setEditTitle] = useState('');
    const [editDescription, setEditDescription] = useState('');
    const [editIsPublic, setEditIsPublic] = useState(true);
    const [isUpdating, setIsUpdating] = useState(false);

    // Delete dialog state
    const [deleteTarget, setDeleteTarget] = useState<Playlist | null>(null);
    const [isDeleting, setIsDeleting] = useState(false);

    // Load playlists
    const loadPlaylists = async () => {
        setLoading(true);
        setError(null);
        try {
            const response = await adminPlaylistApi.list(getParams());
            const playlistList = extractList<Playlist>(response);
            setPlaylists(playlistList);
            if ((response as any)?.total !== undefined) {
                setTotal((response as any).total);
            }
        } catch (err) {
            setError('Failed to load playlists');
            console.error('Error loading playlists:', err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadPlaylists();
    }, [page]);

    const filteredPlaylists = playlists.filter(playlist => {
        const matchesSearch = playlist.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
            (playlist.description && playlist.description.toLowerCase().includes(searchTerm.toLowerCase()));
        return matchesSearch;
    });

    const totalPlaylists = playlists.length;
    const publicCount = playlists.filter(p => p.is_public).length;
    const privateCount = playlists.filter(p => !p.is_public).length;
    const totalViews = 0;

    const getVisibilityBadge = (visibility: 'public' | 'private' | 'unlisted' | string) => {
        const configs = {
            public: {icon: Globe, label: t('admin.pub'), variant: 'default' as const},
            private: {icon: Lock, label: t('admin.priv'), variant: 'secondary' as const},
            unlisted: {icon: Eye, label: t('admin.unlisted'), variant: 'outline' as const},
        };
        const config = configs[visibility as keyof typeof configs] || configs.public;
        const Icon = config.icon;
        return (
            <Badge variant={config.variant}>
                <Icon className="mr-1 h-3 w-3"/>
                {config.label}
            </Badge>
        );
    };

    const formatNumber = (num: number | undefined | null) => {
        if (num === undefined || num === null) return '0';
        if (num >= 10000) return (num / 10000).toFixed(1) + t('common.wan');
        return num.toString();
    };

    const handleCreate = async () => {
        if (!createTitle.trim() || !createUserId.trim()) return;
        try {
            setIsCreating(true);
            await adminPlaylistApi.create({
                title: createTitle.trim(),
                description: createDescription.trim(),
                user_id: createUserId.trim(),
                is_public: createIsPublic,
            });
            setShowCreateDialog(false);
            setCreateTitle('');
            setCreateDescription('');
            setCreateUserId('');
            setCreateIsPublic(true);
            loadPlaylists();
        } catch (err) {
            console.error('Failed to create playlist:', err);
        } finally {
            setIsCreating(false);
        }
    };

    const handleEdit = (playlist: Playlist) => {
        setEditTarget(playlist);
        setEditTitle(playlist.title);
        setEditDescription(playlist.description || '');
        setEditIsPublic(playlist.is_public);
    };

    const handleSaveEdit = async () => {
        if (!editTarget) return;
        try {
            setIsUpdating(true);
            await adminPlaylistApi.update(editTarget.id, {
                title: editTitle,
                description: editDescription,
                is_public: editIsPublic,
            });
            setEditTarget(null);
            loadPlaylists();
        } catch (err) {
            console.error('Failed to update playlist:', err);
        } finally {
            setIsUpdating(false);
        }
    };

    const handleDelete = async () => {
        if (!deleteTarget) return;
        try {
            setIsDeleting(true);
            await adminPlaylistApi.delete(deleteTarget.id);
            setDeleteTarget(null);
            loadPlaylists();
        } catch (err) {
            console.error('Failed to delete playlist:', err);
        } finally {
            setIsDeleting(false);
        }
    };

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* Action bar */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.playlists')}</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage your playlists
                                </p>
                            </div>
                        </div>

                        <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1 min-w-[120px] max-w-[400px]">
                                <div className="relative w-full">
                                    <Search
                                        className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                    <Input
                                        placeholder={t('admin.search') || t('admin.playlists') + '...'}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={visibilityFilter} onValueChange={setVisibilityFilter}>
                                    <SelectTrigger className="w-[140px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {visibilityFilter === 'all' ? (
                                                <span className="text-muted-foreground">Visibility</span>
                                            ) : (
                                                <SelectValue placeholder="Visibility"/>
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                        <SelectItem value="public">{t('admin.pub')}</SelectItem>
                                        <SelectItem value="private">{t('admin.priv')}</SelectItem>
                                        <SelectItem value="unlisted">{t('admin.unlisted')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => {
                                            setSearchTerm('');
                                            setVisibilityFilter('all');
                                        }}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-2"/>
                                        Reset
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Stats cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <PlayCircle className="h-5 w-5 text-indigo-600"/>
                            <div>
                                <div className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">{totalPlaylists}</div>
                                <p className="text-sm text-muted-foreground">{t('admin.playlistTotal')}</p>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-indigo-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-success dark:text-green-400">{publicCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.publicLists')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{privateCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.privateLists')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-warning w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-cyan-600 dark:text-cyan-400">{formatNumber(totalViews)}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalViews')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-cyan-500 w-full opacity-10"/>
                </Card>
            </div>

            {/* Playlist table */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.playlistList')}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button size="sm" onClick={() => setShowCreateDialog(true)}>
                                <Plus className="mr-2 h-4 w-4"/>
                                {t('admin.newPlaylist')}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>{t('admin.name')}</TableHead>
                                <TableHead>{t('admin.creator')}</TableHead>
                                <TableHead className="text-right">{t('admin.videoCount')}</TableHead>
                                <TableHead>{t('admin.visibility')}</TableHead>
                                <TableHead>{t('admin.createdAt')}</TableHead>
                                <TableHead className="text-right">{t('admin.actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8">
                                        <div className="animate-pulse">Loading playlists...</div>
                                    </TableCell>
                                </TableRow>
                            ) : error ? (
                                <TableRow>
                                    <TableCell colSpan={7} className="text-center py-8">
                                        <div className="text-destructive">{error}</div>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            className="mt-2"
                                            onClick={loadPlaylists}
                                        >
                                            Retry
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ) : filteredPlaylists.length === 0 ? (
                                <TableRow key="empty">
                                    <TableCell colSpan={7} className="text-center py-8">
                                        No playlists found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredPlaylists.map((playlist) => (
                                    <TableRow key={playlist.id}>
                                        <TableCell className="font-medium text-xs">{playlist.id?.substring(0, 8)}...</TableCell>
                                        <TableCell>
                                            <div>
                                                <div className="font-medium">{playlist.title}</div>
                                                {playlist.description && (
                                                    <div className="text-xs text-muted-foreground">{playlist.description}</div>
                                                )}
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="flex items-center gap-2">
                                                <Avatar className="h-6 w-6">
                                                    <AvatarFallback className="text-xs">
                                                        <User className="h-3 w-3"/>
                                                    </AvatarFallback>
                                                </Avatar>
                                                <span className="text-muted-foreground text-xs">User ID: {playlist.user_id?.substring(0, 8)}...</span>
                                            </div>
                                        </TableCell>
                                        <TableCell className="text-right">{playlist.media_items?.length || 0}</TableCell>
                                        <TableCell>
                                            {getVisibilityBadge(playlist.is_public ? 'public' : 'private')}
                                        </TableCell>
                                        <TableCell className="text-muted-foreground">
                                            {formatDateTime(playlist.create_time)}
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
                                                    <DropdownMenuItem asChild>
                                                        <Link to="/playlist/$token" params={{token: playlist.short_token || playlist.id}}>
                                                            <Eye className="mr-2 h-4 w-4"/>
                                                            {t('admin.view')}
                                                        </Link>
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => handleEdit(playlist)}>
                                                        <Edit className="mr-2 h-4 w-4"/>
                                                        {t('admin.edit')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setDeleteTarget(playlist)}>
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

            {/* Create Playlist Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('admin.newPlaylist') || 'Create Playlist'}</DialogTitle>
                        <DialogDescription>Create a new playlist for a user</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 mt-4">
                        <div>
                            <label className="text-sm font-medium mb-1 block">Title *</label>
                            <Input value={createTitle} onChange={(e) => setCreateTitle(e.target.value)}
                                   placeholder="Enter playlist title"/>
                        </div>
                        <div>
                            <label className="text-sm font-medium mb-1 block">Description</label>
                            <Input value={createDescription} onChange={(e) => setCreateDescription(e.target.value)}
                                   placeholder="Enter playlist description"/>
                        </div>
                        <div>
                            <label className="text-sm font-medium mb-1 block">User ID *</label>
                            <Input value={createUserId} onChange={(e) => setCreateUserId(e.target.value)}
                                   placeholder="Enter user ID (UUID)"/>
                        </div>
                        <div className="flex items-center gap-2">
                            <input type="checkbox" id="create-is-public" checked={createIsPublic}
                                   onChange={(e) => setCreateIsPublic(e.target.checked)} className="rounded"/>
                            <label htmlFor="create-is-public" className="text-sm">Public</label>
                        </div>
                        <div className="flex justify-end gap-2">
                            <Button variant="outline" onClick={() => setShowCreateDialog(false)} disabled={isCreating}>Cancel</Button>
                            <Button onClick={handleCreate} disabled={!createTitle.trim() || !createUserId.trim() || isCreating}>
                                {isCreating ? <Loader2 className="w-4 h-4 mr-1 animate-spin"/> : <Plus className="w-4 h-4 mr-1"/>}
                                Create
                            </Button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Edit Playlist Dialog */}
            <Dialog open={!!editTarget} onOpenChange={() => setEditTarget(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Edit Playlist</DialogTitle>
                        <DialogDescription>Update playlist details</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 mt-4">
                        <div>
                            <label className="text-sm font-medium mb-1 block">Title</label>
                            <Input value={editTitle} onChange={(e) => setEditTitle(e.target.value)}/>
                        </div>
                        <div>
                            <label className="text-sm font-medium mb-1 block">Description</label>
                            <Input value={editDescription} onChange={(e) => setEditDescription(e.target.value)}/>
                        </div>
                        <div className="flex items-center gap-2">
                            <input type="checkbox" id="edit-is-public" checked={editIsPublic}
                                   onChange={(e) => setEditIsPublic(e.target.checked)} className="rounded"/>
                            <label htmlFor="edit-is-public" className="text-sm">Public</label>
                        </div>
                        <div className="flex justify-end gap-2">
                            <Button variant="outline" onClick={() => setEditTarget(null)} disabled={isUpdating}>Cancel</Button>
                            <Button onClick={handleSaveEdit} disabled={isUpdating}>
                                {isUpdating ? <Loader2 className="w-4 h-4 mr-1 animate-spin"/> : null}
                                Save
                            </Button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Delete Playlist Dialog */}
            <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Delete Playlist</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to delete "{deleteTarget?.title}"? This action cannot be undone.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-2 mt-4">
                        <Button variant="outline" onClick={() => setDeleteTarget(null)} disabled={isDeleting}>Cancel</Button>
                        <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
                            {isDeleting ? <Loader2 className="w-4 h-4 mr-1 animate-spin"/> : null}
                            Delete
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default Playlists;
