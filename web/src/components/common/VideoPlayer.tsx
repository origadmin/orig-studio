import React, {useState, useRef, useEffect, useCallback, useMemo, forwardRef, useImperativeHandle} from 'react';
import {
    Play, Pause, Volume2, VolumeX, Maximize, Minimize,
    SkipBack, SkipForward, Settings, Subtitles, PictureInPicture,
    MonitorPlay, AlertCircle
} from 'lucide-react';
import Hls from 'hls.js';
import {Button} from '@/components/ui/button';
import {formatDuration} from '@/lib/format';
import {getFullUrl} from '@/lib/utils';
import {usePlayerSettings} from '@/hooks/usePlayerSettings';
import SpritePreview from './SpritePreview';

// TypeScript 类型定义
export interface QualityOption {
    name: string;
    url?: string;
    height?: number;
    bitrate?: number;
    isRecommended?: boolean;
}

export interface VideoPlayerHandle {
    play: () => void;
    pause: () => void;
    seek: (time: number) => void;
    getCurrentTime: () => number;
    getDuration: () => number;
}

interface VideoPlayerProps {
    src: string;
    hlsSrc?: string;
    poster?: string;
    autoPlay?: boolean;
    onTimeUpdate?: (time: number) => void;
    onPlay?: () => void;
    onPause?: () => void;
    onEnded?: () => void;
    onError?: (error: Error) => void;
    className?: string;
    qualities?: QualityOption[];
    subtitles?: Array<{ label: string; src: string; language: string }>;
    hasAudioTracks?: boolean;
    isPlaying?: boolean;
    currentTime?: number;
    onPlayingChange?: (playing: boolean) => void;
    onTimeChange?: (time: number) => void;
    onAutoPlayNext?: () => void;
    /** When true, the video is still being processed (transcoding) and
     *  should not attempt playback. Shows a processing overlay instead. */
    isProcessing?: boolean;
    /** WebVTT sprite sheet URL for progress bar hover preview */
    spriteVttUrl?: string;
    /** Whether to enable sprite preview on progress bar hover (default: true) */
    enableSpritePreview?: boolean;
}

const VideoPlayer = forwardRef<VideoPlayerHandle, VideoPlayerProps>(({
                                                                             src,
                                                                             hlsSrc,
                                                                             poster,
                                                                             autoPlay = false,
                                                                             onTimeUpdate,
                                                                             onPlay,
                                                                             onPause,
                                                                             onEnded,
                                                                             onError,
                                                                             className = '',
                                                                             qualities: externalQualities,
                                                                             subtitles,
                                                                             hasAudioTracks = false,
                                                                             // 受控模式
                                                                             isPlaying: controlledIsPlaying,
                                                                             currentTime: controlledCurrentTime,
                                                                             onPlayingChange,
                                                                             onTimeChange,
                                                                             onAutoPlayNext,
                                                                             isProcessing = false,
                                                                             spriteVttUrl,
                                                                             enableSpritePreview = true,
                                                                         }, ref) => {

    const videoRef = useRef<HTMLVideoElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const hlsRef = useRef<Hls | null>(null);
    const controlsTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const settingsMenuRef = useRef<HTMLDivElement>(null);
    const lastClickTimeRef = useRef<number>(0);
    const centerOverlayTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // 状态管理
    const [isPlaying, setIsPlaying] = useState(false);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [showControls, setShowControls] = useState(true);
    const [showPlaybackMenu, setShowPlaybackMenu] = useState(false);
    const [showSettingsMenu, setShowSettingsMenu] = useState(false);
    const [buffered, setBuffered] = useState(0);
    const [currentQuality, setCurrentQuality] = useState<string>('auto');
    const [showCenterOverlay, setShowCenterOverlay] = useState(false);
    const [centerOverlayIcon, setCenterOverlayIcon] = useState<'play' | 'pause'>('play');

    // Sprite preview state
    const [hoverTime, setHoverTime] = useState<number | null>(null);
    const [hoverRatio, setHoverRatio] = useState<number>(0);
    const [progressBarRect, setProgressBarRect] = useState<DOMRect | null>(null);
    const [playerRect, setPlayerRect] = useState<DOMRect | null>(null);
    const spriteRafRef = useRef<number>(0);
    const isDraggingProgress = useRef(false);
    const [hasError, setHasError] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    const [hlsQualities, setHlsQualities] = useState<QualityOption[]>([]);
    const [isBuffering, setIsBuffering] = useState(false);
    const [autoResolution, setAutoResolution] = useState<string>('');
    const wasPlayingBeforeQualitySwitch = useRef(false);

    // Use global player settings
    const {
        volume,
        isMuted,
        playbackRate,
        autoPlayNext,
        setVolume,
        setIsMuted,
        setPlaybackRate,
        setAutoPlayNext,
    } = usePlayerSettings();

    // 暴露方法给父组件
    useImperativeHandle(ref, () => ({
        play: () => {
            videoRef.current?.play();
        },
        pause: () => {
            videoRef.current?.pause();
        },
        seek: (time: number) => {
            if (videoRef.current) {
                videoRef.current.currentTime = time;
                setCurrentTime(time);
            }
        },
        getCurrentTime: () => videoRef.current?.currentTime || 0,
        getDuration: () => videoRef.current?.duration || 0,
    }));

    // 合并外部和 HLS 质量选项
    const allQualities = useMemo(() => {
        if (externalQualities && externalQualities.length > 0) {
            return externalQualities;
        }
        if (hlsQualities.length > 0) {
            return hlsQualities;
        }
        return [];
    }, [externalQualities, hlsQualities]);

    // 检测功能是否可用
    const hasSubtitles = useMemo(() => subtitles && subtitles.length > 0, [subtitles]);
    const hasQualityOptions = useMemo(() => allQualities.length > 0, [allQualities]);
    const supportsPiP = useMemo(() => typeof document !== 'undefined' && 'pictureInPictureEnabled' in document, []);
    const supportsFullscreen = useMemo(() => typeof document !== 'undefined' && !!document.fullscreenEnabled, []);

    // Validate HLS source: must be a non-empty path that looks like an HLS manifest
    const isValidHlsSrc = useCallback((src?: string): boolean => {
        if (!src) return false;
        // Accept paths containing "hls/" or ending with ".m3u8" — these are
        // the only valid HLS sources produced by the backend transcoder.
        return src.includes('hls/') || src.endsWith('.m3u8');
    }, []);

    // Initialize HLS player with quality levels
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        // When the video is still being processed (transcoding), do NOT
        // attempt to load any source.  Loading the raw upload (e.g. MKV/AVI)
        // causes DEMUXER_ERROR_COULD_NOT_OPEN because the browser cannot
        // demux non-web-friendly containers.
        if (isProcessing) {
            // Clear any previous source to avoid stale playback attempts
            video.removeAttribute('src');
            video.load();
            if (hlsRef.current) {
                hlsRef.current.destroy();
                hlsRef.current = null;
            }
            setHasError(false);
            setErrorMessage('');
            return;
        }

        const validHls = isValidHlsSrc(hlsSrc) ? hlsSrc : undefined;
        const fullHlsSrc = validHls ? getFullUrl(validHls) : undefined;
        const fullSrc = getFullUrl(src);

        // Reset error state
        setHasError(false);
        setErrorMessage('');

        if (validHls && fullHlsSrc && Hls.isSupported()) {
            if (hlsRef.current) {
                hlsRef.current.destroy();
            }

            const hls = new Hls({
                // Worker: enable for offloaded demuxing
                enableWorker: true,
                // VOD mode: lowLatencyMode is for live streams; disable for on-demand
                // to avoid aggressive small-segment fetching that causes stuttering
                lowLatencyMode: false,

                // === Buffer configuration (optimized for VOD playback) ===
                // Maximum forward buffer length in seconds (default 30s is too short for VOD;
                // 60s provides smoother playback without excessive memory usage)
                maxBufferLength: 60,
                // Absolute maximum buffer length cap in seconds (default 600s is excessive;
                // 120s prevents memory bloat on long videos while ensuring smooth seeking)
                maxMaxBufferLength: 120,
                // Maximum buffer size in bytes (default 60MB; 80MB for higher quality streams)
                maxBufferSize: 80 * 1024 * 1024,
                // Maximum inter-buffer hole tolerance in seconds (default 0.1s is too strict;
                // 0.5s avoids unnecessary rebuffering on small segment gaps after seek)
                maxBufferHole: 0.5,
                // Back buffer length in seconds — clear already-played content from memory
                // (default Infinity causes unbounded memory growth on long videos; 30s is sufficient)
                backBufferLength: 30,

                // === ABR (Adaptive Bitrate) configuration ===
                // Default bandwidth estimate in bps (default 500kbps is too conservative;
                // 1Mbps starts at a reasonable quality instead of always beginning at 240p)
                abrEwmaDefaultEstimate: 1000000,
                // Maximum default estimate cap (prevents overestimation on fast connections)
                abrEwmaDefaultEstimateMax: 5000000,

                // === Startup optimization ===
                // Prefetch the first fragment before attaching to reduce time-to-first-frame
                startFragPrefetch: true,
                // Start from the highest matching quality level instead of lowest
                // (let ABR downgrade if needed rather than always starting low)
                startLevel: -1,

                // === Network resilience ===
                // Fragment loading timeout in ms (default 20s; 30s for slower connections)
                fragLoadingTimeOut: 30000,
                // Fragment loading max retries (default 6; increase for unreliable networks)
                fragLoadingMaxRetry: 8,
                // Fragment loading retry delay in ms (default 1s)
                fragLoadingRetryDelay: 1000,
                // Manifest loading timeout in ms (default 15s; 20s for slower connections)
                manifestLoadingTimeOut: 20000,
                // Manifest loading max retries (default 1; increase for reliability)
                manifestLoadingMaxRetry: 4,
                // Level loading timeout in ms (for loading child playlists in multi-variant streams)
                levelLoadingTimeOut: 20000,

                // === Buffer watchdog ===
                // High buffer watchdog period in seconds (default 2s; increase to reduce
                // unnecessary buffer trimming on momentary speed fluctuations)
                highBufferWatchdogPeriod: 3,
            });

            hls.loadSource(fullHlsSrc);
            hls.attachMedia(video);

            // Extract quality levels from HLS manifest
            hls.on(Hls.Events.MANIFEST_PARSED, (_event, data) => {
                const qualities: QualityOption[] = data.levels.map((level, index) => ({
                    name: `${level.height}p`,
                    height: level.height,
                    bitrate: level.bitrate,
                    isRecommended: index === data.levels.length - 2, // Second highest as recommended
                }));

                // Sort by quality (highest first)
                qualities.sort((a, b) => (b.height || 0) - (a.height || 0));
                setHlsQualities(qualities);

                // Auto play if requested
                if (autoPlay) {
                    video.play().catch(console.error);
                }
            });

            // Handle errors with enhanced recovery strategy
            hls.on(Hls.Events.ERROR, (_event, data) => {
                // Log non-fatal errors for diagnostics but don't disrupt playback
                if (!data.fatal) {
                    console.warn('[HLS] Non-fatal error:', data.type, data.details);
                    return;
                }

                console.error('[HLS] Fatal error:', data.type, data.details);

                switch (data.type) {
                    case Hls.ErrorTypes.NETWORK_ERROR:
                        // Network errors: retry loading with a short delay to allow
                        // transient network issues to resolve
                        console.error('[HLS] Fatal network error, retrying load in 2s...');
                        setTimeout(() => {
                            if (hlsRef.current) {
                                hlsRef.current.startLoad();
                            }
                        }, 2000);
                        break;
                    case Hls.ErrorTypes.MEDIA_ERROR:
                        // Media errors: attempt recovery via hls.js built-in mechanism
                        console.error('[HLS] Fatal media error, attempting recovery...');
                        hls.recoverMediaError();
                        break;
                    default:
                        // Unrecoverable errors: destroy player and show error UI
                        console.error('[HLS] Unrecoverable fatal error, destroying player.');
                        hls.destroy();
                        setHasError(true);
                        setErrorMessage('Failed to load video. Please try again.');
                        onError?.(new Error(data.type));
                        break;
                }
            });

            // Handle quality level changes
            // Only update currentQuality when user manually selected a specific level.
            // In auto mode, only update autoResolution (shown as "Auto (720p)").
            // Also clear buffering state — HLS level switches do NOT trigger the
            // HTML5 video `canplay` event, so the spinner set by handleQualityChange
            // would otherwise remain visible forever.
            hls.on(Hls.Events.LEVEL_SWITCHED, (_event, data) => {
                const level = hls.levels[data.level];
                if (level) {
                    const resolution = `${level.height}p`;
                    setAutoResolution(resolution);
                    // currentLevel === -1 means auto mode; do NOT overwrite currentQuality
                    if (hls.currentLevel !== -1) {
                        setCurrentQuality(resolution);
                    }
                }

                // Resume playback if the video was playing before the quality switch.
                // The buffering spinner is managed entirely by the HTML5 video
                // `waiting`/`playing` events — no manual setIsBuffering needed.
                if (wasPlayingBeforeQualitySwitch.current) {
                    wasPlayingBeforeQualitySwitch.current = false;
                    video.play().catch(console.error);
                }
            });

            hlsRef.current = hls;

            return () => {
                if (hlsRef.current) {
                    hlsRef.current.destroy();
                    hlsRef.current = null;
                }
            };
        } else if (video.canPlayType('application/vnd.apple.mpegurl') && fullHlsSrc) {
            video.src = fullHlsSrc;
            if (autoPlay) {
                video.play().catch(console.error);
            }
        } else if (fullSrc) {
            video.src = fullSrc;
            if (autoPlay) {
                video.play().catch(console.error);
            }
        }

        return () => {};
    }, [src, hlsSrc, autoPlay, onError, isProcessing]);

    // Apply global settings to video element
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        video.volume = volume;
        video.muted = isMuted;
        video.playbackRate = playbackRate;
    }, [volume, isMuted, playbackRate]);

    // Sync controlled state
    useEffect(() => {
        if (controlledIsPlaying !== undefined) {
            setIsPlaying(controlledIsPlaying);
        }
    }, [controlledIsPlaying]);

    useEffect(() => {
        if (controlledCurrentTime !== undefined && videoRef.current) {
            const diff = Math.abs(videoRef.current.currentTime - controlledCurrentTime);
            if (diff > 0.5) { // Only seek if difference is significant
                videoRef.current.currentTime = controlledCurrentTime;
            }
        }
    }, [controlledCurrentTime]);

    // Click outside to close menu
    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (settingsMenuRef.current && !settingsMenuRef.current.contains(e.target as Node)) {
                setShowSettingsMenu(false);
                setShowPlaybackMenu(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Show center overlay icon
    const showCenterIcon = useCallback((icon: 'play' | 'pause') => {
        setCenterOverlayIcon(icon);
        setShowCenterOverlay(true);

        if (centerOverlayTimeoutRef.current) {
            clearTimeout(centerOverlayTimeoutRef.current);
        }

        centerOverlayTimeoutRef.current = setTimeout(() => {
            setShowCenterOverlay(false);
        }, 1500);
    }, []);

    // ========== 基础操作函数 (必须在 handleVideoClick 之前定义) ==========

    // Play/Pause toggle
    // Read video.paused directly instead of relying on the React `isPlaying`
    // state.  `isPlaying` is captured in the useCallback closure and can be
    // stale when the user clicks rapidly — the state update from the previous
    // click may not have re-rendered yet, causing every click to take the same
    // branch (e.g. always play, never pause).  The DOM property `paused` is
    // always in sync with the actual playback state.
    const togglePlay = useCallback(async () => {
        const video = videoRef.current;
        if (!video) return;
        if (video.paused) {
            try {
                await video.play();
            } catch (err) {
                console.error('Play error:', err);
            }
        } else {
            video.pause();
        }
    }, []);

    // Mute toggle
    const toggleMute = useCallback(() => {
        if (!videoRef.current) return;
        const newMuted = !isMuted;
        videoRef.current.muted = newMuted;
        setIsMuted(newMuted);
        if (!newMuted && volume === 0) {
            videoRef.current.volume = 0.5;
            setVolume(0.5);
        }
    }, [isMuted, volume, setIsMuted, setVolume]);

    // Fullscreen toggle
    const toggleFullscreen = useCallback(async () => {
        if (!containerRef.current || !supportsFullscreen) return;

        try {
            if (!document.fullscreenElement) {
                await containerRef.current.requestFullscreen();
                setIsFullscreen(true);
            } else {
                await document.exitFullscreen();
                setIsFullscreen(false);
            }
        } catch (err) {
            console.log('Fullscreen error:', err);
        }
    }, [supportsFullscreen]);

    // Picture-in-Picture
    const togglePiP = useCallback(async () => {
        if (!videoRef.current || !supportsPiP) return;

        try {
            if (document.pictureInPictureElement) {
                await document.exitPictureInPicture();
            } else {
                await videoRef.current.requestPictureInPicture();
            }
        } catch (err) {
            console.log('PiP error:', err);
        }
    }, [supportsPiP]);

    // ========== 复合交互函数 (依赖上面的基础函数) ==========

    // Handle single click for play/pause and double click for fullscreen
    const handleVideoClick = useCallback((e: React.MouseEvent<HTMLVideoElement>) => {
        e.preventDefault();

        const now = Date.now();
        const timeDiff = now - lastClickTimeRef.current;

        if (timeDiff < 300) {
            // Double click - toggle fullscreen
            toggleFullscreen();
            lastClickTimeRef.current = 0;
        } else {
            // Single click - toggle play/pause
            lastClickTimeRef.current = now;
            // Read video.paused directly to determine the correct icon.
            // Using the `isPlaying` state here would show the wrong icon on
            // rapid clicks because the React state update hasn't re-rendered yet.
            const video = videoRef.current;
            const wasPaused = video ? video.paused : true;
            togglePlay();
            // After toggle: if it was paused, we're now playing → show 'play' icon
            // (meaning "play action just happened"); if it was playing, we're now
            // paused → show 'pause' icon (meaning "pause action just happened").
            showCenterIcon(wasPaused ? 'play' : 'pause');
        }
    }, [togglePlay, toggleFullscreen, showCenterIcon]);

    // Handle video events
    const handleTimeUpdate = useCallback(() => {
        if (!videoRef.current) return;
        const time = videoRef.current.currentTime;
        setCurrentTime(time);
        onTimeUpdate?.(time);
        onTimeChange?.(time);

        // Update buffered progress — only update state when the change is
        // significant (>1%) to avoid excessive re-renders on every timeUpdate tick
        if (videoRef.current.buffered.length > 0) {
            const bufferedEnd = videoRef.current.buffered.end(videoRef.current.buffered.length - 1);
            const duration = videoRef.current.duration;
            if (duration > 0) {
                const newBuffered = bufferedEnd / duration;
                setBuffered(prev => {
                    // Skip update if change is less than 1% to reduce re-renders
                    if (Math.abs(newBuffered - prev) < 0.01) return prev;
                    return newBuffered;
                });
            }
        }
    }, [onTimeUpdate, onTimeChange]);

    const handleLoadedMetadata = useCallback(() => {
        if (!videoRef.current) return;
        const d = videoRef.current.duration;
        if (d && isFinite(d)) {
            setDuration(d);
        }
    }, []);

    const handleDurationChange = useCallback(() => {
        if (!videoRef.current) return;
        const d = videoRef.current.duration;
        if (d && isFinite(d)) {
            setDuration(d);
        }
    }, []);

    const handleEnded = useCallback(() => {
        setIsPlaying(false);
        onPlayingChange?.(false);
        onEnded?.();
        if (autoPlayNext) {
            onAutoPlayNext?.();
        }
    }, [onPlayingChange, onEnded, autoPlayNext, onAutoPlayNext]);

    const handleWaiting = useCallback(() => {
        setIsBuffering(true);
    }, []);

    const handlePlaying = useCallback(() => {
        setIsBuffering(false);
    }, []);

    const handleCanPlay = useCallback(() => {
        setIsBuffering(false);
    }, []);

    const handleError = useCallback((e: React.SyntheticEvent<HTMLVideoElement>) => {
        const video = e.currentTarget;
        const error = video.error;
        setHasError(true);
        setErrorMessage(error ? `Video error: ${error.message}` : 'Failed to load video');
        onError?.(new Error(errorMessage));
    }, [onError, errorMessage]);

    // Seek
    const getProgressRatio = useCallback((clientX: number, bar: DOMRect) => {
        return Math.max(0, Math.min(1, (clientX - bar.left) / bar.width));
    }, []);

    const handleProgressMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
        if (!enableSpritePreview) return;
        const bar = e.currentTarget.getBoundingClientRect();
        const clientX = e.clientX;

        if (spriteRafRef.current) cancelAnimationFrame(spriteRafRef.current);

        spriteRafRef.current = requestAnimationFrame(() => {
            const ratio = getProgressRatio(clientX, bar);
            setHoverRatio(ratio);
            setHoverTime(ratio * duration);
            setProgressBarRect(bar);
            setPlayerRect(containerRef.current?.getBoundingClientRect() ?? bar);
        });
    }, [enableSpritePreview, duration, getProgressRatio]);

    const handleProgressMouseLeave = useCallback(() => {
        if (spriteRafRef.current) cancelAnimationFrame(spriteRafRef.current);
        setHoverTime(null);
    }, []);

    const handleProgressPointerDown = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
        if (!videoRef.current || !duration) return;
        isDraggingProgress.current = true;
        (e.target as HTMLElement).setPointerCapture(e.pointerId);
        const bar = e.currentTarget.getBoundingClientRect();
        const ratio = getProgressRatio(e.clientX, bar);
        const seekTime = ratio * duration;
        videoRef.current.currentTime = seekTime;
        setCurrentTime(seekTime);
        onTimeChange?.(seekTime);
    }, [duration, onTimeChange, getProgressRatio]);

    const handleProgressPointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
        if (!isDraggingProgress.current || !videoRef.current || !duration) return;
        const bar = e.currentTarget.getBoundingClientRect();
        const ratio = getProgressRatio(e.clientX, bar);
        const seekTime = ratio * duration;
        videoRef.current.currentTime = seekTime;
        setCurrentTime(seekTime);
        onTimeChange?.(seekTime);
    }, [duration, onTimeChange, getProgressRatio]);

    const handleProgressPointerUp = useCallback(() => {
        isDraggingProgress.current = false;
    }, []);

    // Volume
    const handleVolumeChange = useCallback((value: number[]) => {
        if (!videoRef.current) return;
        const newVolume = value[0];
        videoRef.current.volume = newVolume;
        setVolume(newVolume);
    }, [setVolume]);

    useEffect(() => {
        const handleFullscreenChange = () => {
            setIsFullscreen(!!document.fullscreenElement);
        };

        document.addEventListener('fullscreenchange', handleFullscreenChange);
        return () => document.removeEventListener('fullscreenchange', handleFullscreenChange);
    }, []);

    // Playback speed
    const handlePlaybackRateChange = useCallback((rate: number) => {
        if (!videoRef.current) return;
        videoRef.current.playbackRate = rate;
        setPlaybackRate(rate);
        setShowPlaybackMenu(false);
    }, [setPlaybackRate]);

    // Quality selection with smooth switching
    const handleQualityChange = useCallback((quality: string) => {
        if (!hlsRef.current) {
            setCurrentQuality(quality);
            setShowSettingsMenu(false);
            return;
        }

        // Read video.paused directly instead of the `isPlaying` state to avoid
        // stale closure values.
        const video = videoRef.current;
        wasPlayingBeforeQualitySwitch.current = video ? !video.paused : false;
        // Do NOT manually set setIsBuffering(true) here.  The HTML5 video
        // element's `waiting`/`playing` events naturally manage the buffering
        // spinner.  If the quality switch causes a stall, `waiting` fires and
        // the spinner appears; when playback resumes, `playing` fires and the
        // spinner disappears.  If the switch is seamless (no stall), no
        // spinner should appear at all — this matches YouTube's behavior.

        if (quality === 'auto') {
            hlsRef.current.currentLevel = -1;
            setCurrentQuality('auto');
        } else {
            const height = parseInt(quality.replace('p', ''), 10);
            const levelIndex = hlsRef.current.levels.findIndex(level => level.height === height);

            if (levelIndex !== -1) {
                hlsRef.current.currentLevel = levelIndex;
                setCurrentQuality(quality);
            }
        }

        setShowSettingsMenu(false);
    }, []);

    // Skip forward/backward
    const skip = useCallback((seconds: number) => {
        if (!videoRef.current) return;
        const newTime = Math.max(0, Math.min(duration, videoRef.current.currentTime + seconds));
        videoRef.current.currentTime = newTime;
        setCurrentTime(newTime);
        onTimeChange?.(newTime);
    }, [duration, onTimeChange]);

    // Controls visibility with improved logic
    const showControlsTemporarily = useCallback(() => {
        setShowControls(true);
        if (controlsTimeoutRef.current) {
            clearTimeout(controlsTimeoutRef.current);
        }
        if (isPlaying) {
            controlsTimeoutRef.current = setTimeout(() => {
                setShowControls(false);
            }, 3000);
        }
    }, [isPlaying]);

    const hideControlsImmediately = useCallback(() => {
        if (controlsTimeoutRef.current) {
            clearTimeout(controlsTimeoutRef.current);
        }
        if (isPlaying) {
            setShowControls(false);
        }
    }, [isPlaying]);

    useEffect(() => {
        const handleMouseMove = () => {
            showControlsTemporarily();
        };

        const handleMouseLeave = () => {
            hideControlsImmediately();
        };

        const container = containerRef.current;
        if (container) {
            container.addEventListener('mousemove', handleMouseMove);
            container.addEventListener('mouseleave', handleMouseLeave);
        }

        // Touch support for mobile
        const handleTouchStart = () => {
            showControlsTemporarily();
        };

        if (container) {
            container.addEventListener('touchstart', handleTouchStart);
        }

        return () => {
            if (container) {
                container.removeEventListener('mousemove', handleMouseMove);
                container.removeEventListener('mouseleave', handleMouseLeave);
                container.removeEventListener('touchstart', handleTouchStart);
            }
            if (controlsTimeoutRef.current) {
                clearTimeout(controlsTimeoutRef.current);
            }
        };
    }, [showControlsTemporarily, hideControlsImmediately]);

    // Cleanup center overlay timeout on unmount
    useEffect(() => {
        return () => {
            if (centerOverlayTimeoutRef.current) {
                clearTimeout(centerOverlayTimeoutRef.current);
            }
        };
    }, []);

    // Keyboard shortcuts
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            // Only handle shortcuts when player is focused or in fullscreen
            const container = containerRef.current;
            if (!container || !document.activeElement?.closest('.video-player-container')) {
                if (!document.fullscreenElement) return;
            }

            switch (e.key.toLowerCase()) {
                case ' ':
                case 'k':
                    e.preventDefault();
                    {
                        const video = videoRef.current;
                        const wasPaused = video ? video.paused : true;
                        togglePlay();
                        showCenterIcon(wasPaused ? 'play' : 'pause');
                    }
                    break;
                case 'm':
                    e.preventDefault();
                    toggleMute();
                    break;
                case 'f':
                    e.preventDefault();
                    toggleFullscreen();
                    break;
                case 'arrowleft':
                    e.preventDefault();
                    skip(-10);
                    break;
                case 'arrowright':
                    e.preventDefault();
                    skip(10);
                    break;
                case 'arrowup':
                    e.preventDefault();
                    if (videoRef.current) {
                        const newVol = Math.min(1, videoRef.current.volume + 0.1);
                        videoRef.current.volume = newVol;
                        setVolume(newVol);
                    }
                    break;
                case 'arrowdown':
                    e.preventDefault();
                    if (videoRef.current) {
                        const newVol = Math.max(0, videoRef.current.volume - 0.1);
                        videoRef.current.volume = newVol;
                        setVolume(newVol);
                    }
                    break;
                case 'p':
                    e.preventDefault();
                    togglePiP();
                    break;
            }
        };

        document.addEventListener('keydown', handleKeyDown);
        return () => document.removeEventListener('keydown', handleKeyDown);
    }, [togglePlay, toggleMute, toggleFullscreen, togglePiP, skip, showCenterIcon, setVolume]);

    const playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 2];

    return (
        <div
            ref={containerRef}
            className={`relative bg-black overflow-hidden aspect-video group video-player-container ${className}`}
            tabIndex={0}
            role="application"
            aria-label="Video Player"
        >
            {/* Video Element */}
            <video
                ref={videoRef}
                onClick={handleVideoClick}
                onTimeUpdate={handleTimeUpdate}
                onLoadedMetadata={handleLoadedMetadata}
                onDurationChange={handleDurationChange}
                onEnded={handleEnded}
                onWaiting={handleWaiting}
                onPlaying={handlePlaying}
                onCanPlay={handleCanPlay}
                onPlay={() => {
                    setIsPlaying(true);
                    onPlayingChange?.(true);
                    onPlay?.();
                }}
                onPause={() => {
                    setIsPlaying(false);
                    onPlayingChange?.(false);
                    onPause?.();
                    setShowControls(true);
                }}
                onError={handleError}
                poster={poster ? getFullUrl(poster) : undefined}
                className="w-full h-full cursor-pointer"
                playsInline
                preload="metadata"
            />

            {/* Center overlay icon (shows on click) */}
            {showCenterOverlay && !isBuffering && (
                <div
                    className="absolute inset-0 flex items-center justify-center pointer-events-none z-30"
                    aria-hidden="true"
                >
                    <div className="w-[clamp(3rem,15vw,6rem)] h-[clamp(3rem,15vw,6rem)] bg-black/40 backdrop-blur-sm rounded-full flex items-center justify-center animate-fade-in">
                        {centerOverlayIcon === 'play' ? (
                            <Play className="text-white fill-white ml-2" style={{width: 'clamp(1.5rem,8vw,3.5rem)', height: 'clamp(1.5rem,8vw,3.5rem)'}}/>
                        ) : (
                            <Pause className="text-white fill-white" style={{width: 'clamp(1.5rem,8vw,3.5rem)', height: 'clamp(1.5rem,8vw,3.5rem)'}}/>
                        )}
                    </div>
                </div>
            )}

            {/* Buffering indicator */}
            {isBuffering && (
                <div
                    className="absolute inset-0 flex items-center justify-center pointer-events-none z-30"
                    aria-hidden="true"
                >
                    <div className="w-16 h-16 border-4 border-white/30 border-t-white rounded-full animate-spin"/>
                </div>
            )}

            {/* Center play button overlay when paused - only shows when controls are visible */}
            {!isPlaying && !hasError && showControls && !showCenterOverlay && (
                <div
                    className="absolute inset-0 flex items-center justify-center pointer-events-none z-20"
                >
                    <div
                        className="w-[clamp(2.5rem,12vw,5rem)] h-[clamp(2.5rem,12vw,5rem)] bg-white/20 backdrop-blur-sm rounded-full flex items-center justify-center cursor-pointer pointer-events-auto transition-transform hover:scale-110"
                        onClick={(e) => {
                            e.stopPropagation();
                            togglePlay();
                            showCenterIcon('play');
                        }}
                        role="button"
                        aria-label="Play video"
                        tabIndex={0}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter' || e.key === ' ') {
                                e.preventDefault();
                                togglePlay();
                                showCenterIcon('play');
                            }
                        }}
                    >
                        <Play className="text-white fill-white ml-2" style={{width: 'clamp(1.25rem,7vw,3rem)', height: 'clamp(1.25rem,7vw,3rem)'}}/>
                    </div>
                </div>
            )}

            {/* Error overlay */}
            {hasError && (
                <div
                    className="absolute inset-0 flex flex-col items-center justify-center bg-black/80 z-20 gap-4 p-8"
                    role="alert"
                    aria-live="assertive"
                >
                    <AlertCircle size={64} className="text-destructive"/>
                    <p className="text-white text-lg font-medium text-center max-w-md">{errorMessage}</p>
                    <Button
                        variant="secondary"
                        onClick={() => window.location.reload()}
                        className="mt-2"
                    >
                        Retry
                    </Button>
                </div>
            )}

            {/* Processing overlay — shown when video is still transcoding */}
            {isProcessing && (
                <div
                    className="absolute inset-0 flex flex-col items-center justify-center bg-black/70 z-20 gap-4 p-8"
                    role="status"
                    aria-live="polite"
                    aria-label="Video is being processed"
                >
                    <div className="w-16 h-16 border-4 border-white/30 border-t-white rounded-full animate-spin"/>
                    <p className="text-white text-lg font-medium text-center">Video is being processed</p>
                    <p className="text-white/60 text-sm text-center max-w-sm">
                        This video is currently being transcoded and will be available for playback once processing is complete.
                    </p>
                </div>
            )}

            {/* Controls overlay */}
            <div
                className={`absolute inset-0 flex flex-col justify-end transition-opacity duration-300 z-10 pointer-events-none ${
                    showControls || !isPlaying ? 'opacity-100' : 'opacity-0'
                }`}
                role="toolbar"
                aria-label="Video Controls"
            >
                {/* Gradient background */}
                <div
                    className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent pointer-events-none"/>

                {/* Top controls area (can add more here) */}
                <div className="relative flex-1"/>

                {/* Bottom controls */}
                <div className="relative px-4 pb-4 pt-12 pointer-events-auto">
                    {/* Progress bar */}
                    <div
                        className="relative h-1.5 group/hover:h-2.5 transition-all mb-4 cursor-pointer"
                        role="slider"
                        aria-label="Video progress"
                        aria-valuemin={0}
                        aria-valuemax={Math.floor(duration)}
                        aria-valuenow={Math.floor(currentTime)}
                        tabIndex={0}
                        onPointerDown={handleProgressPointerDown}
                        onPointerMove={handleProgressPointerMove}
                        onPointerUp={handleProgressPointerUp}
                        onMouseMove={enableSpritePreview ? handleProgressMouseMove : undefined}
                        onMouseLeave={enableSpritePreview ? handleProgressMouseLeave : undefined}
                    >
                        {/* Buffered progress */}
                        <div className="absolute inset-0 bg-white/30 rounded-full overflow-hidden">
                            <div
                                className="h-full bg-white/50 rounded-full transition-all"
                                style={{width: `${buffered * 100}%`}}
                                aria-hidden="true"
                            />
                        </div>
                        {/* Played progress */}
                        <div className="absolute inset-0 flex items-center">
                            <div
                                className="h-full bg-red-600 rounded-full transition-all"
                                style={{width: `${(currentTime / (duration || 1)) * 100}%`}}
                            >
                                <div
                                    className="absolute right-0 top-1/2 -translate-y-1/2 w-3.5 h-3.5 bg-red-600 rounded-full opacity-0 group-hover/hover:opacity-100 transition-opacity shadow-lg border-2 border-white"
                                />
                            </div>
                        </div>
                    </div>

                    {/* Control buttons */}
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            {/* Play/Pause */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    togglePlay();
                                }}
                                aria-label={isPlaying ? 'Pause' : 'Play'}
                            >
                                {isPlaying ?
                                    <Pause size={24} fill="currentColor" aria-hidden="true"/> :
                                    <Play size={24} fill="currentColor" className="ml-1" aria-hidden="true"/>
                                }
                            </Button>

                            {/* Skip backward */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    skip(-10);
                                }}
                                aria-label="Rewind 10 seconds"
                            >
                                <SkipBack size={20} aria-hidden="true"/>
                            </Button>

                            {/* Skip forward */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    skip(10);
                                }}
                                aria-label="Forward 10 seconds"
                            >
                                <SkipForward size={20} aria-hidden="true"/>
                            </Button>

                            {/* Volume */}
                            <div className="flex items-center gap-2 group/volume">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        toggleMute();
                                    }}
                                    aria-label={isMuted || volume === 0 ? 'Unmute' : 'Mute'}
                                >
                                    {isMuted || volume === 0 ?
                                        <VolumeX size={20} aria-hidden="true"/> :
                                        <Volume2 size={20} aria-hidden="true"/>
                                    }
                                </Button>
                                <div
                                    className="w-0 group-hover/volume:w-24 overflow-hidden transition-all duration-300"
                                    role="slider"
                                    aria-label="Volume"
                                    aria-valuemin={0}
                                    aria-valuemax={100}
                                    aria-valuenow={Math.round((isMuted ? 0 : volume) * 100)}
                                >
                                    <input
                                        type="range"
                                        value={isMuted ? 0 : volume}
                                        min={0}
                                        max={1}
                                        step={0.01}
                                        onChange={(e) => handleVolumeChange([parseFloat(e.target.value)])}
                                        className="w-24 h-1.5 bg-white/30 rounded-full appearance-none cursor-pointer accent-white"
                                        aria-hidden="true"
                                    />
                                </div>
                            </div>

                            {/* Time display */}
                            <span
                                className="text-white text-sm font-medium min-w-[100px] tabular-nums"
                                aria-label={`Time: ${formatDuration(Math.floor(currentTime))} of ${formatDuration(Math.floor(duration))}`}
                            >
                                {formatDuration(Math.floor(currentTime))} / {formatDuration(Math.floor(duration))}
                            </span>
                        </div>

                        <div className="flex items-center gap-2" ref={settingsMenuRef}>
                            {/* Auto-play next toggle */}
                            <Button
                                variant="ghost"
                                size="sm"
                                className={`h-8 px-2 text-xs font-medium rounded-full transition-colors ${
                                    autoPlayNext
                                        ? 'text-blue-400 hover:text-blue-300 hover:bg-info/10'
                                        : 'text-white/50 hover:text-white/80 hover:bg-white/10'
                                }`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    setAutoPlayNext(!autoPlayNext);
                                }}
                                title={autoPlayNext ? 'Auto-play next: ON' : 'Auto-play next: OFF'}
                                aria-label={autoPlayNext ? 'Disable auto-play next' : 'Enable auto-play next'}
                            >
                                <SkipForward size={14} className="mr-1"/>
                                {autoPlayNext ? 'ON' : 'OFF'}
                            </Button>

                            {/* Subtitles - conditionally rendered */}
                            {hasSubtitles ? (
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white"
                                    title="Subtitles"
                                    aria-label="Subtitles"
                                >
                                    <Subtitles size={20} aria-hidden="true"/>
                                </Button>
                            ) : null}

                            {/* Audio tracks - conditionally rendered */}
                            {hasAudioTracks ? (
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white"
                                    title="Audio tracks"
                                    aria-label="Audio tracks"
                                >
                                    <MonitorPlay size={20} aria-hidden="true"/>
                                </Button>
                            ) : null}

                            {/* Playback speed */}
                            <div className="relative">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-10 px-3 text-white hover:bg-white/10 rounded-full text-sm font-medium focus-visible:ring-2 focus-visible:ring-white"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowPlaybackMenu(!showPlaybackMenu);
                                        setShowSettingsMenu(false);
                                    }}
                                    aria-label={`Playback speed: ${playbackRate}x`}
                                    aria-expanded={showPlaybackMenu}
                                    aria-haspopup="menu"
                                >
                                    {playbackRate}x
                                </Button>
                                {showPlaybackMenu && (
                                    <div
                                        className="absolute bottom-full right-0 mb-2 bg-black/90 backdrop-blur-md border border-white/10 rounded-lg overflow-hidden min-w-[120px] shadow-xl z-50"
                                        role="menu"
                                        aria-label="Playback speed options"
                                    >
                                        {playbackRates.map((rate) => (
                                            <button
                                                key={rate}
                                                role="menuitemradio"
                                                aria-checked={playbackRate === rate}
                                                className={`w-full text-left px-4 py-2 text-white hover:bg-white/10 transition-colors text-sm ${
                                                    playbackRate === rate ? 'bg-white/10 font-semibold' : ''
                                                }`}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handlePlaybackRateChange(rate);
                                                }}
                                            >
                                                {rate}x{rate === 1 ? ' (Normal)' : ''}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* Settings (Quality) */}
                            <div className="relative">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className={`h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white ${
                                        !hasQualityOptions ? 'opacity-50 cursor-not-allowed' : ''
                                    }`}
                                    onClick={(e) => {
                                        if (!hasQualityOptions) return;
                                        e.stopPropagation();
                                        setShowSettingsMenu(!showSettingsMenu);
                                        setShowPlaybackMenu(false);
                                    }}
                                    disabled={!hasQualityOptions}
                                    title={hasQualityOptions ? 'Quality settings' : 'Quality options not available'}
                                    aria-label="Quality settings"
                                    aria-expanded={showSettingsMenu}
                                    aria-haspopup="menu"
                                >
                                    <Settings size={20} aria-hidden="true"/>
                                </Button>
                                {showSettingsMenu && hasQualityOptions && (
                                    <div
                                        className="absolute bottom-full right-0 mb-2 bg-black/90 backdrop-blur-md border border-white/10 rounded-lg overflow-hidden min-w-[160px] shadow-xl z-50"
                                        role="menu"
                                        aria-label="Quality options"
                                    >
                                        {/* Quality header */}
                                        <div
                                            className="px-3 py-2 text-xs font-bold text-muted-foreground uppercase tracking-wider border-b border-white/10">
                                            Quality
                                        </div>

                                        {/* Auto option */}
                                        <button
                                            role="menuitemradio"
                                            aria-checked={currentQuality === 'auto'}
                                            className={`w-full text-left px-4 py-2 text-white hover:bg-white/10 transition-colors text-sm flex items-center justify-between ${
                                                currentQuality === 'auto' ? 'bg-white/10 font-semibold' : ''
                                            }`}
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleQualityChange('auto');
                                            }}
                                        >
                                            <span>Auto{autoResolution ? ` (${autoResolution})` : ''}</span>
                                            {currentQuality === 'auto' && (
                                                <span className="text-blue-400 text-xs">✓</span>
                                            )}
                                        </button>

                                        {/* Quality options */}
                                        {allQualities.map((q) => (
                                            <button
                                                key={q.name}
                                                role="menuitemradio"
                                                aria-checked={currentQuality === q.name}
                                                className={`w-full text-left px-4 py-2 text-white hover:bg-white/10 transition-colors text-sm flex items-center justify-between ${
                                                    currentQuality === q.name ? 'bg-white/10 font-semibold' : ''
                                                }`}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleQualityChange(q.name);
                                                }}
                                            >
                                                <span>{q.name}</span>
                                                {currentQuality === q.name ? (
                                                    <span className="text-blue-400 text-xs">✓</span>
                                                ) : q.isRecommended ? (
                                                    <span className="text-[10px] bg-info/20 text-blue-400 px-1.5 py-0.5 rounded">推荐</span>
                                                ) : null}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* Picture-in-Picture - conditionally enabled */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className={`h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white ${
                                    !supportsPiP ? 'opacity-50 cursor-not-allowed' : ''
                                }`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    togglePiP();
                                }}
                                disabled={!supportsPiP}
                                title={supportsPiP ? 'Picture-in-Picture' : 'Picture-in-Picture not supported'}
                                aria-label="Picture-in-Picture mode"
                            >
                                <PictureInPicture size={20} aria-hidden="true"/>
                            </Button>

                            {/* Fullscreen - conditionally enabled */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className={`h-10 w-10 text-white hover:bg-white/10 rounded-full focus-visible:ring-2 focus-visible:ring-white ${
                                    !supportsFullscreen ? 'opacity-50 cursor-not-allowed' : ''
                                }`}
                                onClick={(e) => {
                                    e.stopPropagation();
                                    toggleFullscreen();
                                }}
                                disabled={!supportsFullscreen}
                                title={supportsFullscreen ? (isFullscreen ? 'Exit fullscreen' : 'Fullscreen') : 'Fullscreen not supported'}
                                aria-label={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
                            >
                                {isFullscreen ?
                                    <Minimize size={20} aria-hidden="true"/> :
                                    <Maximize size={20} aria-hidden="true"/>
                                }
                            </Button>
                        </div>
                    </div>
                </div>
            </div>

            {/* Sprite preview on progress bar hover — rendered at player container level
                so that absolute positioning is relative to the player container, not the
                bottom controls area. This avoids padding/offset miscalculations. */}
            {hoverTime !== null && enableSpritePreview && progressBarRect && playerRect && (
                <SpritePreview
                    hoverTime={hoverTime}
                    hoverRatio={hoverRatio}
                    progressBarRect={progressBarRect}
                    playerRect={playerRect}
                    vttUrl={spriteVttUrl ?? null}
                    duration={duration}
                />
            )}
        </div>
    );
});

VideoPlayer.displayName = 'VideoPlayer';

export default VideoPlayer;
