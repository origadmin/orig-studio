import React, {useState, useEffect} from 'react';
import {MessageCircle, ThumbsUp, ThumbsDown, Send, Loader2, LogIn, MoreVertical, List} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Textarea} from '@/components/ui/textarea';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDate} from '@/lib/format';
import {commentApi} from '@/lib/api/comment';
import {useAuth} from '@/hooks/useAuth';
import {useNavigate, Link} from '@tanstack/react-router';

interface CommentSectionProps {
    mediaId: string;
}

interface Comment {
    id: string;
    media_id?: string;
    user_id?: string;
    username?: string;
    avatar?: string;
    parent_id?: string | null;
    content: string;
    status?: string;
    create_time?: string;
    update_time?: string;
    like_count?: number;
    is_liked?: boolean;
    replies?: Comment[];
}

const CommentSection: React.FC<CommentSectionProps> = ({mediaId}) => {
    const {t} = useTranslation();
    const {isAuthenticated, user} = useAuth();
    const navigate = useNavigate();
    const [comments, setComments] = useState<Comment[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [commentText, setCommentText] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [replyingTo, setReplyingTo] = useState<string | null>(null);
    const [replyText, setReplyText] = useState('');
    const [isSubmittingReply, setIsSubmittingReply] = useState(false);

    useEffect(() => {
        fetchComments();
    }, [mediaId]);

    const fetchComments = async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await commentApi.getAll({media_id: mediaId});
            const commentMap = new Map<string, Comment>();
            const topLevelComments: Comment[] = [];

            const commentsList = response?.comments || [];
            commentsList.forEach((comment: any) => {
                const formattedComment: Comment = {
                    id: comment.id || '',
                    media_id: comment.media_id || '',
                    user_id: comment.user_id || '',
                    username: comment.username || 'Anonymous',
                    avatar: comment.avatar || '',
                    parent_id: comment.parent_id || null,
                    content: comment.content || '',
                    status: comment.status || '',
                    create_time: comment.create_time || '',
                    update_time: comment.update_time || '',
                    like_count: comment.like_count || 0,
                    is_liked: comment.is_liked || false,
                    replies: comment.replies || []
                };
                commentMap.set(formattedComment.id, formattedComment);

                if (!formattedComment.parent_id) {
                    topLevelComments.push(formattedComment);
                } else {
                    const parent = commentMap.get(formattedComment.parent_id);
                    if (parent) {
                        parent.replies?.push(formattedComment);
                    }
                }
            });

            setComments(topLevelComments);
        } catch (err) {
            setError('Failed to fetch comments');
            console.error('Failed to fetch comments:', err);
        } finally {
            setLoading(false);
        }
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
            await fetchComments();
        } catch (err: any) {
            console.error('Failed to submit comment:', err);
            setError(err.message || 'Failed to submit comment');
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleSubmitReply = async (parentId: string) => {
        if (!replyText.trim()) return;
        if (!isAuthenticated) {
            navigate({to: '/auth/signin'});
            return;
        }

        try {
            setIsSubmittingReply(true);
            setError(null);
            await commentApi.create({media_id: mediaId, parent_id: parentId, content: replyText});
            setReplyText('');
            setReplyingTo(null);
            await fetchComments();
        } catch (err: any) {
            console.error('Failed to submit reply:', err);
            setError(err.message || 'Failed to submit reply');
        } finally {
            setIsSubmittingReply(false);
        }
    };

    const handleLikeComment = async (_commentId: string) => {
        try {
            await fetchComments();
        } catch (err) {
            console.error('Failed to like comment:', err);
        }
    };

    const renderComment = (comment: Comment, isReply: boolean = false) => (
        <div key={comment.id} className={`flex gap-3 ${isReply ? '' : 'py-4'}`}>
            <Avatar className={`${isReply ? 'h-8 w-8' : 'h-10 w-10'} flex-shrink-0`}>
                <AvatarImage src={comment.avatar}/>
                <AvatarFallback className="bg-gray-200 text-gray-600 text-sm">
                    {comment.username?.[0]?.toUpperCase() || 'U'}
                </AvatarFallback>
            </Avatar>
            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                    <Link
                        to="/@$username"
                        params={{username: comment.username || 'anonymous'}}
                        className={`font-medium text-gray-900 dark:text-white hover:text-blue-600 cursor-pointer ${isReply ? 'text-sm' : ''}`}
                    >
                        @{comment.username || 'Anonymous'}
                    </Link>
                    <span className={`text-gray-500 dark:text-gray-400 ${isReply ? 'text-xs' : 'text-xs'}`}>
                        {formatDate(comment.create_time)}
                        {comment.update_time && comment.update_time !== comment.create_time && (
                            <span className="ml-1">(edited)</span>
                        )}
                    </span>
                </div>
                <p className={`text-gray-900 dark:text-gray-100 mt-1 whitespace-pre-wrap ${isReply ? 'text-sm' : ''}`}>
                    {comment.content}
                </p>
                <div className="flex items-center gap-4 mt-2">
                    <button
                        className="flex items-center gap-1.5 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full p-1.5 transition-colors"
                        onClick={() => handleLikeComment(comment.id)}
                    >
                        <ThumbsUp className="w-4 h-4"/>
                    </button>
                    <button className="flex items-center gap-1.5 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full p-1.5 transition-colors">
                        <ThumbsDown className="w-4 h-4"/>
                    </button>
                    <button
                        className="flex items-center gap-1.5 text-gray-500 hover:text-blue-600 font-medium text-sm px-2 py-1 rounded-full hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
                        onClick={() => setReplyingTo(comment.id)}
                    >
                        {t('common.reply') || 'Reply'}
                    </button>
                </div>
                {replyingTo === comment.id && (
                    <div className="mt-4 space-y-3 border-l-2 border-gray-200 dark:border-gray-700 pl-4">
                        {isAuthenticated ? (
                            <>
                                <div className="flex gap-3 items-start">
                                    <Avatar className="h-8 w-8 flex-shrink-0">
                                        <AvatarImage src={user?.avatar}/>
                                        <AvatarFallback className="bg-gray-200 text-gray-600 text-xs">
                                            {user?.username?.[0]?.toUpperCase() || 'U'}
                                        </AvatarFallback>
                                    </Avatar>
                                    <div className="flex-1 space-y-2">
                                        <textarea
                                            placeholder={t('watch.addComment') || 'Add a comment...'}
                                            value={replyText}
                                            onChange={(e) => setReplyText(e.target.value)}
                                            className="w-full min-h-[60px] px-3 py-2 border border-transparent focus:border-blue-500 focus:outline-none resize-none rounded-lg bg-transparent hover:bg-gray-50 dark:hover:bg-gray-800"
                                            rows={2}
                                        />
                                        <div className="flex justify-end gap-2">
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                onClick={() => {setReplyingTo(null); setReplyText('');}}
                                                className="text-gray-500"
                                            >
                                                {t('common.cancel') || 'Cancel'}
                                            </Button>
                                            <Button
                                                onClick={() => handleSubmitReply(comment.id)}
                                                disabled={isSubmittingReply || !replyText.trim()}
                                                size="sm"
                                                className="bg-blue-600 hover:bg-blue-700 text-white disabled:opacity-50"
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
                            </>
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
            {comment.replies && comment.replies.length > 0 && (
                <div className="ml-12 space-y-0 mt-2">
                    {comment.replies.map(reply => renderComment(reply, true))}
                </div>
            )}
        </div>
    );

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
            {/* Header */}
            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-gray-900 dark:text-white">
                    {comments.length} {t('watch.comments') || 'Comments'}
                </h3>
                <button className="flex items-center gap-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 px-3 py-1.5 rounded-full transition-colors">
                    <List className="w-4 h-4"/>
                    {t('watch.sortBy') || 'Sort by'}
                </button>
            </div>

            {/* Comment Input */}
            <div className="mb-6">
                {isAuthenticated ? (
                    <div className="flex gap-3 items-start">
                        <Avatar className="h-10 w-10 flex-shrink-0">
                            <AvatarImage src={user?.avatar}/>
                            <AvatarFallback className="bg-gray-200 text-gray-600">
                                {user?.username?.[0]?.toUpperCase() || 'U'}
                            </AvatarFallback>
                        </Avatar>
                        <div className="flex-1">
                            <textarea
                                placeholder={t('watch.addComment') || 'Add a comment...'}
                                value={commentText}
                                onChange={(e) => setCommentText(e.target.value)}
                                className="w-full min-h-[80px] px-4 py-3 border border-transparent bg-gray-50 dark:bg-gray-800 rounded-xl focus:border-blue-500 focus:bg-white dark:focus:bg-gray-900 focus:outline-none resize-none transition-colors"
                                rows={3}
                            />
                            {(commentText.trim()) && (
                                <div className="flex justify-end gap-2 mt-2">
                                    <Button
                                        variant="ghost"
                                        size="sm"
                                        onClick={() => setCommentText('')}
                                        className="text-gray-500"
                                    >
                                        {t('common.cancel') || 'Cancel'}
                                    </Button>
                                    <Button
                                        onClick={handleSubmitComment}
                                        disabled={isSubmitting || !commentText.trim()}
                                        size="sm"
                                        className="bg-blue-600 hover:bg-blue-700 text-white disabled:opacity-50"
                                    >
                                        {isSubmitting ? (
                                            <Loader2 className="w-4 h-4 animate-spin"/>
                                        ) : (
                                            <>
                                                {t('watch.postComment') || 'Comment'}
                                            </>
                                        )}
                                    </Button>
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

            {/* Comments List */}
            {comments.length === 0 ? (
                <div className="text-center py-12 text-gray-500 dark:text-gray-400">
                    <MessageCircle className="w-12 h-12 mx-auto mb-3 opacity-30"/>
                    <p>{t('watch.noComments') || 'No comments yet. Be the first to comment!'}</p>
                </div>
            ) : (
                <div className="divide-y divide-gray-100 dark:divide-gray-800">
                    {comments.map(comment => renderComment(comment))}
                </div>
            )}
        </div>
    );
};

export default CommentSection;
