/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 视频播放页 - 对接真实数据
 */

import React, {useState, useEffect, useRef} from 'react';
import {useSearch, useNavigate, Link} from '@tanstack/react-router';
import {
    Loader2, RefreshCw, AlertTriangle, Trash2, FileText, Eye, Pencil
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent} from '@/components/ui/card';
import {Skeleton} from '@/components/ui/skeleton';
import {formatViews, formatDate, formatDuration} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {publicMediaApi, adminMediaApi, encodingApi} from '@/lib/api/media';
import {spriteApi} from '@/lib/api/sprite';
import {commentApi} from '@/lib/api/comment';
import {usePublicMediaDetail, useMediaList, useDeleteMedia} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import ErrorPage from '@/components/common/ErrorPage';
import SubscribeButton from '@/components/common/SubscribeButton';
import CommentSection from '@/components/common/CommentSection';
import InteractionBar from '@/components/common/InteractionBar';
import VideoPlayer, {VideoPlayerHandle} from '@/components/common/VideoPlayer';
import {DeleteConfirmDialog} from '@/components/common/DeleteConfirmDialog';
import {toast} from 'sonner';

const WatchPage = () => {
    const {t} = useTranslation();
    const {v: shortToken} = useSearch({strict: false});
    const navigate = useNavigate();
    // ✅ 使用新的 usePublicMediaDetail hook (short_token based)
    const {data: media, isLoading: isMediaLoading, error: mediaError} = usePublicMediaDetail(shortToken as string);
    const {user, isAdmin} = useAuth();
    const deleteMutation = useDeleteMedia();

    const [retrying, setRetrying] = useState(false);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const commentSectionRef = useRef<HTMLDivElement>(null);
    const viewCountedRef = useRef(false);

    // Reset view count tracking when shortToken changes
    useEffect(() => {
        viewCountedRef.current = false;
    }, [shortToken]);

    // Video player ref for external control
    const videoPlayerRef = useRef<VideoPlayerHandle>(null);

    const {data: recData} = useMediaList({
        page_size: 10,
        category_id: media?.edges?.category?.id || undefined,
        status: 'active'
    });

    // 推荐视频过滤：使用 short_token 过滤当前视频
    const recommendations = recData?.items?.filter(m => m.short_token !== shortToken) || [];
    const loading = isMediaLoading;
    const error = mediaError ? t('watch.failedToLoad') : null;

    // Handle media deletion
    const handleDeleteMedia = async () => {
        if (!media) return;

        try {
            // 使用 media.id 删除（Admin API 需要 ID）
            await deleteMutation.mutateAsync(media.id);
            window.location.href = '/';
        } catch (err) {
            console.error('Failed to delete media:', err);
        }
    };

    // Retry transcoding handler
    const handleRetry = async () => {
        if (!media || retrying) return;
        setRetrying(true);
        try {
            // 使用 encodingApi 的 admin 路径重试所有失败任务
            // mediaApi.encoding.retry 使用了不存在的 public 路径，且 :taskId 未替换
            await encodingApi.retryAllFailed(media.id);
            setTimeout(() => window.location.reload(), 1000);
        } catch {
        } finally {
            setRetrying(false);
        }
    };

    if (loading) {
        return (
            <div className="flex flex-col lg:flex-row gap-6 animate-pulse">
                <div className="flex-1 space-y-4">
                    <Skeleton className="aspect-video w-full rounded-2xl"/>
                    <Skeleton className="h-8 w-3/4"/>
                    <div className="flex justify-between items-center">
                        <div className="flex items-center gap-3">
                            <Skeleton className="h-12 w-12 rounded-full"/>
                            <div className="space-y-2">
                                <Skeleton className="h-4 w-24"/>
                                <Skeleton className="h-3 w-16"/>
                            </div>
                        </div>
                        <Skeleton className="h-10 w-32 rounded-full"/>
                    </div>
                </div>
                <div className="lg:w-80 xl:w-96 space-y-4">
                    {Array.from({length: 5}).map((_, i) => (
                        <div key={i} className="flex gap-3">
                            <Skeleton className="w-40 aspect-video rounded-lg shrink-0"/>
                            <div className="flex-1 space-y-2 py-1">
                                <Skeleton className="h-4 w-full"/>
                                <Skeleton className="h-3 w-2/3"/>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    if (error || !media) {
        return <ErrorPage
            statusCode={404}
            title={error || t('watch.videoNotFound')}
            message={t('error.404Message')}
        />;
    }

    const mediaUser = media.edges?.user?.[0];
    const isProcessing = media.encoding_status !== 'success';

    return (
        <div className="flex flex-col lg:flex-row gap-6 relative">
            {/* Main Content: Player & Details */}
            <div className="flex-1 min-w-0">
                {/* Player Container with new YouTube-style VideoPlayer */}
                <div className="relative">
                    <VideoPlayer
                        ref={videoPlayerRef}
                        src={media.url || ''}
                        hlsSrc={media.hls_file}
                        poster={media.poster || media.thumbnail}
                        spriteVttUrl={media.sprite_status === 'success' && media.short_token ? spriteApi.getVttUrl(media.short_token) : undefined}
                        enableSpritePreview={true}
                        onPlay={() => {
                            if (!viewCountedRef.current && media.short_token) {
                                viewCountedRef.current = true;
                                publicMediaApi.incrementViewCount(media.short_token).catch(() => {});
                            }
                        }}
                        onError={(error) => {
                            console.error('Video player error:', error);
                        }}
                        onAutoPlayNext={() => {
                            if (recData?.items && recData.items.length > 0) {
                                const currentIndex = recData.items.findIndex((v: any) => v.short_token === shortToken);
                                const nextIndex = currentIndex >= 0 ? currentIndex + 1 : 0;
                                if (nextIndex < recData.items.length) {
                                    navigate({to: '/watch', search: {v: recData.items[nextIndex].short_token}});
                                }
                            }
                        }}
                        autoPlay={false}
                    />
                    
                    {/* Encoding Status Indicator */}
                    {isProcessing && (
                        <div className="absolute top-3 left-3 z-20 flex items-center gap-2">
                            <Badge
                                variant="secondary"
                                className="gap-1 bg-black/60 text-white border-white/20 backdrop-blur-md text-[10px] px-1.5 py-0 h-5 whitespace-nowrap"
                            >
                                {media.encoding_status === 'processing' ? (
                                    <><Loader2 size={9}
                                               className="animate-spin"/>{t('watch.transcoding') || 'Transcoding...'}</>
                                ) : media.encoding_status === 'failed' ? (
                                    <><AlertTriangle size={9}/>{t('watch.failed') || 'Failed'}</>
                                ) : media.encoding_status === 'pending' ? (
                                    <><Eye size={9}/>{t('watch.optimizing') || 'Queued'}</>
                                ) : (
                                    <><Eye size={9}/>{t('watch.partial') || 'Partial'}</>
                                )}
                            </Badge>

                            {/* Retry button for failed status */}
                            {media.encoding_status === 'failed' && (
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    className="gap-1 bg-black/60 hover:bg-red-600/80 text-white border-white/20 backdrop-blur-md text-[10px] px-1.5 h-5"
                                    onClick={handleRetry}
                                    disabled={retrying}
                                >
                                    <RefreshCw size={9} className={retrying ? 'animate-spin' : ''}/>
                                    {retrying ? 'Retrying...' : 'Retry'}
                                </Button>
                            )}

                            {/* Fallback to MP4 indicator for non-success states */}
                            {media.encoding_status !== 'success' && media.url && (
                                <Badge
                                    variant="outline"
                                    className="gap-1 bg-black/60 text-yellow-300 border-yellow-500/30 backdrop-blur-md text-[10px] px-1.5 py-0 h-5"
                                >
                                    MP4 Fallback
                                </Badge>
                            )}
                        </div>
                    )}
                </div>

                {/* Video Info */}
                <div className="mt-6 space-y-4">
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white line-clamp-2">
                        {media.title}
                    </h1>

                    {(user && (String(user.id) === String(media.user_id) || isAdmin)) && (
                        <div className="flex items-center gap-2 mt-2">
                            <Button
                                variant="ghost"
                                size="sm"
                                className="gap-1.5"
                                onClick={() => navigate({to: '/media/$shortToken/edit', params: {shortToken: media.short_token || (shortToken as string)}} as any)}
                            >
                                <Pencil className="w-4 h-4"/>
                                {t('common.edit')}
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="gap-1.5"
                                disabled
                                onClick={() => toast.info(t('watch.subtitleComingSoon') || '字幕功能即将上线')}
                            >
                                <FileText className="w-4 h-4"/>
                                {t('common.subtitles')}
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="gap-1.5 text-destructive hover:text-destructive"
                                onClick={() => setShowDeleteConfirm(true)}
                            >
                                <Trash2 className="w-4 h-4"/>
                                {t('common.delete')}
                            </Button>
                        </div>
                    )}

                    <div
                        className="flex flex-wrap items-center justify-between gap-4 py-2 border-b dark:border-gray-800">
                        <div className="flex flex-col gap-3">
                            <div className="flex items-center gap-4">
                                {mediaUser ? (
                                    <Link to={`/@${mediaUser.username}`}>
                                        <Avatar className="h-12 w-12 ring-2 ring-gray-100 dark:ring-gray-800">
                                            <AvatarImage src={getImageUrl(mediaUser.avatar, 'avatar')} loading="lazy"
                                                         onError={(e) => handleImageError(e, 'avatar')}/>
                                            <AvatarFallback>{mediaUser.username?.[0] || 'U'}</AvatarFallback>
                                        </Avatar>
                                    </Link>
                                ) : (
                                    <Avatar className="h-12 w-12 ring-2 ring-gray-100 dark:ring-gray-800">
                                        <AvatarFallback>?</AvatarFallback>
                                    </Avatar>
                                )}
                                <div>
                                    {mediaUser ? (
                                        <>
                                            <Link to={`/@${mediaUser.username}`}
                                                  className="font-bold text-gray-900 dark:text-white hover:text-info transition-colors">
                                                {mediaUser.nickname || mediaUser.username}
                                            </Link>
                                            <p className="text-xs text-gray-500 dark:text-muted-foreground">{formatViews(mediaUser.subscriber_count || 0)} {t('common.subscribers')}</p>
                                        </>
                                    ) : (
                                        <span className="font-bold text-muted-foreground dark:text-gray-500">{t('watch.deletedUser') || 'Deleted User'}</span>
                                    )}
                                </div>
                                <SubscribeButton
                                    channelId={media.channel_id || ''}
                                    className="ml-4 rounded-full"
                                />
                            </div>
                        </div>

                        <div className="flex items-center">
                            <InteractionBar
                                mediaId={String(media.id)}
                                shortToken={media.short_token || (shortToken as string)}
                                commentCount={media.comment_count}
                                onCommentClick={() => {
                                    commentSectionRef.current?.scrollIntoView({behavior: 'smooth', block: 'start'});
                                }}
                            />
                        </div>
                    </div>

                    {/* Meta & Description */}
                    <Card
                        className="bg-gray-100 dark:bg-gray-800 border-none shadow-none rounded-xl overflow-hidden mt-4">
                        <CardContent className="p-4 space-y-2">
                            <div className="flex gap-3 text-sm font-bold text-gray-900 dark:text-white">
                                <span>{formatViews(media.view_count)} {t('watch.views')}</span>
                                <span>{formatDate(media.create_time)}</span>
                                {media.tags?.map(tag => (
                                    <span key={tag}
                                          className="text-info dark:text-blue-400 cursor-pointer hover:underline">#{tag}</span>
                                ))}
                            </div>
                            <p className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed">
                                {media.description || t('watch.noDescription')}
                            </p>
                        </CardContent>
                    </Card>

                    {/* Comments Section */}
                    <div className="mt-8" ref={commentSectionRef}>
                        <CommentSection mediaId={String(media.id)}/>
                    </div>
                </div>
            </div>

            {/* Sidebar: Recommendations */}
            <div className="lg:w-80 xl:w-96 shrink-0 space-y-4">
                <h3 className="font-bold text-lg text-gray-900 dark:text-white flex items-center gap-2 mb-4">
                    {t('watch.nextVideos')}
                </h3>

                <div className="space-y-4">
                    {recommendations.length === 0 ? (
                        <p className="text-sm text-gray-500 py-4 italic">{t('watch.noRecommendations')}</p>
                    ) : (
                        recommendations.map((item) => {
                            const recUser = item.edges?.user?.[0];
                            const recThumb = getImageUrl(item.thumbnail, 'thumbnail');

                            return (
                                <Link
                                    key={item.id}
                                    to="/watch"
                                    search={{v: item.short_token}}
                                    className="flex gap-3 group"
                                >
                                    <div className="relative w-40 aspect-video rounded-lg overflow-hidden shrink-0">
                                        <img
                                            src={recThumb}
                                            alt={item.title}
                                            loading="lazy"
                                            onError={(e) => handleImageError(e, 'thumbnail')}
                                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                                        />
                                        <div
                                            className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1 rounded">
                                            {formatDuration(item.duration)}
                                        </div>
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <h4 className="text-sm font-bold text-gray-900 dark:text-white line-clamp-2 leading-snug group-hover:text-info transition-colors">
                                            {item.title}
                                        </h4>
                                        <p className="text-xs text-gray-500 mt-1">{recUser?.nickname || recUser?.username || 'Unknown'}</p>
                                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                            <span>{formatViews(item.view_count)} views</span>
                                            <span>·</span>
                                            <span>{formatDate(item.create_time)}</span>
                                        </div>
                                    </div>
                                </Link>
                            );
                        })
                    )}
                </div>
            </div>

            <DeleteConfirmDialog
                open={showDeleteConfirm}
                onOpenChange={setShowDeleteConfirm}
                title={media.title}
                isDeleting={deleteMutation.isPending}
                onConfirm={handleDeleteMedia}
            />
        </div>
    );
};

export default WatchPage;
