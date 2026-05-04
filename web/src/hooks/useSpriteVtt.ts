/**
 * Hook for loading and parsing sprite sheet WebVTT files.
 *
 * Uses TanStack Query for caching - same vttUrl only triggers one network request.
 * staleTime is set to Infinity since VTT content does not change.
 */

import {useQuery} from '@tanstack/react-query';
import {parseWebVTT, ParsedSpriteVTT} from '@/lib/parseWebVTT';

/** Result of the useSpriteVtt hook. */
interface UseSpriteVttResult {
    /** Parsed VTT data, null if not loaded or failed */
    parsed: ParsedSpriteVTT | null;
    /** Whether the VTT is currently loading */
    loading: boolean;
    /** Error if VTT loading/parsing failed */
    error: Error | null;
}

/**
 * Loads and parses a sprite sheet WebVTT file.
 *
 * - Uses TanStack Query cache: same vttUrl only requests once
 * - staleTime: Infinity (VTT content does not change)
 * - enabled: only when vttUrl is non-empty
 *
 * @param vttUrl - WebVTT file URL, or null/undefined to disable
 */
export function useSpriteVtt(vttUrl: string | null | undefined): UseSpriteVttResult {
    const {data, isLoading, error} = useQuery({
        queryKey: ['sprite-vtt', vttUrl],
        queryFn: async (): Promise<ParsedSpriteVTT | null> => {
            if (!vttUrl) return null;

            const response = await fetch(vttUrl);
            if (!response.ok) {
                throw new Error(`Failed to fetch VTT: ${response.status}`);
            }

            const text = await response.text();
            const basePath = vttUrl.substring(0, vttUrl.lastIndexOf('/') + 1);
            const baseUrl = new URL(basePath, window.location.origin).href;
            const result = parseWebVTT(text, baseUrl);

            if (!result) {
                throw new Error('Failed to parse VTT content');
            }

            return result;
        },
        enabled: !!vttUrl,
        staleTime: Infinity,
        retry: 1,
    });

    return {
        parsed: data ?? null,
        loading: isLoading,
        error: error as Error | null,
    };
}
