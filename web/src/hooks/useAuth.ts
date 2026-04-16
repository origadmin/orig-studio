/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useState, useCallback, useEffect} from 'react';
import {isTokenExpired, clearAuth} from '@/lib/request';

export interface User {
    id: number;
    username: string;
    displayName: string;
    avatarUrl?: string;
    roles: string[];
}

export interface AuthState {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    isAdmin: boolean;
}

export interface UseAuthReturn extends AuthState {
    login: (token: string, refreshToken: string, user: User) => void;
    logout: () => void;
}

const TOKEN_KEY = 'origcms_token';
const USER_KEY = 'origcms_user';

/** Pure helpers used by router beforeLoad (no hooks allowed there) */
export function getStoredToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
}

export function getStoredUser(): User | null {
    try {
        const raw = localStorage.getItem(USER_KEY);
        return raw ? (JSON.parse(raw) as User) : null;
    } catch {
        return null;
    }
}

/**
 * useAuth provides authentication state and actions.
 * State is initialised from localStorage so it survives page refresh.
 */
export function useAuth(): UseAuthReturn {
    const [token, setToken] = useState<string | null>(() => {
        const storedToken = getStoredToken();
        if (storedToken && isTokenExpired()) {
            clearAuth();
            return null;
        }
        return storedToken;
    });

    const [user, setUser] = useState<User | null>(() => {
        const storedToken = getStoredToken();
        if (storedToken && isTokenExpired()) {
            return null;
        }
        return getStoredUser();
    });

    const login = useCallback((newToken: string, newRefreshToken: string, newUser: User) => {
        localStorage.setItem(TOKEN_KEY, newToken);
        localStorage.setItem('origcms_refresh_token', newRefreshToken);
        localStorage.setItem(USER_KEY, JSON.stringify(newUser));

        // 同时设置 token_expires_at 以保持一致性
        try {
            const payload = JSON.parse(atob(newToken.split('.')[1]));
            if (payload.exp) {
                localStorage.setItem("token_expires_at", String(payload.exp * 1000));
            }
        } catch {
            // 忽略解析错误
        }
        
        setToken(newToken);
        setUser(newUser);
    }, []);

    const logout = useCallback(() => {
        clearAuth();
        setToken(null);
        setUser(null);
    }, []);

    // 监听 storage 事件，接收 token 更新通知
    useEffect(() => {
        const handleStorageChange = (e: StorageEvent) => {
            if (e.key === TOKEN_KEY) {
                if (e.newValue) {
                    setToken(e.newValue);
                    // 同时也更新 user，因为刷新 token 时也会更新 user
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

    // 只在挂载时检查一次 token 状态
    useEffect(() => {
        if (token && isTokenExpired()) {
            clearAuth();
            setToken(null);
            setUser(null);
        }
    }, [token]);

    const isAdmin = user?.roles?.includes('admin') ?? false;

    return {
        user,
        token,
        isAuthenticated: !!token && !!user && !isTokenExpired(),
        isAdmin,
        login,
        logout,
    };
}
