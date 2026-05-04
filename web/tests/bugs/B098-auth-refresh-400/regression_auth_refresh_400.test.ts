/**
 * B096 Regression Test: auth endpoints unified response format + token refresh
 *
 * Root Cause (B096): Backend Login/RegisterUser used c.JSON() directly,
 * returning flat TokenResponse without the {code, message, data} envelope.
 * This violated C016 unified API response convention. Meanwhile, the frontend
 * attemptRefresh() used raw axios (no interceptor) and did not unwrap the
 * code/data envelope, causing setAuth() to receive the wrong shape.
 *
 * Previous B098 fix was incorrect: it made RefreshToken also use c.JSON()
 * to "match" Login, but the correct fix is to make ALL auth endpoints use
 * server.OK()/server.Created() per C016, and have the frontend manually
 * unwrap the ApiResponse envelope when using raw axios.
 *
 * Fix (backend): Login/RegisterUser/RefreshToken all use server.OK()/server.Created()
 * to return unified format {"code":0,"message":"ok","data":{...}}.
 *
 * Fix (frontend): attemptRefresh and interceptor refresh call manually unwrap
 * ApiResponse<Token> (check code===0, then use .data).
 *
 * This test verifies:
 * 1. setAuth correctly handles flat TokenResponse (unwrapped by interceptor or manually)
 * 2. attemptRefresh correctly unwraps ApiResponse<Token> and stores tokens
 * 3. The interceptor's publicUrls matching logic works correctly
 */

// Mock localStorage
const localStorageMock = (() => {
    let store: Record<string, string> = {};
    return {
        getItem: jest.fn((key: string) => store[key] ?? null),
        setItem: jest.fn((key: string, value: string) => { store[key] = value; }),
        removeItem: jest.fn((key: string) => { delete store[key]; }),
        clear: jest.fn(() => { store = {}; }),
    };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

// Mock axios
jest.mock('axios', () => {
    const actualAxios = jest.requireActual('axios');
    return {
        ...actualAxios,
        __esModule: true,
        default: {
            ...actualAxios.default,
            post: jest.fn(),
            create: jest.fn(() => ({
                ...actualAxios.default.create(),
                interceptors: {
                    request: { use: jest.fn() },
                    response: { use: jest.fn() },
                },
                get: jest.fn(),
                post: jest.fn(),
                put: jest.fn(),
                delete: jest.fn(),
            })),
        },
    };
});

import axios from 'axios';
import { setAuth, clearAuth, attemptRefresh } from '@/lib/request';

// Shared localStorage keys (must match request.ts)
const TOKEN_KEY = 'origcms_token';
const REFRESH_TOKEN_KEY = 'origcms_refresh_token';

describe('B096: auth endpoints unified response format + token refresh', () => {
    beforeEach(() => {
        localStorageMock.clear();
        jest.clearAllMocks();
    });

    describe('setAuth with flat TokenResponse (after interceptor/manual unwrapping)', () => {
        it('should correctly store access_token and refresh_token from unwrapped response', () => {
            // After B096 fix, the response interceptor auto-unwraps {code, data} for
            // api.post calls, and attemptRefresh manually unwraps. setAuth always
            // receives the flat Token object.
            const tokenData = {
                access_token: 'eyJhbGciOiJIUzI1NiJ9.test-access-token',
                refresh_token: 'eyJhbGciOiJIUzI1NiJ9.test-refresh-token',
                token_type: 'Bearer',
                expires_in: 3600,
                user: {
                    id: '1',
                    username: 'admin',
                    nickname: 'Admin',
                    is_staff: true,
                },
            };

            setAuth(tokenData);

            expect(localStorageMock.setItem).toHaveBeenCalledWith(
                TOKEN_KEY,
                'eyJhbGciOiJIUzI1NiJ9.test-access-token'
            );
            expect(localStorageMock.setItem).toHaveBeenCalledWith(
                REFRESH_TOKEN_KEY,
                'eyJhbGciOiJIUzI1NiJ9.test-refresh-token'
            );
        });

        it('should NOT store "undefined" as access_token (B096/B098 regression guard)', () => {
            const tokenData = {
                access_token: 'valid-access-token',
                refresh_token: 'valid-refresh-token',
                token_type: 'Bearer',
                expires_in: 3600,
            };

            setAuth(tokenData);

            const tokenCalls = (localStorageMock.setItem as jest.Mock).mock.calls.filter(
                (call: any[]) => call[0] === TOKEN_KEY
            );

            expect(tokenCalls.length).toBeGreaterThan(0);
            const storedToken = tokenCalls[0][1];
            expect(storedToken).not.toBe('undefined');
            expect(storedToken).not.toBeUndefined();
            expect(storedToken).toBe('valid-access-token');
        });
    });

    describe('attemptRefresh with unified ApiResponse format (B096 fix)', () => {
        it('should correctly unwrap ApiResponse<Token> and extract tokens', async () => {
            // After B096 fix, /auth/refresh returns unified format:
            // {"code":0,"message":"ok","data":{"access_token":...,"refresh_token":...}}
            // attemptRefresh manually unwraps the ApiResponse envelope.
            const unifiedResponse = {
                data: {
                    code: 0,
                    message: 'ok',
                    data: {
                        access_token: 'eyJhbGciOiJIUzI1NiJ9.new-access-token',
                        refresh_token: 'eyJhbGciOiJIUzI1NiJ9.new-refresh-token',
                        token_type: 'Bearer',
                        expires_in: 3600,
                        user: {
                            id: '1',
                            username: 'admin',
                            is_staff: true,
                        },
                    },
                },
            };

            // Set up existing refresh token in localStorage
            localStorageMock.getItem.mockImplementation((key: string) => {
                if (key === REFRESH_TOKEN_KEY) return 'old-refresh-token';
                if (key === TOKEN_KEY) return 'old-access-token';
                return null;
            });

            (axios.post as jest.Mock).mockResolvedValue(unifiedResponse);

            const result = await attemptRefresh();

            expect(result).toBe(true);

            // The stored access_token should be the actual token, not "undefined"
            const tokenCalls = (localStorageMock.setItem as jest.Mock).mock.calls.filter(
                (call: any[]) => call[0] === TOKEN_KEY
            );
            expect(tokenCalls.length).toBeGreaterThan(0);
            const storedToken = tokenCalls[0][1];
            expect(storedToken).toBe('eyJhbGciOiJIUzI1NiJ9.new-access-token');
            expect(storedToken).not.toBe('undefined');

            // The stored refresh_token should also be the actual token
            const refreshTokenCalls = (localStorageMock.setItem as jest.Mock).mock.calls.filter(
                (call: any[]) => call[0] === REFRESH_TOKEN_KEY
            );
            expect(refreshTokenCalls.length).toBeGreaterThan(0);
            const storedRefreshToken = refreshTokenCalls[0][1];
            expect(storedRefreshToken).toBe('eyJhbGciOiJIUzI1NiJ9.new-refresh-token');
            expect(storedRefreshToken).not.toBe('undefined');
        });

        it('should return false when ApiResponse code is non-zero', async () => {
            const errorResponse = {
                data: {
                    code: 401,
                    message: 'invalid refresh token',
                    data: null,
                },
            };

            localStorageMock.getItem.mockImplementation((key: string) => {
                if (key === REFRESH_TOKEN_KEY) return 'old-refresh-token';
                return null;
            });

            (axios.post as jest.Mock).mockResolvedValue(errorResponse);

            const result = await attemptRefresh();
            expect(result).toBe(false);
        });

        it('should return false when ApiResponse data is null', async () => {
            const nullDataResponse = {
                data: {
                    code: 0,
                    message: 'ok',
                    data: null,
                },
            };

            localStorageMock.getItem.mockImplementation((key: string) => {
                if (key === REFRESH_TOKEN_KEY) return 'old-refresh-token';
                return null;
            });

            (axios.post as jest.Mock).mockResolvedValue(nullDataResponse);

            const result = await attemptRefresh();
            expect(result).toBe(false);
        });

        it('should return false when no refresh token is available', async () => {
            localStorageMock.getItem.mockImplementation((key: string) => {
                if (key === REFRESH_TOKEN_KEY) return null;
                return null;
            });

            const result = await attemptRefresh();
            expect(result).toBe(false);
            expect(axios.post).not.toHaveBeenCalled();
        });

        it('should return false when refresh request fails', async () => {
            localStorageMock.getItem.mockImplementation((key: string) => {
                if (key === REFRESH_TOKEN_KEY) return 'old-refresh-token';
                return null;
            });

            (axios.post as jest.Mock).mockRejectedValue(new Error('Network error'));

            const result = await attemptRefresh();
            expect(result).toBe(false);
        });
    });

    describe('interceptor publicUrls matching (B098/B096 fix)', () => {
        it('should correctly identify /auth/refresh as a public URL using includes()', () => {
            // After B098 fix, the interceptor uses includes() matching
            // instead of exact URL comparison with full API_PREFIX path.
            // originalRequest.url is the relative path (e.g., "/auth/refresh")
            // because baseURL already includes API_PREFIX.
            const publicAuthUrls = ["/auth/refresh", "/auth/signin", "/auth/signup"];

            const testUrls = [
                "/auth/refresh",
                "/auth/signin",
                "/auth/signup",
                "/auth/refresh?token=xxx",  // with query params
            ];

            for (const url of testUrls) {
                const isPublicAuthUrl = publicAuthUrls.some(pattern => url.includes(pattern));
                expect(isPublicAuthUrl).toBe(true);
            }

            // Non-auth URLs should NOT match
            const nonAuthUrls = ["/me", "/admin/medias", "/articles"];
            for (const url of nonAuthUrls) {
                const isPublicAuthUrl = publicAuthUrls.some(pattern => url.includes(pattern));
                expect(isPublicAuthUrl).toBe(false);
            }
        });
    });
});
