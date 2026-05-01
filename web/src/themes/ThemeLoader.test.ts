/**
 * Unit tests for ThemeLoader
 *
 * Covers: loadRegistry, loadTheme, unloadTheme, clearCache, isLoaded
 */
import { themeLoader, STORAGE_KEYS, STYLE_IDS } from './index';
import type { ThemeRegistry } from './types';

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: jest.fn((key: string) => store[key] ?? null),
    setItem: jest.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: jest.fn((key: string) => {
      delete store[key];
    }),
    clear: jest.fn(() => {
      store = {};
    }),
    get length() {
      return Object.keys(store).length;
    },
    key: jest.fn((index: number) => Object.keys(store)[index] ?? null),
  };
})();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock document.head.appendChild
const appendChildSpy = jest.spyOn(document.head, 'appendChild').mockImplementation((node) => node);
const removeChildSpy = jest.spyOn(document.head, 'removeChild').mockImplementation((node) => node);

// Helper to create a mock registry
function createMockRegistry(): ThemeRegistry {
  return {
    version: 1,
    defaultTheme: 'default',
    themes: [
      {
        id: 'default',
        name: 'Default Blue',
        nameEn: 'Default Blue',
        category: 'professional',
        preview: { primary: '#0B6CFA' },
      },
      {
        id: 'feishu-blue',
        name: 'Feishu Blue',
        nameEn: 'Feishu Blue',
        category: 'professional',
        preview: { primary: '#3370ff' },
      },
    ],
  };
}

beforeEach(() => {
  jest.clearAllMocks();
  mockFetch.mockReset();
  localStorageMock.clear();
  // Reset the singleton's internal state
  (themeLoader as any).registry = null;
  (themeLoader as any).loadedThemes = new Set();
  // Remove any style elements that might exist
  document.querySelectorAll('style[data-theme-id]').forEach((el) => el.remove());
  const cacheStyle = document.getElementById(STYLE_IDS.cache);
  if (cacheStyle) cacheStyle.remove();
});

describe('ThemeLoader', () => {
  describe('loadRegistry', () => {
    it('should fetch and return registry from server', async () => {
      const mockRegistry = createMockRegistry();
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });

      const result = await themeLoader.loadRegistry();

      expect(mockFetch).toHaveBeenCalledWith('/themes/registry.json');
      expect(result.themes).toHaveLength(2);
      expect(result.defaultTheme).toBe('default');
    });

    it('should cache registry to localStorage after fetch', async () => {
      const mockRegistry = createMockRegistry();
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });

      await themeLoader.loadRegistry();

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        STORAGE_KEYS.registryCache,
        JSON.stringify(mockRegistry)
      );
    });

    it('should return cached registry on second call', async () => {
      const mockRegistry = createMockRegistry();
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });

      await themeLoader.loadRegistry();
      const result = await themeLoader.loadRegistry();

      // fetch should only be called once
      expect(mockFetch).toHaveBeenCalledTimes(1);
      expect(result.themes).toHaveLength(2);
    });

    it('should fall back to localStorage cache when fetch fails', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));
      localStorageMock.getItem.mockImplementationOnce(
        () => JSON.stringify(createMockRegistry())
      );

      const result = await themeLoader.loadRegistry();

      expect(result.themes).toHaveLength(2);
    });

    it('should return minimal registry when all sources fail', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));
      // localStorage returns null

      const result = await themeLoader.loadRegistry();

      expect(result.themes).toHaveLength(1);
      expect(result.themes[0].id).toBe('default');
      expect(result.defaultTheme).toBe('default');
    });

    it('should reject invalid registry format', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ version: 1 }), // missing themes array
      });

      localStorageMock.getItem.mockReturnValueOnce(null);

      const result = await themeLoader.loadRegistry();

      // Should fall back to minimal registry
      expect(result.themes).toHaveLength(1);
      expect(result.themes[0].id).toBe('default');
    });
  });

  describe('loadTheme', () => {
    it('should return compiled source for default theme without fetching', async () => {
      const result = await themeLoader.loadTheme('default');

      expect(result.status).toBe('loaded');
      expect(result.source).toBe('compiled');
      expect(mockFetch).not.toHaveBeenCalledWith(expect.stringContaining('/themes/default/'));
    });

    it('should fetch theme CSS from server when not cached', async () => {
      const cssText = '[data-theme="feishu-blue"] { --primary: 220 90% 56%; }';
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(cssText),
      });

      const result = await themeLoader.loadTheme('feishu-blue');

      expect(result.status).toBe('loaded');
      expect(result.source).toBe('fetch');
      expect(mockFetch).toHaveBeenCalledWith('/themes/feishu-blue/index.css');
    });

    it('should load theme CSS from localStorage cache', async () => {
      const cachedCss = '[data-theme="feishu-blue"] { --primary: 220 90% 56%; }';
      localStorageMock.getItem.mockImplementationOnce(() => cachedCss);

      const result = await themeLoader.loadTheme('feishu-blue');

      expect(result.status).toBe('loaded');
      expect(result.source).toBe('cache');
      // Should not fetch from server
      expect(mockFetch).not.toHaveBeenCalledWith('/themes/feishu-blue/index.css');
    });

    it('should return existing source when theme CSS is already in DOM', async () => {
      // First load
      const cssText = '[data-theme="feishu-blue"] { --primary: 220 90% 56%; }';
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(cssText),
      });
      await themeLoader.loadTheme('feishu-blue');

      // Second load - should find existing style element
      const result = await themeLoader.loadTheme('feishu-blue');

      expect(result.status).toBe('loaded');
      expect(result.source).toBe('existing');
    });

    it('should return error when fetch fails', async () => {
      // Load registry first so whitelist check passes
      const mockRegistry = createMockRegistry();
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });
      await themeLoader.loadRegistry();

      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        text: () => Promise.resolve('Not Found'),
      });

      const result = await themeLoader.loadTheme('nonexistent');

      expect(result.status).toBe('error');
      expect(result.error).toBeDefined();
    });

    it('should cache fetched CSS to localStorage', async () => {
      // Load registry first so whitelist check passes
      const mockRegistry = createMockRegistry();
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });
      await themeLoader.loadRegistry();

      const cssText = '[data-theme="feishu-blue"] { --primary: 220 90% 56%; }';
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(cssText),
      });

      await themeLoader.loadTheme('feishu-blue');

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        STORAGE_KEYS.themeCss('feishu-blue'),
        cssText
      );
    });

    it('should inject CSS via textContent (not innerHTML) for XSS prevention', async () => {
      // Load registry with test-theme included
      const testRegistry = createMockRegistry();
      testRegistry.themes.push({
        id: 'test-theme',
        name: 'Test Theme',
        nameEn: 'Test Theme',
        category: 'professional',
        preview: { primary: '#000000' },
      });
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(testRegistry),
      });
      await themeLoader.loadRegistry();

      const cssText = '[data-theme="test-theme"] { --primary: 0 0% 0%; }';
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(cssText),
      });

      // Mock createElement to capture the created style element
      const createdElements: HTMLStyleElement[] = [];
      const origCreateElement = document.createElement.bind(document);
      jest.spyOn(document, 'createElement').mockImplementation((tag: string) => {
        const el = origCreateElement(tag);
        if (tag === 'style') createdElements.push(el as HTMLStyleElement);
        return el;
      });

      await themeLoader.loadTheme('test-theme');

      expect(createdElements.length).toBeGreaterThan(0);
      const styleEl = createdElements[0];
      expect(styleEl.textContent).toBe(cssText);
      expect(styleEl.getAttribute('data-theme-id')).toBe('test-theme');

      (document.createElement as jest.Mock).mockRestore();
    });

    it('should reject invalid theme IDs (path traversal prevention)', async () => {
      const maliciousIds = ['../etc/passwd', '..%2F..%2Fetc', 'theme/../../etc', 'THEME', 'a b c', ''];
      for (const id of maliciousIds) {
        const result = await themeLoader.loadTheme(id);
        expect(result.status).toBe('error');
        if (id) {
          expect(result.error?.message).toContain('Invalid theme ID');
        }
      }
    });

    it('should reject theme not in registry (whitelist enforcement)', async () => {
      const mockRegistry = createMockRegistry();
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });
      await themeLoader.loadRegistry();

      const result = await themeLoader.loadTheme('unregistered-theme');
      expect(result.status).toBe('error');
      expect(result.error?.message).toContain('not in the registry');
    });

    it('should allow loading theme when registry is not yet loaded', async () => {
      // When registry is null, skip whitelist check (registry will load later)
      const cssText = '[data-theme="feishu-blue"] { --primary: 220 90% 56%; }';
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(cssText),
      });

      const result = await themeLoader.loadTheme('feishu-blue');
      expect(result.status).toBe('loaded');
    });
  });

  describe('unloadTheme', () => {
    it('should not unload default theme', () => {
      expect(() => themeLoader.unloadTheme('default')).not.toThrow();
    });

    it('should remove style element for non-default theme', async () => {
      const cssText = '[data-theme="feishu-blue"] { --primary: 220 90% 56%; }';
      mockFetch.mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve(cssText),
      });
      await themeLoader.loadTheme('feishu-blue');

      themeLoader.unloadTheme('feishu-blue');

      expect(themeLoader.isLoaded('feishu-blue')).toBe(false);
    });
  });

  describe('clearCache', () => {
    it('should clear specific theme cache', () => {
      themeLoader.clearCache('feishu-blue');

      expect(localStorageMock.removeItem).toHaveBeenCalledWith(
        STORAGE_KEYS.themeCss('feishu-blue')
      );
    });

    it('should clear all theme CSS caches when no themeId provided', () => {
      // Directly set items in the mock's internal store
      localStorageMock.setItem('origcms-theme-css-feishu-blue', 'some css');
      localStorageMock.setItem('origcms-theme-css-stripe-indigo', 'some css');
      localStorageMock.setItem('other-key', 'other value');

      // Reset mocks to use actual store behavior for iteration
      localStorageMock.removeItem.mockClear();

      themeLoader.clearCache();

      // Should have called removeItem for theme CSS keys
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('origcms-theme-css-feishu-blue');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('origcms-theme-css-stripe-indigo');
      // Should NOT remove non-theme keys
      expect(localStorageMock.removeItem).not.toHaveBeenCalledWith('other-key');
    });
  });

  describe('isLoaded', () => {
    it('should return false for unloaded theme', () => {
      expect(themeLoader.isLoaded('feishu-blue')).toBe(false);
    });

    it('should return true after loading', async () => {
      await themeLoader.loadTheme('default');
      expect(themeLoader.isLoaded('default')).toBe(true);
    });
  });
});
