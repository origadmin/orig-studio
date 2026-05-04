import React, {useState} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {Clock, Eye, MoreVertical, Pencil, BarChart3, Play, ListPlus, Share2, Flag} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface Video {
    id: string;
    short_token?: string;
    title: string;
    thumbnail?: string;
    duration?: number;
    view_count?: number;
    published_at?: string;
    create_time?: string;
    progress?: number;
    user?: {
        id?: string;
        username?: string;
        nickname?: string;
        avatar?: string;
    };
}

interface VideoCardProps {
    video: Video;
    showChannelInfo?: boolean;
    showProgress?: boolean;
    isOwner?: boolean;
    size?: 'normal' | 'compact' | 'large';
    onEdit?: (videoId: string) => void;
    onViewStats?: (videoId: string) => void;
    onAddToPlaylist?: (videoId: string) => void;
    onShare?: (videoId: string) => void;
}

const VideoCard: React.FC<VideoCardProps> = ({
    video,
    showChannelInfo = false,
    showProgress = false,
    isOwner = false,
    size = 'normal',
    onEdit,
    onViewStats,
    onAddToPlaylist,
    onShare,
}) => {
    const navigate = useNavigate();
    const {t} = useTranslation();
    const [isHovered, setIsHovered] = useState(false);

    const formatDuration = (seconds: number): string => {
        const hrs = Math.floor(seconds / 3600);
        const mins = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;

        if (hrs > 0) {
            return `${hrs}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
        }
        return `${mins}:${secs.toString().padStart(2, '0')}`;
    };

    const formatCount = (num: number): string => {
        if (!num) return '0';
        if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
        if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
        return num.toString();
    };

    const timeAgo = (dateStr: string): string => {
        const date = new Date(dateStr);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMins / 60);
        const diffDays = Math.floor(diffHours / 24);

        if (diffMins < 1) return t('common.justNow');
        if (diffMins < 60) return t('common.minutesAgo', {count: diffMins});
        if (diffHours < 24) return t('common.hoursAgo', {count: diffHours});
        if (diffDays < 7) return t('common.daysAgo', {count: diffDays});
        if (diffDays < 30) return t('common.weeksAgo', {count: Math.floor(diffDays / 7)});
        if (diffDays < 365) return t('common.monthsAgo', {count: Math.floor(diffDays / 30)});
        return date.toLocaleDateString();
    };

    const sizeClasses = {
        compact: 'max-w-[200px]',
        normal: '',
        large: 'max-w-none',
    };

    return (
        <div
            className={`group cursor-pointer ${sizeClasses[size]}`}
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
            onClick={() => navigate({to: '/watch', search: {v: video.short_token || String(video.id)}})}
        >
            {/* Thumbnail */}
            <div className="relative aspect-video bg-muted rounded-lg overflow-hidden shadow-sm group-hover:shadow-md transition-shadow duration-200">
                {video.thumbnail ? (
                    <img
                        src={video.thumbnail}
                        alt={video.title}
                        className={`w-full h-full object-cover transition-transform duration-300 ${isHovered ? 'scale-105' : 'scale-100'}`}
                        loading="lazy"
                    />
                ) : (
                    <div className="w-full h-full flex items-center justify-center text-muted-foreground bg-muted">
                        <Film className="w-10 h-10 opacity-50"/>
                    </div>
                )}

                {/* Duration badge */}
                {video.duration && (
                    <span className="absolute bottom-2 right-2 px-1.5 py-0.5 bg-black/80 text-white text-xs rounded font-medium flex items-center gap-1">
                        <Clock className="w-3 h-3"/>
                        {formatDuration(video.duration)}
                    </span>
                )}

                {/* Play overlay on hover */}
                {isHovered && (
                    <div className="absolute inset-0 bg-black/30 flex items-center justify-center transition-opacity duration-200">
                        <div className="w-[clamp(2rem,6vw,3rem)] h-[clamp(2rem,6vw,3rem)] rounded-full bg-black/60 flex items-center justify-center backdrop-blur-sm">
                            <Play className="text-white fill-white ml-0.5" style={{width: 'clamp(0.75rem,3vw,1.5rem)', height: 'clamp(0.75rem,3vw,1.5rem)'}}/>
                        </div>
                    </div>
                )}

                {/* Quick action buttons for owner on hover */}
                {isOwner && isHovered && (
                    <div className="absolute top-2 right-2 flex items-center gap-1.5 transition-opacity duration-200">
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onEdit?.(video.id);
                            }}
                            className="p-1.5 bg-black/60 hover:bg-black/80 rounded-full backdrop-blur-sm transition-colors"
                            title="Edit video"
                        >
                            <Pencil className="w-3.5 h-3.5 text-white"/>
                        </button>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onViewStats?.(video.id);
                            }}
                            className="p-1.5 bg-black/60 hover:bg-black/80 rounded-full backdrop-blur-sm transition-colors"
                            title="View stats"
                        >
                            <BarChart3 className="w-3.5 h-3.5 text-white"/>
                        </button>
                    </div>
                )}

                {/* Progress bar */}
                {showProgress && video.progress !== undefined && video.progress > 0 && (
                    <div className="absolute bottom-0 left-0 right-0 h-1 bg-black/30">
                        <div
                            className="h-full bg-red-600 transition-all duration-300"
                            style={{width: `${Math.min(video.progress, 100)}%`}}
                        />
                    </div>
                )}
            </div>

            {/* Video info */}
            <div className="mt-2.5 flex gap-2.5">
                {!showChannelInfo && (
                    <div className="hidden sm:block w-9 h-9 rounded-full bg-primary/10 flex-shrink-0 mt-0.5 overflow-hidden">
                        {video.user?.avatar ? (
                            <img src={video.user.avatar} alt="" className="w-full h-full object-cover"/>
                        ) : (
                            <div className="w-full h-full rounded-full bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center text-xs font-bold text-primary">
                                {(video.user?.nickname || video.user?.username || 'U').charAt(0).toUpperCase()}
                            </div>
                        )}
                    </div>
                )}
                <div className="flex-1 min-w-0">
                    <h3 className="font-medium text-sm line-clamp-2 group-hover:text-primary transition-colors leading-snug">
                        {video.title}
                    </h3>
                    <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                        {video.user?.username && (
                            <span>@{video.user.nickname || video.user.username}</span>
                        )}
                        {video.user?.username && video.view_count !== undefined && (
                            <span>·</span>
                        )}
                        {video.view_count !== undefined && (
                            <span className="flex items-center gap-1">
                                <Eye className="w-3 h-3"/>
                                {formatCount(video.view_count)}
                            </span>
                        )}
                        {(video.published_at || video.create_time) && (
                            <>
                                <span>·</span>
                                <span>{timeAgo(video.published_at || video.create_time!)}</span>
                            </>
                        )}
                    </div>
                </div>

                {/* More options dropdown */}
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <button
                            onClick={(e) => e.stopPropagation()}
                            className="opacity-0 group-hover:opacity-100 p-1 hover:bg-accent rounded-full transition-all self-start flex-shrink-0 h-6 w-6 flex items-center justify-center"
                        >
                            <MoreVertical className="h-3 w-3"/>
                        </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-48">
                        <DropdownMenuItem onClick={(e) => {
                            e.stopPropagation();
                            onAddToPlaylist?.(video.id);
                        }}>
                            <ListPlus className="h-4 w-4 mr-2"/>
                            {t('watch.saveToPlaylist')}
                        </DropdownMenuItem>
                        <DropdownMenuItem onClick={(e) => {
                            e.stopPropagation();
                            onShare?.(video.id);
                        }}>
                            <Share2 className="h-4 w-4 mr-2"/>
                            {t('watch.share')}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator/>
                        <DropdownMenuItem onClick={(e) => {
                            e.stopPropagation();
                            console.log('Report video:', video.id);
                        }}>
                            <Flag className="h-4 w-4 mr-2"/>
                            {t('channel.reportChannel')}
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
        </div>
    );
};

// Simple Film icon for placeholder
function Film({className}: {className?: string}) {
    return (
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
            <rect width="20" height="20" x="2" y="2" rx="2.18" ry="2.18"/>
            <line x1="7" x2="7" y1="2" y2="22"/>
            <line x1="17" x2="17" y1="2" y2="22"/>
            <line x1="2" x2="22" y1="12" y2="12"/>
            <line x1="2" x2="7" y1="7" y2="7"/>
            <line x1="2" x2="7" y1="17" y2="17"/>
            <line x1="17" x2="22" y1="7" y2="7"/>
            <line x1="17" x2="22" y1="17" y2="17"/>
        </svg>
    );
}

export default VideoCard;
