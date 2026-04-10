import {authApi} from './auth';

// Mock request module
jest.mock('../request', () => ({
    api: {
        post: jest.fn(),
        get: jest.fn(),
    },
    setAuth: jest.fn(),
    clearAuth: jest.fn(),
    getAccessToken: jest.fn(),
    getRefreshToken: jest.fn(),
    isTokenExpired: jest.fn().mockReturnValue(false),
}));

// Import the mocked functions
import {api, getAccessToken} from '../request';

describe('Auth API', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        localStorage.clear();
    });

    describe('login', () => {
        it('should login successfully and store token', async () => {
            const mockResponse = {
                access_token: 'test-token-123',
                token_type: 'Bearer',
                expires_in: 3600,
                user: {
                    id: 1,
                    username: 'testuser',
                    email: 'test@example.com',
                    is_staff: false,
                },
            };

            (api.post as jest.Mock).mockResolvedValueOnce(mockResponse);

            const result = await authApi.login('testuser', 'password123');

            expect(result).toEqual(mockResponse);
            expect(api.post).toHaveBeenCalledWith('/auth/signin', {
                username: 'testuser',
                password: 'password123',
            });
        });

        it('should handle login failure', async () => {
            const mockError = {
                response: {
                    status: 401,
                    data: {error: 'invalid credentials'},
                },
            };

            (api.post as jest.Mock).mockRejectedValueOnce(new Error('invalid credentials'));

            await expect(
                authApi.login('wrong', 'wrong')
            ).rejects.toThrow('invalid credentials');
        });
    });

    describe('register', () => {
        it('should register successfully', async () => {
            const mockResponse = {
                access_token: 'new-token-456',
                token_type: 'Bearer',
                expires_in: 3600,
                user: {
                    id: 2,
                    username: 'newuser',
                    email: 'new@example.com',
                },
            };

            (api.post as jest.Mock).mockResolvedValueOnce(mockResponse);

            const result = await authApi.register('newuser', 'new@example.com', 'password123');

            expect(result).toEqual(mockResponse);
            expect(api.post).toHaveBeenCalledWith('/auth/signup', {
                username: 'newuser',
                password: 'password123',
                email: 'new@example.com',
            });
        });

        it('should handle duplicate username', async () => {
            const mockError = {
                response: {
                    status: 409,
                    data: {error: 'username already exists'},
                },
            };

            (api.post as jest.Mock).mockRejectedValueOnce(new Error('username already exists'));

            await expect(
                authApi.register('existing', 'existing@example.com', 'pass123')
            ).rejects.toThrow('username already exists');
        });
    });

    describe('getCurrentUser', () => {
        it('should get current user info', async () => {
            const mockUser = {
                id: 1,
                username: 'testuser',
                email: 'test@example.com',
                is_staff: true,
                role: 'admin',
            };

            (getAccessToken as jest.Mock).mockReturnValueOnce('test-token-123');
            (api.get as jest.Mock).mockResolvedValueOnce(mockUser);

            const result = await authApi.getCurrentUser();

            expect(result).toEqual(mockUser);
            expect(api.get).toHaveBeenCalledWith('/auth/me');
        });

        it('should handle 401 when not authenticated', async () => {
            const mockError = {
                response: {
                    status: 401,
                    data: {error: 'unauthorized'},
                },
            };

            (getAccessToken as jest.Mock).mockReturnValueOnce(null);

            await expect(authApi.getCurrentUser()).rejects.toThrow('Not authenticated');
        });
    });

    describe('logout', () => {
        it('should logout successfully', async () => {
            const result = await authApi.logout();

            expect(result).toBeUndefined();
        });
    });
});
