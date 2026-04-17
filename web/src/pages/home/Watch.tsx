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
import {mediaApi} from '@/lib/api/media';
import {commentApi} from '@/lib/api/comment';
import {useMediaDetail, useMediaList, useDeleteMedia} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import ErrorPage from '@/components/common/ErrorPage';
import SubscribeButton from '@/components/common/SubscribeButton';
import CommentSection from '@/components/common/CommentSection';
import InteractionBar from '@/components/common/InteractionBar';
import VideoPlayer from '@/components/common/VideoPlayer';

const WatchPage = () =&gt; {
    const {t} = useTranslation();
    const {v: id} = useSearch({strict: false});
    const {data: media, isLoading: isMediaLoading, error: mediaError} = useMediaDetail(id as string);
    const {user} = useAuth();
    const deleteMutation = useDeleteMedia();

    const [retrying, setRetrying] = useState(false);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

    const {data: recData} = useMediaList({
        page_size: 10,
        category_id: media?.edges?.category?.id || undefined,
        status: 'active'
    });

    const recommendations = recData?.items?.filter(m =&gt; m.id !== Number(id)) || [];
    const loading = isMediaLoading;
    const error = mediaError ? t('watch.failedToLoad') : null;

    // Handle media deletion
    const handleDeleteMedia = async () =&gt; {
        if (!media) return;

        try {
            await deleteMutation.mutateAsync(id as string);
            // Redirect to home page after deletion
            window.location.href = '/';
        } catch (err) {
            console.error('Failed to delete media:', err);
        }
    };

    // Retry transcoding handler
    const handleRetry = async () =&gt; {
        if (!media || retrying) return;
        setRetrying(true);
        try {
            await mediaApi.encoding.retry(media.id);
            // Reload the page data after a short delay to show processing state
            setTimeout(() =&gt; window.location.reload(), 1000);
        } catch {
            // Error silently — user can see the button still there
        } finally {
            setRetrying(false);
        }
    };

    if (loading) {
        return (
            &lt;div className="flex flex-col lg:flex-row gap-6 animate-pulse"&gt;
                &lt;div className="flex-1 space-y-4"&gt;
                    &lt;Skeleton className="aspect-video w-full rounded-2xl"/&gt;
                    &lt;Skeleton className="h-8 w-3/4"/&gt;
                    &lt;div className="flex justify-between items-center"&gt;
                        &lt;div className="flex items-center gap-3"&gt;
                            &lt;Skeleton className="h-12 w-12 rounded-full"/&gt;
                            &lt;div className="space-y-2"&gt;
                                &lt;Skeleton className="h-4 w-24"/&gt;
                                &lt;Skeleton className="h-3 w-16"/&gt;
                            &lt;/div&gt;
                        &lt;/div&gt;
                        &lt;Skeleton className="h-10 w-32 rounded-full"/&gt;
                    &lt;/div&gt;
                &lt;/div&gt;
                &lt;div className="lg:w-80 xl:w-96 space-y-4"&gt;
                    {Array.from({length: 5}).map((_, i) =&gt; (
                        &lt;div key={i} className="flex gap-3"&gt;
                            &lt;Skeleton className="w-40 aspect-video rounded-lg shrink-0"/&gt;
                            &lt;div className="flex-1 space-y-2 py-1"&gt;
                                &lt;Skeleton className="h-4 w-full"/&gt;
                                &lt;Skeleton className="h-3 w-2/3"/&gt;
                            &lt;/div&gt;
                        &lt;/div&gt;
                    ))}
                &lt;/div&gt;
            &lt;/div&gt;
        );
    }

    if (error || !media) {
        return &lt;ErrorPage
            statusCode={404}
            title={error || t('watch.videoNotFound')}
            message={t('error.404Message')}
        /&gt;;
    }

    const mediaUser = media.edges?.user?.[0];
    const isProcessing = media.encoding_status !== 'success';

    return (
        &lt;div className="flex flex-col lg:flex-row gap-6 relative"&gt;
            {/* Main Content: Player &amp; Details */}
            &lt;div className="flex-1 min-w-0"&gt;
                {/* Player Container with new YouTube-style VideoPlayer */}
                &lt;div className="relative"&gt;
                    &lt;VideoPlayer
                        src={media.url}
                        hlsSrc={media.hls_file}
                        poster={media.poster || media.thumbnail}
                    /&gt;
                    
                    {/* Encoding Status Indicator */}
                    {isProcessing &amp;&amp; (
                        &lt;div className="absolute top-3 left-3 z-20 flex items-center gap-2"&gt;
                            &lt;Badge
                                variant="secondary"
                                className="gap-1 bg-black/60 text-white border-white/20 backdrop-blur-md text-[10px] px-1.5 py-0 h-5 whitespace-nowrap"
                            &gt;
                                {media.encoding_status === 'processing' ? (
                                    &lt;&gt;&lt;Loader2 size={9}
                                               className="animate-spin"/&gt;{t('watch.transcoding') || 'Transcoding...'}&lt;/&gt;
                                ) : media.encoding_status === 'failed' ? (
                                    &lt;&gt;&lt;AlertTriangle size={9}/&gt;{t('watch.failed') || 'Failed'}&lt;/&gt;
                                ) : media.encoding_status === 'pending' ? (
                                    &lt;&gt;&lt;Eye size={9}/&gt;{t('watch.optimizing') || 'Queued'}&lt;/&gt;
                                ) : (
                                    &lt;&gt;&lt;Eye size={9}/&gt;{t('watch.partial') || 'Partial'}&lt;/&gt;
                                )}
                            &lt;/Badge&gt;

                            {/* Retry button for failed status */}
                            {media.encoding_status === 'failed' &amp;&amp; (
                                &lt;Button
                                    variant="secondary"
                                    size="sm"
                                    className="gap-1 bg-black/60 hover:bg-red-600/80 text-white border-white/20 backdrop-blur-md text-[10px] px-1.5 h-5"
                                    onClick={handleRetry}
                                    disabled={retrying}
                                &gt;
                                    &lt;RefreshCw size={9} className={retrying ? 'animate-spin' : ''}/&gt;
                                    {retrying ? 'Retrying...' : 'Retry'}
                                &lt;/Button&gt;
                            )}

                            {/* Fallback to MP4 indicator for non-success states */}
                            {media.encoding_status !== 'success' &amp;&amp; media.url &amp;&amp; (
                                &lt;Badge
                                    variant="outline"
                                    className="gap-1 bg-black/60 text-yellow-300 border-yellow-500/30 backdrop-blur-md text-[10px] px-1.5 py-0 h-5"
                                &gt;
                                    MP4 Fallback
                                &lt;/Badge&gt;
                            )}
                        &lt;/div&gt;
                    )}
                &lt;/div&gt;

                {/* Video Info */}
                &lt;div className="mt-6 space-y-4"&gt;
                    &lt;h1 className="text-2xl font-bold text-gray-900 dark:text-white line-clamp-2"&gt;
                        {media.title}
                    &lt;/h1&gt;

                    &lt;div
                        className="flex flex-wrap items-center justify-between gap-4 py-2 border-b dark:border-gray-800"&gt;
                        &lt;div className="flex flex-col gap-3"&gt;
                            &lt;div className="flex items-center gap-4"&gt;
                                &lt;Link to={`/members?u=${media.user_id}`}&gt;
                                    &lt;Avatar className="h-12 w-12 ring-2 ring-gray-100 dark:ring-gray-800"&gt;
                                        &lt;AvatarImage src={getImageUrl(mediaUser?.avatar, 'avatar')} loading="lazy"
                                                     onError={(e) =&gt; handleImageError(e, 'avatar')}/&gt;
                                        &lt;AvatarFallback&gt;{mediaUser?.username?.[0] || 'U'}&lt;/AvatarFallback&gt;
                                    &lt;/Avatar&gt;
                                &lt;/Link&gt;
                                &lt;div&gt;
                                    &lt;Link to={`/members?u=${media.user_id}`}
                                          className="font-bold text-gray-900 dark:text-white hover:text-blue-600 transition-colors"&gt;
                                        {mediaUser?.nickname || mediaUser?.username || media?.username || 'Unknown User'}
                                    &lt;/Link&gt;
                                    &lt;p className="text-xs text-gray-500 dark:text-gray-400"&gt;{formatViews(mediaUser?.subscriber_count || 0)} {t('common.subscribers')}&lt;/p&gt;
                                &lt;/div&gt;
                                &lt;SubscribeButton
                                    userId={media.user_id?.toString() || ''}
                                    className="ml-4 rounded-full"
                                /&gt;
                            &lt;/div&gt;

                            {/* Media owner controls */}
                            {user &amp;&amp; media &amp;&amp; user.id === media.user_id?.toString() &amp;&amp; (
                                &lt;div className="flex items-center gap-2 flex-nowrap"&gt;
                                    &lt;Button variant="secondary" size="sm"
                                            className="gap-1 text-xs h-8 px-3 flex-shrink-0"&gt;
                                        &lt;Edit className="w-3.5 h-3.5"/&gt;
                                        &lt;span&gt;{t('common.edit')}&lt;/span&gt;
                                    &lt;/Button&gt;
                                    &lt;Button variant="secondary" size="sm"
                                            className="gap-1 text-xs h-8 px-3 flex-shrink-0"&gt;
                                        &lt;FileText className="w-3.5 h-3.5"/&gt;
                                        &lt;span&gt;{t('common.subtitles')}&lt;/span&gt;
                                    &lt;/Button&gt;
                                    &lt;Button
                                        variant="destructive"
                                        size="sm"
                                        className="gap-1 text-xs h-8 px-3 flex-shrink-0"
                                        onClick={() =&gt; setShowDeleteConfirm(true)}
                                    &gt;
                                        &lt;Trash2 className="w-3.5 h-3.5"/&gt;
                                        &lt;span&gt;{t('common.delete')}&lt;/span&gt;
                                    &lt;/Button&gt;
                                &lt;/div&gt;
                            )}
                        &lt;/div&gt;

                        &lt;div className="flex items-center"&gt;
                            &lt;InteractionBar
                                mediaId={id as string}
                                commentCount={media.comment_count}
                            /&gt;
                        &lt;/div&gt;
                    &lt;/div&gt;

                    {/* Meta &amp; Description */}
                    &lt;Card
                        className="bg-gray-100 dark:bg-gray-800 border-none shadow-none rounded-xl overflow-hidden mt-4"&gt;
                        &lt;CardContent className="p-4 space-y-2"&gt;
                            &lt;div className="flex gap-3 text-sm font-bold text-gray-900 dark:text-white"&gt;
                                &lt;span&gt;{formatViews(media.view_count)} {t('watch.views')}&lt;/span&gt;
                                &lt;span&gt;{formatDate(media.created_at)}&lt;/span&gt;
                                {media.tags?.map(tag =&gt; (
                                    &lt;span key={tag}
                                          className="text-blue-600 dark:text-blue-400 cursor-pointer hover:underline"&gt;#{tag}&lt;/span&gt;
                                ))}
                            &lt;/div&gt;
                            &lt;p className="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed"&gt;
                                {media.description || t('watch.noDescription')}
                            &lt;/p&gt;
                        &lt;/CardContent&gt;
                    &lt;/Card&gt;

                    {/* Comments Section */}
                    &lt;div className="mt-8"&gt;
                        &lt;CommentSection mediaId={id as string}/&gt;
                    &lt;/div&gt;
                &lt;/div&gt;
            &lt;/div&gt;

            {/* Sidebar: Recommendations */}
            &lt;div className="lg:w-80 xl:w-96 shrink-0 space-y-4"&gt;
                &lt;h3 className="font-bold text-lg text-gray-900 dark:text-white flex items-center gap-2 mb-4"&gt;
                    {t('watch.nextVideos')}
                &lt;/h3&gt;

                &lt;div className="space-y-4"&gt;
                    {recommendations.length === 0 ? (
                        &lt;p className="text-sm text-gray-500 py-4 italic"&gt;{t('watch.noRecommendations')}&lt;/p&gt;
                    ) : (
                        recommendations.map((item) =&gt; {
                            const recUser = item.edges?.user?.[0];
                            const recThumb = getImageUrl(item.thumbnail, 'thumbnail');

                            return (
                                &lt;Link
                                    key={item.id}
                                    to="/watch"
                                    search={{v: item.friendly_token || String(item.id)}}
                                    className="flex gap-3 group"
                                &gt;
                                    &lt;div className="relative w-40 aspect-video rounded-lg overflow-hidden shrink-0"&gt;
                                        &lt;img
                                            src={recThumb}
                                            alt={item.title}
                                            loading="lazy"
                                            onError={(e) =&gt; handleImageError(e, 'thumbnail')}
                                            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                                        /&gt;
                                        &lt;div
                                            className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1 rounded"&gt;
                                            {formatDuration(item.duration)}
                                        &lt;/div&gt;
                                    &lt;/div&gt;
                                    &lt;div className="flex-1 min-w-0"&gt;
                                        &lt;h4 className="text-sm font-bold text-gray-900 dark:text-white line-clamp-2 leading-snug group-hover:text-blue-600 transition-colors"&gt;
                                            {item.title}
                                        &lt;/h4&gt;
                                        &lt;p className="text-xs text-gray-500 mt-1"&gt;{recUser?.nickname || recUser?.username || 'Unknown'}&lt;/p&gt;
                                        &lt;div className="flex items-center gap-2 text-xs text-gray-400"&gt;
                                            &lt;span&gt;{formatViews(item.view_count)} views&lt;/span&gt;
                                            &lt;span&gt;·&lt;/span&gt;
                                            &lt;span&gt;{formatDate(item.created_at)}&lt;/span&gt;
                                        &lt;/div&gt;
                                    &lt;/div&gt;
                                &lt;/Link&gt;
                            );
                        })
                    )}
                &lt;/div&gt;
            &lt;/div&gt;

            {/* Custom Delete Confirmation Dialog */}
            {showDeleteConfirm &amp;&amp; (
                &lt;div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"&gt;
                    &lt;div className="bg-white dark:bg-gray-900 rounded-lg shadow-xl p-6 max-w-md w-full"&gt;
                        &lt;h3 className="text-lg font-bold text-gray-900 dark:text-white mb-2"&gt;{t('common.delete')}&lt;/h3&gt;
                        &lt;p className="text-gray-600 dark:text-gray-400 mb-4"&gt;{t('watch.confirmDelete') || 'Are you sure you want to delete this video? This action cannot be undone.'}&lt;/p&gt;
                        &lt;div className="flex justify-end gap-2"&gt;
                            &lt;Button
                                variant="secondary"
                                size="sm"
                                onClick={() =&gt; setShowDeleteConfirm(false)}
                            &gt;
                                {t('common.cancel')}
                            &lt;/Button&gt;
                            &lt;Button
                                variant="destructive"
                                size="sm"
                                onClick={() =&gt; {
                                    setShowDeleteConfirm(false);
                                    handleDeleteMedia();
                                }}
                            &gt;
                                {t('common.delete')}
                            &lt;/Button&gt;
                        &lt;/div&gt;
                    &lt;/div&gt;
                &lt;/div&gt;
            )}
        &lt;/div&gt;
    );
};

export default WatchPage;
