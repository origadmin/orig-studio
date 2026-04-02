/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React from 'react';
import {Link} from '@tanstack/react-router';
import {Play, Eye, Calendar, Loader2} from 'lucide-react';
import {MediaItem} from '@/types/media';
import {formatDuration, formatViews, formatDate} from '@/lib/format';

const VideoCard = ({video}: { video: MediaItem }) => {
    const isProcessing = video.encoding_status !== 'success';

    return (
        <div
            className="group cursor-pointer rounded-[2rem] bg-white border border-gray-100 hover:border-blue-100 transition-all overflow-hidden shadow-sm hover:shadow-2xl hover:-translate-y-2 duration-500 ease-out">
            <Link to="/watch" search={{v: String(video.id)}} className="block relative aspect-video overflow-hidden">
                <img
                    src={video.thumbnail || 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?auto=format&fit=crop&q=80&w=800'}
                    alt={video.title}
                    className="object-cover w-full h-full group-hover:scale-110 transition-transform duration-700 ease-in-out"
                />

                {/* Processing Badge */}
                {isProcessing && (
                    <div className="absolute top-3 left-3 z-10">
                        <div
                            className="bg-black/60 backdrop-blur-md text-white text-[10px] font-black px-2 py-1 rounded-lg flex items-center gap-1.5 shadow-lg border border-white/10 uppercase tracking-wider">
                            {video.encoding_status === 'processing' ? (
                                <>
                                    <Loader2 size={10} className="animate-spin"/>
                                    Processing
                                </>
                            ) : (
                                <>
                                    <Eye size={10}/>
                                    Optimizing
                                </>
                            )}
                        </div>
                    </div>
                )}

                {/* Play Icon Overlay */}
                <div
                    className="absolute inset-0 bg-blue-600/10 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                    <div
                        className="w-14 h-14 bg-white/20 backdrop-blur-md rounded-full flex items-center justify-center border border-white/30 text-white shadow-xl transform scale-75 group-hover:scale-100 transition-transform duration-500">
                        <Play size={28} className="fill-current ml-1"/>
                    </div>
                </div>
                {/* Duration Badge */}
                <div
                    className="absolute bottom-3 right-3 bg-black/80 text-white text-[10px] font-black px-2 py-1 rounded-lg leading-none backdrop-blur-sm shadow-lg">
                    {formatDuration(video.duration)}
                </div>
            </Link>

            <div className="p-6 space-y-4">
                <h3 className="font-black text-slate-900 line-clamp-2 leading-tight group-hover:text-blue-600 transition-colors text-lg tracking-tight">
                    <Link to="/watch" search={{v: String(video.id)}}>{video.title}</Link>
                </h3>

                <div className="flex items-center space-x-4 border-t border-gray-50 pt-4">
                    <div
                        className="w-10 h-10 rounded-2xl bg-blue-50 overflow-hidden ring-2 ring-white shadow-sm shrink-0 border border-blue-100 group-hover:rotate-6 transition-transform">
                        <img
                            src={video.author_avatar || `https://i.pravatar.cc/100?u=${video.user_id}`}
                            alt={video.author_name}
                            className="object-cover w-full h-full"
                        />
                    </div>
                    <div className="min-w-0 flex-1">
                        <p className="font-black text-slate-800 truncate hover:text-blue-500 transition-colors cursor-pointer text-sm">
                            {video.author_name || 'OrigAdmin Contributor'}
                        </p>
                        <div
                            className="flex items-center space-x-2 text-[10px] font-black text-slate-400 uppercase tracking-widest mt-0.5">
                            <span className="flex items-center gap-1"><Eye size={12}
                                                                           className="text-blue-500"/> {formatViews(video.view_count)}</span>
                            <span>•</span>
                            <span className="flex items-center gap-1"><Calendar size={12}
                                                                                className="text-blue-500"/> {new Date(video.create_time).toLocaleDateString('zh-CN')}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default VideoCard;
