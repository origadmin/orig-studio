/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * My Playlists Page - lists the current user's playlists with create/delete functionality.
 */

import React, {useState} from 'react';
import {ListVideo, Plus, Play, Video, Trash2, Globe, Lock} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useQuery, useQueryClient, useMutation} from '@tanstack/react-query';
import {playlistApi, type Playlist} from '@/lib/api/playlist';
import {formatDate} from '@/lib/format';
import {Spinner} from '@/components/ui/spinner';
import {Link} from '@tanstack/react-router';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from '@/components/ui/dialog';

const PlaylistsPage = () => {
    const {t} = useTranslation();
    const {user} = useAuth();
    const queryClient = useQueryClient();

    // Create dialog state
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [newTitle, setNewTitle] = useState('');
    const [newDescription, setNewDescription] = useState('');
    const [newIsPublic, setNewIsPublic] = useState(true);
    const [isCreating, setIsCreating] = useState(false);

    // Delete dialog state
    const [deleteTarget, setDeleteTarget] = useState<Playlist | null>(null);
    const [isDeleting, setIsDeleting] = useState(false);

    const {data, isLoading, error} = useQuery({
        queryKey: ['playlists', user?.id],
        queryFn: async () => {
            if (!user) throw new Error('User not logged in');
            const response = await playlistApi.getMyPlaylists();
            return response.items || [];
        },
        enabled: !!user
    });

    const playlists: Playlist[] = data || [];

    const handleCreate = async () => {
        if (!newTitle.trim()) return;
        try {
            setIsCreating(true);
            await playlistApi.create({title: newTitle.trim(), description: newDescription.trim(), is_public: newIsPublic});
            setShowCreateDialog(false);
            setNewTitle('');
            setNewDescription('');
            setNewIsPublic(true);
            queryClient.invalidateQueries({queryKey: ['playlists', user?.id]});
        } catch (err) {
            console.error('Failed to create playlist:', err);
        } finally {
            setIsCreating(false);
        }
    };

    const handleDelete = async () => {
        if (!deleteTarget) return;
        try {
            setIsDeleting(true);
            await playlistApi.delete(deleteTarget.id);
            setDeleteTarget(null);
            queryClient.invalidateQueries({queryKey: ['playlists', user?.id]});
        } catch (err) {
            console.error('Failed to delete playlist:', err);
        } finally {
            setIsDeleting(false);
        }
    };

    const visibilityLabel = (v: string) => {
        const map: Record<string, string> = {
            public: t('common.public'),
            private: t('common.private'),
            unlisted: t('common.unlisted')
        };
        return map[v] || v;
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner/>
            </div>
        );
    }

    if (error || !user) {
        return (
            <div className="text-center py-20 text-muted-foreground">
                <ListVideo size={48} className="mx-auto mb-3 opacity-30"/>
                <p className="text-lg mb-1">{t('playlists.empty')}</p>
                <p className="text-sm">{t('playlists.emptyDesc')}</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Title + Create button */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <ListVideo size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('playlists.title')}</h1>
                    <span className="text-sm text-gray-500">{t('playlists.listCount', {count: playlists.length})}</span>
                </div>
                <Button
                    onClick={() => setShowCreateDialog(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 transition-colors">
                    <Plus size={16}/> {t('playlists.newList')}
                </Button>
            </div>

            {/* Playlist cards */}
            {playlists.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                    {playlists.map((pl) => (
                        <div
                            key={pl.id}
                            className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-100 dark:border-gray-700 hover:shadow-lg transition-all group relative"
                        >
                            {/* Cover - clickable to detail */}
                            <Link to="/playlist/$token" params={{token: pl.short_token || pl.id}}>
                                <div className="relative aspect-video overflow-hidden bg-gray-100 dark:bg-gray-700 cursor-pointer">
                                    <div className="absolute inset-0 flex items-center justify-center">
                                        <ListVideo size={48} className="text-gray-300 dark:text-gray-600"/>
                                    </div>
                                    <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent"/>
                                    <div className="absolute bottom-3 left-3 flex items-center gap-2">
                                        <Video size={14} className="text-white/80"/>
                                        <span
                                            className="text-white text-sm">{pl.media_items?.length || 0} {t('common.videos_count')}</span>
                                    </div>
                                    <div className="absolute top-3 right-3">
                                        <span className={`text-xs px-2 py-0.5 rounded-full ${
                                            pl.is_public
                                                ? 'bg-emerald-500/80 text-white'
                                                : 'bg-gray-600/80 text-white'
                                        }`}>
                                            {visibilityLabel(pl.is_public ? 'public' : 'private')}
                                        </span>
                                    </div>
                                    {/* Play all overlay */}
                                    <div
                                        className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                        <div
                                            className="w-12 h-12 bg-white/90 rounded-full flex items-center justify-center shadow-lg">
                                            <Play className="w-5 h-5 text-gray-900 ml-0.5" fill="currentColor"/>
                                        </div>
                                    </div>
                                </div>
                            </Link>

                            {/* Info */}
                            <div className="p-4">
                                <Link to="/playlist/$token" params={{token: pl.short_token || pl.id}}>
                                    <h3 className="font-semibold text-gray-900 dark:text-white mb-1 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors cursor-pointer">
                                        {pl.title}
                                    </h3>
                                </Link>
                                <p className="text-sm text-gray-500 dark:text-muted-foreground line-clamp-2">{pl.description}</p>
                                <div className="flex items-center justify-between mt-2">
                                    <p className="text-xs text-muted-foreground">{t('playlists.updated', {date: formatDate(pl.update_time)})}</p>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-7 w-7 text-muted-foreground hover:text-destructive"
                                        onClick={() => setDeleteTarget(pl)}
                                        title={t('common.delete') || 'Delete'}
                                    >
                                        <Trash2 className="w-3.5 h-3.5"/>
                                    </Button>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 text-muted-foreground">
                    <ListVideo size={48} className="mx-auto mb-3 opacity-30"/>
                    <p className="text-lg mb-1">{t('playlists.empty')}</p>
                    <p className="text-sm mb-4">{t('playlists.emptyDesc')}</p>
                    <Button
                        onClick={() => setShowCreateDialog(true)}
                        className="bg-emerald-600 hover:bg-emerald-700"
                    >
                        <Plus className="w-4 h-4 mr-2"/>
                        {t('playlists.newList') || 'Create Playlist'}
                    </Button>
                </div>
            )}

            {/* Create Playlist Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <ListVideo className="w-5 h-5 text-emerald-600"/>
                            {t('playlists.createPlaylist') || 'Create Playlist'}
                        </DialogTitle>
                        <DialogDescription>
                            {t('playlists.createPlaylistDesc') || 'Create a new playlist to organize your videos'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 mt-4">
                        <div>
                            <label className="text-sm font-medium mb-1 block">{t('playlists.title') || 'Title'}</label>
                            <Input
                                value={newTitle}
                                onChange={(e) => setNewTitle(e.target.value)}
                                placeholder={t('playlists.titlePlaceholder') || 'Enter playlist title'}
                                autoFocus
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' && newTitle.trim() && !isCreating) {
                                        handleCreate();
                                    }
                                }}
                            />
                        </div>
                        <div>
                            <label className="text-sm font-medium mb-1 block">{t('playlists.description') || 'Description'}</label>
                            <Input
                                value={newDescription}
                                onChange={(e) => setNewDescription(e.target.value)}
                                placeholder={t('playlists.descriptionPlaceholder') || 'Enter playlist description (optional)'}
                            />
                        </div>
                        <div className="flex items-center gap-2">
                            <input
                                type="checkbox"
                                id="new-is-public"
                                checked={newIsPublic}
                                onChange={(e) => setNewIsPublic(e.target.checked)}
                                className="rounded border-gray-300"
                            />
                            <label htmlFor="new-is-public" className="text-sm">
                                {t('playlists.makePublic') || 'Make playlist public'}
                            </label>
                        </div>
                        <div className="flex justify-end gap-2">
                            <Button variant="outline" onClick={() => {
                                setShowCreateDialog(false);
                                setNewTitle('');
                                setNewDescription('');
                                setNewIsPublic(true);
                            }} disabled={isCreating}>
                                {t('common.cancel')}
                            </Button>
                            <Button
                                onClick={handleCreate}
                                disabled={!newTitle.trim() || isCreating}
                                className="bg-emerald-600 hover:bg-emerald-700"
                            >
                                {isCreating ? <Spinner className="w-4 h-4 mr-1"/> : <Plus className="w-4 h-4 mr-1"/>}
                                {t('watch.create') || 'Create'}
                            </Button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Delete Playlist Dialog */}
            <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('playlists.deletePlaylist') || 'Delete Playlist'}</DialogTitle>
                        <DialogDescription>
                            {t('playlists.deleteConfirm', {title: deleteTarget?.title}) || `Are you sure you want to delete "${deleteTarget?.title}"? This action cannot be undone.`}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-2 mt-4">
                        <Button variant="outline" onClick={() => setDeleteTarget(null)} disabled={isDeleting}>
                            {t('common.cancel')}
                        </Button>
                        <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
                            {isDeleting ? <Spinner className="w-4 h-4 mr-1"/> : null}
                            {t('common.delete') || 'Delete'}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default PlaylistsPage;
