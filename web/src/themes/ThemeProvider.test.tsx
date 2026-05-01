/**
 * Unit tests for useTheme hook and ThemeProvider
 *
 * Covers: context value, theme switching, color mode, dark mode toggle
 */
import { describe, it, expect, beforeEach, jest } from '@jest/globals';
import { renderHook, act } from '@testing-library/react';
import { ThemeProvider } from './ThemeProvider';
import { useTheme } from './useTheme';

// Mock fetch for registry loading
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

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

// Mock document.head.appendChild
jest.spyOn(document.head, 'appendChild').mockImplementation((node) => node);
jest.spyOn(document.head, 'removeChild').mockImplementation((node) => node);

const mockRegistry = {
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

beforeEach(() => {
  jest.clearAllMocks();
  localStorageMock.clear();
  document.documentElement.removeAttribute('data-theme');
  document.documentElement.classList.remove('dark');

  // Default: registry fetch succeeds
  mockFetch.mockImplementation((url: string) => {
    if (url === '/themes/registry.json') {
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve(mockRegistry),
      });
    }
    if (url.includes('/index.css')) {
      return Promise.resolve({
        ok: true,
        text: () => Promise.resolve('[data-theme="feishu-blue"] { --primary: 220 90% 56%; }'),
      });
    }
    return Promise.reject(new Error('Unknown URL'));
  });
});

describe('useTheme', () => {
  it('should throw error when used outside ThemeProvider', () => {
    // Suppress console.error for this test
    const spy = jest.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      renderHook(() => useTheme());
    }).toThrow('useTheme must be used within a ThemeProvider');

    spy.mockRestore();
  });

  it('should provide default theme context values', async () => {
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="system">
          {children}
        </ThemeProvider>
      ),
    });

    // Wait for async initialization
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(result.current.themeId).toBe('default');
    expect(result.current.colorMode).toBe('system');
    expect(result.current.themes).toHaveLength(2);
    expect(result.current.isLoading).toBe(false);
    expect(result.current.loadError).toBeNull();
  });

  it('should resolve system mode to light when system prefers light', async () => {
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="system">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    // matchMedia mock returns matches: false (light mode)
    expect(result.current.resolvedMode).toBe('light');
    expect(result.current.isDark).toBe(false);
  });

  it('should set color mode and persist to localStorage', async () => {
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="system">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    act(() => {
      result.current.setColorMode('dark');
    });

    expect(result.current.colorMode).toBe('dark');
    expect(result.current.isDark).toBe(true);
    expect(localStorageMock.setItem).toHaveBeenCalledWith(
      expect.stringContaining('color-mode'),
      'dark'
    );
  });

  it('should toggle dark mode', async () => {
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="light">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(result.current.isDark).toBe(false);

    act(() => {
      result.current.toggleDark();
    });

    expect(result.current.isDark).toBe(true);

    act(() => {
      result.current.toggleDark();
    });

    expect(result.current.isDark).toBe(false);
  });

  it('should switch theme and persist to localStorage', async () => {
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="light">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    await act(async () => {
      await result.current.setTheme('feishu-blue');
    });

    expect(result.current.themeId).toBe('feishu-blue');
    expect(localStorageMock.setItem).toHaveBeenCalledWith(
      expect.stringContaining('theme'),
      'feishu-blue'
    );
  });

  it('should apply data-theme attribute to document element', async () => {
    renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="light">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(document.documentElement.getAttribute('data-theme')).toBe('default');
  });

  it('should toggle dark class on document element', async () => {
    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="light">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(document.documentElement.classList.contains('dark')).toBe(false);

    act(() => {
      result.current.setColorMode('dark');
    });

    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('should restore saved theme from localStorage', async () => {
    localStorageMock.getItem.mockImplementation((key: string) => {
      if (key === 'origcms-theme') return 'feishu-blue';
      if (key === 'origcms-color-mode') return 'dark';
      return null;
    });

    const { result } = renderHook(() => useTheme(), {
      wrapper: ({ children }) => (
        <ThemeProvider defaultTheme="default" defaultColorMode="system">
          {children}
        </ThemeProvider>
      ),
    });

    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100));
    });

    expect(result.current.themeId).toBe('feishu-blue');
    expect(result.current.colorMode).toBe('dark');
  });
});
