/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 视频播放页 - 对接真实数据
 */

import React, {useState, useEffect, useRef} from 'react';
import {useSearch, Link} from '@tanstack/react-router';
import {
    Loader2, RefreshCw, AlertTriangle, Edit, Trash2, FileText, Eye
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent} from '@/components/ui/card';
import {Skeleton} from '@/components/ui/skeleton';
import {formatViews, formatDate, formatDuration} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {publicMediaApi, mediaApi} from '@/lib/api/media';
import {commentApi} from '@/lib/api/comment';
import {usePublicMediaDetail, useMediaList, useDeleteMedia} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import ErrorPage from '@/components/common/ErrorPage';
import SubscribeButton from '@/components/common/SubscribeButton';
import CommentSection from '@/components/common/CommentSection';
import InteractionBar from '@/components/common/InteractionBar';
import VideoPlayer, {VideoPlayerHandle} from '@/components/common/VideoPlayer';

const WatchPage = () => {
    const {t} = useTranslation();
    const {v: shortToken} = useSearch({strict: false});
    // ✅ 使用新的 usePublicMediaDetail hook (short_token based)
    const {data: media, isLoading: isMediaLoading, error: mediaError} = usePublicMediaDetail(shortToken as string);
    const {user} = useAuth();
    const deleteMutation = useDeleteMedia();

    const [retrying, setRetrying] = useState(false);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

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
            // 使用 publicMediaApi 的转码重试（如果支持）或保留旧的 mediaApi
            await mediaApi.encoding.retry(media.id);
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
                        onError={(error) => {
                            console.error('Video player error:', error);
                        }}
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

                    <div
                        className="flex flex-wrap items-center justify-between gap-4 py-2 border-b dark:border-gray-800">
                        <div className="flex flex-col gap-3">
                            <div className="flex items-center gap-4">
                                <Link to={`/members?u=${media.user_id}`}>
                                    <Avatar className="h-12 w-12 ring-2 ring-gray-100 dark:ring-gray-800">
                                        <AvatarImage src={getImageUrl(mediaUser?.avatar, 'avatar')} loading="lazy"
                                                     onError={(e) => handleImageError(e, 'avatar')}/>
                                        <AvatarFallback>{mediaUser?.username?.[0] || 'U'}</AvatarFallback>
                                    </Avatar>
                                </Link>
                                <div>
                                    <Link to={`/members?u=${media.user_id}`}
                                          className="font-bold text-gray-900 dark:text-white hover:text-blue-600 transition-colors">
                                        {mediaUser?.nickname || mediaUser?.username || media?.username || 'Unknown User'}
                                    </Link>
                                    <p className="text-xs text-gray-500 dark:text-gray-400">{formatViews(mediaUser?.subscriber_count || 0)} {t('common.subscribers')}</p>
                                </div>
                                <SubscribeButton
                                    channelId={media.channel_id || ''}
                                    className="ml-4 rounded-full"
                                />
                            </div>

                            {/* Media owner controls */}
                            {user && media && user.id === media.user_id?.toString() && (
                                <div className="flex items-center gap-2 flex-nowrap">
                                    <Button variant="secondary" size="sm"
                                            className="gap-1 text-xs h-8 px-3 flex-shrink-0">
                                        <Edit className="w-3.5 h-3.5"/>
                                        <span>{t('common.edit')}</span>
                                    </Button>
                                    <Button variant="secondary" size="sm"
                                            className="gap-1 text-xs h-8 px-3 flex-shrink-0">
                                        <FileText className="w-3.5 h-3.5"/>
                                        <span>{t('common.subtitles')}</span>
                                    </Button>
                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        className="gap-1 text-xs h-8 px-3 flex-shrink-0"
                                        onClick={() => setShowDeleteConfirm(true)}
                                    >
                                        <Trash2 className="w-3.5 h-3.5"/>
                                        <span>{t('common.delete')}</span>
                                    </Button>
                                </div>
                            )}
                        </div>

                        <div className="flex items-center">
                            <InteractionBar
                                mediaId={String(media.id)}
                                shortToken={media.short_token || (shortToken as string)}
                                commentCount={media.comment_count}
                            />
                        </div>
                    </div>

                    {/* Meta & Description */}
                    <Card
                        className="bg-gray-100 dark:bg-gray-800 border-none shadow-none rounded-xl overflow-hidden mt-4">
                        <CardContent className="p-4 space-y-2">
                            <div className="flex gap-3 text-sm font-bold text-gray-900 dark:text-white">
                                <span>{formatViews(media.view_count)} {t('watch.views')}</span>
                                <span>{formatDate(media.created_at)}</span>
                                {media.tags?.map(tag => (
                                    <span key={tag}
                                          className="text-blue-600 dark:text-blue-400 cursor-pointer hover:underline">#{tag}</span>
                                ))}
                            </div>
                            <p className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed">
                                {media.description || t('watch.noDescription')}
                            </p>
                        </CardContent>
                    </Card>

                    {/* Comments Section */}
                    <div className="mt-8">
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
                                        <h4 className="text-sm font-bold text-gray-900 dark:text-white line-clamp-2 leading-snug group-hover:text-blue-600 transition-colors">
                                            {item.title}
                                        </h4>
                                        <p className="text-xs text-gray-500 mt-1">{recUser?.nickname || recUser?.username || 'Unknown'}</p>
                                        <div className="flex items-center gap-2 text-xs text-gray-400">
                                            <span>{formatViews(item.view_count)} views</span>
                                            <span>·</span>
                                            <span>{formatDate(item.created_at)}</span>
                                        </div>
                                    </div>
                                </Link>
                            );
                        })
                    )}
                </div>
            </div>

            {/* Custom Delete Confirmation Dialog */}
            {showDeleteConfirm && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl p-6 max-w-md w-full">
                        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-2">{t('common.delete')}</h3>
                        <p className="text-gray-600 dark:text-gray-400 mb-4">{t('watch.confirmDelete') || 'Are you sure you want to delete this video? This action cannot be undone.'}</p>
                        <div className="flex justify-end gap-2">
                            <Button
                                variant="secondary"
                                size="sm"
                                onClick={() => setShowDeleteConfirm(false)}
                            >
                                {t('common.cancel')}
                            </Button>
                            <Button
                                variant="destructive"
                                size="sm"
                                onClick={() => {
                                    setShowDeleteConfirm(false);
                                    handleDeleteMedia();
                                }}
                            >
                                {t('common.delete')}
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default WatchPage;
