/**
 * Theme system type definitions for OrigCMS dynamic theme system.
 * Supports runtime theme loading from public/themes/ directory.
 */

/** Theme category types */
export type ThemeCategory = 'professional' | 'social' | 'creative' | 'minimal' | 'custom';

/** Color mode strategy */
export type ColorMode = 'light' | 'dark' | 'system';

/** Theme metadata (from meta.json / registry.json) */
export interface ThemeMeta {
  id: string;
  name: string;
  nameEn?: string;
  description?: string;
  category: ThemeCategory;
  preview: {
    primary: string;
    accent?: string;
    success?: string;
    warning?: string;
    info?: string;
    background?: string;
    surface?: string;
    /** Color scale preview values (HEX format).
     * Key format: "{family}-{level}" e.g. "primary-50", "success-600"
     * Families: primary, brand, success, warning, info, destructive
     * Levels: 50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950
     */
    [key: string]: string | undefined;
  };
  tags?: string[];
  author?: string;
  version?: string;
  thumbnail?: string;
}

/** Theme registry (from registry.json) */
export interface ThemeRegistry {
  version: number;
  defaultTheme: string;
  themes: ThemeMeta[];
}

/** Theme load result */
export interface ThemeLoadResult {
  status: 'loaded' | 'error';
  source?: 'compiled' | 'cache' | 'fetch' | 'existing';
  error?: Error;
}

/** Theme context value exposed by ThemeProvider */
export interface ThemeContextValue {
  /** Current theme ID */
  themeId: string;
  /** Current color mode setting */
  colorMode: ColorMode;
  /** Resolved mode (system resolves to light/dark) */
  resolvedMode: 'light' | 'dark';
  /** Whether dark mode is active */
  isDark: boolean;
  /** Switch theme (async, waits for CSS loading) */
  setTheme: (themeId: string) => Promise<void>;
  /** Set color mode */
  setColorMode: (mode: ColorMode) => void;
  /** Toggle between light and dark */
  toggleDark: () => void;
  /** All available themes from registry */
  themes: ThemeMeta[];
  /** Whether a theme is currently loading */
  isLoading: boolean;
  /** Last loading error */
  loadError: Error | null;
}

/** ThemeProvider component props */
export interface ThemeProviderProps {
  children: React.ReactNode;
  /** Default theme ID (fallback when no localStorage value) */
  defaultTheme?: string;
  /** Default color mode */
  defaultColorMode?: ColorMode;
  /** localStorage key prefix */
  storageKey?: string;
}
