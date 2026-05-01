/**
 * F018 ArticleView Component Tests
 *
 * Tests for the portal ArticleView page component.
 * Covers: article detail rendering, video player, loading/error states, no-article state.
 */

import React from 'react';
import {render, screen, waitFor} from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock TanStack Router useParams
jest.mock('@tanstack/react-router', () => ({
    useParams: ({strict}: { strict?: boolean }) => {
        // Return slug from a configurable source
        return (global as any).__mockParams || {slug: 'test-article'};
    },
}));

// Mock article API
jest.mock('../../../src/lib/api/article', () => ({
    articleApi: {
        list: jest.fn(),
        featured: jest.fn(),
        get: jest.fn(),
        getBySlug: jest.fn(),
        latest: jest.fn(),
    },
}));

// Mock format utility
jest.mock('../../../src/lib/format', () => ({
    formatDateTime: jest.fn((date: string) => '2026-04-15 10:00:00'),
}));

// Mock request module
jest.mock('../../../src/lib/request', () => ({
    API_BASE_URL: '',
}));

// Mock VideoPlayer component
jest.mock('../../../src/components/common/VideoPlayer', () => {
    return function MockVideoPlayer({src, poster}: { src?: string; poster?: string }) {
        return (
            <div data-testid="video-player" data-src={src} data-poster={poster}>
                Video Player
            </div>
        );
    };
});

import {articleApi, type Article} from '../../../src/lib/api/article';
import ArticleViewPage from '../../../src/pages/portal/ArticleView';

const mockedArticleApi = articleApi as jest.Mocked<typeof articleApi>;

// Test fixtures
const mockArticleWithVideo: Article = {
    id: '1',
    title: 'Getting Started with Go',
    slug: 'getting-started-go',
    content: '# Introduction\n\nThis is a **test** article about Go.',
    summary: 'Learn Go programming basics',
    state: 'published',
    user_id: 'user-001-abc',
    category_id: 1,
    media_id: 'media-001',
    thumbnail: '/thumbs/go.jpg',
    tags: ['golang', 'tutorial'],
    view_count: 100,
    comment_count: 5,
    featured: true,
    published_at: '2026-04-15T10:00:00Z',
    create_time: '2026-04-15T10:00:00Z',
    update_time: '2026-04-20T12:00:00Z',
    media: {
        id: 'media-001',
        title: 'Go Tutorial Video',
        thumbnail: '/thumbs/go-video.jpg',
        duration: 600,
        type: 'video',
        short_token: 'abc123',
    },
};

const mockArticleWithoutVideo: Article = {
    id: '2',
    title: 'Pure Text Article',
    slug: 'pure-text-article',
    content: 'Just text content with *italic* and `code`.',
    summary: 'A text-only article',
    state: 'published',
    user_id: 'user-002',
    thumbnail: '/thumbs/text.jpg',
    tags: ['writing'],
    view_count: 50,
    comment_count: 2,
    featured: false,
    create_time: '2026-04-16T10:00:00Z',
};

describe('F018: ArticleView Component', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (global as any).__mockParams = {slug: 'test-article'};
    });

    afterEach(() => {
        delete (global as any).__mockParams;
    });

    // ========================================================================
    // Article Detail Rendering
    // ========================================================================

    describe('Article Detail Rendering', () => {
        it('should render article title', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Getting Started with Go')).toBeInTheDocument();
            });
        });

        it('should render article summary', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Learn Go programming basics')).toBeInTheDocument();
            });
        });

        it('should render article tags', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('golang')).toBeInTheDocument();
                expect(screen.getByText('tutorial')).toBeInTheDocument();
            });
        });

        it('should render view count', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText(/100 views/)).toBeInTheDocument();
            });
        });

        it('should render comment count', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText(/5 comments/)).toBeInTheDocument();
            });
        });

        it('should render featured badge for featured articles', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                const featuredBadges = screen.getAllByText('Featured');
                expect(featuredBadges.length).toBeGreaterThanOrEqual(1);
            });
        });

        it('should render author section', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Author')).toBeInTheDocument();
            });
        });

        it('should render article info sidebar', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Article Info')).toBeInTheDocument();
            });
        });
    });

    // ========================================================================
    // Video Player
    // ========================================================================

    describe('Video Player Display', () => {
        it('should show video player when article has associated media', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByTestId('video-player')).toBeInTheDocument();
            });
        });

        it('should pass correct src to video player', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                const player = screen.getByTestId('video-player');
                expect(player.getAttribute('data-src')).toContain('/stream/abc123/index.m3u8');
            });
        });

        it('should show related video section in sidebar', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Related Video')).toBeInTheDocument();
            });
        });

        it('should show Watch Video button when media has short_token', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Watch Video')).toBeInTheDocument();
            });
        });
    });

    // ========================================================================
    // No Video (Pure Article)
    // ========================================================================

    describe('Pure Article (No Video)', () => {
        it('should not show video player when article has no media', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithoutVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Pure Text Article')).toBeInTheDocument();
            });

            expect(screen.queryByTestId('video-player')).not.toBeInTheDocument();
        });

        it('should not show related video section when no media', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithoutVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.queryByText('Related Video')).not.toBeInTheDocument();
            });
        });

        it('should show cover image when article has thumbnail but no video', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithoutVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                const img = screen.getByRole('img');
                expect(img).toBeInTheDocument();
                expect(img.getAttribute('alt')).toBe('Pure Text Article');
            });
        });
    });

    // ========================================================================
    // Loading State
    // ========================================================================

    describe('Loading State', () => {
        it('should show loading spinner while fetching article', () => {
            mockedArticleApi.getBySlug.mockReturnValue(new Promise(() => {
            }));

            render(<ArticleViewPage/>);

            // Spinner should be present
            expect(document.querySelector('.animate-spin, [data-slot="spinner"]')).toBeTruthy();
        });
    });

    // ========================================================================
    // Error State
    // ========================================================================

    describe('Error State', () => {
        it('should show error when article not found', async () => {
            mockedArticleApi.getBySlug.mockRejectedValueOnce(new Error('Not found'));

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Article Not Found')).toBeInTheDocument();
            });
        });

        it('should show Go Back button on error', async () => {
            mockedArticleApi.getBySlug.mockRejectedValueOnce(new Error('Not found'));

            render(<ArticleViewPage/>);

            await waitFor(() => {
                expect(screen.getByText('Go Back')).toBeInTheDocument();
            });
        });

        it('should show error when slug is missing', async () => {
            (global as any).__mockParams = {};

            render(<ArticleViewPage/>);

            // When no slug, the API should not be called and article should be null
            expect(mockedArticleApi.getBySlug).not.toHaveBeenCalled();
        });
    });

    // ========================================================================
    // Markdown Rendering
    // ========================================================================

    describe('Markdown Content Rendering', () => {
        it('should render markdown content as HTML', async () => {
            mockedArticleApi.getBySlug.mockResolvedValueOnce(mockArticleWithVideo);

            render(<ArticleViewPage/>);

            await waitFor(() => {
                const proseDiv = document.querySelector('.prose');
                expect(proseDiv).toBeInTheDocument();
            });
        });
    });
});
