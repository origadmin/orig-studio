import React from 'react';
import {render, screen, waitFor} from '@testing-library/react';
import {RouterProvider, createMemoryHistory, createRouter} from '@tanstack/react-router';
import {routeTree} from '@/router';
import WatchPage from './Watch';

// Mock the necessary hooks and APIs
jest.mock('@/hooks/queries', () => ({
    useMediaDetail: () => ({
        data: {
            id: 1,
            title: 'Test Video',
            description: 'Test Description',
            view_count: 1000,
            like_count: 50,
            favorite_count: 20,
            created_at: '2024-01-01T00:00:00Z',
            encoding_status: 'success',
            url: '/api/v1/media/1',
            hls_file: '/api/v1/media/1/hls.m3u8',
            thumbnail: '/api/v1/media/1/thumbnail.jpg',
            poster: '/api/v1/media/1/poster.jpg',
            preview_file_path: '/api/v1/media/1/preview.jpg',
            duration: 300,
            user_id: 1,
            tags: ['test', 'video'],
            edges: {
                user: [{
                    id: 1,
                    username: 'testuser',
                    nickname: 'Test User',
                    avatar: '/api/v1/users/1/avatar.jpg',
                    subscriber_count: 100
                }],
                category: {
                    id: 1,
                    name: 'Test Category'
                }
            }
        },
        isLoading: false,
        error: null
    }),
    useMediaList: () => ({
        data: {
            list: [
                {
                    id: 2,
                    title: 'Recommended Video 1',
                    view_count: 500,
                    created_at: '2024-01-02T00:00:00Z',
                    duration: 180,
                    thumbnail: '/api/v1/media/2/thumbnail.jpg',
                    edges: {
                        user: [{
                            id: 2,
                            username: 'otheruser',
                            nickname: 'Other User'
                        }]
                    }
                },
                {
                    id: 3,
                    title: 'Recommended Video 2',
                    view_count: 300,
                    created_at: '2024-01-03T00:00:00Z',
                    duration: 240,
                    thumbnail: '/api/v1/media/3/thumbnail.jpg',
                    edges: {
                        user: [{
                            id: 3,
                            username: 'anotheruser',
                            nickname: 'Another User'
                        }]
                    }
                }
            ]
        }
    }),
    useDeleteMedia: () => ({
        mutateAsync: jest.fn()
    })
}));

jest.mock('@/hooks/useAuth', () => ({
    useAuth: () => ({
        user: {
            id: 1,
            username: 'testuser'
        },
        isAuthenticated: true
    })
}));

// Mock i18n
jest.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: (key: string) => key,
        i18n: {
            changeLanguage: jest.fn()
        }
    })
}));

// Mock format functions
jest.mock('@/lib/format', () => ({
    formatViews: (views: number) => {
        if (views >= 1000) {
            return '1K views';
        }
        return `${views} views`;
    },
    formatDate: (date: string) => '2024-01-01',
    formatDuration: (seconds: number) => '5:00'
}));

// Mock InteractionBar component
jest.mock('@/components/common/InteractionBar', () => ({
    __esModule: true,
    default: () => (
        <div>
            <button>Save</button>
            <button>Download</button>
        </div>
    )
}));

// Mock SubscribeButton component
jest.mock('@/components/common/SubscribeButton', () => ({
    __esModule: true,
    default: () => <button>Subscribe</button>
}));

// Mock CommentSection component
jest.mock('@/components/common/CommentSection', () => ({
    __esModule: true,
    default: () => <div>Comments</div>
}));

// Mock VideoPreview component
jest.mock('@/components/common/VideoPreview', () => ({
    __esModule: true,
    default: () => <div>Video Preview</div>
}));

// Mock VideoPlayer component
jest.mock('@/components/common/VideoPlayer', () => ({
    __esModule: true,
    default: () => <div>Video Player</div>
}));

// Mock NotificationBadge component
jest.mock('@/components/common/NotificationBadge', () => ({
    __esModule: true,
    default: () => <div>Notification Badge</div>
}));

// Mock NotificationCenter component
jest.mock('@/components/common/NotificationCenter', () => ({
    __esModule: true,
    default: () => <div>Notification Center</div>
}));

// Mock Header component
jest.mock('@/components/portal/Header', () => ({
    __esModule: true,
    default: () => <header>Header</header>
}));

// Mock Sidebar component
jest.mock('@/components/portal/Sidebar', () => ({
    __esModule: true,
    default: () => <aside>Sidebar</aside>
}));

jest.mock('@/lib/api/media', () => ({
    mediaApi: {
        getVariants: jest.fn().mockResolvedValue({
            data: {
                variants: [
                    {
                        status: 'success',
                        output_path: '/api/v1/media/1/720p.m3u8',
                        resolution: '720p',
                        profile_name: '720p'
                    },
                    {
                        status: 'success',
                        output_path: '/api/v1/media/1/480p.m3u8',
                        resolution: '480p',
                        profile_name: '480p'
                    }
                ]
            }
        }),
        retryTranscode: jest.fn().mockResolvedValue({})
    }
}));

jest.mock('@/lib/api/comment', () => ({
    commentApi: {
        getAll: jest.fn().mockResolvedValue({
            list: [
                {
                    id: 1,
                    body: 'Test comment',
                    username: 'testuser',
                    created_at: '2024-01-01T01:00:00Z'
                }
            ]
        }),
        create: jest.fn().mockResolvedValue({})
    }
}));

jest.mock('@/lib/api/like', () => ({
    likeApi: {
        getStatus: jest.fn().mockResolvedValue({
            is_liked: false
        }),
        toggle: jest.fn().mockResolvedValue({})
    }
}));

jest.mock('@/lib/api/favorite', () => ({
    favoriteApi: {
        getStatus: jest.fn().mockResolvedValue({
            is_favorited: false
        }),
        toggle: jest.fn().mockResolvedValue({})
    }
}));

jest.mock('hls.js', () => {
    const mockHls = jest.fn().mockImplementation(() => ({
        loadSource: jest.fn(),
        attachMedia: jest.fn(),
        on: jest.fn(),
        destroy: jest.fn(),
        levels: [
            {width: 1280, height: 720},
            {width: 854, height: 480}
        ],
        currentLevel: -1
    }));
    mockHls.isSupported = jest.fn().mockReturnValue(true);
    mockHls.Events = {
        MANIFEST_PARSED: 'MANIFEST_PARSED',
        LEVEL_SWITCHED: 'LEVEL_SWITCHED'
    };
    return mockHls;
});

describe('WatchPage', () => {
    beforeEach(() => {
        // Clear all mocks before each test
        jest.clearAllMocks();
    });

    // Helper to render with router
    const renderWithRouter = (initialUrl: string) => {
        const memoryHistory = createMemoryHistory({
            initialEntries: [initialUrl],
        });
        const router = createRouter({routeTree, history: memoryHistory});
        return render(<RouterProvider router={router}/>);
    };

    it('should render video title', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            expect(screen.getByText('Test Video')).toBeInTheDocument();
        });
    });

    it('should render video description', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            expect(screen.getByText('Test Description')).toBeInTheDocument();
        });
    });

    it('should render view count', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            expect(screen.getByText(/1K views/)).toBeInTheDocument();
        });
    });

    it('should render username', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            expect(screen.getByText('Test User')).toBeInTheDocument();
        });
    });

    it('should render recommendations', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            expect(screen.getByText('Recommended Video 1')).toBeInTheDocument();
            expect(screen.getByText('Recommended Video 2')).toBeInTheDocument();
        });
    });

    it('should render tags', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            expect(screen.getByText(/test/)).toBeInTheDocument();
            expect(screen.getByText(/video/)).toBeInTheDocument(); 
        });
    });

    it('should render interaction buttons', async () => {
        renderWithRouter('/watch?v=1');

        await waitFor(() => {
            // Check that interaction buttons are rendered
            expect(screen.getByText('Save')).toBeInTheDocument();   
            expect(screen.getByText('Download')).toBeInTheDocument();
        });
    });
});
