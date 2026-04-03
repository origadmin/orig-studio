/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 视频播放页 - 对接真实数据
 */

import React, {useState, useEffect, useRef, useCallback} from 'react';
import {useSearch, Link} from '@tanstack/react-router';
import {
    ThumbsUp, ThumbsDown, Share2, MessageCircle,
    MoreHorizontal, UserPlus, Eye, Loader2, Settings, RefreshCw, AlertTriangle
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent} from '@/components/ui/card';
import {Skeleton} from '@/components/ui/skeleton';
import {formatViews, formatDate, formatDuration} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {type Media, type VariantInfo} from '@/lib/api/media';
import {mediaApi} from '@/lib/api/media';
import {useMediaDetail, useMediaList} from '@/hooks/queries';
import Hls from 'hls.js';

const WatchPage = () => {
    const {t} = useTranslation();
    const {v: id} = useSearch({strict: false});
    const {data: media, isLoading: isMediaLoading, error: mediaError} = useMediaDetail(id as string);
    const videoRef = useRef<HTMLVideoElement>(null);
    const hlsRef = useRef<Hls | null>(null);

    // Quality switcher state
    const [variants, setVariants] = useState<VariantInfo[]>([]);
    const [currentLevel, setCurrentLevel] = useState(-1); // -1 = auto
    const [showQualityMenu, setShowQualityMenu] = useState(false);
    const [retrying, setRetrying] = useState(false);

    const {data: recData} = useMediaList({
        page_size: 10,
        category_id: media?.edges?.category?.id || undefined,
        status: 'active'
    });

    const recommendations = recData?.list?.filter(m => m.id !== Number(id)) || [];
    const loading = isMediaLoading;
    const error = mediaError ? t('watch.failedToLoad') : null;

    // Fetch variants for quality switcher (only for successfully transcoded videos)
    useEffect(() => {
        if (!media || media.encoding_status !== 'success' && media.encoding_status !== 'partial') {
            return;
        }
        const mediaId = Number(id);
        if (!mediaId || mediaId <= 0) return;

        mediaApi.getVariants(mediaId)
            .then(res => {
                if (res.data?.variants) {
                    // Filter to only successful video variants for quality switching
                    const successful = res.data.variants.filter(
                        v => v.status === 'success' && (
                            v.output_path?.includes('.m3u8') ||
                            v.profile_name?.includes('playlist')
                        )
                    );
                    // Sort by resolution descending (highest first)
                    successful.sort((a, b) => {
                        const parseRes = (r: string) => parseInt(r.replace(/\D/g, '')) || 0;
                        return parseRes(b.resolution) - parseRes(a.resolution);
                    });
                    setVariants(successful);
                }
            })
            .catch(() => { /* variants are optional */
            });
    }, [media?.id, media?.encoding_status, id]);

    // HLS player setup with quality level control
    useEffect(() => {
        if (!media || !videoRef.current) return;

        const video = videoRef.current;
        const API_BASE_URL = (import.meta as any).env.VITE_API_BASE_URL || "http://localhost:9090";
        const getFullUrl = (path?: string) => {
            if (!path) return '';
            if (path.startsWith('http')) return path;
            const base = API_BASE_URL.replace(/\/$/, '');
            const sep = path.startsWith('/') ? '' : '/';
            return `${base}${sep}${path}`;
        };

        const hlsUrl = media.hls_file ? getFullUrl(media.hls_file) : null;
        const originalUrl = getFullUrl(media.url);

        // Cleanup previous HLS instance
        if (hlsRef.current) {
            hlsRef.current.destroy();
            hlsRef.current = null;
        }

        if (hlsUrl && Hls.isSupported()) {
            const hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true,
            });
            hls.loadSource(hlsUrl);
            hls.attachMedia(video);
            hls.on(Hls.Events.MANIFEST_PARSED, (_event, data) => {
                video.play().catch(() => {
                    // Autoplay might be blocked
                    console.log("Autoplay blocked");
                });
                // Store available levels for quality switching
                if (data.levels.length > 1) {
                    setCurrentLevel(-1); // auto by default
                }
            });
            // Track level changes for UI sync
            hls.on(Hls.Events.LEVEL_SWITCHED, (_event, data) => {
                setCurrentLevel(data.level);
            });
            hlsRef.current = hls;

            return () => {
                hls.destroy();
                hlsRef.current = null;
            };
        } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
            // Native HLS support (Safari)
            video.src = hlsUrl || originalUrl;
        } else {
            // Fallback to original MP4
            video.src = originalUrl;
        }
    }, [media]);

    // Quality level switch handler
    const handleQualityChange = useCallback((levelIndex: number) => {
        if (hlsRef.current) {
            // Map variant index to HLS level (levels are ordered by HLS, usually descending bitrate)
            // levelIndex -1 = auto, 0+ = specific level
            hlsRef.current.currentLevel = levelIndex;
            setCurrentLevel(levelIndex);
        }
        setShowQualityMenu(false);
    }, []);

    // Retry transcoding handler
    const handleRetry = useCallback(async () => {
        if (!media || retrying) return;
        setRetrying(true);
        try {
            await mediaApi.retryTranscode(media.id);
            // Reload the page data after a short delay to show processing state
            setTimeout(() => window.location.reload(), 1000);
        } catch {
            // Error silently — user can see the button still there
        } finally {
            setRetrying(false);
        }
    }, [media, retrying]);

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
        return (
            <div className="py-20 text-center space-y-4">
                <div className="text-red-500 text-lg">{error || "Video not found"}</div>
                <Link to="/">
                    <Button variant="outline">{t('common.backToHome')}</Button>
                </Link>
            </div>
        );
    }

    const API_BASE_URL = (import.meta as any).env.VITE_API_BASE_URL || "http://localhost:9090";

    const getFullUrl = (path?: string) => {
        if (!path) return '';
        if (path.startsWith('http')) return path;
        const base = API_BASE_URL.replace(/\/$/, '');
        const sep = path.startsWith('/') ? '' : '/';
        return `${base}${sep}${path}`;
    };

    const user = media.edges?.user?.[0];
    const isProcessing = media.encoding_status !== 'success';

    return (
        <div className="flex flex-col lg:flex-row gap-6">
            {/* Main Content: Player & Details */}
            <div className="flex-1 min-w-0">
                {/* Player Container */}
                <div className="bg-black rounded-2xl overflow-hidden aspect-video shadow-2xl relative group">
                    <video
                        ref={videoRef}
                        controls
                        className="w-full h-full"
                        poster={media.poster ? getFullUrl(media.poster) : (media.thumbnail ? getFullUrl(media.thumbnail) : undefined)}
                    >
                        Your browser does not support the video tag.
                    </video>

                    {/* Quality Switcher — only when HLS has multiple levels or variants available */}
                    {(variants.length > 0 || (hlsRef.current && (hlsRef.current.levels?.length ?? 0) > 1)) && (
                        <div className="absolute bottom-16 right-3 z-10">
                            <div className="relative">
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    className="bg-black/70 hover:bg-black/90 text-white border-white/20 backdrop-blur-md text-xs gap-1.5 h-7 px-2"
                                    onClick={() => setShowQualityMenu(!showQualityMenu)}
                                >
                                    <Settings size={12}/>
                                    {currentLevel === -1 ? 'AUTO' :
                                        variants[currentLevel]?.resolution || `${currentLevel}p`}
                                </Button>
                                {showQualityMenu && (
                                    <div
                                        className="absolute bottom-full right-0 mb-1 bg-black/90 backdrop-blur-md border border-white/20 rounded-lg overflow-hidden min-w-[120px] shadow-xl">
                                        <button
                                            className={`w-full text-left text-xs px-3 py-1.5 text-white hover:bg-white/10 transition-colors flex justify-between items-center ${currentLevel === -1 ? 'bg-blue-600/40 font-semibold' : ''}`}
                                            onClick={() => handleQualityChange(-1)}
                                        >
                                            <span>AUTO</span>
                                            {currentLevel === -1 && <span>✓</span>}
                                        </button>
                                        {hlsRef.current?.levels?.map((level, idx) => {
                                            const width = level.width;
                                            const height = level.height;
                                            const label = height >= 720 ? `${height}p` : `${width}x${height}`;
                                            return (
                                                <button
                                                    key={idx}
                                                    className={`w-full text-left text-xs px-3 py-1.5 text-white hover:bg-white/10 transition-colors flex justify-between items-center ${currentLevel === idx ? 'bg-blue-600/40 font-semibold' : ''}`}
                                                    onClick={() => handleQualityChange(idx)}
                                                >
                                                    <span>{label}</span>
                                                    {currentLevel === idx && <span>✓</span>}
                                                </button>
                                            );
                                        }) || variants.map((v, idx) => (
                                            <button
                                                key={v.task_id || idx}
                                                className={`w-full text-left text-xs px-3 py-1.5 text-white hover:bg-white/10 transition-colors flex justify-between items-center ${currentLevel === idx ? 'bg-blue-600/40 font-semibold' : ''}`}
                                                onClick={() => handleQualityChange(idx)}
                                            >
                                                <span>{v.profile_name || v.resolution}</span>
                                                {v.status === 'success' && <span className="text-green-400">✓</span>}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Encoding Status Indicator */}
                    {isProcessing && (
                        <div className="absolute top-3 left-3 z-10 flex items-center gap-2">
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
                        <div className="flex items-center gap-4">
                            <Link to="/u/$id" params={{id: String(media.user_id)}}>
                                <Avatar className="h-12 w-12 ring-2 ring-gray-100 dark:ring-gray-800">
                                    <AvatarImage src={user?.avatar}/>
                                    <AvatarFallback>{user?.username?.[0] || 'U'}</AvatarFallback>
                                </Avatar>
                            </Link>
                            <div>
                                <Link to="/u/$id" params={{id: String(media.user_id)}}
                                      className="font-bold text-gray-900 dark:text-white hover:text-blue-600 transition-colors">
                                    {user?.nickname || user?.username || 'Unknown Gopher'}
                                </Link>
                                <p className="text-xs text-gray-500 dark:text-gray-400">1.2M {t('watch.subscribers')}</p>
                            </div>
                            <Button
                                className="ml-4 rounded-full bg-gray-900 dark:bg-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-200">
                                {t('watch.subscribe')}
                            </Button>
                        </div>

                        <div className="flex items-center bg-gray-100 dark:bg-gray-800 rounded-full p-1">
                            <Button variant="ghost"
                                    className="rounded-l-full gap-2 px-4 hover:bg-gray-200 dark:hover:bg-gray-700">
                                <ThumbsUp className="w-5 h-5"/>
                                <span className="text-sm font-medium">{formatViews(media.like_count)}</span>
                            </Button>
                            <div className="w-[1px] h-6 bg-gray-300 dark:bg-gray-600"/>
                            <Button variant="ghost"
                                    className="rounded-r-full px-4 hover:bg-gray-200 dark:hover:bg-gray-700">
                                <ThumbsDown className="w-5 h-5"/>
                            </Button>
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
                            const recThumb = item.thumbnail
                                ? getFullUrl(item.thumbnail)
                                : 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=400&h=225';

                            return (
                                <Link
                                    key={item.id}
                                    to="/watch"
                                    search={{v: String(item.id)}}
                                    className="flex gap-3 group"
                                >
                                    <div className="relative w-40 aspect-video rounded-lg overflow-hidden shrink-0">
                                        <img
                                            src={recThumb}
                                            alt={item.title}
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
        </div>
    );
};

export default WatchPage;
