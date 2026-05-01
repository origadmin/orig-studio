import {renderHook, act} from '@testing-library/react';
import {useMediaList} from './useMedia';

// Mock fetch
global.fetch = jest.fn();

describe('useMedia Hook', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('useMediaList', () => {
        it('should fetch media list successfully', async () => {
            const mockData = {
                sections: [
                    {
                        items: [
                            {
                                id: 1,
                                title: 'Test Video 1',
                                description: 'Test Description 1',
                                thumbnail_url: 'http://example.com/thumb1.jpg',
                                view_count: 100,
                                author_id: 1,
                                create_time: '2024-01-01T00:00:00Z',
                            },
                        ],
                    },
                ],
            };

            (global.fetch as jest.Mock).mockResolvedValue({
                ok: true,
                json: jest.fn().mockResolvedValue(mockData),
            });

            const {result} = renderHook(() => useMediaList());

            // Initially loading should be true
            expect(result.current.loading).toBe(true);

            // Wait for the fetch to complete
            await act(async () => {
                await new Promise((resolve) => setTimeout(resolve, 0));
            });

            expect(result.current.loading).toBe(false);
            expect(result.current.error).toBeNull();
            expect(result.current.items).toHaveLength(1);
            expect(result.current.items[0].title).toBe('Test Video 1');
            expect(result.current.items[0].description).toBe('Test Description 1');
            expect(result.current.items[0].viewCount).toBe(100);
        });

        it('should handle fetch errors', async () => {
            const errorMessage = 'Failed to load media';
            (global.fetch as jest.Mock).mockRejectedValue(new Error(errorMessage));

            const {result} = renderHook(() => useMediaList());

            await act(async () => {
                await new Promise((resolve) => setTimeout(resolve, 0));
            });

            expect(result.current.loading).toBe(false);
            expect(result.current.error).toBe(errorMessage);
            expect(result.current.items).toEqual([]);
        });

        it('should handle HTTP errors', async () => {
            (global.fetch as jest.Mock).mockResolvedValue({
                ok: false,
                status: 500,
            });

            const {result} = renderHook(() => useMediaList());

            await act(async () => {
                await new Promise((resolve) => setTimeout(resolve, 0));
            });

            expect(result.current.loading).toBe(false);
            expect(result.current.error).toBe('HTTP 500');
            expect(result.current.items).toEqual([]);
        });

        it('should refresh media list', async () => {
            const mockData1 = {
                sections: [
                    {
                        items: [
                            {
                                id: 1,
                                title: 'Test Video 1',
                                description: 'Test Description 1',
                                thumbnail_url: 'http://example.com/thumb1.jpg',
                                view_count: 100,
                                author_id: 1,
                                create_time: '2024-01-01T00:00:00Z',
                            },
                        ],
                    },
                ],
            };

            const mockData2 = {
                sections: [
                    {
                        items: [
                            {
                                id: 1,
                                title: 'Test Video 1',
                                description: 'Test Description 1',
                                thumbnail_url: 'http://example.com/thumb1.jpg',
                                view_count: 100,
                                author_id: 1,
                                create_time: '2024-01-01T00:00:00Z',
                            },
                            {
                                id: 2,
                                title: 'Test Video 2',
                                description: 'Test Description 2',
                                thumbnail_url: 'http://example.com/thumb2.jpg',
                                view_count: 200,
                                author_id: 2,
                                create_time: '2024-01-02T00:00:00Z',
                            },
                        ],
                    },
                ],
            };

            // First fetch
            (global.fetch as jest.Mock).mockResolvedValueOnce({
                ok: true,
                json: jest.fn().mockResolvedValue(mockData1),
            });

            const {result} = renderHook(() => useMediaList());

            await act(async () => {
                await new Promise((resolve) => setTimeout(resolve, 0));
            });

            expect(result.current.items).toHaveLength(1);

            // Second fetch - refresh
            (global.fetch as jest.Mock).mockResolvedValueOnce({
                ok: true,
                json: jest.fn().mockResolvedValue(mockData2),
            });

            await act(async () => {
                result.current.refresh();
                await new Promise((resolve) => setTimeout(resolve, 0));
            });

            expect(result.current.items).toHaveLength(2);
        });

        it('should respect pagination params', async () => {
            const mockData = {
                sections: [],
            };

            (global.fetch as jest.Mock).mockResolvedValue({
                ok: true,
                json: jest.fn().mockResolvedValue(mockData),
            });

            const {result} = renderHook(() => useMediaList({page: 2, pageSize: 20}));

            await act(async () => {
                await new Promise((resolve) => setTimeout(resolve, 0));
            });

            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('page=2'),
            );
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('page_size=20'),
            );
        });
    });
});
