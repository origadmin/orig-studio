import {renderHook, act} from '@testing-library/react';
import {useAuth} from './useAuth';
import {authApi} from '../lib/api/auth';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import React from 'react';

// Mock API
jest.mock('../lib/api/auth', () => ({
    authApi: {
        login: jest.fn(),
        register: jest.fn(),
        getCurrentUser: jest.fn(),
        logout: jest.fn(),
    },
}));

let queryClient: QueryClient;

const createWrapper = () => {
    queryClient = new QueryClient({
        defaultOptions: {
            queries: {
                retry: false,
                gcTime: 0,
            },
        },
    });
    return ({children}: { children: React.ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
};

describe('useAuth Hook', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        localStorage.clear();
    });

    afterEach(() => {
        queryClient?.clear();
    });

    describe('login', () => {
        it('should login and update auth state', async () => {
            const mockUser = {
                id: 1,
                username: 'testuser',
                displayName: 'Test User',
                roles: ['user']
            };
            const token = 'token-123';

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            await act(async () => {
                result.current.login(token, mockUser);
            });

            expect(localStorage.getItem('origcms_token')).toBe(token);
            expect(localStorage.getItem('origcms_user')).toBe(JSON.stringify(mockUser));
            expect(result.current.token).toBe(token);
            expect(result.current.user).toEqual(mockUser);
            expect(result.current.isAuthenticated).toBe(true);
        });
    });

    describe('logout', () => {
        it('should logout and clear state', async () => {
            // Setup: logged in state
            localStorage.setItem('origcms_token', 'token-123');
            localStorage.setItem('origcms_user', JSON.stringify({
                id: 1,
                username: 'testuser',
                displayName: 'Test User',
                roles: ['user']
            }));

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            await act(async () => {
                result.current.logout();
            });

            expect(localStorage.getItem('origcms_token')).toBeNull();
            expect(localStorage.getItem('origcms_user')).toBeNull();
            expect(result.current.token).toBeNull();
            expect(result.current.user).toBeNull();
            expect(result.current.isAuthenticated).toBe(false);
        });
    });

    describe('isAuthenticated', () => {
        it('should return true when token exists', () => {
            localStorage.setItem('origcms_token', 'token-123');
            localStorage.setItem('origcms_user', JSON.stringify({
                id: 1,
                username: 'testuser',
                displayName: 'Test User',
                roles: ['user']
            }));

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isAuthenticated).toBe(true);
        });

        it('should return false when no token', () => {
            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isAuthenticated).toBe(false);
        });
    });

    describe('isAdmin', () => {
        it('should return true for admin user', () => {
            localStorage.setItem(
                'origcms_user',
                JSON.stringify({id: 1, username: 'admin', displayName: 'Admin', roles: ['admin']})
            );

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isAdmin).toBe(true);
        });

        it('should return false for regular user', () => {
            localStorage.setItem(
                'origcms_user',
                JSON.stringify({id: 2, username: 'user', displayName: 'User', roles: ['user']})
            );

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isAdmin).toBe(false);
        });
    });
});
