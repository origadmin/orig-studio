import {subscriptionApi, mediaApi, commentApi, userApi} from './api';

// Mock request module
jest.mock('./request', () => ({
    api: {
        get: jest.fn(),
        post: jest.fn(),
        del: jest.fn(),
    },
    getAccessToken: jest.fn(),
}));

// Import the mocked functions
import {api, getAccessToken} from './request';

describe('API Tests', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('Subscription API', () => {
        test('getStatus should return subscription status', async () => {
            const mockResponse = {
                is_subscribed: false,
                subscriber_count: 0
            };

            (api.get as jest.Mock).mockResolvedValueOnce(mockResponse);

            const result = await subscriptionApi.getStatus('1');
            expect(result).toEqual(mockResponse);
            expect(api.get).toHaveBeenCalledWith('/users/1/subscription');
        });

        test('subscribe should return success', async () => {
            (api.post as jest.Mock).mockResolvedValueOnce(undefined);

            await subscriptionApi.subscribe('1');
            expect(api.post).toHaveBeenCalledWith('/users/1/subscribe');
        });

        test('unsubscribe should return success', async () => {
            (api.del as jest.Mock).mockResolvedValueOnce(undefined);

            await subscriptionApi.unsubscribe('1');
            expect(api.del).toHaveBeenCalledWith('/users/1/subscribe');
        });
    });

    describe('Share API', () => {
        test('getShareUrl should return share URL', async () => {
            const mockResponse = {
                url: 'https://localhost:8080/watch?v=1'
            };

            (api.get as jest.Mock).mockResolvedValueOnce(mockResponse);

            const result = await mediaApi.getShareUrl('1');
            expect(result).toEqual(mockResponse);
            expect(api.get).toHaveBeenCalledWith('/media/1/share');
        });

        test('share should return success', async () => {
            const mockResponse = {success: true};

            (api.post as jest.Mock).mockResolvedValueOnce(mockResponse);

            const result = await mediaApi.share('1');
            expect(result).toEqual(mockResponse);
            expect(api.post).toHaveBeenCalledWith('/media/1/share');
        });
    });

    describe('Comment API', () => {
        test('getAll should return comments', async () => {
            const mockResponse = {
                list: [],
                total: 0,
                page: 1,
                limit: 20
            };

            (api.get as jest.Mock).mockResolvedValueOnce(mockResponse);

            const result = await commentApi.getAll({media_id: '1'});
            expect(result).toEqual(mockResponse);
            expect(api.get).toHaveBeenCalledWith('/comments', {media_id: '1'});
        });
    });
});
