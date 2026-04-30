import {createContext, useContext} from 'react';
import type {AuthContextValue} from './types';

/**
 * Authentication context.
 *
 * Default value is undefined to detect usage outside of AuthProvider.
 * Using undefined (instead of an empty object) ensures a clear error
 * when useAuth is called without an AuthProvider ancestor.
 */
export const AuthContext = createContext<AuthContextValue | undefined>(undefined);

/**
 * useAuth hook - consume authentication state from AuthProvider.
 *
 * Must be called within an AuthProvider. Returns the same AuthContextValue
 * instance regardless of which component calls it, solving the multi-instance
 * state split problem (P1).
 *
 * Backward compatible: return type includes all original fields
 * (user, token, isAuthenticated, isAdmin, login, logout).
 * New fields: isRefreshing, isInitialized, refreshAuth.
 */
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
