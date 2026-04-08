import {useState, useEffect, useCallback} from 'react';

interface PlayerSettings {
    volume: number;
    isMuted: boolean;
    playbackRate: number;
    quality?: string;
}

const STORAGE_KEY = 'origcms_player_settings';

const DEFAULT_SETTINGS: PlayerSettings = {
    volume: 1,
    isMuted: false,
    playbackRate: 1,
};

export function usePlayerSettings() {
    const [settings, setSettings] = useState<PlayerSettings>(() => {
        try {
            const stored = localStorage.getItem(STORAGE_KEY);
            if (stored) {
                return {...DEFAULT_SETTINGS, ...JSON.parse(stored)};
            }
        } catch {
            // 使用默认值
        }
        return DEFAULT_SETTINGS;
    });

    // 保存到 localStorage
    useEffect(() => {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
        } catch {
            // 忽略存储错误
        }
    }, [settings]);

    const setVolume = useCallback((volume: number) => {
        setSettings(prev => ({...prev, volume, isMuted: volume === 0 ? true : prev.isMuted}));
    }, []);

    const setIsMuted = useCallback((isMuted: boolean) => {
        setSettings(prev => ({...prev, isMuted}));
    }, []);

    const setPlaybackRate = useCallback((rate: number) => {
        setSettings(prev => ({...prev, playbackRate: rate}));
    }, []);

    const setQuality = useCallback((quality: string) => {
        setSettings(prev => ({...prev, quality}));
    }, []);

    return {
        ...settings,
        setVolume,
        setIsMuted,
        setPlaybackRate,
        setQuality,
    };
}
