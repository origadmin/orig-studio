import {useState, useCallback, useEffect, useRef} from 'react';
import {AuthContext} from './AuthContext';
import {
    isTokenExpired,
    clearAuth,
    registerAuthCallback,
    attemptRefresh,
    hasRefreshToken,
} from '@/lib/request';
import {Spinner} from '@/components/ui/spinner';
import type {User, AuthContextValue, AuthProviderProps} from './types';

// Shared localStorage keys (must match request.ts)
const TOKEN_KEY = 'origstudio_token';
const USER_KEY = 'origstudio_user';
const REFRESH_TOKEN_KEY = 'origstudio_refresh_token';
const EXPIRES_KEY = 'token_expires_at';

/** Read token from localStorage */
function getStoredToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
}

/** Read user from localStorage */
function getStoredUser(): User | null {
    try {
        const raw = localStorage.getItem(USER_KEY);
        return raw ? (JSON.parse(raw) as User) : null;
    } catch {
        return null;
    }
}

/**
 * AuthProvider - single source of truth for authentication state.
 *
 * Wraps the application and provides AuthContext to all children.
 * Solves the multi-instance state split problem (P1) by holding
 * a single React state instance that both page components and
 * route guards read from.
 */
export function AuthProvider({children}: AuthProviderProps) {
    const [token, setToken] = useState<string | null>(null);
    const [user, setUser] = useState<User | null>(null);
    const [isRefreshing, setIsRefreshing] = useState(false);
    const [isInitialized, setIsInitialized] = useState(false);
    const refreshingRef = useRef(false);

    // Initialize: restore state from localStorage
    useEffect(() => {
        const storedToken = getStoredToken();
        if (!storedToken) {
            setIsInitialized(true);
            return;
        }
        if (!isTokenExpired()) {
            setToken(storedToken);
            setUser(getStoredUser());
            setIsInitialized(true);
            return;
        }
        if (hasRefreshToken()) {
            attemptRefresh().then(success => {
                if (success) {
                    setToken(getStoredToken());
                    setUser(getStoredUser());
                }
                setIsInitialized(true);
            });
            return;
        }
        clearAuth();
        setIsInitialized(true);
    }, []);

    // Scheduled token refresh: refresh 60 seconds before expiration
    useEffect(() => {
        if (!token) return;
        if (isTokenExpired()) {
            // Already expired, attempt refresh
            if (refreshingRef.current) return;
            refreshingRef.current = true;
            attemptRefresh().then(success => {
                refreshingRef.current = false;
                if (success) {
                    setToken(getStoredToken());
                    setUser(getStoredUser());
                } else {
                    setToken(null);
                    setUser(null);
                }
            });
            return;
        }

        // Schedule refresh 60 seconds before expiration
        let timer: ReturnType<typeof setTimeout> | undefined;
        try {
            const payload = JSON.parse(atob(token.split('.')[1]));
            if (payload.exp) {
                const refreshAt = (payload.exp - 60) * 1000;
                const delay = refreshAt - Date.now();
                if (delay > 0) {
                    timer = setTimeout(() => {
                        if (refreshingRef.current) return;
                        refreshingRef.current = true;
                        attemptRefresh().then(success => {
                            refreshingRef.current = false;
                            if (success) {
                                setToken(getStoredToken());
                                setUser(getStoredUser());
                            } else {
                                setToken(null);
                                setUser(null);
                            }
                        });
                    }, delay);
                }
            }
        } catch {
            // Ignore JWT parse errors
        }
        return () => {
            if (timer) clearTimeout(timer);
        };
    }, [token]);

    // Cross-tab synchronization via native StorageEvent
    // (only fires when a different tab modifies localStorage)
    useEffect(() => {
        const handleStorageChange = (e: StorageEvent) => {
            if (e.key === TOKEN_KEY) {
                if (e.newValue) {
                    setToken(e.newValue);
                    setUser(getStoredUser());
                } else {
                    setToken(null);
                    setUser(null);
                }
            } else if (e.key === USER_KEY) {
                setUser(getStoredUser());
            }
        };
        window.addEventListener('storage', handleStorageChange);
        return () => window.removeEventListener('storage', handleStorageChange);
    }, []);

    // Register callback: let request.ts setAuth/clearAuth notify AuthProvider
    useEffect(() => {
        registerAuthCallback((newToken, newUser) => {
            if (newToken) {
                setToken(newToken);
                setUser(newUser);
            } else {
                setToken(null);
                setUser(null);
            }
        });
        return () => registerAuthCallback(null);
    }, []);

    const login = useCallback((newToken: string, newRefreshToken: string, newUser: User) => {
        localStorage.setItem(TOKEN_KEY, newToken);
        localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken);
        localStorage.setItem(USER_KEY, JSON.stringify(newUser));
        try {
            const payload = JSON.parse(atob(newToken.split('.')[1]));
            if (payload.exp) {
                localStorage.setItem(EXPIRES_KEY, String(payload.exp * 1000));
            }
        } catch {
            // Ignore JWT parse errors
        }
        setToken(newToken);
        setUser(newUser);
    }, []);

    const logout = useCallback(() => {
        clearAuth();
        setToken(null);
        setUser(null);
    }, []);

    const refreshAuthFn = useCallback(async (): Promise<boolean> => {
        if (refreshingRef.current) return false;
        refreshingRef.current = true;
        setIsRefreshing(true);
        try {
            const success = await attemptRefresh();
            if (success) {
                setToken(getStoredToken());
                setUser(getStoredUser());
            } else {
                setToken(null);
                setUser(null);
            }
            return success;
        } finally {
            refreshingRef.current = false;
            setIsRefreshing(false);
        }
    }, []);

    const isAuthenticated = !!token && !!user && !isTokenExpired();
    const isAdmin = user?.roles?.includes('admin') ?? false;

    const value: AuthContextValue = {
        token, user, isAuthenticated, isAdmin,
        isRefreshing, isInitialized,
        login, logout, refreshAuth: refreshAuthFn,
    };

    // Show loading spinner until initialization completes
    if (!isInitialized) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <Spinner/>
            </div>
        );
    }

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
