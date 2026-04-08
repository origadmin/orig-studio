import React, {useState, useEffect} from 'react';
import {ThumbsUp, ThumbsDown, Share2, MessageCircle, Loader2, Save, Download, LogIn} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {formatViews} from '@/lib/format';
import {likeApi} from '@/lib/api/like';
import {shareApi} from '@/lib/api/share';
import {playlistApi} from '@/lib/api/playlist';
import {useAuth} from '@/hooks/useAuth';
import {useNavigate} from '@tanstack/react-router';
import {Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription} from '@/components/ui/dialog';

interface InteractionBarProps {
    mediaId: string;
}

const InteractionBar: React.FC<InteractionBarProps> = ({mediaId}) => {
    const {t} = useTranslation();
    const {isAuthenticated} = useAuth();
    const navigate = useNavigate();
    const [likeCount, setLikeCount] = useState(0);
    const [dislikeCount, setDislikeCount] = useState(0);
    const [isLiked, setIsLiked] = useState(false);
    const [isDisliked, setIsDisliked] = useState(false);
    const [isLiking, setIsLiking] = useState(false);
    const [isDisliking, setIsDisliking] = useState(false);
    const [isSharing, setIsSharing] = useState(false);
    const [shareUrl, setShareUrl] = useState('');
    const [showShareModal, setShowShareModal] = useState(false);
    const [showLoginDialog, setShowLoginDialog] = useState(false);
    const [isSaving, setIsSaving] = useState(false);
    const [showSaveModal, setShowSaveModal] = useState(false);
    const [playlists, setPlaylists] = useState<{ id: string, name: string }[]>([]);
    const [isDownloading, setIsDownloading] = useState(false);

    useEffect(() => {
        fetchLikeStatus();
    }, [mediaId]);

    const fetchPlaylists = async () => {
        try {
            const playlists = await playlistApi.getAll();
            setPlaylists(playlists.map(p => ({id: p.id, name: p.name})));
        } catch (err) {
            console.error('Failed to fetch playlists:', err);
        }
    };

    const fetchLikeStatus = async () => {
        try {
            const likeStatus = await likeApi.getStatus(mediaId);
            setLikeCount(likeStatus.like_count);
            setDislikeCount(likeStatus.dislike_count);
            setIsLiked(likeStatus.is_liked);
            setIsDisliked(likeStatus.is_disliked);
        } catch (err) {
            console.error('Failed to fetch like status:', err);
        }
    };

    const handleLike = async () => {
        if (!isAuthenticated) {
            setShowLoginDialog(true);
            return;
        }
        try {
            setIsLiking(true);
            const likeStatus = await likeApi.toggle(mediaId);
            setLikeCount(likeStatus.like_count);
            setDislikeCount(likeStatus.dislike_count);
            setIsLiked(likeStatus.is_liked);
            setIsDisliked(likeStatus.is_disliked);
        } catch (err) {
            console.error('Failed to toggle like:', err);
        } finally {
            setIsLiking(false);
        }
    };

    const handleDislike = async () => {
        if (!isAuthenticated) {
            setShowLoginDialog(true);
            return;
        }
        try {
            setIsDisliking(true);
            const likeStatus = await likeApi.toggleDislike(mediaId);
            setLikeCount(likeStatus.like_count);
            setDislikeCount(likeStatus.dislike_count);
            setIsLiked(likeStatus.is_liked);
            setIsDisliked(likeStatus.is_disliked);
        } catch (err) {
            console.error('Failed to toggle dislike:', err);
        } finally {
            setIsDisliking(false);
        }
    };

    const handleShare = async () => {
        try {
            setIsSharing(true);
            const shareResponse = await shareApi.getShareUrl(mediaId);
            setShareUrl(shareResponse.url);
            setShowShareModal(true);
            // 复制分享链接到剪贴板
            if (navigator.clipboard) {
                await navigator.clipboard.writeText(shareResponse.url);
                // 可以添加一个复制成功的提示
            }
        } catch (err) {
            console.error('Failed to get share url:', err);
        } finally {
            setIsSharing(false);
        }
    };

    const handleSave = async () => {
        if (!isAuthenticated) {
            setShowLoginDialog(true);
            return;
        }
        try {
            setIsSaving(true);
            await fetchPlaylists();
            setShowSaveModal(true);
        } catch (err) {
            console.error('Failed to save to playlist:', err);
        } finally {
            setIsSaving(false);
        }
    };

    const handleAddToPlaylist = async (playlistId: string) => {
        try {
            await playlistApi.addMedia(playlistId, mediaId);
            setShowSaveModal(false);
            // 可以添加一个保存成功的提示
        } catch (err) {
            console.error('Failed to add to playlist:', err);
        }
    };

    const handleDownload = async () => {
        try {
            setIsDownloading(true);
            // 这里应该调用下载 API 或者直接打开下载链接
            // 由于没有直接的下载 API，我们可以模拟一个下载过程
            setTimeout(() => {
                setIsDownloading(false);
                // 可以添加一个下载成功的提示
            }, 1000);
        } catch (err) {
            console.error('Failed to download:', err);
            setIsDownloading(false);
        }
    };

    return (
        <div className="flex items-center gap-4">
            {/* Like Button */}
            <Button
                variant="ghost"
                className={`flex items-center gap-2 ${isLiked ? 'text-blue-500' : 'text-gray-600 dark:text-gray-300'}`}
                onClick={handleLike}
                disabled={isLiking}
            >
                {isLiking ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <ThumbsUp className={`w-5 h-5 ${isLiked ? 'fill-blue-500' : ''}`}/>
                )}
                <span>{formatViews(likeCount)}</span>
            </Button>

            {/* Dislike Button */}
            <Button
                variant="ghost"
                className={`flex items-center gap-2 ${isDisliked ? 'text-red-500' : 'text-gray-600 dark:text-gray-300'}`}
                onClick={handleDislike}
                disabled={isDisliking}
            >
                {isDisliking ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <ThumbsDown className={`w-5 h-5 ${isDisliked ? 'fill-red-500' : ''}`}/>
                )}
                <span>{formatViews(dislikeCount)}</span>
            </Button>

            {/* Comment Button */}
            <Button
                variant="ghost"
                className="flex items-center gap-2 text-gray-600 dark:text-gray-300"
            >
                <MessageCircle className="w-5 h-5"/>
                <span>{t('watch.comments')}</span>
            </Button>

            {/* Share Button */}
            <Button
                variant="ghost"
                className="flex items-center gap-2 text-gray-600 dark:text-gray-300"
                onClick={handleShare}
                disabled={isSharing}
            >
                {isSharing ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <Share2 className="w-5 h-5"/>
                )}
                <span>{t('watch.share')}</span>
            </Button>

            {/* Save to Playlist Button */}
            <Button
                variant="ghost"
                className="flex items-center gap-2 text-gray-600 dark:text-gray-300"
                onClick={handleSave}
                disabled={isSaving}
            >
                {isSaving ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <Save className="w-5 h-5"/>
                )}
                <span>{t('watch.save')}</span>
            </Button>

            {/* Download Button */}
            <Button
                variant="ghost"
                className="flex items-center gap-2 text-gray-600 dark:text-gray-300"
                onClick={handleDownload}
                disabled={isDownloading}
            >
                {isDownloading ? (
                    <Loader2 className="w-4 h-4 animate-spin"/>
                ) : (
                    <Download className="w-5 h-5"/>
                )}
                <span>{t('watch.download')}</span>
            </Button>

            {/* Share Modal */}
            {showShareModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full">
                        <h3 className="text-lg font-bold mb-4">{t('watch.shareVideo')}</h3>
                        <div className="mb-4">
                            <p className="text-sm text-gray-600 dark:text-gray-300 mb-2">{t('watch.shareLink')}</p>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    value={shareUrl}
                                    readOnly
                                    className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-gray-50 dark:bg-gray-900"
                                />
                                <Button
                                    variant="default"
                                    className="bg-blue-600 hover:bg-blue-700"
                                    onClick={() => {
                                        navigator.clipboard.writeText(shareUrl);
                                    }}
                                >
                                    {t('watch.copyLink')}
                                </Button>
                            </div>
                        </div>
                        <div className="flex justify-end">
                            <Button
                                variant="default"
                                className="bg-gray-600 hover:bg-gray-700"
                                onClick={() => setShowShareModal(false)}
                            >
                                {t('common.close')}
                            </Button>
                        </div>
                    </div>
                </div>
            )}

            {/* Save to Playlist Modal */}
            {showSaveModal && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full">
                        <h3 className="text-lg font-bold mb-4">{t('watch.saveToPlaylist')}</h3>
                        <div className="mb-4">
                            {playlists.length > 0 ? (
                                <div className="space-y-2">
                                    {playlists.map(playlist => (
                                        <Button
                                            key={playlist.id}
                                            variant="default"
                                            className="w-full justify-start"
                                            onClick={() => handleAddToPlaylist(playlist.id)}
                                        >
                                            {playlist.name}
                                        </Button>
                                    ))}
                                </div>
                            ) : (
                                <p className="text-sm text-gray-600 dark:text-gray-300">{t('watch.noPlaylists')}</p>
                            )}
                        </div>
                        <div className="flex justify-end">
                            <Button
                                variant="default"
                                className="bg-gray-600 hover:bg-gray-700"
                                onClick={() => setShowSaveModal(false)}
                            >
                                {t('common.close')}
                            </Button>
                        </div>
                    </div>
                </div>
            )}

            {/* Login Required Dialog */}
            <Dialog open={showLoginDialog} onOpenChange={setShowLoginDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <LogIn className="w-5 h-5"/>
                            {t('auth.loginRequired') || 'Login Required'}
                        </DialogTitle>
                        <DialogDescription>
                            {t('watch.pleaseLoginToInteract') || 'Please log in to like, dislike, or save videos.'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-3 mt-4">
                        <Button
                            variant="default"
                            onClick={() => setShowLoginDialog(false)}
                        >
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button
                            variant="default"
                            className="bg-blue-600 hover:bg-blue-700"
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
        </div>
    );
};

export default InteractionBar;
