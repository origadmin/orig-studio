import React from 'react';
import {useSpriteVtt} from '@/hooks/useSpriteVtt';
import {findCueAtTime} from '@/lib/parseWebVTT';
import {formatDuration} from '@/lib/format';
import SpriteThumbnail from './SpriteThumbnail';

interface SpritePreviewProps {
    hoverTime: number;
    hoverRatio: number;
    progressBarRect: DOMRect;
    playerRect: DOMRect;
    vttUrl: string | null;
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

    const cue = parsed ? findCueAtTime(parsed.cues, hoverTime) : null;

    const thumbnailWidth = cue?.w ?? 160;
    const thumbnailHeight = cue?.h ?? 90;
    const gap = 8;

    const mouseAbsoluteX = progressBarRect.left + hoverRatio * progressBarRect.width;
    const playerLeft = playerRect.left;
    const playerWidth = playerRect.width;

    let left = mouseAbsoluteX - playerLeft - thumbnailWidth / 2;
    if (left < 0) left = 0;
    if (left + thumbnailWidth > playerWidth) left = playerWidth - thumbnailWidth;

    const thumbnailBottom = playerRect.bottom - progressBarRect.top + gap;
    const timeLabelTop = progressBarRect.bottom - playerRect.top + 2;

    const displayTime = cue ? cue.startTime : hoverTime;
    const hasThumbnail = cue && parsed && !error && !loading;

    return (
        <>
            {hasThumbnail && (
                <div
                    className="absolute pointer-events-none z-50"
                    style={{
                        left: `${left}px`,
                        bottom: `${thumbnailBottom}px`,
                        width: `${thumbnailWidth}px`,
                    }}
                >
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
                </div>
            )}

            <div
                className="absolute pointer-events-none z-50"
                style={{
                    left: `${left}px`,
                    top: `${timeLabelTop}px`,
                    width: `${thumbnailWidth}px`,
                }}
            >
                <div className="text-center">
                    <span className="bg-black/80 text-white text-xs px-1.5 py-0.5 rounded">
                        <time dateTime={`PT${Math.floor(displayTime)}S`}>
                            {formatDuration(Math.floor(displayTime))}
                        </time>
                    </span>
                </div>
            </div>
        </>
    );
};

export default SpritePreview;
