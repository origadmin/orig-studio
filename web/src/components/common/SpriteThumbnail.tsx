/**
 * SpriteThumbnail - Displays a single frame from a sprite sheet image.
 *
 * Uses CSS background-image + background-position + background-size
 * to show a specific frame from a larger sprite sheet, avoiding
 * the need to create multiple <img> elements.
 */

import React from 'react';

interface SpriteThumbnailProps {
    /** Full URL of the sprite sheet image */
    imageUrl: string;
    /** X offset of the frame in the sprite sheet (px) */
    x: number;
    /** Y offset of the frame in the sprite sheet (px) */
    y: number;
    /** Frame width (px) */
    w: number;
    /** Frame height (px) */
    h: number;
    /** Total width of the sprite sheet (px) */
    totalWidth: number;
    /** Total height of the sprite sheet (px) */
    totalHeight: number;
    /** Additional CSS class name */
    className?: string;
}

const SpriteThumbnail: React.FC<SpriteThumbnailProps> = ({
    imageUrl,
    x,
    y,
    w,
    h,
    totalWidth,
    totalHeight,
    className,
}) => {
    return (
        <div
            className={className}
            role="img"
            aria-label="Video preview thumbnail"
            style={{
                width: `${w}px`,
                height: `${h}px`,
                backgroundImage: `url(${imageUrl})`,
                backgroundSize: `${totalWidth}px ${totalHeight}px`,
                backgroundPosition: `-${x}px -${y}px`,
                backgroundRepeat: 'no-repeat',
            }}
        />
    );
};

export default SpriteThumbnail;
