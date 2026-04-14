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

    // Periodically check token expiry
    useEffect(() => {
        const checkToken = async () => {
            if (token && isTokenExpired()) {
                // 尝试使用refresh token刷新
                try {
                    const refreshToken = localStorage.getItem('origcms_refresh_token');
                    if (refreshToken) {
                        const response = await fetch('/api/v1/auth/refresh', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                            },
                            body: JSON.stringify({ refresh_token: refreshToken }),
                        });
                        
                        if (response.ok) {
                            const data = await response.json();
                            // 保存新token
                            localStorage.setItem('origcms_token', data.access_token);
                            if (data.refresh_token) {
                                localStorage.setItem('origcms_refresh_token', data.refresh_token);
                            }
                            localStorage.setItem('token_expires_at', String(Date.now() + data.expires_in * 1000));
                            // 更新状态
                            setToken(data.access_token);
                        } else {
                            // 刷新失败，登出
                            logout();
                        }
                    } else {
                        // 没有refresh token，登出
                        logout();
                    }
                } catch (error) {
                    // 刷新失败，登出
                    logout();
                }
            }
        };

        // Check on mount
        checkToken();

        // Check every minute
        const interval = setInterval(checkToken, 60000);

        return () => clearInterval(interval);
    }, [token, logout]);

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
