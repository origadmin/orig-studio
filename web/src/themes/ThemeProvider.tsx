/**
 * ThemeProvider - React context provider for the OrigCMS dynamic theme system.
 *
 * Manages theme ID, color mode (light/dark/system), and provides
 * theme switching functionality with async CSS loading support.
 */
import React, { createContext, useEffect, useState, useCallback, useContext } from 'react';
import type { ThemeContextValue, ThemeProviderProps, ColorMode, ThemeMeta } from './types';
import { themeLoader } from './index';

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

const STORAGE_THEME_KEY = 'origcms-theme';
const STORAGE_MODE_KEY = 'origcms-color-mode';

function getSystemPreference(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function resolveMode(mode: ColorMode): 'light' | 'dark' {
  return mode === 'system' ? getSystemPreference() : mode;
}

export function ThemeProvider({
  children,
  defaultTheme = 'default',
  defaultColorMode = 'system',
  storageKey = 'origcms',
}: ThemeProviderProps) {
  const [themeId, setThemeIdState] = useState<string>(() => {
    try {
      return localStorage.getItem(`${storageKey}-theme`) || defaultTheme;
    } catch {
      return defaultTheme;
    }
  });

  const [colorMode, setColorModeState] = useState<ColorMode>(() => {
    try {
      const saved = localStorage.getItem(`${storageKey}-color-mode`) as ColorMode;
      return saved || defaultColorMode;
    } catch {
      return defaultColorMode;
    }
  });

  const [themes, setThemes] = useState<ThemeMeta[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<Error | null>(null);

  const resolvedMode = resolveMode(colorMode);
  const isDark = resolvedMode === 'dark';

  // Load registry and current theme CSS on mount
  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        // Clean up stale default theme CSS cache from localStorage
        try {
          localStorage.removeItem(`${storageKey}-theme-css-default`);
        } catch {
          // localStorage unavailable
        }
        const staleDefaultStyle = document.getElementById('theme-css-default');
        if (staleDefaultStyle) staleDefaultStyle.remove();

        // 1. Load registry
        const registry = await themeLoader.loadRegistry();
        if (cancelled) return;
        setThemes(registry.themes);

        // 2. Load current theme CSS (if not default)
        const savedTheme = localStorage.getItem(`${storageKey}-theme`) || defaultTheme;
        if (savedTheme !== 'default') {
          const result = await themeLoader.loadTheme(savedTheme);
          if (cancelled) return;
          if (result.status === 'error') {
            console.warn(
              `[ThemeProvider] Failed to load theme "${savedTheme}", falling back to default`
            );
            setThemeIdState('default');
          }
        }

        // 3. Remove inline script cache tag (FOUC prevention handoff)
        const cacheStyle = document.getElementById('theme-cache');
        if (cacheStyle) cacheStyle.remove();

        setIsLoading(false);
      } catch (error) {
        if (cancelled) return;
        setLoadError(error as Error);
        setIsLoading(false);
      }
    }

    init();
    return () => {
      cancelled = true;
    };
  }, [defaultTheme, storageKey]);

  // Apply theme to DOM
  useEffect(() => {
    const root = document.documentElement;
    root.setAttribute('data-theme', themeId);
    root.classList.toggle('dark', isDark);
  }, [themeId, isDark]);

  // Listen for system color scheme changes when in 'system' mode
  useEffect(() => {
    if (colorMode !== 'system') return;
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => {
      // Trigger re-render by updating state (resolvedMode will recalculate)
      setColorModeState('system');
    };
    mq.addEventListener('change', handler);
    return () => mq.removeEventListener('change', handler);
  }, [colorMode]);

  // Switch theme (async - may need to fetch CSS)
  const setTheme = useCallback(
    async (id: string) => {
      if (!themeLoader.isLoaded(id)) {
        setIsLoading(true);
        const result = await themeLoader.loadTheme(id);
        setIsLoading(false);

        if (result.status === 'error') {
          setLoadError(result.error || new Error(`Failed to load theme "${id}"`));
          return;
        }
      }

      if (themeId !== 'default' && themeId !== id) {
        themeLoader.unloadTheme(themeId);
      }

      setThemeIdState(id);
      setLoadError(null);
      try {
        localStorage.setItem(`${storageKey}-theme`, id);
      } catch {
        // localStorage unavailable
      }
    },
    [themeId, storageKey]
  );

  const setColorMode = useCallback(
    (mode: ColorMode) => {
      setColorModeState(mode);
      try {
        localStorage.setItem(`${storageKey}-color-mode`, mode);
      } catch {
        // localStorage unavailable
      }
    },
    [storageKey]
  );

  const toggleDark = useCallback(() => {
    const next: ColorMode = isDark ? 'light' : 'dark';
    setColorMode(next);
  }, [isDark, setColorMode]);

  const value: ThemeContextValue = {
    themeId,
    colorMode,
    resolvedMode,
    isDark,
    setTheme,
    setColorMode,
    toggleDark,
    themes,
    isLoading,
    loadError,
  };

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export { ThemeContext };
