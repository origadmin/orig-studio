/**
 * useTheme - Custom hook to access the OrigCMS theme context.
 *
 * Must be used within a ThemeProvider.
 * Provides theme ID, color mode, switching functions, and theme list.
 */
import { useContext } from 'react';
import { ThemeContext } from './ThemeProvider';
import type { ThemeContextValue } from './types';

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return ctx;
}
