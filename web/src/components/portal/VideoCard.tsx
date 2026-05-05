/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Calendar, Loader2} from 'lucide-react';
import {Badge} from '@/components/ui/badge';
import {MediaItem} from '@/types/media';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';

const VideoCard = ({video}: { video: MediaItem }) => {
    const isProcessing = video.encoding_status !== 'success';

    // Resolve user info from edges.user (populated by normalizeMedia) or flat user field.
    // Fallback to deprecated author_name/author_avatar for backward compatibility.
    const edgeUser = video.edges?.user?.[0] || video.user;
    const authorName = edgeUser?.nickname || edgeUser?.username || video.author_name || 'OrigAdmin Contributor';
    const authorAvatar = edgeUser?.avatar || video.author_avatar;

    return (
        <div
            className="group cursor-pointer rounded-[2rem] bg-card border border-border hover:border-brand-muted transition-all overflow-hidden shadow-sm hover:shadow-2xl hover:-translate-y-2 duration-500 ease-out">
            <Link to="/watch" search={{v: video.short_token}} className="block relative aspect-video overflow-hidden">
                <img
                    src={getImageUrl(video.thumbnail, 'thumbnail')}
                    alt={video.title}
                    loading="lazy"
                    onError={(e) => handleImageError(e, 'thumbnail')}
                    className="object-cover w-full h-full group-hover:scale-110 transition-transform duration-700 ease-in-out"
                />

                {/* Processing indicator */}
                {isProcessing && (
                    <Badge
                        variant="secondary"
                        className="absolute top-2 left-2 z-10 gap-1 bg-black/60 text-white border-white/10 backdrop-blur text-[9px] px-1.5 py-0 h-5 hover:bg-black/60 uppercase tracking-wide"
                    >
                        {video.encoding_status === 'processing' ? (
                            <><Loader2 size={8} className="animate-spin"/>Processing</>
                        ) : (
                            <><Eye size={8}/>Optimizing</>
                        )}
                    </Badge>
                )}

                {/* Play Icon Overlay */}
                <div
                    className="absolute inset-0 bg-brand/10 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                    <div
                        className="w-14 h-14 bg-white/20 backdrop-blur-md rounded-full flex items-center justify-center border border-white/30 text-white shadow-xl transform scale-75 group-hover:scale-100 transition-transform duration-500">
                        <Play size={28} className="fill-current ml-1"/>
                    </div>
                </div>
                {/* Duration */}
                <Badge
                    variant="secondary"
                    className="absolute bottom-3 right-3 bg-black/80 text-white text-[10px] font-black px-2 py-0 h-5 backdrop-blur-sm hover:bg-black/80 leading-none"
                >
                    {formatDuration(video.duration)}
                </Badge>
            </Link>

            <div className="p-6 space-y-4">
                <h3 className="font-black text-foreground line-clamp-2 leading-tight group-hover:text-info transition-colors text-lg tracking-tight">
                    <Link to="/watch" search={{v: video.short_token}}>{video.title}</Link>
                </h3>

                <div className="flex items-center space-x-4 border-t border-border pt-4">
                    <div
                        className="w-10 h-10 rounded-2xl bg-brand-muted overflow-hidden ring-2 ring-card shadow-sm shrink-0 border border-brand-muted group-hover:rotate-6 transition-transform">
                        <img
                            src={getImageUrl(authorAvatar, 'avatar')}
                            alt={authorName}
                            loading="lazy"
                            onError={(e) => handleImageError(e, 'avatar')}
                            className="object-cover w-full h-full"
                        />
                    </div>
                    <div className="min-w-0 flex-1">
                        <p className="font-black text-foreground truncate hover:text-info transition-colors cursor-pointer text-sm">
                            {authorName}
                        </p>
                        <div
                            className="flex items-center space-x-2 text-[10px] font-black text-muted-foreground uppercase tracking-widest mt-0.5">
                            <span className="flex items-center gap-1"><Eye size={12}
                                                                           className="text-info"/> {formatViews(video.view_count)}</span>
                            <span>•</span>
                            <span className="flex items-center gap-1"><Calendar size={12}
                                                                                className="text-info"/> {formatDate(video.create_time)}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default VideoCard;
