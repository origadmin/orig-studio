/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Backward-compatible re-export from contexts/auth.
// Existing import paths (`@/hooks/useAuth`) continue to work.
// The useAuth hook now reads from AuthContext (single state instance)
// instead of creating independent useState instances.
import type {User} from '@/contexts/auth/types';

export {useAuth} from '@/contexts/auth';
export type {User, AuthState, AuthContextValue as UseAuthReturn} from '@/contexts/auth/types';

// Preserve localStorage read functions for backward compatibility.
// All route guards have been migrated to context.auth in R3.
// These functions will be removed in a future version.
const TOKEN_KEY = 'origcms_token';
const USER_KEY = 'origcms_user';

/**
 * @deprecated Use useAuth() instead. All route guards now read from context.auth.
 * This function will be removed in a future version.
 */
export function getStoredToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
}

/**
 * @deprecated Use useAuth() instead. All route guards now read from context.auth.
 * This function will be removed in a future version.
 */
export function getStoredUser(): User | null {
    try {
        const raw = localStorage.getItem(USER_KEY);
        return raw ? (JSON.parse(raw) as User) : null;
    } catch {
        return null;
    }
}
