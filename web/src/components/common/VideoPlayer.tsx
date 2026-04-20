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
    // 受控模式 props
    isPlaying?: boolean;
    currentTime?: number;
    onPlayingChange?: (playing: boolean) => void;
    onTimeChange?: (time: number) => void;
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
    const [hasError, setHasError] = useState(false);
    const [errorMessage, setErrorMessage] = useState('');
    const [hlsQualities, setHlsQualities] = useState<QualityOption[]>([]);

    // Use global player settings
    const {
        volume,
        isMuted,
        playbackRate,
        setVolume,
        setIsMuted,
        setPlaybackRate,
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

    // Initialize HLS player with quality levels
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const fullHlsSrc = hlsSrc ? getFullUrl(hlsSrc) : undefined;
        const fullSrc = getFullUrl(src);

        // Reset error state
        setHasError(false);
        setErrorMessage('');

        if (hlsSrc && fullHlsSrc && Hls.isSupported()) {
            if (hlsRef.current) {
                hlsRef.current.destroy();
            }

            const hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true,
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

            // Handle errors
            hls.on(Hls.Events.ERROR, (_event, data) => {
                if (data.fatal) {
                    switch (data.type) {
                        case Hls.ErrorTypes.NETWORK_ERROR:
                            console.error('Fatal network error encountered, trying to recover...');
                            hls.startLoad();
                            break;
                        case Hls.ErrorTypes.MEDIA_ERROR:
                            console.error('Fatal media error encountered, trying to recover...');
                            hls.recoverMediaError();
                            break;
                        default:
                            console.error('Fatal error:', data);
                            hls.destroy();
                            setHasError(true);
                            setErrorMessage('Failed to load video. Please try again.');
                            onError?.(new Error(data.type));
                            break;
                    }
                }
            });

            // Handle quality level changes
            hls.on(Hls.Events.LEVEL_SWITCHED, (_event, data) => {
                const level = hls.levels[data.level];
                if (level) {
                    setCurrentQuality(`${level.height}p`);
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
    }, [src, hlsSrc, autoPlay, onError]);

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
    const togglePlay = useCallback(async () => {
        if (!videoRef.current) return;
        if (isPlaying) {
            videoRef.current.pause();
            setIsPlaying(false);
            onPlayingChange?.(false);
        } else {
            try {
                await videoRef.current.play();
                setIsPlaying(true);
                onPlayingChange?.(true);
            } catch (err) {
                console.error('Play error:', err);
            }
        }
    }, [isPlaying, onPlayingChange]);

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
            togglePlay();
            showCenterIcon(isPlaying ? 'pause' : 'play');
        }
    }, [isPlaying, togglePlay, toggleFullscreen, showCenterIcon]);

    // Handle video events
    const handleTimeUpdate = useCallback(() => {
        if (!videoRef.current) return;
        const time = videoRef.current.currentTime;
        setCurrentTime(time);
        onTimeUpdate?.(time);
        onTimeChange?.(time);

        // Update buffered
        if (videoRef.current.buffered.length > 0) {
            const bufferedEnd = videoRef.current.buffered.end(videoRef.current.buffered.length - 1);
            setBuffered(bufferedEnd / videoRef.current.duration);
        }
    }, [onTimeUpdate, onTimeChange]);

    const handleLoadedMetadata = useCallback(() => {
        if (!videoRef.current) return;
        setDuration(videoRef.current.duration);
    }, []);

    const handleEnded = useCallback(() => {
        setIsPlaying(false);
        onPlayingChange?.(false);
        onEnded?.();
    }, [onPlayingChange, onEnded]);

    const handleError = useCallback((e: React.SyntheticEvent<HTMLVideoElement>) => {
        const video = e.currentTarget;
        const error = video.error;
        setHasError(true);
        setErrorMessage(error ? `Video error: ${error.message}` : 'Failed to load video');
        onError?.(new Error(errorMessage));
    }, [onError, errorMessage]);

    // Seek
    const handleSeek = useCallback((value: number[]) => {
        if (!videoRef.current) return;
        const time = value[0];
        videoRef.current.currentTime = time;
        setCurrentTime(time);
        onTimeChange?.(time);
    }, [onTimeChange]);

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
            // If no HLS instance, just update state for display
            setCurrentQuality(quality);
            setShowSettingsMenu(false);
            return;
        }

        if (quality === 'auto') {
            hlsRef.current.currentLevel = -1; // Auto quality
            setCurrentQuality('auto');
        } else {
            // Find the matching level by height
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
        // Only auto-hide when playing
        if (isPlaying) {
            controlsTimeoutRef.current = setTimeout(() => {
                setShowControls(false);
            }, 3000);
        }
    }, [isPlaying]);

    const hideControlsImmediately = useCallback(() => {
        if (isPlaying) {
            if (controlsTimeoutRef.current) {
                clearTimeout(controlsTimeoutRef.current);
            }
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
            if (centerOverlayTimeoutRef.current) {
                clearTimeout(centerOverlayTimeoutRef.current);
            }
        };
    }, [showControlsTemporarily, hideControlsImmediately]);

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
                    togglePlay();
                    showCenterIcon(isPlaying ? 'pause' : 'play');
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
    }, [togglePlay, toggleMute, toggleFullscreen, togglePiP, skip, isPlaying, showCenterIcon, setVolume]);

    const playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 2];

    return (
        <div
            ref={containerRef}
            className={`relative bg-black rounded-2xl overflow-hidden aspect-video group video-player-container ${className}`}
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
                onEnded={handleEnded}
                onPlay={() => {
                    setIsPlaying(true);
                    onPlayingChange?.(true);
                    onPlay?.();
                }}
                onPause={() => {
                    setIsPlaying(false);
                    onPlayingChange?.(false);
                    onPause?.();
                }}
                onError={handleError}
                poster={poster ? getFullUrl(poster) : undefined}
                className="w-full h-full cursor-pointer"
                playsInline
            />

            {/* Center overlay icon (shows on click) */}
            {showCenterOverlay && (
                <div
                    className="absolute inset-0 flex items-center justify-center pointer-events-none z-30"
                    aria-hidden="true"
                >
                    <div className="w-24 h-24 bg-black/40 backdrop-blur-sm rounded-full flex items-center justify-center animate-fade-in">
                        {centerOverlayIcon === 'play' ? (
                            <Play size={56} className="text-white fill-white ml-2"/>
                        ) : (
                            <Pause size={56} className="text-white fill-white"/>
                        )}
                    </div>
                </div>
            )}

            {/* Center play button overlay when paused */}
            {!isPlaying && !hasError && (
                <div
                    className="absolute inset-0 flex items-center justify-center bg-black/20 opacity-0 group-hover:opacity-100 transition-opacity duration-300 cursor-pointer z-10"
                    onClick={(e) => {
                        e.stopPropagation();
                        togglePlay();
                    }}
                    role="button"
                    aria-label="Play video"
                    tabIndex={0}
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault();
                            togglePlay();
                        }
                    }}
                >
                    <div
                        className="w-20 h-20 bg-white/20 backdrop-blur-sm rounded-full flex items-center justify-center transition-transform hover:scale-110">
                        <Play size={48} className="text-white fill-white ml-2"/>
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
                    <AlertCircle size={64} className="text-red-500"/>
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

            {/* Controls overlay */}
            <div
                className={`absolute inset-0 flex flex-col justify-end transition-opacity duration-300 z-10 ${
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
                <div className="relative px-4 pb-4 pt-12">
                    {/* Progress bar */}
                    <div
                        className="relative h-1.5 group/hover:h-2.5 transition-all mb-4 cursor-pointer"
                        role="slider"
                        aria-label="Video progress"
                        aria-valuemin={0}
                        aria-valuemax={Math.floor(duration)}
                        aria-valuenow={Math.floor(currentTime)}
                        tabIndex={0}
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
                            <input
                                type="range"
                                value={currentTime}
                                min={0}
                                max={duration || 100}
                                step={0.1}
                                onChange={(e) => handleSeek([parseFloat(e.target.value)])}
                                className="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
                                aria-label="Seek video"
                            />
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
                                            className="px-3 py-2 text-xs font-bold text-gray-400 uppercase tracking-wider border-b border-white/10">
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
                                            <span>Auto</span>
                                            {currentQuality === 'auto' && (
                                                <span className="text-xs text-blue-400">Active</span>
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
                                                <span className="flex items-center gap-2">
                                                    {q.name}
                                                    {q.isRecommended && (
                                                        <span
                                                            className="text-[10px] bg-blue-500/20 text-blue-400 px-1.5 py-0.5 rounded">
                                                            Recommended
                                                        </span>
                                                    )}
                                                </span>
                                                {currentQuality === q.name && (
                                                    <span className="text-xs text-blue-400">Active</span>
                                                )}
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
        </div>
    );
});

VideoPlayer.displayName = 'VideoPlayer';

export default VideoPlayer;
