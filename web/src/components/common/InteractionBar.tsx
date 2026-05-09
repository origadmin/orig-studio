import React, {useState, useEffect} from 'react';
import {
    ThumbsUp,
    ThumbsDown,
    Share2,
    MessageCircle,
    Loader2,
    Bookmark,
    Download,
    LogIn,
    Check,
    Link2,
    BookmarkPlus,
    Flag
} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {formatViews} from '@/lib/format';
import {mediaApi, publicMediaApi, LikeResponse, FavoriteResponse, ShareResponse} from '@/lib/api/media';
import {playlistApi} from '@/lib/api/playlist';
import {useAuth} from '@/hooks/useAuth';
import {useNavigate} from '@tanstack/react-router';
import {toast} from 'sonner';
import ReportDialog from '@/components/common/ReportDialog';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription
} from '@/components/ui/dialog';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface InteractionBarProps {
    mediaId: string;
    shortToken?: string;  // 可选：使用 short_token 调用 publicMediaApi (推荐)
    commentCount?: number;
    onCommentClick?: () => void;
}

const InteractionBar: React.FC<InteractionBarProps> = ({mediaId, shortToken, commentCount = 0, onCommentClick}) => {
    const {t} = useTranslation();
    const {isAuthenticated} = useAuth();
    const navigate = useNavigate();

    // 决定使用哪个 API：优先使用 publicMediaApi (short_token based)
    const usePublicApi = !!shortToken && shortToken.trim().length > 0;
    const apiIdentifier = (shortToken && shortToken.trim()) || mediaId;

    // Like state
    const [likeCount, setLikeCount] = useState(0);
    const [dislikeCount, setDislikeCount] = useState(0);
    const [isLiked, setIsLiked] = useState(false);
    const [isDisliked, setIsDisliked] = useState(false);
    const [isLiking, setIsLiking] = useState(false);
    const [isDisliking, setIsDisliking] = useState(false);

    // Favorite state
    const [favoriteCount, setFavoriteCount] = useState(0);
    const [isFavorited, setIsFavorited] = useState(false);
    const [isFavoriting, setIsFavoriting] = useState(false);

    // Share state
    const [isSharing, setIsSharing] = useState(false);
    const [shareData, setShareData] = useState<ShareResponse | null>(null);
    const [showShareModal, setShowShareModal] = useState(false);
    const [copied, setCopied] = useState(false);

    // Save/Playlist state
    const [isSaving, setIsSaving] = useState(false);
    const [showSaveModal, setShowSaveModal] = useState(false);
    const [playlists, setPlaylists] = useState<{ id: string; name: string }[]>([]);
    const [addedPlaylistIds, setAddedPlaylistIds] = useState<Set<string>>(new Set());
    const [newPlaylistName, setNewPlaylistName] = useState('');
    const [isCreatingPlaylist, setIsCreatingPlaylist] = useState(false);
    const [showCreateForm, setShowCreateForm] = useState(false);

    // Download state
    const [isDownloading, setIsDownloading] = useState(false);

    // Login dialog
    const [showLoginDialog, setShowLoginDialog] = useState(false);
    const [loginAction, setLoginAction] = useState<string>('');

    // Report dialog
    const [showReportDialog, setShowReportDialog] = useState(false);

    // Fetch initial status
    useEffect(() => {
        if (!apiIdentifier) return;

        const fetchStatus = async () => {
            try {
                // Fetch like status - 优先使用 publicMediaApi
                const likeStatus: LikeResponse = usePublicApi
                    ? await publicMediaApi.likes.getStatus(apiIdentifier)
                    : await mediaApi.likes.getStatus(mediaId);
                setLikeCount(likeStatus.like_count);
                setDislikeCount(likeStatus.dislike_count);
                setIsLiked(likeStatus.is_liked);
                setIsDisliked(likeStatus.is_disliked);
            } catch (err) {
                console.error('Failed to fetch like status:', err);
            }

            try {
                // Fetch favorite status - 优先使用 publicMediaApi
                const favStatus: FavoriteResponse = usePublicApi
                    ? await publicMediaApi.favorites.getStatus(apiIdentifier)
                    : await mediaApi.favorites.getStatus(mediaId);
                setFavoriteCount(favStatus.favorite_count);
                setIsFavorited(favStatus.is_favorited);
            } catch (err) {
                console.error('Failed to fetch favorite status:', err);
            }
        };

        fetchStatus();
    }, [apiIdentifier, usePublicApi, mediaId]);

    const requireAuth = (action: string): boolean => {
        if (!isAuthenticated) {
            setLoginAction(action);
            setShowLoginDialog(true);
            return false;
        }
        return true;
    };

    const handleLike = async () => {
        if (!requireAuth('like')) return;
        if (isLiking) return; // 防止重复点击

        try {
            setIsLiking(true);

            // 乐观更新：立即更新 UI
            const prevLiked = isLiked;
            const prevDisliked = isDisliked;
            const prevLikeCount = likeCount;
            const prevDislikeCount = dislikeCount;

            if (isLiked) {
                // 取消点赞
                setIsLiked(false);
                setLikeCount(Math.max(0, prevLikeCount - 1));
            } else {
                // 点赞（如果之前是踩，先取消踩）
                setIsLiked(true);
                setLikeCount(prevLikeCount + 1);
                if (isDisliked) {
                    setIsDisliked(false);
                    setDislikeCount(Math.max(0, prevDislikeCount - 1));
                }
            }

            // 调用 API
            const response: LikeResponse = usePublicApi
                ? await publicMediaApi.likes.toggle(apiIdentifier)
                : await mediaApi.likes.toggle(mediaId);

            // 使用服务器返回的最终状态更新
            setLikeCount(response.like_count);
            setDislikeCount(response.dislike_count);
            setIsLiked(response.is_liked);
            setIsDisliked(response.is_disliked);
        } catch (err) {
            console.error('Failed to toggle like:', err);
            // 回滚到之前的状态（乐观更新失败）
            // 注意：这里简化处理，实际应该保存 prevValues 并回滚
        } finally {
            setIsLiking(false);
        }
    };

    const handleDislike = async () => {
        if (!requireAuth('dislike')) return;
        if (isDisliking) return; // 防止重复点击

        try {
            setIsDisliking(true);

            // 乐观更新：立即更新 UI
            if (isDisliked) {
                // 取消点踩
                setIsDisliked(false);
                setDislikeCount(Math.max(0, dislikeCount - 1));
            } else {
                // 点踩（如果之前是赞，先取消赞）
                setIsDisliked(true);
                setDislikeCount(dislikeCount + 1);
                if (isLiked) {
                    setIsLiked(false);
                    setLikeCount(Math.max(0, likeCount - 1));
                }
            }

            // 调用 API
            const response: LikeResponse = usePublicApi
                ? await publicMediaApi.likes.toggleDislike(apiIdentifier)
                : await mediaApi.likes.toggleDislike(mediaId);

            // 使用服务器返回的最终状态更新
            setLikeCount(response.like_count);
            setDislikeCount(response.dislike_count);
            setIsLiked(response.is_liked);
            setIsDisliked(response.is_disliked);
        } catch (err) {
            console.error('Failed to toggle dislike:', err);
        } finally {
            setIsDisliking(false);
        }
    };

    const handleFavorite = async () => {
        if (!requireAuth('favorite')) return;

        try {
            setIsFavoriting(true);
            // 使用 publicMediaApi 或 mediaApi
            const response: FavoriteResponse = usePublicApi
                ? await publicMediaApi.favorites.toggle(apiIdentifier)
                : await mediaApi.favorites.toggle(mediaId);
            setFavoriteCount(response.favorite_count);
            setIsFavorited(response.is_favorited);
        } catch (err) {
            console.error('Failed to toggle favorite:', err);
        } finally {
            setIsFavoriting(false);
        }
    };

    const handleShare = async () => {
        try {
            setIsSharing(true);
            // 使用 publicMediaApi 或 mediaApi
            const response: ShareResponse = usePublicApi
                ? await publicMediaApi.shares.getShareUrl(apiIdentifier)
                : await mediaApi.shares.getShareUrl(mediaId);
            setShareData(response);
            setShowShareModal(true);
        } catch (err) {
            console.error('Failed to get share URL:', err);
        } finally {
            setIsSharing(false);
        }
    };

    const handleCopyLink = async () => {
        if (shareData?.url) {
            try {
                await navigator.clipboard.writeText(shareData.url);
                setCopied(true);
                setTimeout(() => setCopied(false), 2000);
            } catch (err) {
                console.error('Failed to copy:', err);
            }
        }
    };

    const handleNativeShare = async () => {
        if (navigator.share && shareData) {
            try {
                await navigator.share({
                    title: shareData.title,
                    url: shareData.url,
                });
            } catch (err) {
                // User cancelled or share failed
            }
        }
    };

    const fetchPlaylists = async () => {
        if (!isAuthenticated) {
            setLoginAction('save');
            setShowLoginDialog(true);
            return;
        }
        try {
            setIsSaving(true);
            const response = await playlistApi.getMyPlaylists();
            const items = response.items || [];
            setPlaylists(items.map((p: any) => ({id: String(p.id), name: p.title})));
            setAddedPlaylistIds(new Set()); // Reset added state when opening a new dialog
            setShowSaveModal(true);
        } catch (err) {
            console.error('Failed to fetch playlists:', err);
        } finally {
            setIsSaving(false);
        }
    };

    const handleAddToPlaylist = async (playlistId: string) => {
        // Skip if already added to this playlist
        if (addedPlaylistIds.has(playlistId)) return;
        try {
            setIsSaving(true);
            await playlistApi.addMedia(playlistId, mediaId);
            setAddedPlaylistIds(prev => new Set(prev).add(playlistId));
        } catch (err) {
            console.error('Failed to add to playlist:', err);
        } finally {
            setIsSaving(false);
        }
    };

    const handleCreatePlaylist = async () => {
        if (!newPlaylistName.trim()) return;

        try {
            setIsCreatingPlaylist(true);
            const result = await playlistApi.create({title: newPlaylistName.trim()});
            const newPlaylist = result.playlist;
            if (newPlaylist && newPlaylist.id) {
                await playlistApi.addMedia(newPlaylist.id, mediaId);
                const playlistId = String(newPlaylist.id);
                setPlaylists(prev => [...prev, {id: playlistId, name: newPlaylist.title || newPlaylistName.trim()}]);
                setAddedPlaylistIds(prev => new Set(prev).add(playlistId));
            }
            setNewPlaylistName('');
            setShowCreateForm(false);
        } catch (err) {
            console.error('Failed to create playlist:', err);
        } finally {
            setIsCreatingPlaylist(false);
        }
    };

    const handleDownload = async () => {
        try {
            setIsDownloading(true);
            const response = await mediaApi.download(mediaId);
            if (response.download_url) {
                window.open(response.download_url, '_blank');
            }
        } catch (err) {
            console.error('Failed to download:', err);
        } finally {
            setIsDownloading(false);
        }
    };

    const handleReport = async (data: { reason: string; description?: string }) => {
        try {
            if (usePublicApi) {
                await publicMediaApi.report(apiIdentifier, data);
            } else {
                await mediaApi.report(mediaId, data);
            }
            toast.success(t('report.submitted') || 'Report submitted successfully');
        } catch (err: any) {
            throw err;
        }
    };

    const handleOpenReportDialog = () => {
        if (!requireAuth('report')) return;
        setShowReportDialog(true);
    };

    return (
        <div className="flex items-center gap-2 flex-wrap">
            {/* Like Button */}
            <div className="flex items-center bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                <Button
                    variant="ghost"
                    size="sm"
                    className={`flex items-center gap-2 rounded-none px-4 ${
                        isLiked ? 'text-info bg-blue-50 dark:bg-blue-900/20' : 'text-gray-700 dark:text-gray-300'
                    }`}
                    onClick={handleLike}
                    disabled={isLiking}
                >
                    {isLiking ? (
                        <Loader2 className="w-4 h-4 animate-spin"/>
                    ) : (
                        <ThumbsUp className={`w-4 h-4 ${isLiked ? 'fill-current' : ''}`}/>
                    )}
                    <span className="font-medium">{formatViews(likeCount)}</span>
                </Button>
                <div className="w-px h-4 bg-gray-300 dark:bg-gray-600"/>
                <Button
                    variant="ghost"
                    size="sm"
                    className={`flex items-center rounded-none px-3 ${
                        isDisliked ? 'text-destructive bg-red-50 dark:bg-red-900/20' : 'text-gray-700 dark:text-gray-300'
                    }`}
                    onClick={handleDislike}
                    disabled={isDisliking}
                >
                    {isDisliking ? (
                        <Loader2 className="w-4 h-4 animate-spin"/>
                    ) : (
                        <ThumbsDown className={`w-4 h-4 ${isDisliked ? 'fill-current' : ''}`}/>
                    )}
                </Button>
            </div>

            {/* Favorite Button */}
            <Button
                variant="ghost"
                size="sm"
                className={`flex items-center gap-2 rounded-full px-4 ${
                    isFavorited ? 'text-success bg-green-50 dark:bg-green-900/20' : 'text-gray-700 dark:text-gray-300'
                }`}
                onClick={handleFavorite}
                disabled={isFavoriting}
            >
                {isFavoriting ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <Bookmark className={`w-4 h-4 ${isFavorited ? 'fill-current' : ''}`}/>
                )}
                <span
                    className="font-medium">{isFavorited ? t('watch.favorited') || 'Favorited' : t('watch.favorite') || 'Favorite'}</span>
            </Button>

            {/* Share Button */}
            <Button
                variant="ghost"
                size="sm"
                className="flex items-center gap-2 rounded-full px-4 text-gray-700 dark:text-gray-300"
                onClick={handleShare}
                disabled={isSharing}
            >
                {isSharing ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <Share2 className="w-4 h-4"/>
                )}
                <span className="font-medium">{t('watch.share')}</span>
            </Button>

            {/* More Actions Dropdown */}
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button
                        variant="ghost"
                        size="sm"
                        className="rounded-full px-3 text-gray-700 dark:text-gray-300"
                    >
                        <span className="sr-only">More actions</span>
                        <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                            <path
                                d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
                        </svg>
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={fetchPlaylists}>
                        <Bookmark className="w-4 h-4 mr-2"/>
                        {t('watch.saveToPlaylist') || 'Save to playlist'}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={handleDownload} disabled={isDownloading}>
                        <Download className="w-4 h-4 mr-2"/>
                        {isDownloading ? t('common.loading') : (t('watch.download') || 'Download')}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={handleOpenReportDialog} className="text-amber-600 focus:text-amber-600">
                        <Flag className="w-4 h-4 mr-2"/>
                        {t('report.reportVideo') || 'Report Video'}
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            {/* Share Modal */}
            <Dialog open={showShareModal} onOpenChange={setShowShareModal}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('watch.shareVideo') || 'Share Video'}</DialogTitle>
                        <DialogDescription>
                            {t('watch.shareDescription') || 'Share this video with your friends'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        {/* Share Link */}
                        <div className="flex items-center gap-2">
                            <div
                                className="flex-1 flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
                                <Link2 className="w-4 h-4 text-gray-500"/>
                                <input
                                    type="text"
                                    value={shareData?.url || ''}
                                    readOnly
                                    className="flex-1 bg-transparent text-sm text-gray-700 dark:text-gray-300 outline-none"
                                />
                            </div>
                            <Button
                                size="sm"
                                onClick={handleCopyLink}
                                className={copied ? 'bg-green-600 hover:bg-green-700' : 'bg-emerald-600 hover:bg-emerald-700'}
                            >
                                {copied ? <Check className="w-4 h-4"/> : t('watch.copyLink') || 'Copy'}
                            </Button>
                        </div>

                        {/* Social Share Buttons */}
                        <div className="grid grid-cols-5 gap-2">
                            {shareData?.twitter && (
                                <a
                                    href={shareData.twitter}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    <div className="w-10 h-10 bg-black rounded-full flex items-center justify-center">
                                        <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                            <path
                                                d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                                        </svg>
                                    </div>
                                    <span className="text-xs text-gray-600 dark:text-muted-foreground">X</span>
                                </a>
                            )}
                            {shareData?.facebook && (
                                <a
                                    href={shareData.facebook}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    <div
                                        className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center">
                                        <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                            <path
                                                d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
                                        </svg>
                                    </div>
                                    <span className="text-xs text-gray-600 dark:text-muted-foreground">Facebook</span>
                                </a>
                            )}
                            {shareData?.whatsapp && (
                                <a
                                    href={shareData.whatsapp}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    <div
                                        className="w-10 h-10 bg-success rounded-full flex items-center justify-center">
                                        <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                            <path
                                                d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                                        </svg>
                                    </div>
                                    <span className="text-xs text-gray-600 dark:text-muted-foreground">WhatsApp</span>
                                </a>
                            )}
                            {shareData?.telegram && (
                                <a
                                    href={shareData.telegram}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    <div
                                        className="w-10 h-10 bg-info rounded-full flex items-center justify-center">
                                        <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                            <path
                                                d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
                                        </svg>
                                    </div>
                                    <span className="text-xs text-gray-600 dark:text-muted-foreground">Telegram</span>
                                </a>
                            )}
                            {'share' in navigator && (
                                <button
                                    onClick={handleNativeShare}
                                    className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    <div
                                        className="w-10 h-10 bg-muted dark:bg-gray-700 rounded-full flex items-center justify-center">
                                        <Share2 className="w-5 h-5 text-gray-700 dark:text-gray-300"/>
                                    </div>
                                    <span className="text-xs text-gray-600 dark:text-muted-foreground">More</span>
                                </button>
                            )}
                        </div>
                    </div>
                </DialogContent>
            </Dialog>

            {/* Save to Playlist Modal */}
            <Dialog open={showSaveModal} onOpenChange={setShowSaveModal}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <Bookmark className="w-5 h-5 text-emerald-600"/>
                            {t('watch.saveToPlaylist') || 'Save to Playlist'}
                        </DialogTitle>
                        <DialogDescription>
                            {t('watch.selectPlaylist') || 'Select a playlist to save this video'}
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-3 mt-4">
                        {/* Existing playlists */}
                        {playlists.length > 0 && (
                            <div className="space-y-2 max-h-48 overflow-y-auto">
                                {playlists.map(playlist => {
                                    const isAdded = addedPlaylistIds.has(playlist.id);
                                    return (
                                        <Button
                                            key={playlist.id}
                                            variant="outline"
                                            className={`w-full justify-start h-auto py-2.5 px-3 ${
                                                isAdded
                                                    ? 'border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-900/20'
                                                    : ''
                                            }`}
                                            onClick={() => handleAddToPlaylist(playlist.id)}
                                            disabled={isSaving || isAdded}
                                        >
                                            {isAdded ? (
                                                <Check className="w-4 h-4 mr-2 flex-shrink-0 text-emerald-600"/>
                                            ) : (
                                                <Bookmark className="w-4 h-4 mr-2 flex-shrink-0"/>
                                            )}
                                            <span className={`truncate text-left ${isAdded ? 'text-emerald-700 dark:text-emerald-400' : ''}`}>
                                                {playlist.name}
                                            </span>
                                            {isAdded && (
                                                <span className="ml-auto text-xs text-emerald-600 dark:text-emerald-400 flex-shrink-0">
                                                    {t('watch.added') || 'Added'}
                                                </span>
                                            )}
                                            {isSaving && !isAdded && (
                                                <Loader2 className="w-4 h-4 ml-auto animate-spin flex-shrink-0"/>
                                            )}
                                        </Button>
                                    );
                                })}
                            </div>
                        )}

                        {/* Divider */}
                        {playlists.length > 0 && (
                            <div className="border-t border-gray-200 dark:border-gray-700 pt-3">
                                <span className="text-xs text-gray-500 uppercase tracking-wider px-1">
                                    {t('watch.orCreateNew') || 'Or create new'}
                                </span>
                            </div>
                        )}

                        {/* Create new playlist form */}
                        {!showCreateForm ? (
                            <Button
                                variant="outline"
                                className="w-full justify-start border-dashed border-2 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
                                onClick={() => setShowCreateForm(true)}
                            >
                                <BookmarkPlus className="w-4 h-4 mr-2"/>
                                {t('watch.createNewPlaylist') || '+ Create new playlist'}
                            </Button>
                        ) : (
                            <div className="space-y-2 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                                <input
                                    type="text"
                                    value={newPlaylistName}
                                    onChange={(e) => setNewPlaylistName(e.target.value)}
                                    placeholder={t('watch.playlistNamePlaceholder') || 'Enter playlist name...'}
                                    className="w-full px-3 py-2 rounded-md border border-input dark:border-gray-600 bg-white dark:bg-gray-900 text-sm focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none"
                                    autoFocus
                                    onKeyDown={(e) => {
                                        if (e.key === 'Enter' && !isCreatingPlaylist) {
                                            handleCreatePlaylist();
                                        }
                                    }}
                                />
                                <div className="flex gap-2">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="flex-1"
                                        onClick={() => {
                                            setShowCreateForm(false);
                                            setNewPlaylistName('');
                                        }}
                                        disabled={isCreatingPlaylist}
                                    >
                                        {t('common.cancel')}
                                    </Button>
                                    <Button
                                        size="sm"
                                        className="flex-1 bg-emerald-600 hover:bg-emerald-700"
                                        onClick={handleCreatePlaylist}
                                        disabled={!newPlaylistName.trim() || isCreatingPlaylist}
                                    >
                                        {isCreatingPlaylist ? (
                                            <Loader2 className="w-4 h-4 animate-spin mr-1"/>
                                        ) : null}
                                        {t('watch.create') || 'Create'}
                                    </Button>
                                </div>
                            </div>
                        )}

                        {/* Empty state with CTA */}
                        {playlists.length === 0 && !showCreateForm && (
                            <div className="text-center py-6">
                                <Bookmark className="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3"/>
                                <p className="text-sm text-gray-500 mb-4">
                                    {t('watch.noPlaylists') || "You don't have any playlists yet"}
                                </p>
                                <p className="text-xs text-muted-foreground mb-4">
                                    {t('watch.createFirstPlaylist') || 'Create your first playlist to organize your videos'}
                                </p>
                                <Button
                                    onClick={() => setShowCreateForm(true)}
                                    className="bg-emerald-600 hover:bg-emerald-700"
                                >
                                    <BookmarkPlus className="w-4 h-4 mr-2"/>
                                    {t('watch.createPlaylist') || 'Create Playlist'}
                                </Button>
                            </div>
                        )}
                    </div>
                </DialogContent>
            </Dialog>

            {/* Login Required Dialog */}
            <Dialog open={showLoginDialog} onOpenChange={setShowLoginDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <LogIn className="w-5 h-5"/>
                            {t('auth.loginRequired') || 'Login Required'}
                        </DialogTitle>
                        <DialogDescription>
                            {loginAction === 'like' && (t('watch.loginToLike') || 'Please login to like this video')}
                            {loginAction === 'dislike' && (t('watch.loginToDislike') || 'Please login to dislike this video')}
                            {loginAction === 'favorite' && (t('watch.loginToFavorite') || 'Please login to save this video')}
                            {loginAction === 'save' && (t('watch.loginToSave') || 'Please login to save to playlist')}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-3 mt-4">
                        <Button
                            variant="outline"
                            onClick={() => setShowLoginDialog(false)}
                        >
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button
                            className="bg-emerald-600 hover:bg-emerald-700"
                            onClick={() => {
                                setShowLoginDialog(false);
                                navigate({to: '/auth/signin'});
                            }}
                        >
                            <LogIn className="w-4 h-4 mr-2"/>
                            {t('auth.signin') || 'Sign In'}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            <ReportDialog
                open={showReportDialog}
                onOpenChange={setShowReportDialog}
                targetId={usePublicApi ? apiIdentifier : String(mediaId)}
                targetType="media"
                onSubmit={handleReport}
            />
        </div>
    );
};

export default InteractionBar;
