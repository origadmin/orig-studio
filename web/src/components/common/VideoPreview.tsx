import React, {useState, useEffect, useRef} from 'react';

interface VideoPreviewProps {
    previewFile: string;
    duration: number;
    currentTime: number;
    width: number;
    height: number;
    thumbnailCount?: number;
}

const VideoPreview: React.FC<VideoPreviewProps> = ({
                                                       previewFile,
                                                       duration,
                                                       currentTime,
                                                       width,
                                                       height,
                                                       thumbnailCount = 10,
                                                   }) => {
    const [hoverTime, setHoverTime] = useState<number | null>(null);
    const [thumbnailWidth, setThumbnailWidth] = useState(0);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (containerRef.current) {
            setThumbnailWidth(containerRef.current.offsetWidth / thumbnailCount);
        }
    }, [thumbnailCount]);

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        if (!containerRef.current) return;

        const rect = containerRef.current.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const ratio = x / rect.width;
        const time = ratio * duration;
        setHoverTime(time);
    };

    const handleMouseLeave = () => {
        setHoverTime(null);
    };

    const getThumbnailIndex = (time: number) => {
        const index = Math.floor((time / duration) * (thumbnailCount - 1));
        return Math.min(index, thumbnailCount - 1);
    };

    const renderThumbnail = (time: number) => {
        const index = getThumbnailIndex(time);
        const left = -index * thumbnailWidth;

        return (
            <div
                key={time}
                className="absolute top-0 left-0 w-full h-full overflow-hidden"
                style={{
                    backgroundImage: `url(${previewFile})`,
                    backgroundSize: `${thumbnailCount * 100}% 100%`,
                    backgroundPosition: `${left}px 0`,
                }}
            />
        );
    };

    return (
        <div
            ref={containerRef}
            className="relative w-full h-16 bg-gray-100 dark:bg-gray-800 rounded-md overflow-hidden"
            onMouseMove={handleMouseMove}
            onMouseLeave={handleMouseLeave}
        >
            {/* Default thumbnail (current time) */}
            {renderThumbnail(currentTime)}

            {/* Hover thumbnail */}
            {hoverTime !== null && renderThumbnail(hoverTime)}

            {/* Time indicator */}
            {hoverTime !== null && (
                <div
                    className="absolute top-0 left-0 w-full h-full flex items-center justify-center pointer-events-none">
                    <div className="bg-black/70 text-white text-xs px-2 py-1 rounded">
                        {Math.floor(hoverTime / 60)}:{String(Math.floor(hoverTime % 60)).padStart(2, '0')}
                    </div>
                </div>
            )}
        </div>
    );
};

export default VideoPreview;
