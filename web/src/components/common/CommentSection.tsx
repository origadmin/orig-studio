import React, {useState, useEffect} from 'react';
import {MessageCircle, ThumbsUp, Send, Loader2} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Textarea} from '@/components/ui/textarea';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDate, formatViews} from '@/lib/format';
import {commentApi} from '@/lib/api/comment';
import {likeApi} from '@/lib/api/like';
import ErrorPage from '@/components/common/ErrorPage';

interface CommentSectionProps {
    mediaId: string;
}

interface Comment {
    id: string;
    media_id?: string;
    user_id: string;
    username: string;
    avatar?: string;
    parent_id?: string;
    body: string;
    status: string;
    created_at: string;
    updated_at: string;
    like_count?: number;
    is_liked?: boolean;
    replies?: Comment[];
}

const CommentSection: React.FC<CommentSectionProps> = ({mediaId}) => {
    const {t} = useTranslation();
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
            // 处理评论和回复的嵌套结构
            const commentMap = new Map<string, Comment>();
            const topLevelComments: Comment[] = [];

            response.forEach(comment => {
                const formattedComment: Comment = {
                    ...comment,
                    replies: [],
                    like_count: 0,
                    is_liked: false
                };
                commentMap.set(comment.id, formattedComment);

                if (!comment.parent_id) {
                    topLevelComments.push(formattedComment);
                } else {
                    const parent = commentMap.get(comment.parent_id);
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

        try {
            setIsSubmitting(true);
            await commentApi.create({media_id: mediaId, body: commentText});
            setCommentText('');
            // 重新获取评论
            await fetchComments();
        } catch (err) {
            console.error('Failed to submit comment:', err);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleSubmitReply = async (parentId: string) => {
        if (!replyText.trim()) return;

        try {
            setIsSubmittingReply(true);
            await commentApi.create({media_id: mediaId, parent_id: parentId, body: replyText});
            setReplyText('');
            setReplyingTo(null);
            // 重新获取评论
            await fetchComments();
        } catch (err) {
            console.error('Failed to submit reply:', err);
        } finally {
            setIsSubmittingReply(false);
        }
    };

    const handleLikeComment = async (commentId: string) => {
        try {
            // 这里应该调用评论点赞的 API，暂时使用视频点赞 API 作为示例
            // 实际项目中需要实现评论点赞的 API
            // await commentApi.like(commentId);
            // 重新获取评论
            await fetchComments();
        } catch (err) {
            console.error('Failed to like comment:', err);
        }
    };

    const renderComment = (comment: Comment) => (
        <div key={comment.id} className="space-y-4">
            <div className="flex gap-3">
                <Avatar className="h-10 w-10">
                    <AvatarImage src={comment.avatar}/>
                    <AvatarFallback>{comment.username?.[0] || 'U'}</AvatarFallback>
                </Avatar>
                <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                        <span className="font-medium text-gray-900 dark:text-white">{comment.username}</span>
                        <span
                            className="text-xs text-gray-500 dark:text-gray-400">{formatDate(comment.created_at)}</span>
                    </div>
                    <p className="text-gray-800 dark:text-gray-200">{comment.body}</p>
                    <div className="flex items-center gap-4">
                        <button
                            className="flex items-center gap-1 text-gray-500 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400"
                            onClick={() => handleLikeComment(comment.id)}
                        >
                            <ThumbsUp className="w-4 h-4"/>
                            <span className="text-xs">{formatViews(comment.like_count || 0)}</span>
                        </button>
                        <button
                            className="flex items-center gap-1 text-gray-500 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400"
                            onClick={() => setReplyingTo(comment.id)}
                        >
                            <MessageCircle className="w-4 h-4"/>
                            <span className="text-xs">{t('common.reply')}</span>
                        </button>
                    </div>
                    {replyingTo === comment.id && (
                        <div className="mt-3 space-y-2">
                            <Textarea
                                placeholder={t('watch.addComment')}
                                value={replyText}
                                onChange={(e) => setReplyText(e.target.value)}
                                className="min-h-[80px]"
                            />
                            <div className="flex justify-end">
                                <Button
                                    onClick={() => handleSubmitReply(comment.id)}
                                    disabled={isSubmittingReply || !replyText.trim()}
                                    className="bg-blue-600 hover:bg-blue-700"
                                >
                                    {isSubmittingReply ? (
                                        <>
                                            <Loader2 className="w-4 h-4 mr-2 animate-spin"/>
                                            {t('common.submitting')}
                                        </>
                                    ) : (
                                        <>
                                            <Send className="w-4 h-4 mr-2"/>
                                            {t('watch.postComment')}
                                        </>
                                    )}
                                </Button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
            {comment.replies && comment.replies.length > 0 && (
                <div className="pl-12 space-y-4">
                    {comment.replies.map(reply => renderComment(reply))}
                </div>
            )}
        </div>
    );

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[200px]">
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    return (
        <div className="space-y-6">
            <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                {t('watch.comments')} ({comments.length})
            </h3>

            {/* Comment Form */}
            <div className="space-y-4">
                <Textarea
                    placeholder={t('watch.addComment')}
                    value={commentText}
                    onChange={(e) => setCommentText(e.target.value)}
                    className="min-h-[100px]"
                />
                <div className="flex justify-end">
                    <Button
                        onClick={handleSubmitComment}
                        disabled={isSubmitting || !commentText.trim()}
                        className="bg-blue-600 hover:bg-blue-700"
                    >
                        {isSubmitting ? (
                            <>
                                <Loader2 className="w-4 h-4 mr-2 animate-spin"/>
                                {t('common.submitting')}
                            </>
                        ) : (
                            <>
                                <Send className="w-4 h-4 mr-2"/>
                                {t('watch.postComment')}
                            </>
                        )}
                    </Button>
                </div>
            </div>

            {/* Comments List */}
            {comments.length === 0 ? (
                <div className="text-center py-12 text-gray-500 dark:text-gray-400">
                    {t('watch.noComments')}
                </div>
            ) : (
                <div className="space-y-6">
                    {comments.map(comment => renderComment(comment))}
                </div>
            )}
        </div>
    );
};

export default CommentSection;
