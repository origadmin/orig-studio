/**
 * WebVTT sprite sheet parser for video preview thumbnails.
 *
 * Parses WebVTT files that describe sprite sheet thumbnails, where each CUE
 * maps a time range to a region (xywh) of a sprite sheet image.
 */

/** Single WebVTT CUE corresponding to one frame in the sprite sheet. */
export interface SpriteCue {
    /** CUE start time in seconds */
    startTime: number;
    /** CUE end time in seconds */
    endTime: number;
    /** X offset in the sprite sheet (px) */
    x: number;
    /** Y offset in the sprite sheet (px) */
    y: number;
    /** Frame width (px) */
    w: number;
    /** Frame height (px) */
    h: number;
}

/** Parsed sprite sheet VTT data. */
export interface ParsedSpriteVTT {
    /** All CUE entries, sorted by startTime ascending */
    cues: SpriteCue[];
    /** Full URL of the sprite sheet image */
    imageUrl: string;
    /** Total width of the sprite sheet (px) */
    totalWidth: number;
    /** Total height of the sprite sheet (px) */
    totalHeight: number;
}

// Regex for VTT timestamp: HH:MM:SS.mmm or MM:SS.mmm
const TIMESTAMP_REGEX = /(\d{1,2}:)?(\d{2}):(\d{2})\.(\d{3})/;

// Regex for CUE time line
const CUE_TIME_REGEX = /(\d{1,2}:\d{2}:\d{2}\.\d{3}|\d{2}:\d{2}\.\d{3})\s*-->\s*(\d{1,2}:\d{2}:\d{2}\.\d{3}|\d{2}:\d{2}\.\d{3})/;

// Regex for xywh fragment in CUE payload
const XYWH_REGEX = /^(.+)#xywh=(\d+),(\d+),(\d+),(\d+)$/;

/**
 * Parses a VTT timestamp string into seconds.
 * Supports both HH:MM:SS.mmm and MM:SS.mmm formats.
 */
export function parseVTTTime(timeStr: string): number {
    const match = timeStr.match(TIMESTAMP_REGEX);
    if (!match) return 0;

    const hours = match[1] ? parseInt(match[1].replace(':', ''), 10) : 0;
    const minutes = parseInt(match[2], 10);
    const seconds = parseInt(match[3], 10);
    const millis = parseInt(match[4], 10);

    return hours * 3600 + minutes * 60 + seconds + millis / 1000;
}

/**
 * Parses a WebVTT sprite sheet file content.
 *
 * @param vttText - Raw VTT file text
 * @param baseUrl - Base URL of the VTT file, used to resolve relative image paths
 * @returns Parsed sprite VTT data, or null if parsing fails
 */
export function parseWebVTT(vttText: string, baseUrl: string): ParsedSpriteVTT | null {
    if (!vttText || !vttText.trim()) return null;

    // Remove BOM if present
    let text = vttText;
    if (text.charCodeAt(0) === 0xFEFF) {
        text = text.slice(1);
    }

    // Validate WEBVTT header
    const firstLine = text.split('\n')[0].trim();
    if (!firstLine.startsWith('WEBVTT')) return null;

    // Split into blocks by double newline
    const blocks = text.split('\n\n');

    const cues: SpriteCue[] = [];
    let imageSrc = '';

    for (const block of blocks) {
        const lines = block.trim().split('\n');

        // Find the time line in the block
        let timeLineIdx = -1;
        for (let i = 0; i < lines.length; i++) {
            if (CUE_TIME_REGEX.test(lines[i].trim())) {
                timeLineIdx = i;
                break;
            }
        }

        if (timeLineIdx === -1) continue; // Not a CUE block (could be header, NOTE, etc.)

        const timeLine = lines[timeLineIdx].trim();
        const timeMatch = timeLine.match(CUE_TIME_REGEX);
        if (!timeMatch) continue;

        const startTime = parseVTTTime(timeMatch[1]);
        const endTime = parseVTTTime(timeMatch[2]);

        // Find the payload line (after the time line)
        let payloadLine = '';
        for (let i = timeLineIdx + 1; i < lines.length; i++) {
            const trimmed = lines[i].trim();
            if (trimmed) {
                payloadLine = trimmed;
                break;
            }
        }

        if (!payloadLine) continue;

        // Parse xywh from payload
        const xywhMatch = payloadLine.match(XYWH_REGEX);
        if (!xywhMatch) continue;

        const src = xywhMatch[1];
        const x = parseInt(xywhMatch[2], 10);
        const y = parseInt(xywhMatch[3], 10);
        const w = parseInt(xywhMatch[4], 10);
        const h = parseInt(xywhMatch[5], 10);

        if (!imageSrc && src) {
            imageSrc = src;
        }

        cues.push({startTime, endTime, x, y, w, h});
    }

    if (cues.length === 0) return null;

    // Sort cues by startTime
    cues.sort((a, b) => a.startTime - b.startTime);

    // Calculate total dimensions from all cues
    let totalWidth = 0;
    let totalHeight = 0;
    for (const cue of cues) {
        const right = cue.x + cue.w;
        const bottom = cue.y + cue.h;
        if (right > totalWidth) totalWidth = right;
        if (bottom > totalHeight) totalHeight = bottom;
    }

    // Resolve image URL
    let imageUrl: string;
    try {
        imageUrl = new URL(imageSrc, baseUrl).href;
    } catch {
        // If URL resolution fails, use as-is
        imageUrl = imageSrc;
    }

    return {
        cues,
        imageUrl,
        totalWidth,
        totalHeight,
    };
}

/**
 * Finds the CUE that contains the given time using binary search.
 * Time complexity: O(log n)
 *
 * @param cues - CUE list sorted by startTime ascending
 * @param time - Target time in seconds
 * @returns Matching CUE, or null if not found
 */
export function findCueAtTime(cues: SpriteCue[], time: number): SpriteCue | null {
    if (cues.length === 0 || time < 0) return null;

    let low = 0;
    let high = cues.length - 1;
    let result: SpriteCue | null = null;

    while (low <= high) {
        const mid = (low + high) >>> 1;
        const cue = cues[mid];

        if (cue.startTime <= time) {
            if (time < cue.endTime) {
                result = cue;
                break;
            }
            low = mid + 1;
        } else {
            high = mid - 1;
        }
    }

    return result;
}
