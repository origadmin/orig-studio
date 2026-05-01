/**
 * SpritePreview - Progress bar hover preview container.
 *
 * Combines useSpriteVtt + SpriteThumbnail + time label
 * to display a thumbnail preview when hovering over the video progress bar.
 * Handles positioning logic (follows mouse + boundary constraints).
 */

import React from 'react';
import {useSpriteVtt} from '@/hooks/useSpriteVtt';
import {findCueAtTime} from '@/lib/parseWebVTT';
import {formatDuration} from '@/lib/format';
import SpriteThumbnail from './SpriteThumbnail';

interface SpritePreviewProps {
    /** Hover time point in seconds */
    hoverTime: number;
    /** Mouse position ratio on the progress bar (0-1) */
    hoverRatio: number;
    /** Progress bar DOMRect */
    progressBarRect: DOMRect;
    /** Player container DOMRect */
    playerRect: DOMRect;
    /** WebVTT file URL */
    vttUrl: string | null;
    /** Video total duration in seconds */
    duration: number;
}

const SpritePreview: React.FC<SpritePreviewProps> = ({
    hoverTime,
    hoverRatio,
    progressBarRect,
    playerRect,
    vttUrl,
    duration,
}) => {
    const {parsed, loading, error} = useSpriteVtt(vttUrl);

    // Find the CUE for the current hover time
    const cue = parsed ? findCueAtTime(parsed.cues, hoverTime) : null;

    // Calculate thumbnail dimensions
    const thumbnailWidth = cue?.w ?? 160;
    const thumbnailHeight = cue?.h ?? 90;
    const timeLabelHeight = 24;
    const gap = 8;

    // Horizontal positioning: follow mouse, center-aligned, clamp to player bounds
    const mouseAbsoluteX = progressBarRect.left + hoverRatio * progressBarRect.width;
    const playerLeft = playerRect.left;
    const playerWidth = playerRect.width;

    let left = mouseAbsoluteX - playerLeft - thumbnailWidth / 2;
    // Left boundary
    if (left < 0) left = 0;
    // Right boundary
    if (left + thumbnailWidth > playerWidth) left = playerWidth - thumbnailWidth;

    // Vertical positioning: above the progress bar
    const bottom = playerRect.bottom - progressBarRect.top + gap;

    const hasThumbnail = cue && parsed && !error && !loading;

    return (
        <div
            className="absolute pointer-events-none z-50"
            style={{
                left: `${left}px`,
                bottom: `${bottom}px`,
                width: `${thumbnailWidth}px`,
            }}
        >
            {/* Thumbnail image */}
            {hasThumbnail && (
                <SpriteThumbnail
                    imageUrl={parsed.imageUrl}
                    x={cue.x}
                    y={cue.y}
                    w={cue.w}
                    h={cue.h}
                    totalWidth={parsed.totalWidth}
                    totalHeight={parsed.totalHeight}
                    className="rounded-sm overflow-hidden"
                />
            )}

            {/* Time label */}
            <div className="text-center mt-1">
                <span className="bg-black/80 text-white text-xs px-1.5 py-0.5 rounded">
                    <time dateTime={`PT${Math.floor(hoverTime)}S`}>
                        {formatDuration(Math.floor(hoverTime))}
                    </time>
                </span>
            </div>
        </div>
    );
};

export default SpritePreview;
