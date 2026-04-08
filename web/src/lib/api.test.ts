import {subscriptionApi, mediaApi, commentApi, userApi} from './api';

// Mock fetch
global.fetch = jest.fn();

const mockFetch = global.fetch as jest.MockedFunction<typeof fetch>;

describe('API Tests', () => {
    beforeEach(() => {
        mockFetch.mockClear();
    });

    describe('Subscription API', () => {
        test('getStatus should return subscription status', async () => {
            const mockResponse = {
                is_subscribed: false,
                subscriber_count: 0
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse
            } as Response);

            const result = await subscriptionApi.getStatus('1');
            expect(result).toEqual(mockResponse);
            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9090/api/v1/users/1/subscription', {
                method: 'GET',
                headers: expect.any(Object)
            });
        });

        test('subscribe should return success', async () => {
            const mockResponse = {success: true};

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse
            } as Response);

            await subscriptionApi.subscribe('1');
            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9090/api/v1/users/1/subscribe', {
                method: 'POST',
                headers: expect.any(Object)
            });
        });

        test('unsubscribe should return success', async () => {
            const mockResponse = {success: true};

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse
            } as Response);

            await subscriptionApi.unsubscribe('1');
            expect(mockFetch).toHaveBeenCalledWith('http://localhost:9090/api/v1/users/1/subscribe', {
                method: 'DELETE',
                headers: expect.any(Object)
            });
        });
    });

    describe('Share API', () => {
        test('getShareUrl should return share URL', async () => {
            const mockResponse = {
                url: 'https://localhost:9090/watch?v=1'
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse
            } as Response);

            const result = await mediaApi.getShareUrl('1');
            expect(result).toEqual(mockResponse);
        });

        test('share should return success', async () => {
            const mockResponse = {success: true};

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse
            } as Response);

            const result = await mediaApi.share('1');
            expect(result).toEqual(mockResponse);
        });
    });

    describe('Comment API', () => {
        test('getAll should return comments', async () => {
            const mockResponse = [];

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse
            } as Response);

            const result = await commentApi.getAll({media_id: '1'});
            expect(result).toEqual(mockResponse);
        });
    });
});
