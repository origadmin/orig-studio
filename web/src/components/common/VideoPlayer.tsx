import React, {useState, useRef, useEffect, useCallback} from 'react';
import {
    Play, Pause, Volume2, VolumeX, Maximize, Minimize,
    SkipBack, SkipForward, Settings, Subtitles, PictureInPicture,
    MonitorPlay, RotateCcw
} from 'lucide-react';
import Hls from 'hls.js';
import {Button} from '@/components/ui/button';
import {formatDuration} from '@/lib/format';
import {getFullUrl} from '@/lib/utils';
import {usePlayerSettings} from '@/hooks/usePlayerSettings';

interface VideoPlayerProps {
    src: string;
    hlsSrc?: string;
    poster?: string;
    autoPlay?: boolean;
    onTimeUpdate?: (time: number) => void;
    className?: string;
    qualities?: Array<{ name: string; url: string }>;
}

const VideoPlayer: React.FC<VideoPlayerProps> = ({
                                                     src,
                                                     hlsSrc,
                                                     poster,
                                                     autoPlay = false,
                                                     onTimeUpdate,
                                                     className = '',
                                                     qualities
                                                 }) => {
    const videoRef = useRef<HTMLVideoElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const hlsRef = useRef<Hls | null>(null);
    const controlsTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const settingsMenuRef = useRef<HTMLDivElement>(null);

    const [isPlaying, setIsPlaying] = useState(false);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [showControls, setShowControls] = useState(true);
    const [showPlaybackMenu, setShowPlaybackMenu] = useState(false);
    const [showSettingsMenu, setShowSettingsMenu] = useState(false);
    const [buffered, setBuffered] = useState(0);
    const [currentQuality, setCurrentQuality] = useState<string>('auto');

    // 使用全局播放器设置
    const {
        volume,
        isMuted,
        playbackRate,
        setVolume,
        setIsMuted,
        setPlaybackRate,
    } = usePlayerSettings();

    // Initialize HLS player
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        const fullHlsSrc = hlsSrc ? getFullUrl(hlsSrc) : undefined;
        const fullSrc = getFullUrl(src);

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
            hlsRef.current = hls;
        } else if (video.canPlayType('application/vnd.apple.mpegurl') && fullHlsSrc) {
            video.src = fullHlsSrc;
        } else if (fullSrc) {
            video.src = fullSrc;
        }

        return () => {
            if (hlsRef.current) {
                hlsRef.current.destroy();
                hlsRef.current = null;
            }
        };
    }, [src, hlsSrc]);

    // 应用全局设置到视频元素
    useEffect(() => {
        const video = videoRef.current;
        if (!video) return;

        video.volume = volume;
        video.muted = isMuted;
        video.playbackRate = playbackRate;
    }, [volume, isMuted, playbackRate]);

    // 点击外部关闭菜单
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

    // Play/Pause
    const togglePlay = useCallback(() => {
        if (!videoRef.current) return;
        if (videoRef.current.paused) {
            videoRef.current.play();
            setIsPlaying(true);
        } else {
            videoRef.current.pause();
            setIsPlaying(false);
        }
    }, []);

    // Handle video events
    const handleTimeUpdate = useCallback(() => {
        if (!videoRef.current) return;
        setCurrentTime(videoRef.current.currentTime);
        onTimeUpdate?.(videoRef.current.currentTime);

        // Update buffered
        if (videoRef.current.buffered.length > 0) {
            const bufferedEnd = videoRef.current.buffered.end(videoRef.current.buffered.length - 1);
            setBuffered(bufferedEnd / videoRef.current.duration);
        }
    }, [onTimeUpdate]);

    const handleLoadedMetadata = useCallback(() => {
        if (!videoRef.current) return;
        setDuration(videoRef.current.duration);
    }, []);

    const handleEnded = useCallback(() => {
        setIsPlaying(false);
    }, []);

    // Seek
    const handleSeek = useCallback((value: number[]) => {
        if (!videoRef.current) return;
        const time = value[0];
        videoRef.current.currentTime = time;
        setCurrentTime(time);
    }, []);

    // Volume
    const handleVolumeChange = useCallback((value: number[]) => {
        if (!videoRef.current) return;
        const newVolume = value[0];
        videoRef.current.volume = newVolume;
        setVolume(newVolume);
    }, [setVolume]);

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

    // Fullscreen
    const toggleFullscreen = useCallback(() => {
        if (!containerRef.current) return;

        if (!document.fullscreenElement) {
            containerRef.current.requestFullscreen().catch(err => {
                console.log('Fullscreen error:', err);
            });
            setIsFullscreen(true);
        } else {
            document.exitFullscreen();
            setIsFullscreen(false);
        }
    }, []);

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

    // 清晰度选择
    const handleQualityChange = useCallback((quality: string) => {
        setCurrentQuality(quality);
        setShowSettingsMenu(false);
    }, []);

    // Skip forward/backward
    const skip = useCallback((seconds: number) => {
        if (!videoRef.current) return;
        videoRef.current.currentTime = Math.max(0, Math.min(duration, videoRef.current.currentTime + seconds));
    }, [duration]);

    // Controls visibility
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

    useEffect(() => {
        const handleMouseMove = () => {
            showControlsTemporarily();
        };

        const container = containerRef.current;
        if (container) {
            container.addEventListener('mousemove', handleMouseMove);
        }

        return () => {
            if (container) {
                container.removeEventListener('mousemove', handleMouseMove);
            }
            if (controlsTimeoutRef.current) {
                clearTimeout(controlsTimeoutRef.current);
            }
        };
    }, [showControlsTemporarily]);

    const playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 2];

    return (
        <div
            ref={containerRef}
            className={`relative bg-black rounded-2xl overflow-hidden aspect-video group ${className}`}
            onMouseEnter={showControlsTemporarily}
            onMouseLeave={() => isPlaying && setShowControls(false)}
        >
            <video
                ref={videoRef}
                onClick={togglePlay}
                onTimeUpdate={handleTimeUpdate}
                onLoadedMetadata={handleLoadedMetadata}
                onEnded={handleEnded}
                onPlay={() => setIsPlaying(true)}
                onPause={() => setIsPlaying(false)}
                poster={poster ? getFullUrl(poster) : undefined}
                className="w-full h-full cursor-pointer"
            />

            {/* Center play button overlay when paused */}
            {!isPlaying && (
                <div
                    className="absolute inset-0 flex items-center justify-center bg-black/30 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
                    onClick={togglePlay}
                >
                    <div
                        className="w-20 h-20 bg-white/20 backdrop-blur-sm rounded-full flex items-center justify-center">
                        <Play size={48} className="text-white fill-white ml-2"/>
                    </div>
                </div>
            )}

            {/* Controls overlay */}
            <div
                className={`absolute inset-0 flex flex-col justify-end transition-opacity duration-300 ${
                    showControls ? 'opacity-100' : 'opacity-0'
                }`}
            >
                {/* Gradient background */}
                <div
                    className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent pointer-events-none"/>

                {/* Top controls - can add more here */}
                <div className="relative flex-1"/>

                {/* Bottom controls */}
                <div className="relative px-4 pb-4 pt-12">
                    {/* Progress bar */}
                    <div className="relative h-1.5 group-hover:h-2.5 transition-all mb-4">
                        {/* Buffered progress */}
                        <div className="absolute inset-0 bg-white/30 rounded-full overflow-hidden">
                            <div
                                className="h-full bg-white/50 rounded-full"
                                style={{width: `${buffered * 100}%`}}
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
                            />
                            <div
                                className="h-full bg-red-600 rounded-full transition-all"
                                style={{width: `${(currentTime / (duration || 1)) * 100}%`}}
                            >
                                <div
                                    className="absolute right-0 top-1/2 -translate-y-1/2 w-3 h-3 bg-red-600 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
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
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    togglePlay();
                                }}
                            >
                                {isPlaying ? <Pause size={24} fill="currentColor"/> :
                                    <Play size={24} fill="currentColor" className="ml-1"/>}
                            </Button>

                            {/* Skip backward */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    skip(-10);
                                }}
                            >
                                <SkipBack size={20}/>
                            </Button>

                            {/* Skip forward */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    skip(10);
                                }}
                            >
                                <SkipForward size={20}/>
                            </Button>

                            {/* Volume */}
                            <div className="flex items-center gap-2 group/volume">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-10 w-10 text-white hover:bg-white/10 rounded-full"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        toggleMute();
                                    }}
                                >
                                    {isMuted || volume === 0 ? <VolumeX size={20}/> : <Volume2 size={20}/>}
                                </Button>
                                <div
                                    className="w-0 group-hover/volume:w-24 overflow-hidden transition-all duration-300">
                                    <input
                                        type="range"
                                        value={isMuted ? 0 : volume}
                                        min={0}
                                        max={1}
                                        step={0.01}
                                        onChange={(e) => handleVolumeChange([parseFloat(e.target.value)])}
                                        className="w-24 h-1.5 bg-white/30 rounded-full appearance-none cursor-pointer accent-white"
                                    />
                                </div>
                            </div>

                            {/* Time display - 修复小数问题 */}
                            <span className="text-white text-sm font-medium min-w-[100px]">
                                {formatDuration(Math.floor(currentTime))} / {formatDuration(Math.floor(duration))}
                            </span>
                        </div>

                        <div className="flex items-center gap-2" ref={settingsMenuRef}>
                            {/* 字幕 */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full opacity-50 cursor-not-allowed"
                                disabled
                                title="字幕 (开发中)"
                            >
                                <Subtitles size={20}/>
                            </Button>

                            {/* 播放速度 */}
                            <div className="relative">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    className="h-10 px-3 text-white hover:bg-white/10 rounded-full text-sm font-medium"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowPlaybackMenu(!showPlaybackMenu);
                                        setShowSettingsMenu(false);
                                    }}
                                >
                                    {playbackRate}x
                                </Button>
                                {showPlaybackMenu && (
                                    <div
                                        className="absolute bottom-full right-0 mb-2 bg-black/90 backdrop-blur-md border border-white/10 rounded-lg overflow-hidden min-w-[120px] shadow-xl">
                                        {playbackRates.map((rate) => (
                                            <button
                                                key={rate}
                                                className={`w-full text-left px-4 py-2 text-white hover:bg-white/10 transition-colors text-sm ${
                                                    playbackRate === rate ? 'bg-white/10 font-semibold' : ''
                                                }`}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handlePlaybackRateChange(rate);
                                                }}
                                            >
                                                {rate}x
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* 设置 */}
                            <div className="relative">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-10 w-10 text-white hover:bg-white/10 rounded-full"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowSettingsMenu(!showSettingsMenu);
                                        setShowPlaybackMenu(false);
                                    }}
                                >
                                    <Settings size={20}/>
                                </Button>
                                {showSettingsMenu && (
                                    <div
                                        className="absolute bottom-full right-0 mb-2 bg-black/90 backdrop-blur-md border border-white/10 rounded-lg overflow-hidden min-w-[160px] shadow-xl">
                                        {/* 清晰度 */}
                                        <div
                                            className="px-3 py-2 text-xs font-bold text-gray-400 uppercase tracking-wider border-b border-white/10">
                                            清晰度
                                        </div>
                                        <button
                                            className={`w-full text-left px-4 py-2 text-white hover:bg-white/10 transition-colors text-sm ${
                                                currentQuality === 'auto' ? 'bg-white/10 font-semibold' : ''
                                            }`}
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleQualityChange('auto');
                                            }}
                                        >
                                            自动
                                        </button>
                                        {qualities?.map((q) => (
                                            <button
                                                key={q.name}
                                                className={`w-full text-left px-4 py-2 text-white hover:bg-white/10 transition-colors text-sm ${
                                                    currentQuality === q.name ? 'bg-white/10 font-semibold' : ''
                                                }`}
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleQualityChange(q.name);
                                                }}
                                            >
                                                {q.name}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* 画中画 */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full opacity-50 cursor-not-allowed"
                                disabled
                                title="画中画 (开发中)"
                            >
                                <PictureInPicture size={20}/>
                            </Button>

                            {/* 迷你播放器 */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full opacity-50 cursor-not-allowed"
                                disabled
                                title="迷你播放器 (开发中)"
                            >
                                <MonitorPlay size={20}/>
                            </Button>

                            {/* 全屏 */}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-10 w-10 text-white hover:bg-white/10 rounded-full"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    toggleFullscreen();
                                }}
                            >
                                {isFullscreen ? <Minimize size={20}/> : <Maximize size={20}/>}
                            </Button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default VideoPlayer;
