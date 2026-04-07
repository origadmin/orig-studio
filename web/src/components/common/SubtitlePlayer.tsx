import React, {useState, useEffect, useRef} from 'react';

interface Subtitle {
    id: number;
    startTime: number;
    endTime: number;
    text: string;
}

interface SubtitlePlayerProps {
    subtitleUrl: string;
    currentTime: number;
    onError?: (error: Error) => void;
}

const SubtitlePlayer: React.FC<SubtitlePlayerProps> = ({
                                                           subtitleUrl,
                                                           currentTime,
                                                           onError
                                                       }) => {
    const [subtitles, setSubtitles] = useState<Subtitle[]>([]);
    const [currentSubtitle, setCurrentSubtitle] = useState<Subtitle | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<Error | null>(null);
    const subtitleRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const fetchSubtitles = async () => {
            try {
                setLoading(true);
                setError(null);
                const response = await fetch(subtitleUrl);
                if (!response.ok) {
                    throw new Error(`Failed to fetch subtitles: ${response.status}`);
                }
                const text = await response.text();
                const parsedSubtitles = parseSRT(text);
                setSubtitles(parsedSubtitles);
            } catch (err) {
                const error = err as Error;
                setError(error);
                if (onError) {
                    onError(error);
                }
            } finally {
                setLoading(false);
            }
        };

        if (subtitleUrl) {
            fetchSubtitles();
        }
    }, [subtitleUrl, onError]);

    useEffect(() => {
        if (subtitles.length === 0) return;

        // Find the current subtitle based on currentTime
        const current = subtitles.find(subtitle =>
            currentTime >= subtitle.startTime && currentTime <= subtitle.endTime
        );

        setCurrentSubtitle(current || null);
    }, [currentTime, subtitles]);

    const parseSRT = (text: string): Subtitle[] => {
        const lines = text.split('\n');
        const subtitles: Subtitle[] = [];
        let currentSubtitle: Partial<Subtitle> = {};
        let currentText: string[] = [];

        for (const line of lines) {
            const trimmedLine = line.trim();

            if (trimmedLine === '') {
                // End of current subtitle
                if (currentSubtitle.id && currentSubtitle.startTime !== undefined && currentSubtitle.endTime !== undefined) {
                    subtitles.push({
                        id: currentSubtitle.id,
                        startTime: currentSubtitle.startTime,
                        endTime: currentSubtitle.endTime,
                        text: currentText.join('\n')
                    });
                }
                currentSubtitle = {};
                currentText = [];
            } else if (!currentSubtitle.id) {
                // Subtitle ID
                currentSubtitle.id = parseInt(trimmedLine, 10);
            } else if (!currentSubtitle.startTime) {
                // Time range
                const timeMatch = trimmedLine.match(/^(\d{2}):(\d{2}):(\d{2}),(\d{3}) --> (\d{2}):(\d{2}):(\d{2}),(\d{3})$/);
                if (timeMatch) {
                    currentSubtitle.startTime = parseTime(timeMatch[1], timeMatch[2], timeMatch[3], timeMatch[4]);
                    currentSubtitle.endTime = parseTime(timeMatch[5], timeMatch[6], timeMatch[7], timeMatch[8]);
                }
            } else {
                // Subtitle text
                currentText.push(trimmedLine);
            }
        }

        // Add the last subtitle
        if (currentSubtitle.id && currentSubtitle.startTime !== undefined && currentSubtitle.endTime !== undefined) {
            subtitles.push({
                id: currentSubtitle.id,
                startTime: currentSubtitle.startTime,
                endTime: currentSubtitle.endTime,
                text: currentText.join('\n')
            });
        }

        return subtitles;
    };

    const parseTime = (hours: string, minutes: string, seconds: string, milliseconds: string): number => {
        return parseInt(hours, 10) * 3600 +
            parseInt(minutes, 10) * 60 +
            parseInt(seconds, 10) +
            parseInt(milliseconds, 10) / 1000;
    };

    if (loading) {
        return null; // Don't show anything while loading
    }

    if (error) {
        return null; // Don't show anything on error
    }

    if (!currentSubtitle) {
        return null; // Don't show anything if no subtitle is active
    }

    return (
        <div
            ref={subtitleRef}
            className="absolute bottom-20 left-0 right-0 flex justify-center pointer-events-none"
        >
            <div className="bg-black/70 text-white text-center px-6 py-2 rounded-lg max-w-2xl mx-auto">
                {currentSubtitle.text}
            </div>
        </div>
    );
};

export default SubtitlePlayer;
