import React from 'react';
import {useNavigate} from '@tanstack/react-router';
import {ListVideo, Clock, MoreVertical, Play, Pencil, Trash2} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type {ChannelPlaylist} from '@/lib/api/channel';

interface PlaylistCardProps {
    playlist: ChannelPlaylist;
    isOwner?: boolean;
    onEdit?: (playlistId: string) => void;
    onDelete?: (playlistId: string) => void;
}

const PlaylistCard: React.FC<PlaylistCardProps> = ({
    playlist,
    isOwner = false,
    onEdit,
    onDelete,
}) => {
    const navigate = useNavigate();

    const displayName = playlist.title || playlist.name || 'Untitled Playlist';
    const videoCount = playlist.video_count || playlist.media_count || 0;

    const timeAgo = (dateStr: string): string => {
        if (!dateStr) return '';
        const date = new Date(dateStr);
        if (isNaN(date.getTime())) return dateStr;

        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / 60000);
        const diffHours = Math.floor(diffMins / 60);
        const diffDays = Math.floor(diffHours / 24);

        if (diffMins < 1) return 'just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        if (diffHours < 24) return `${diffHours}h ago`;
        if (diffDays < 7) return `${diffDays}d ago`;
        if (diffDays < 30) return `${Math.floor(diffDays / 7)}w ago`;
        return date.toLocaleDateString();
    };

    return (
        <div
            className="group cursor-pointer border rounded-lg hover:border-primary/30 hover:shadow-md transition-all bg-card overflow-hidden"
            onClick={() => navigate({to: '/me/playlists'})}
        >
            {/* Cover image grid */}
            <div className="relative aspect-video overflow-hidden">
                {playlist.thumbnail ? (
                    <img
                        src={playlist.thumbnail}
                        alt={displayName}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                        loading="lazy"
                    />
                ) : (
                    <div className="grid grid-cols-2 gap-0.5 w-full h-full bg-muted">
                        {[...Array(4)].map((_, idx) => (
                            <div
                                key={idx}
                                className={`bg-muted-foreground/5 relative ${
                                    idx === 3 ? 'col-span-2 row-span-2' : ''
                                }`}
                            >
                                {playlist.cover_images?.[idx] ? (
                                    <img
                                        src={playlist.cover_images[idx]}
                                        alt=""
                                        className="w-full h-full object-cover"
                                        loading="lazy"
                                    />
                                ) : (
                                    <div className="w-full h-full flex items-center justify-center text-muted-foreground/30">
                                        <ListVideo className="w-6 h-6"/>
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                )}

                {/* Video count badge */}
                <div className="absolute bottom-0 right-0 bg-black/80 text-white px-2 py-1 text-xs font-medium flex items-center gap-1 rounded-tl-md">
                    <ListVideo className="w-3.5 h-3.5"/>
                    {videoCount}
                </div>

                {/* Play all overlay on hover */}
                <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors duration-200 flex items-center justify-center opacity-0 group-hover:opacity-100">
                    <div className="px-3 py-1.5 bg-black/70 rounded-full text-white text-xs font-medium flex items-center gap-1.5 backdrop-blur-sm">
                        <Play className="w-3.5 h-3.5 fill-white"/>
                        Play all
                    </div>
                </div>
            </div>

            {/* Playlist info */}
            <div className="p-3 sm:p-4">
                <div className="flex items-start justify-between gap-2">
                    <h3 className="font-medium text-sm line-clamp-2 group-hover:text-primary transition-colors">
                        {displayName}
                    </h3>

                    {isOwner && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <button
                                    onClick={(e) => e.stopPropagation()}
                                    className="opacity-0 group-hover:opacity-100 p-1 hover:bg-accent rounded-full transition-all flex-shrink-0 h-6 w-6 flex items-center justify-center"
                                >
                                    <MoreVertical className="h-3 w-3"/>
                                </button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-40">
                                <DropdownMenuItem onClick={(e) => {
                                    e.stopPropagation();
                                    onEdit?.(playlist.id);
                                }}>
                                    <Pencil className="h-4 w-4 mr-2"/>
                                    Edit
                                </DropdownMenuItem>
                                <DropdownMenuSeparator/>
                                <DropdownMenuItem
                                    className="text-destructive focus:text-destructive"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        onDelete?.(playlist.id);
                                    }}
                                >
                                    <Trash2 className="h-4 w-4 mr-2"/>
                                    Delete
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </div>

                <div className="flex items-center gap-3 mt-1.5 text-xs text-muted-foreground">
                    <span className="flex items-center gap-1">
                        <ListVideo className="w-3 h-3"/>
                        {videoCount} {videoCount === 1 ? 'video' : 'videos'}
                    </span>
                    {(playlist.updated_at || playlist.update_time) && (
                        <span className="flex items-center gap-1">
                            <Clock className="w-3 h-3"/>
                            {timeAgo(playlist.updated_at || playlist.update_time)}
                        </span>
                    )}
                </div>

                {playlist.description && (
                    <p className="text-xs text-muted-foreground mt-1.5 line-clamp-1">
                        {playlist.description}
                    </p>
                )}
            </div>
        </div>
    );
};

export default PlaylistCard;
