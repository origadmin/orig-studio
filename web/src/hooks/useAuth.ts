/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useState, useCallback} from 'react';

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
    login: (token: string, user: User) => void;
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
    const [token, setToken] = useState<string | null>(() => getStoredToken());
    const [user, setUser] = useState<User | null>(() => getStoredUser());

    const login = useCallback((newToken: string, newUser: User) => {
        localStorage.setItem(TOKEN_KEY, newToken);
        localStorage.setItem(USER_KEY, JSON.stringify(newUser));
        setToken(newToken);
        setUser(newUser);
    }, []);

    const logout = useCallback(() => {
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(USER_KEY);
        setToken(null);
        setUser(null);
    }, []);

    const isAdmin = user?.roles?.includes('admin') ?? false;

    return {
        user,
        token,
        isAuthenticated: !!token && !!user,
        isAdmin,
        login,
        logout,
    };
}
