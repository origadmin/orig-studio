/**
 * F018 ArticleList Component Tests
 *
 * Tests for the portal ArticleList page component.
 * Covers: rendering, search, pagination, loading/error/empty states.
 */

import React from 'react';
import {render, screen, fireEvent, waitFor, act} from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock TanStack Router Link
jest.mock('@tanstack/react-router', () => ({
    Link: ({children, to, params}: { children: React.ReactNode; to: string; params?: Record<string, string> }) => (
        <a href={to.replace('$slug', params?.slug || '')} data-testid="link">{children}</a>
    ),
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
    adminArticleApi: {
        adminList: jest.fn(),
        get: jest.fn(),
        create: jest.fn(),
        update: jest.fn(),
        delete: jest.fn(),
        updateState: jest.fn(),
    },
}));

// Mock format utility
jest.mock('../../../src/lib/format', () => ({
    formatRelativeTime: jest.fn((date: string) => '2d ago'),
}));

// Mock request module
jest.mock('../../../src/lib/request', () => ({
    API_BASE_URL: '',
}));

import {articleApi, type Article} from '../../../src/lib/api/article';
import ArticleListPage from '../../../src/pages/portal/ArticleList';

const mockedArticleApi = articleApi as jest.Mocked<typeof articleApi>;

// Test fixtures
const mockArticles: Article[] = [
    {
        id: '1',
        title: 'Getting Started with Go',
        slug: 'getting-started-go',
        content: 'Content here',
        summary: 'Learn Go programming',
        state: 'published',
        user_id: 'user-001',
        category_id: 1,
        thumbnail: '/thumbs/go.jpg',
        tags: ['golang', 'tutorial'],
        view_count: 100,
        comment_count: 5,
        featured: false,
        create_time: '2026-04-15T10:00:00Z',
    },
    {
        id: '2',
        title: 'Advanced React Patterns',
        slug: 'advanced-react-patterns',
        content: 'React patterns',
        summary: 'Deep dive into React',
        state: 'published',
        user_id: 'user-002',
        media_id: 'media-001',
        media: {
            id: 'media-001',
            title: 'React Video',
            thumbnail: '/thumbs/react.jpg',
            duration: 600,
            type: 'video',
            short_token: 'tok001',
        },
        tags: ['react', 'frontend'],
        view_count: 200,
        comment_count: 10,
        featured: true,
        create_time: '2026-04-16T10:00:00Z',
    },
    {
        id: '3',
        title: 'No Thumbnail Article',
        slug: 'no-thumb-article',
        content: 'No thumbnail',
        state: 'published',
        user_id: 'user-003',
        view_count: 50,
        comment_count: 2,
        featured: false,
        create_time: '2026-04-17T10:00:00Z',
    },
];

const mockListResponse = {
    items: mockArticles,
    total: 3,
    page: 1,
    page_size: 12,
};

describe('F018: ArticleList Component', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedArticleApi.featured.mockResolvedValue([]);
        mockedArticleApi.list.mockResolvedValue(mockListResponse);
    });

    // ========================================================================
    // Rendering
    // ========================================================================

    describe('Rendering', () => {
        it('should render page header with title and description', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Articles')).toBeInTheDocument();
            });
            expect(screen.getByText(/Discover in-depth articles/)).toBeInTheDocument();
        });

        it('should render article cards with titles', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Getting Started with Go')).toBeInTheDocument();
            });
            expect(screen.getByText('Advanced React Patterns')).toBeInTheDocument();
            expect(screen.getByText('No Thumbnail Article')).toBeInTheDocument();
        });

        it('should render article summaries', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Learn Go programming')).toBeInTheDocument();
            });
            expect(screen.getByText('Deep dive into React')).toBeInTheDocument();
        });

        it('should render view counts', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('100')).toBeInTheDocument();
            });
        });

        it('should render tags for articles', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('golang')).toBeInTheDocument();
            });
            expect(screen.getByText('tutorial')).toBeInTheDocument();
            expect(screen.getByText('react')).toBeInTheDocument();
        });

        it('should render featured badge for featured articles', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                // There are Featured badges (one in article card, possibly one in featured section)
                const featuredBadges = screen.getAllByText('Featured');
                expect(featuredBadges.length).toBeGreaterThanOrEqual(1);
            });
        });

        it('should render search input', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByPlaceholderText('Search articles...')).toBeInTheDocument();
            });
        });
    });

    // ========================================================================
    // Featured Section
    // ========================================================================

    describe('Featured Section', () => {
        it('should render featured section when featured articles exist', async () => {
            mockedArticleApi.featured.mockResolvedValue([mockArticles[1]]);

            render(<ArticleListPage/>);

            await waitFor(() => {
                // Featured section heading + Featured badges in cards
                const featuredElements = screen.getAllByText('Featured');
                expect(featuredElements.length).toBeGreaterThanOrEqual(1);
            });
        });

        it('should not render featured section when no featured articles', async () => {
            mockedArticleApi.featured.mockResolvedValue([]);

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Getting Started with Go')).toBeInTheDocument();
            });

            // "Featured" heading should not be present (only badges in cards)
            const featuredHeadings = screen.queryAllByRole('heading', {name: 'Featured'});
            expect(featuredHeadings.length).toBe(0);
        });
    });

    // ========================================================================
    // Loading State
    // ========================================================================

    describe('Loading State', () => {
        it('should show loading spinner while fetching articles', () => {
            mockedArticleApi.list.mockReturnValue(new Promise(() => {
            }));

            render(<ArticleListPage/>);

            // Spinner should be visible during loading
            expect(document.querySelector('.animate-spin, [data-slot="spinner"]')).toBeTruthy();
        });
    });

    // ========================================================================
    // Error State
    // ========================================================================

    describe('Error State', () => {
        it('should show error message when API fails', async () => {
            mockedArticleApi.list.mockRejectedValue(new Error('Network error'));

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Failed to load articles')).toBeInTheDocument();
            });
        });

        it('should show retry button on error', async () => {
            mockedArticleApi.list.mockRejectedValue(new Error('Network error'));

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Retry')).toBeInTheDocument();
            });
        });
    });

    // ========================================================================
    // Empty State
    // ========================================================================

    describe('Empty State', () => {
        it('should show empty state when no articles exist', async () => {
            mockedArticleApi.list.mockResolvedValue({items: [], total: 0, page: 1, page_size: 12});

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('No articles yet')).toBeInTheDocument();
            });
        });

        it('should show search-specific empty message', async () => {
            mockedArticleApi.list.mockResolvedValue({items: [], total: 0, page: 1, page_size: 12});

            render(<ArticleListPage/>);

            const searchInput = screen.getByPlaceholderText('Search articles...');
            await act(async () => {
                fireEvent.change(searchInput, {target: {value: 'nonexistent'}});
            });

            await waitFor(() => {
                expect(screen.getByText('No articles match your search')).toBeInTheDocument();
            });
        });
    });

    // ========================================================================
    // Search
    // ========================================================================

    describe('Search', () => {
        it('should update search input value on change', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Getting Started with Go')).toBeInTheDocument();
            });

            const searchInput = screen.getByPlaceholderText('Search articles...');
            await act(async () => {
                fireEvent.change(searchInput, {target: {value: 'golang'}});
            });

            expect(searchInput).toHaveValue('golang');
        });

        it('should call API with keyword when searching', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(mockedArticleApi.list).toHaveBeenCalled();
            });

            jest.clearAllMocks();
            mockedArticleApi.list.mockResolvedValue(mockListResponse);

            const searchInput = screen.getByPlaceholderText('Search articles...');
            await act(async () => {
                fireEvent.change(searchInput, {target: {value: 'golang'}});
            });

            await waitFor(() => {
                expect(mockedArticleApi.list).toHaveBeenCalledWith(
                    expect.objectContaining({keyword: 'golang'}),
                );
            });
        });
    });

    // ========================================================================
    // Pagination
    // ========================================================================

    describe('Pagination', () => {
        it('should show pagination when total pages > 1', async () => {
            mockedArticleApi.list.mockResolvedValue({
                items: mockArticles,
                total: 30,
                page: 1,
                page_size: 12,
            });

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText(/Page 1 of 3/)).toBeInTheDocument();
            });
            expect(screen.getByText('Previous')).toBeInTheDocument();
            expect(screen.getByText('Next')).toBeInTheDocument();
        });

        it('should disable Previous button on first page', async () => {
            mockedArticleApi.list.mockResolvedValue({
                items: mockArticles,
                total: 30,
                page: 1,
                page_size: 12,
            });

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Previous')).toBeDisabled();
            });
        });

        it('should not show pagination when only one page', async () => {
            mockedArticleApi.list.mockResolvedValue(mockListResponse);

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('Getting Started with Go')).toBeInTheDocument();
            });

            expect(screen.queryByText('Previous')).not.toBeInTheDocument();
            expect(screen.queryByText('Next')).not.toBeInTheDocument();
        });
    });

    // ========================================================================
    // Article Card Content
    // ========================================================================

    describe('Article Card Content', () => {
        it('should render link to article detail page', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                const links = screen.getAllByTestId('link');
                expect(links.length).toBeGreaterThan(0);
            });
        });

        it('should render video duration for articles with media', async () => {
            render(<ArticleListPage/>);

            await waitFor(() => {
                // Duration 600s = 10:00
                expect(screen.getByText('10:00')).toBeInTheDocument();
            });
        });

        it('should limit tags to 3 with overflow indicator', async () => {
            const articleWithManyTags: Article = {
                ...mockArticles[0],
                id: '4',
                slug: 'many-tags',
                tags: ['tag1', 'tag2', 'tag3', 'tag4', 'tag5'],
            };
            mockedArticleApi.list.mockResolvedValue({
                items: [articleWithManyTags],
                total: 1,
                page: 1,
                page_size: 12,
            });

            render(<ArticleListPage/>);

            await waitFor(() => {
                expect(screen.getByText('tag1')).toBeInTheDocument();
                expect(screen.getByText('tag2')).toBeInTheDocument();
                expect(screen.getByText('tag3')).toBeInTheDocument();
                expect(screen.getByText('+2')).toBeInTheDocument();
            });
        });
    });
});
