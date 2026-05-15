import {renderHook, act} from '@testing-library/react';
import {useAuth} from './useAuth';
import {AuthProvider} from '@/contexts/auth/AuthProvider';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import React from 'react';

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
        <QueryClientProvider client={queryClient}>
            <AuthProvider>{children}</AuthProvider>
        </QueryClientProvider>
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
            // Create a mock JWT token that won't be considered expired
            const header = btoa(JSON.stringify({alg: 'HS256', typ: 'JWT'}));
            const payload = btoa(JSON.stringify({exp: Math.floor(Date.now() / 1000) + 3600, sub: '1'}));
            const token = `${header}.${payload}.mock-signature`;

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            await act(async () => {
                result.current.login(token, 'refresh-token-123', mockUser);
            });

            expect(localStorage.getItem('origstudio_token')).toBe(token);
            expect(localStorage.getItem('origstudio_user')).toBe(JSON.stringify(mockUser));
            expect(result.current.token).toBe(token);
            expect(result.current.user).toEqual(mockUser);
            expect(result.current.isAuthenticated).toBe(true);
        });
    });

    describe('logout', () => {
        it('should logout and clear state', async () => {
            // Setup: logged in state
            const header = btoa(JSON.stringify({alg: 'HS256', typ: 'JWT'}));
            const payload = btoa(JSON.stringify({exp: Math.floor(Date.now() / 1000) + 3600, sub: '1'}));
            const token = `${header}.${payload}.mock-signature`;
            localStorage.setItem('origstudio_token', token);
            localStorage.setItem('origstudio_user', JSON.stringify({
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

            expect(localStorage.getItem('origstudio_token')).toBeNull();
            expect(localStorage.getItem('origstudio_user')).toBeNull();
            expect(result.current.token).toBeNull();
            expect(result.current.user).toBeNull();
            expect(result.current.isAuthenticated).toBe(false);
        });
    });

    describe('isAuthenticated', () => {
        it('should return true when token exists and not expired', () => {
            const header = btoa(JSON.stringify({alg: 'HS256', typ: 'JWT'}));
            const payload = btoa(JSON.stringify({exp: Math.floor(Date.now() / 1000) + 3600, sub: '1'}));
            const token = `${header}.${payload}.mock-signature`;
            localStorage.setItem('origstudio_token', token);
            localStorage.setItem('origstudio_user', JSON.stringify({
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
            const header = btoa(JSON.stringify({alg: 'HS256', typ: 'JWT'}));
            const payload = btoa(JSON.stringify({exp: Math.floor(Date.now() / 1000) + 3600, sub: '1'}));
            const token = `${header}.${payload}.mock-signature`;
            localStorage.setItem('origstudio_token', token);
            localStorage.setItem(
                'origstudio_user',
                JSON.stringify({id: 1, username: 'admin', displayName: 'Admin', roles: ['admin']})
            );

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isAdmin).toBe(true);
        });

        it('should return false for regular user', () => {
            const header = btoa(JSON.stringify({alg: 'HS256', typ: 'JWT'}));
            const payload = btoa(JSON.stringify({exp: Math.floor(Date.now() / 1000) + 3600, sub: '1'}));
            const token = `${header}.${payload}.mock-signature`;
            localStorage.setItem('origstudio_token', token);
            localStorage.setItem(
                'origstudio_user',
                JSON.stringify({id: 2, username: 'user', displayName: 'User', roles: ['user']})
            );

            const {result} = renderHook(() => useAuth(), {
                wrapper: createWrapper(),
            });

            expect(result.current.isAdmin).toBe(false);
        });
    });

    describe('AuthProvider context', () => {
        it('should throw error when useAuth is called outside AuthProvider', () => {
            // Suppress console.error for expected error
            const spy = jest.spyOn(console, 'error').mockImplementation(() => {});
            expect(() => useAuth()).toThrow('useAuth must be used within an AuthProvider');
            spy.mockRestore();
        });
    });
});
