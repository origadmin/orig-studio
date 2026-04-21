import React, {useState, useEffect, useRef} from 'react';
import {MessageCircle, ThumbsUp, ThumbsDown, Send, Loader2, LogIn, MoreVertical, List, Reply, SmilePlus, X} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDate} from '@/lib/format';
import {commentApi, type CommentLikeResponse} from '@/lib/api/comment';
import {useAuth} from '@/hooks/useAuth';
import {useNavigate, Link} from '@tanstack/react-router';
import data from '@emoji-mart/data';
import Picker from '@emoji-mart/react';

interface CommentSectionProps {
    mediaId: string;
}

interface Comment {
    id: string;
    content?: string;
    media_id?: string;
    user_id?: string;
    username?: string;
    avatar?: string;
    parent_id?: string | null;
    status?: string;
    create_time?: string;
    update_time?: string;
    like_count?: number;
    is_liked?: boolean;
    is_reply?: boolean;
    reply_to_comment_id?: string | null;
    reply_to_username?: string | null;
    reply_to_content?: string | null;
}

const CommentSection: React.FC<CommentSectionProps> = ({mediaId}) => {
    const {t} = useTranslation();
    const {isAuthenticated, user} = useAuth();
    const navigate = useNavigate();
    const [comments, setComments] = useState<Comment[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [commentText, setCommentText] = useState('');
    const [isFocused, setIsFocused] = useState(false);
    const [showEmojiPicker, setShowEmojiPicker] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [replyingTo, setReplyingTo] = useState<{id: string; username: string} | null>(null);
    const [replyText, setReplyText] = useState('');
    const [showReplyEmojiPicker, setShowReplyEmojiPicker] = useState(false);
    const [isSubmittingReply, setIsSubmittingReply] = useState(false);
    const [commentLikes, setCommentLikes] = useState<Map<string, CommentLikeResponse>>(new Map());
    const [likingComments, setLikingComments] = useState<Set<string>>(new Set());
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);
    const [loadingMore, setLoadingMore] = useState(false);
    const [total, setTotal] = useState(0);
    const PAGE_SIZE = 10;

    const emojiPickerRef = useRef<HTMLDivElement>(null);
    const replyEmojiPickerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        fetchComments();
    }, [mediaId]);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (emojiPickerRef.current && !emojiPickerRef.current.contains(event.target as Node)) {
                setShowEmojiPicker(false);
            }
            if (replyEmojiPickerRef.current && !replyEmojiPickerRef.current.contains(event.target as Node)) {
                setShowReplyEmojiPicker(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const fetchComments = async (pageNum: number = 1, append: boolean = false) => {
        try {
            if (!append) setLoading(true);
            else setLoadingMore(true);
            setError(null);
            const response = await commentApi.getAll({media_id: mediaId, page: pageNum, page_size: PAGE_SIZE});
            const commentsList = response?.comments || [];
            const formattedComments: Comment[] = commentsList.map((comment: any) => ({
                id: comment.id || '',
                content: comment.content || '',
                media_id: comment.media_id || '',
                user_id: comment.user_id || '',
                username: comment.username || 'Anonymous',
                avatar: comment.avatar || '',
                parent_id: comment.parent_id || null,
                status: comment.status || '',
                create_time: comment.create_time || '',
                update_time: comment.update_time || '',
                like_count: comment.like_count || 0,
                is_liked: comment.is_liked || false,
                is_reply: comment.is_reply || false,
                reply_to_comment_id: comment.reply_to_comment_id || null,
                reply_to_username: comment.reply_to_username || null,
                reply_to_content: comment.reply_to_content || null,
            }));
            if (append) {
                setComments(prev => [...prev, ...formattedComments]);
            } else {
                setComments(formattedComments);
            }
            setPage(pageNum);
            const totalCount = response?.total || 0;
            setTotal(totalCount);
            setHasMore(pageNum * PAGE_SIZE < totalCount);
        } catch (err) {
            setError('Failed to fetch comments');
            console.error('Failed to fetch comments:', err);
        } finally {
            if (!append) setLoading(false);
            else setLoadingMore(false);
        }
    };

    const loadMore = () => {
        if (!hasMore || loadingMore) return;
        fetchComments(page + 1, true);
    };

    const handleSubmitComment = async () => {
        if (!commentText.trim()) return;
        if (!isAuthenticated) {
            navigate({to: '/auth/signin'});
            return;
        }

        try {
            setIsSubmitting(true);
            setError(null);
            await commentApi.create({media_id: mediaId, content: commentText});
            setCommentText('');
            setIsFocused(false);
            setShowEmojiPicker(false);
            await fetchComments();
        } catch (err: any) {
            console.error('Failed to submit comment:', err);
            setError(err.message || 'Failed to submit comment');
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleSubmitReply = async () => {
        if (!replyText.trim()) return;
        if (!isAuthenticated) {
            navigate({to: '/auth/signin'});
            return;
        }

        try {
            setIsSubmittingReply(true);
            setError(null);
            await commentApi.create({
                media_id: mediaId,
                parent_id: replyingTo?.id,
                content: replyText
            });
            setReplyText('');
            setShowReplyEmojiPicker(false);
            setReplyingTo(null);
            await fetchComments();
        } catch (err: any) {
            console.error('Failed to submit reply:', err);
            setError(err.message || 'Failed to submit reply');
        } finally {
            setIsSubmittingReply(false);
        }
    };

    const handleLikeComment = async (commentId: string) => {
        if (!isAuthenticated) {
            navigate({to: '/auth/signin'});
            return;
        }
        if (likingComments.has(commentId)) return;

        try {
            setLikingComments(prev => new Set(prev).add(commentId));
            const prevStatus = commentLikes.get(commentId);
            const newLiked = !prevStatus?.is_liked;

            if (prevStatus) {
                setCommentLikes(prev => {
                    const updated = new Map(prev);
                    updated.set(commentId, {
                        like_count: prevStatus.like_count + (newLiked ? 1 : -1),
                        is_liked: newLiked,
                        is_disliked: false,
                    });
                    return updated;
                });
            }

            const response: CommentLikeResponse = await commentApi.likes.toggle(commentId);

            setCommentLikes(prev => {
                const updated = new Map(prev);
                updated.set(commentId, response);
                return updated;
            });
        } catch (err) {
            console.error('Failed to like comment:', err);
        } finally {
            setLikingComments(prev => {
                const updated = new Set(prev);
                updated.delete(commentId);
                return updated;
            });
        }
    };

    const handleDislikeComment = async (commentId: string) => {
        if (!isAuthenticated) {
            navigate({to: '/auth/signin'});
            return;
        }
        if (likingComments.has(commentId)) return;

        try {
            setLikingComments(prev => new Set(prev).add(commentId));

            const response: CommentLikeResponse = await commentApi.likes.toggleDislike(commentId);

            setCommentLikes(prev => {
                const updated = new Map(prev);
                updated.set(commentId, response);
                return updated;
            });
        } catch (err) {
            console.error('Failed to dislike comment:', err);
        } finally {
            setLikingComments(prev => {
                const updated = new Set(prev);
                updated.delete(commentId);
                return updated;
            });
        }
    };

    const addEmoji = (emoji: any) => {
        setCommentText(prev => prev + emoji.native);
        setShowEmojiPicker(false);
        setIsFocused(true);
    };

    const addReplyEmoji = (emoji: any) => {
        setReplyText(prev => prev + emoji.native);
        setShowReplyEmojiPicker(false);
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center py-12">
                <Loader2 className="w-6 h-6 animate-spin text-blue-600"/>
            </div>
        );
    }

    if (error) {
        return (
            <div className="py-8 text-center text-red-500">
                <p>{error}</p>
                <Button variant="ghost" size="sm" onClick={fetchComments} className="mt-2">
                    {t('common.retry') || 'Retry'}
                </Button>
            </div>
        );
    }

    return (
        <div className="mt-8">
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white">
                    {comments.length} {t('watch.comments') || 'Comments'}
                </h3>
                <button className="flex items-center gap-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 px-3 py-1.5 rounded-full transition-colors">
                    <List className="w-4 h-4"/>
                    {t('watch.sortBy') || 'Sort by'}
                </button>
            </div>

            <div className="mb-6">
                {isAuthenticated ? (
                    <div className="flex gap-3 items-start">
                        <Avatar className="h-10 w-10 flex-shrink-0 mt-0.5">
                            <AvatarImage src={user?.avatar}/>
                            <AvatarFallback className="bg-gray-200 text-gray-600">
                                {user?.username?.[0]?.toUpperCase() || 'U'}
                            </AvatarFallback>
                        </Avatar>
                        <div className="flex-1 min-w-0">
                            {isFocused ? (
                                <div className="border border-blue-300 dark:border-blue-700 rounded-xl bg-white dark:bg-gray-900 overflow-hidden">
                                    <textarea
                                        placeholder={t('watch.addComment') || 'Add a comment...'}
                                        value={commentText}
                                        onChange={(e) => setCommentText(e.target.value)}
                                        onBlur={() => { setTimeout(() => { if (!commentText.trim()) setIsFocused(false); }, 150); }}
                                        onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey && commentText.trim()) { e.preventDefault(); handleSubmitComment(); }}}
                                        className="w-full min-h-[80px] px-4 py-3 resize-none focus:outline-none bg-transparent"
                                        rows={3}
                                        autoFocus
                                    />
                                    <div className="flex items-center justify-between px-3 py-2 border-t border-gray-100 dark:border-gray-800">
                                        <div className="relative" ref={emojiPickerRef}>
                                            <button
                                                type="button"
                                                onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                                                className="flex items-center gap-1 p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 transition-colors"
                                            >
                                                <SmilePlus className="w-5 h-5"/>
                                            </button>
                                            {showEmojiPicker && (
                                                <div className="absolute bottom-full mb-2 left-0 z-50 shadow-lg rounded-xl overflow-hidden">
                                                    <Picker
                                                        data={data}
                                                        onEmojiSelect={addEmoji}
                                                        theme="auto"
                                                        previewPosition="none"
                                                        skinTonePosition="search"
                                                        perLine={8}
                                                        maxFrequentRows={2}
                                                    />
                                                </div>
                                            )}
                                        </div>
                                        <div className="flex items-center gap-2">
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="sm"
                                                onClick={() => { setCommentText(''); setIsFocused(false); }}
                                                className="text-gray-500 hover:text-gray-700 font-medium"
                                            >
                                                {t('common.cancel') || 'Cancel'}
                                            </Button>
                                            <Button
                                                type="button"
                                                onClick={handleSubmitComment}
                                                disabled={isSubmitting || !commentText.trim()}
                                                size="sm"
                                                className={`font-medium px-4 ${
                                                    commentText.trim()
                                                        ? 'bg-blue-600 hover:bg-blue-700 text-white'
                                                        : 'bg-transparent text-gray-400 cursor-not-allowed'
                                                }`}
                                            >
                                                {isSubmitting ? (
                                                    <Loader2 className="w-4 h-4 animate-spin"/>
                                                ) : (
                                                    t('watch.postComment') || 'Comment'
                                                )}
                                            </Button>
                                        </div>
                                    </div>
                                </div>
                            ) : (
                                <div
                                    onClick={() => setIsFocused(true)}
                                    className="w-full px-4 py-2.5 border border-transparent bg-gray-50 dark:bg-gray-800 hover:bg-white dark:hover:bg-gray-900 rounded-full cursor-text transition-colors text-sm text-gray-500 dark:text-gray-400 select-none"
                                >
                                    {t('watch.addComment') || 'Add a comment...'}
                                </div>
                            )}
                        </div>
                    </div>
                ) : (
                    <div className="flex gap-3 items-start p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
                        <Avatar className="h-10 w-10 flex-shrink-0">
                            <AvatarFallback className="bg-gray-300 text-gray-500">?</AvatarFallback>
                        </Avatar>
                        <div className="flex-1">
                            <p className="text-sm text-gray-500 dark:text-gray-400 mb-3">
                                {t('watch.pleaseLoginToComment') || 'Sign in to add a comment...'}
                            </p>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => navigate({to: '/auth/signin'})}
                                className="border-blue-600 text-blue-600 hover:bg-blue-50"
                            >
                                <LogIn className="w-4 h-4 mr-2"/>
                                {t('auth.signin') || 'Sign in'}
                            </Button>
                        </div>
                    </div>
                )}
            </div>

            {comments.length === 0 ? (
                <div className="text-center py-12 text-gray-500 dark:text-gray-400">
                    <MessageCircle className="w-12 h-12 mx-auto mb-3 opacity-30"/>
                    <p>{t('watch.noComments') || 'No comments yet. Be the first to comment!'}</p>
                </div>
            ) : (
                <div className="space-y-0">
                    {comments.map((comment) => (
                        <div key={comment.id} className="flex gap-3 py-3">
                            <Avatar className="h-9 w-9 flex-shrink-0">
                                <AvatarImage src={comment.avatar || undefined}/>
                                <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-600 text-white text-xs font-medium">
                                    {(comment.username || 'U')[0]?.toUpperCase() || 'U'}
                                </AvatarFallback>
                            </Avatar>
                            <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 flex-wrap">
                                    <Link
                                        to="/@$username"
                                        params={{username: comment.username?.trim() || 'anonymous'}}
                                        className="font-semibold text-sm text-gray-900 dark:text-white hover:text-blue-600 cursor-pointer transition-colors"
                                    >
                                        @{comment.username?.trim() || 'Anonymous'}
                                    </Link>
                                    {comment.is_reply && comment.reply_to_username && (
                                        <>
                                            <span className="text-gray-400 text-xs">→</span>
                                            <Link
                                                to="/@$username"
                                                params={{username: comment.reply_to_username?.trim() || ''}}
                                                className="text-blue-600 text-sm font-medium hover:underline"
                                            >
                                                @{comment.reply_to_username}
                                            </Link>
                                        </>
                                    )}
                                    <span className="text-gray-500 dark:text-gray-400 text-xs">
                                        {formatDate(comment.create_time)}
                                    </span>
                                </div>

                                {comment.is_reply && comment.reply_to_content && (
                                    <div className="mt-1 px-3 py-1.5 bg-gray-50 dark:bg-gray-800/50 rounded-lg border-l-2 border-blue-200 dark:border-blue-800">
                                        <p className="text-xs text-gray-500 dark:text-gray-400 italic line-clamp-2">
                                            {comment.reply_to_content}
                                        </p>
                                    </div>
                                )}

                                <p className="text-[15px] text-gray-800 dark:text-gray-200 mt-1.5 leading-relaxed whitespace-pre-wrap break-words">
                                    {comment.content || <span className="text-gray-400 italic">No content</span>}
                                </p>

                                <div className="flex items-center gap-1 mt-2">
                                    <button
                                        className={`flex items-center gap-1.5 rounded-full p-1.5 transition-colors ${
                                            commentLikes.get(comment.id)?.is_liked
                                                ? 'text-blue-600 bg-blue-50 dark:bg-blue-900/20'
                                                : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'
                                        } ${likingComments.has(comment.id) ? 'opacity-50 cursor-not-allowed' : ''}`}
                                        onClick={() => handleLikeComment(comment.id)}
                                        disabled={likingComments.has(comment.id)}
                                    >
                                        {likingComments.has(comment.id) ? (
                                            <Loader2 className="w-4 h-4 animate-spin"/>
                                        ) : (
                                            <ThumbsUp className={`w-4 h-4 ${commentLikes.get(comment.id)?.is_liked ? 'fill-current' : ''}`}/>
                                        )}
                                        <span className={`text-xs font-medium ${
                                            commentLikes.get(comment.id)?.is_liked ? 'text-blue-600' : ''
                                        }`}>
                                            {(commentLikes.get(comment.id)?.like_count ?? comment.like_count ?? 0) || ''}
                                        </span>
                                    </button>

                                    <button
                                        className={`flex items-center gap-1.5 rounded-full p-1.5 transition-colors ${
                                            commentLikes.get(comment.id)?.is_disliked
                                                ? 'text-red-600 bg-red-50 dark:bg-red-900/20'
                                                : 'text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800'
                                        } ${likingComments.has(comment.id) ? 'opacity-50 cursor-not-allowed' : ''}`}
                                        onClick={() => handleDislikeComment(comment.id)}
                                        disabled={likingComments.has(comment.id)}
                                    >
                                        {likingComments.has(comment.id) ? (
                                            <Loader2 className="w-4 h-4 animate-spin"/>
                                        ) : (
                                            <ThumbsDown className={`w-4 h-4 ${commentLikes.get(comment.id)?.is_disliked ? 'fill-current' : ''}`}/>
                                        )}
                                    </button>

                                    <button
                                        className="flex items-center gap-1.5 text-gray-500 hover:text-blue-600 font-medium text-sm px-2 py-1 rounded-full hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
                                        onClick={() => setReplyingTo({id: comment.id, username: comment.username || 'Anonymous'})}
                                    >
                                        <Reply className="w-4 h-4"/>
                                        {t('common.reply') || 'Reply'}
                                    </button>
                                </div>

                                {replyingTo?.id === comment.id && (
                                    <div className="mt-3 space-y-2 border-l-2 border-gray-200 dark:border-gray-700 pl-4">
                                        {isAuthenticated ? (
                                            <div className="flex gap-3 items-start">
                                                <Avatar className="h-8 w-8 flex-shrink-0 mt-0.5">
                                                    <AvatarImage src={user?.avatar}/>
                                                    <AvatarFallback className="bg-gray-200 text-gray-600 text-xs">
                                                        {user?.username?.[0]?.toUpperCase() || 'U'}
                                                    </AvatarFallback>
                                                </Avatar>
                                                <div className="flex-1 space-y-2 min-w-0">
                                                    <div className="text-xs text-gray-500">
                                                        Replying to <span className="font-medium text-blue-600">@{replyingTo.username}</span>
                                                    </div>
                                                    <div className="border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-900 overflow-hidden">
                                                        <textarea
                                                            placeholder={t('watch.addComment') || 'Add a comment...'}
                                                            value={replyText}
                                                            onChange={(e) => setReplyText(e.target.value)}
                                                            onKeyDown={(e) => { if (e.key === 'Enter' && !e.shiftKey && replyText.trim()) { e.preventDefault(); handleSubmitReply(); }}}
                                                            className="w-full min-h-[60px] px-3 py-2 resize-none focus:outline-none bg-transparent"
                                                            rows={2}
                                                            autoFocus
                                                        />
                                                        <div className="flex items-center justify-between px-2.5 py-1.5 border-t border-gray-100 dark:border-gray-800">
                                                            <div className="relative" ref={replyEmojiPickerRef}>
                                                                <button
                                                                    type="button"
                                                                    onClick={() => setShowReplyEmojiPicker(!showReplyEmojiPicker)}
                                                                    className="flex items-center gap-1 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 transition-colors"
                                                                >
                                                                    <SmilePlus className="w-4 h-4"/>
                                                                </button>
                                                                {showReplyEmojiPicker && (
                                                                    <div className="absolute bottom-full mb-2 left-0 z-50 shadow-lg rounded-xl overflow-hidden">
                                                                        <Picker
                                                                            data={data}
                                                                            onEmojiSelect={addReplyEmoji}
                                                                            theme="auto"
                                                                            previewPosition="none"
                                                                            skinTonePosition="search"
                                                                            perLine={7}
                                                                        />
                                                                    </div>
                                                                )}
                                                            </div>
                                                            <div className="flex items-center gap-2">
                                                                <Button
                                                                    type="button"
                                                                    variant="ghost"
                                                                    size="sm"
                                                                    onClick={() => {setReplyingTo(null); setReplyText('');}}
                                                                    className="text-gray-500 hover:text-gray-700 font-medium"
                                                                >
                                                                    {t('common.cancel') || 'Cancel'}
                                                                </Button>
                                                                <Button
                                                                    type="button"
                                                                    onClick={handleSubmitReply}
                                                                    disabled={isSubmittingReply || !replyText.trim()}
                                                                    size="sm"
                                                                    className={`font-medium px-3 ${
                                                                        replyText.trim()
                                                                            ? 'bg-blue-600 hover:bg-blue-700 text-white'
                                                                            : 'bg-transparent text-gray-400 cursor-not-allowed'
                                                                    }`}
                                                                >
                                                                    {isSubmittingReply ? (
                                                                        <Loader2 className="w-4 h-4 animate-spin"/>
                                                                    ) : (
                                                                        t('common.reply') || 'Reply'
                                                                    )}
                                                                </Button>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        ) : (
                                            <div className="py-3">
                                                <p className="text-sm text-gray-500 mb-2">{t('watch.pleaseLoginToReply') || 'Sign in to reply'}</p>
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    onClick={() => navigate({to: '/auth/signin'})}
                                                    className="border-blue-600 text-blue-600 hover:bg-blue-50"
                                                >
                                                    <LogIn className="w-4 h-4 mr-2"/>
                                                    {t('auth.signin') || 'Sign in'}
                                                </Button>
                                            </div>
                                        )}
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {hasMore && !loading && (
                <div className="flex justify-center pt-4">
                    <button
                        onClick={loadMore}
                        disabled={loadingMore}
                        className="flex items-center gap-2 px-6 py-2.5 text-sm font-medium text-blue-600 bg-blue-50 dark:bg-blue-900/20 hover:bg-blue-100 dark:hover:bg-blue-900/40 rounded-full transition-colors disabled:opacity-50"
                    >
                        {loadingMore ? (
                            <>
                                <Loader2 className="w-4 h-4 animate-spin"/>
                                <span>{t('watch.loading') || 'Loading...'}</span>
                            </>
                        ) : (
                            <>
                                {t('watch.loadMore') || 'Load more comments'}
                                <span className="text-xs text-gray-500">({Math.max(0, total - comments.length)} remaining)</span>
                            </>
                        )}
                    </button>
                </div>
            )}
        </div>
    );
};

export default CommentSection;
