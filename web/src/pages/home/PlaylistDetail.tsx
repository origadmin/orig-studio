/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Playlist Detail Page - displays a single playlist with its videos.
 * Accessed via /playlist/:token (portal, public playlists) or /me/playlists -> click (user's own).
 */

import React, {useState} from 'react';
import {useParams, Link, useNavigate} from '@tanstack/react-router';
import {useQuery, useQueryClient} from '@tanstack/react-query';
import {useTranslation} from 'react-i18next';
import {ListVideo, Play, Video, Trash2, Edit3, Globe, Lock, ArrowLeft, MoreHorizontal} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Spinner} from '@/components/ui/spinner';
import {Badge} from '@/components/ui/badge';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from '@/components/ui/dialog';
import {Input} from '@/components/ui/input';
import {playlistApi, type Playlist, type PlaylistMediaItem} from '@/lib/api/playlist';
import {formatDate, formatDuration, formatViews} from '@/lib/format';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {useAuth} from '@/hooks/useAuth';

const PlaylistDetailPage: React.FC = () => {
    const {token} = useParams({strict: false}) as {token?: string};
    const {t} = useTranslation();
    const {user, isAuthenticated} = useAuth();
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    // Edit dialog state
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [editTitle, setEditTitle] = useState('');
    const [editDescription, setEditDescription] = useState('');
    const [editIsPublic, setEditIsPublic] = useState(true);
    const [isUpdating, setIsUpdating] = useState(false);

    // Delete dialog state
    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);

    // Remove media dialog state
    const [removeMediaId, setRemoveMediaId] = useState<string | null>(null);
    const [isRemovingMedia, setIsRemovingMedia] = useState(false);

    const {data: playlistData, isLoading, error} = useQuery({
        queryKey: ['playlist', token],
        queryFn: async () => {
            if (!token) throw new Error('No playlist token provided');
            const response = await playlistApi.get(token);
            return response.playlist as Playlist;
        },
        enabled: !!token,
    });

    const playlist = playlistData;
    const isOwner = isAuthenticated && user && playlist && String(user.id) === String(playlist.user_id);
    const mediaItems: PlaylistMediaItem[] = playlist?.media_details || [];

    const handleEdit = () => {
        if (!playlist) return;
        setEditTitle(playlist.title);
        setEditDescription(playlist.description || '');
        setEditIsPublic(playlist.is_public);
        setShowEditDialog(true);
    };

    const handleSaveEdit = async () => {
        if (!playlist) return;
        try {
            setIsUpdating(true);
            await playlistApi.update(playlist.id, {
                title: editTitle,
                description: editDescription,
                is_public: editIsPublic,
            });
            setShowEditDialog(false);
            queryClient.invalidateQueries({queryKey: ['playlist', token]});
        } catch (err) {
            console.error('Failed to update playlist:', err);
        } finally {
            setIsUpdating(false);
        }
    };

    const handleDelete = async () => {
        if (!playlist) return;
        try {
            setIsDeleting(true);
            await playlistApi.delete(playlist.id);
            setShowDeleteDialog(false);
            navigate({to: '/me/playlists'});
        } catch (err) {
            console.error('Failed to delete playlist:', err);
        } finally {
            setIsDeleting(false);
        }
    };

    const handleRemoveMedia = async () => {
        if (!playlist || !removeMediaId) return;
        try {
            setIsRemovingMedia(true);
            await playlistApi.removeMedia(playlist.id, removeMediaId);
            setRemoveMediaId(null);
            queryClient.invalidateQueries({queryKey: ['playlist', token]});
        } catch (err) {
            console.error('Failed to remove media from playlist:', err);
        } finally {
            setIsRemovingMedia(false);
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Spinner/>
            </div>
        );
    }

    if (error || !playlist) {
        return (
            <div className="text-center py-20 text-muted-foreground">
                <ListVideo size={48} className="mx-auto mb-3 opacity-30"/>
                <p className="text-lg mb-1">{t('playlists.notFound') || 'Playlist not found'}</p>
                <p className="text-sm mb-4">{t('playlists.notFoundDesc') || 'This playlist may be private or has been deleted.'}</p>
                <Link to="/me/playlists">
                    <Button variant="outline">
                        <ArrowLeft className="w-4 h-4 mr-2"/>
                        {t('playlists.backToList') || 'Back to my playlists'}
                    </Button>
                </Link>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-3 mb-2">
                        <Link to="/me/playlists" className="text-muted-foreground hover:text-foreground transition-colors">
                            <ArrowLeft className="w-5 h-5"/>
                        </Link>
                        <ListVideo size={24} className="text-emerald-600 flex-shrink-0"/>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white line-clamp-2">{playlist.title}</h1>
                        <Badge variant={playlist.is_public ? 'default' : 'secondary'} className="flex-shrink-0">
                            {playlist.is_public ? (
                                <><Globe className="w-3 h-3 mr-1"/>{t('common.public')}</>
                            ) : (
                                <><Lock className="w-3 h-3 mr-1"/>{t('common.private')}</>
                            )}
                        </Badge>
                    </div>
                    {playlist.description && (
                        <p className="text-sm text-gray-500 dark:text-muted-foreground ml-8">{playlist.description}</p>
                    )}
                    <div className="flex items-center gap-3 ml-8 mt-2 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                            <Video className="w-4 h-4"/>
                            {mediaItems.length} {t('common.videos_count')}
                        </span>
                        <span>{t('playlists.updated', {date: formatDate(playlist.update_time || playlist.create_time)})}</span>
                    </div>
                </div>

                {/* Owner actions */}
                {isOwner && (
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon">
                                <MoreHorizontal className="w-5 h-5"/>
                            </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={handleEdit}>
                                <Edit3 className="w-4 h-4 mr-2"/>
                                {t('common.edit') || 'Edit'}
                            </DropdownMenuItem>
                            <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setShowDeleteDialog(true)}>
                                <Trash2 className="w-4 h-4 mr-2"/>
                                {t('common.delete') || 'Delete'}
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                )}
            </div>

            {/* Video list */}
            {mediaItems.length > 0 ? (
                <div className="space-y-2">
                    {mediaItems.map((media, index) => (
                        <div
                            key={media.id}
                            className="flex items-center gap-4 p-3 rounded-lg bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 hover:shadow-md transition-all group"
                        >
                            {/* Index */}
                            <span className="text-sm text-muted-foreground w-6 text-center flex-shrink-0">{index + 1}</span>

                            {/* Thumbnail */}
                            <Link to="/watch" search={{v: media.short_token}} className="flex-shrink-0">
                                <div className="relative w-40 aspect-video rounded overflow-hidden bg-gray-100 dark:bg-gray-700">
                                    {media.thumbnail ? (
                                        <img
                                            src={getImageUrl(media.thumbnail, 'thumbnail')}
                                            alt={media.title}
                                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                                            onError={(e) => handleImageError(e, 'thumbnail')}
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center">
                                            <Play className="w-8 h-8 text-gray-300 dark:text-gray-600"/>
                                        </div>
                                    )}
                                    {media.duration > 0 && (
                                        <div className="absolute bottom-1 right-1 bg-black/80 text-white text-xs px-1.5 py-0.5 rounded">
                                            {formatDuration(media.duration)}
                                        </div>
                                    )}
                                    {/* Hover play overlay */}
                                    <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/20">
                                        <Play className="w-8 h-8 text-white" fill="currentColor"/>
                                    </div>
                                </div>
                            </Link>

                            {/* Info */}
                            <Link to="/watch" search={{v: media.short_token}} className="flex-1 min-w-0">
                                <h3 className="font-medium text-gray-900 dark:text-white line-clamp-2 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                    {media.title}
                                </h3>
                                <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                                    <span>{formatViews(media.view_count)} {t('common.views')}</span>
                                    <span>{formatDate(media.create_time)}</span>
                                </div>
                            </Link>

                            {/* Remove button (owner only) */}
                            {isOwner && (
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-destructive"
                                    onClick={() => setRemoveMediaId(media.id)}
                                    title={t('playlists.removeVideo') || 'Remove from playlist'}
                                >
                                    <Trash2 className="w-4 h-4"/>
                                </Button>
                            )}
                        </div>
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 text-muted-foreground">
                    <Video size={48} className="mx-auto mb-3 opacity-30"/>
                    <p className="text-lg mb-1">{t('playlists.emptyPlaylist') || 'This playlist is empty'}</p>
                    <p className="text-sm">{t('playlists.emptyPlaylistDesc') || 'Add videos to this playlist from the video page.'}</p>
                </div>
            )}

            {/* Edit Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('playlists.editPlaylist') || 'Edit Playlist'}</DialogTitle>
                        <DialogDescription>
                            {t('playlists.editPlaylistDesc') || 'Update your playlist details'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 mt-4">
                        <div>
                            <label className="text-sm font-medium mb-1 block">{t('playlists.title') || 'Title'}</label>
                            <Input
                                value={editTitle}
                                onChange={(e) => setEditTitle(e.target.value)}
                                placeholder={t('playlists.titlePlaceholder') || 'Enter playlist title'}
                            />
                        </div>
                        <div>
                            <label className="text-sm font-medium mb-1 block">{t('playlists.description') || 'Description'}</label>
                            <Input
                                value={editDescription}
                                onChange={(e) => setEditDescription(e.target.value)}
                                placeholder={t('playlists.descriptionPlaceholder') || 'Enter playlist description'}
                            />
                        </div>
                        <div className="flex items-center gap-2">
                            <input
                                type="checkbox"
                                id="edit-is-public"
                                checked={editIsPublic}
                                onChange={(e) => setEditIsPublic(e.target.checked)}
                                className="rounded border-gray-300"
                            />
                            <label htmlFor="edit-is-public" className="text-sm">
                                {t('playlists.makePublic') || 'Make playlist public'}
                            </label>
                        </div>
                        <div className="flex justify-end gap-2">
                            <Button variant="outline" onClick={() => setShowEditDialog(false)} disabled={isUpdating}>
                                {t('common.cancel')}
                            </Button>
                            <Button onClick={handleSaveEdit} disabled={!editTitle.trim() || isUpdating}
                                    className="bg-emerald-600 hover:bg-emerald-700">
                                {isUpdating ? <Spinner className="w-4 h-4 mr-1"/> : null}
                                {t('common.save') || 'Save'}
                            </Button>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Delete Dialog */}
            <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('playlists.deletePlaylist') || 'Delete Playlist'}</DialogTitle>
                        <DialogDescription>
                            {t('playlists.deleteConfirm') || 'Are you sure you want to delete this playlist? This action cannot be undone.'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-2 mt-4">
                        <Button variant="outline" onClick={() => setShowDeleteDialog(false)} disabled={isDeleting}>
                            {t('common.cancel')}
                        </Button>
                        <Button variant="destructive" onClick={handleDelete} disabled={isDeleting}>
                            {isDeleting ? <Spinner className="w-4 h-4 mr-1"/> : null}
                            {t('common.delete') || 'Delete'}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Remove Media Dialog */}
            <Dialog open={!!removeMediaId} onOpenChange={() => setRemoveMediaId(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('playlists.removeVideo') || 'Remove Video'}</DialogTitle>
                        <DialogDescription>
                            {t('playlists.removeVideoConfirm') || 'Remove this video from the playlist?'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-2 mt-4">
                        <Button variant="outline" onClick={() => setRemoveMediaId(null)} disabled={isRemovingMedia}>
                            {t('common.cancel')}
                        </Button>
                        <Button variant="destructive" onClick={handleRemoveMedia} disabled={isRemovingMedia}>
                            {isRemovingMedia ? <Spinner className="w-4 h-4 mr-1"/> : null}
                            {t('common.remove') || 'Remove'}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default PlaylistDetailPage;
