/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * VideoCardSkeleton - Loading skeleton for video cards
 */

import React from 'react';
import {Skeleton} from '@/components/ui/skeleton';

interface VideoCardSkeletonProps {
    count?: number;
}

const VideoCardSkeleton: React.FC<VideoCardSkeletonProps> = ({count = 1}) => {
    return (
        <>
            {Array.from({length: count}).map((_, i) => (
                <div key={i}
                     className="rounded-xl bg-white dark:bg-gray-800 overflow-hidden shadow-sm border border-gray-100 dark:border-gray-700">
                    {/* Thumbnail skeleton */}
                    <div className="relative aspect-video">
                        <Skeleton className="w-full h-full rounded-none"/>
                    </div>

                    {/* Info skeleton */}
                    <div className="p-3 space-y-2">
                        {/* Title */}
                        <Skeleton className="h-4 w-full"/>
                        <Skeleton className="h-4 w-3/4"/>

                        {/* Channel info */}
                        <div className="flex items-center gap-2 pt-1">
                            <Skeleton className="h-5 w-5 rounded-full shrink-0"/>
                            <Skeleton className="h-3 w-24"/>
                        </div>

                        {/* Meta info */}
                        <div className="flex items-center gap-3">
                            <Skeleton className="h-3 w-16"/>
                            <Skeleton className="h-3 w-20"/>
                        </div>
                    </div>
                </div>
            ))}
        </>
    );
};

export default VideoCardSkeleton;
