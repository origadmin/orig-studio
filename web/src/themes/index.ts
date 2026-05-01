/**
 * ThemeLoader - Runtime theme discovery and loading for OrigCMS.
 *
 * Loads theme CSS from public/themes/ directory via fetch(),
 * caches CSS text in localStorage for FOUC prevention.
 * Only loads themes registered in registry.json (security).
 */
import type { ThemeMeta, ThemeRegistry, ThemeLoadResult } from './types';

const REGISTRY_URL = '/themes/registry.json';
const THEME_CSS_URL = (id: string) => `/themes/${id}/index.css`;

const THEME_ID_PATTERN = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/;
const MAX_THEME_ID_LENGTH = 64;

function isValidThemeId(id: string): boolean {
  if (!id || id.length > MAX_THEME_ID_LENGTH) return false;
  return THEME_ID_PATTERN.test(id);
}

/** localStorage key names */
const STORAGE_KEYS = {
  theme: 'origcms-theme',
  colorMode: 'origcms-color-mode',
  themeCss: (id: string) => `origcms-theme-css-${id}`,
  registryCache: 'origcms-theme-registry',
} as const;

/** CSS injection style tag IDs */
const STYLE_IDS = {
  cache: 'theme-cache',
  loaded: (id: string) => `theme-css-${id}`,
} as const;

export { STORAGE_KEYS, STYLE_IDS };

/**
 * ThemeLoader - responsible for runtime theme discovery and CSS loading.
 *
 * Flow:
 * 1. loadRegistry() - fetch registry.json to discover available themes
 * 2. loadTheme(id) - load theme CSS (from cache, localStorage, or fetch)
 * 3. unloadTheme(id) - remove theme CSS from DOM
 * 4. clearCache() - clear localStorage CSS cache
 */
class ThemeLoader {
  private registry: ThemeRegistry | null = null;
  private loadedThemes = new Set<string>();

  private isThemeInRegistry(themeId: string): boolean {
    if (!this.registry?.themes) return false;
    return this.registry.themes.some((t) => t.id === themeId);
  }

  /**
   * Load the theme registry.
   * Priority: in-memory cache > fetch > localStorage cache > minimal fallback
   */
  async loadRegistry(): Promise<ThemeRegistry> {
    if (this.registry) return this.registry;

    try {
      const response = await fetch(REGISTRY_URL);
      if (!response.ok) throw new Error(`Registry fetch failed: ${response.status}`);
      const registry: ThemeRegistry = await response.json();

      // Validate registry format
      if (!registry.themes || !Array.isArray(registry.themes)) {
        throw new Error('Invalid registry format');
      }

      // Cache to localStorage for offline fallback
      try {
        localStorage.setItem(STORAGE_KEYS.registryCache, JSON.stringify(registry));
      } catch {
        // localStorage may be full or disabled
      }

      this.registry = registry;
      return registry;
    } catch (fetchError) {
      // Fetch failed, try localStorage cache
      try {
        const cached = localStorage.getItem(STORAGE_KEYS.registryCache);
        if (cached) {
          this.registry = JSON.parse(cached);
          return this.registry!;
        }
      } catch {
        // localStorage unavailable
      }

      // All failed, return minimal registry with only default theme
      this.registry = {
        version: 1,
        defaultTheme: 'default',
        themes: [
          {
            id: 'default',
            name: '默认蓝',
            nameEn: 'Default Blue',
            category: 'professional',
            preview: { primary: '#0B6CFA' },
          },
        ],
      };
      return this.registry;
    }
  }

  /**
   * Load theme CSS.
   * Returns load result with source indicator.
   */
  async loadTheme(themeId: string): Promise<ThemeLoadResult> {
    if (!isValidThemeId(themeId)) {
      return {
        status: 'error',
        error: new Error(`Invalid theme ID: "${themeId}". Only lowercase alphanumeric and hyphens allowed.`),
      };
    }

    if (themeId === 'default') {
      this.loadedThemes.add('default');
      return { status: 'loaded', source: 'compiled' };
    }

    if (this.registry && !this.isThemeInRegistry(themeId)) {
      return {
        status: 'error',
        error: new Error(`Theme "${themeId}" is not in the registry. Only registered themes can be loaded.`),
      };
    }

    const existingStyle = document.getElementById(STYLE_IDS.loaded(themeId));
    if (existingStyle) {
      this.loadedThemes.add(themeId);
      return { status: 'loaded', source: 'existing' };
    }

    try {
      const cachedCss = localStorage.getItem(STORAGE_KEYS.themeCss(themeId));
      if (cachedCss) {
        this.injectStyle(themeId, cachedCss);
        this.loadedThemes.add(themeId);
        return { status: 'loaded', source: 'cache' };
      }
    } catch {
      // localStorage unavailable
    }

    try {
      const response = await fetch(THEME_CSS_URL(themeId));
      if (!response.ok) throw new Error(`Theme CSS fetch failed: ${response.status}`);
      const cssText = await response.text();

      this.injectStyle(themeId, cssText);

      try {
        localStorage.setItem(STORAGE_KEYS.themeCss(themeId), cssText);
      } catch {
        // localStorage may be full or disabled
      }

      this.loadedThemes.add(themeId);
      return { status: 'loaded', source: 'fetch' };
    } catch (error) {
      console.error(`[ThemeLoader] Failed to load theme "${themeId}":`, error);
      return { status: 'error', error: error as Error };
    }
  }

  /**
   * Inject theme CSS into the DOM as a <style> tag.
   * Uses textContent (not innerHTML) for XSS prevention.
   */
  private injectStyle(themeId: string, cssText: string): void {
    // Remove the inline script cache tag if present
    const cacheStyle = document.getElementById(STYLE_IDS.cache);
    if (cacheStyle) cacheStyle.remove();

    // Create new style tag
    const style = document.createElement('style');
    style.id = STYLE_IDS.loaded(themeId);
    style.setAttribute('data-theme-id', themeId);
    style.textContent = cssText;
    document.head.appendChild(style);
  }

  /**
   * Unload theme CSS from DOM (used when switching away from a theme).
   * Never unloads the default theme (compiled into index.css).
   */
  unloadTheme(themeId: string): void {
    if (themeId === 'default') return;
    const style = document.getElementById(STYLE_IDS.loaded(themeId));
    if (style) style.remove();
    this.loadedThemes.delete(themeId);
  }

  /**
   * Clear localStorage CSS cache.
   * @param themeId - If provided, clear only that theme's cache; otherwise clear all.
   */
  clearCache(themeId?: string): void {
    if (themeId) {
      localStorage.removeItem(STORAGE_KEYS.themeCss(themeId));
    } else {
      const keysToRemove: string[] = [];
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key?.startsWith('origcms-theme-css-')) {
          keysToRemove.push(key);
        }
      }
      keysToRemove.forEach((key) => localStorage.removeItem(key));
    }
  }

  /**
   * Check if a theme's CSS is already loaded.
   */
  isLoaded(themeId: string): boolean {
    return this.loadedThemes.has(themeId);
  }
}

/** Singleton theme loader instance */
export const themeLoader = new ThemeLoader();

export { ThemeProvider } from './ThemeProvider';
export { useTheme } from './useTheme';
export { ThemeSwitcher } from './ThemeSwitcher';
