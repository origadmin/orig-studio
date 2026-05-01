/**
 * F018 Article API Client Tests
 *
 * Tests for articleApi and adminArticleApi client methods.
 * Verifies correct HTTP method, URL, and parameter passing.
 */

import {articleApi, adminArticleApi, type Article, type CreateArticleRequest, type UpdateArticleRequest} from '../../../src/lib/api/article';

// Mock request module
jest.mock('../../../src/lib/request', () => ({
    api: {
        get: jest.fn(),
        post: jest.fn(),
        put: jest.fn(),
        patch: jest.fn(),
        del: jest.fn(),
    },
    API_BASE_URL: '',
}));

import {api} from '../../../src/lib/request';

const mockedApi = api as jest.Mocked<typeof api>;

// Test fixtures
const mockArticle: Article = {
    id: 'article-001',
    title: 'Test Article',
    slug: 'test-article',
    content: '# Hello World\n\nThis is a test article.',
    summary: 'A test article summary',
    state: 'published',
    user_id: 'user-001',
    category_id: 1,
    media_id: 'media-001',
    thumbnail: '/thumbnails/test.jpg',
    tags: ['test', 'golang'],
    view_count: 42,
    comment_count: 5,
    featured: true,
    published_at: '2026-04-15T10:00:00Z',
    create_time: '2026-04-15T10:00:00Z',
    update_time: '2026-04-20T12:00:00Z',
    media: {
        id: 'media-001',
        title: 'Test Video',
        thumbnail: '/thumbnails/video.jpg',
        duration: 300,
        type: 'video',
        short_token: 'abc123',
    },
};

const mockListResponse = {
    items: [mockArticle],
    total: 1,
    page: 1,
    page_size: 12,
};

describe('F018: Article API Client', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    // ========================================================================
    // articleApi (public)
    // ========================================================================

    describe('articleApi.list', () => {
        it('should call GET /articles with default state=published', async () => {
            mockedApi.get.mockResolvedValueOnce(mockListResponse);

            const result = await articleApi.list();

            expect(mockedApi.get).toHaveBeenCalledWith('/articles', {state: 'published'});
            expect(result).toEqual(mockListResponse);
        });

        it('should pass pagination and filter params', async () => {
            mockedApi.get.mockResolvedValueOnce(mockListResponse);

            const params = {page: 2, page_size: 20, category_id: 5, keyword: 'golang'};
            await articleApi.list(params);

            expect(mockedApi.get).toHaveBeenCalledWith('/articles', {
                ...params,
                state: 'published',
            });
        });

        it('should use provided state when explicitly set', async () => {
            mockedApi.get.mockResolvedValueOnce(mockListResponse);

            await articleApi.list({state: 'draft'});

            expect(mockedApi.get).toHaveBeenCalledWith('/articles', {state: 'draft'});
        });

        it('should default to state=published when state is empty string', async () => {
            mockedApi.get.mockResolvedValueOnce(mockListResponse);

            await articleApi.list({state: ''});

            expect(mockedApi.get).toHaveBeenCalledWith('/articles', {state: 'published'});
        });
    });

    describe('articleApi.get', () => {
        it('should call GET /articles/:id', async () => {
            mockedApi.get.mockResolvedValueOnce(mockArticle);

            const result = await articleApi.get('article-001');

            expect(mockedApi.get).toHaveBeenCalledWith('/articles/article-001');
            expect(result).toEqual(mockArticle);
        });
    });

    describe('articleApi.getBySlug', () => {
        it('should call GET /articles/slug/:slug', async () => {
            mockedApi.get.mockResolvedValueOnce(mockArticle);

            const result = await articleApi.getBySlug('test-article');

            expect(mockedApi.get).toHaveBeenCalledWith('/articles/slug/test-article');
            expect(result).toEqual(mockArticle);
        });
    });

    describe('articleApi.featured', () => {
        it('should call GET /articles/featured without limit', async () => {
            mockedApi.get.mockResolvedValueOnce([mockArticle]);

            const result = await articleApi.featured();

            expect(mockedApi.get).toHaveBeenCalledWith('/articles/featured', {limit: undefined});
            expect(result).toEqual([mockArticle]);
        });

        it('should pass limit parameter', async () => {
            mockedApi.get.mockResolvedValueOnce([mockArticle]);

            await articleApi.featured(5);

            expect(mockedApi.get).toHaveBeenCalledWith('/articles/featured', {limit: 5});
        });
    });

    describe('articleApi.latest', () => {
        it('should call GET /articles/latest without limit', async () => {
            mockedApi.get.mockResolvedValueOnce([mockArticle]);

            const result = await articleApi.latest();

            expect(mockedApi.get).toHaveBeenCalledWith('/articles/latest', {limit: undefined});
            expect(result).toEqual([mockArticle]);
        });

        it('should pass limit parameter', async () => {
            mockedApi.get.mockResolvedValueOnce([mockArticle]);

            await articleApi.latest(10);

            expect(mockedApi.get).toHaveBeenCalledWith('/articles/latest', {limit: 10});
        });
    });

    // ========================================================================
    // adminArticleApi
    // ========================================================================

    describe('adminArticleApi.adminList', () => {
        it('should call GET /admin/articles with params', async () => {
            mockedApi.get.mockResolvedValueOnce(mockListResponse);

            const params = {page: 1, page_size: 10, state: 'draft', keyword: 'test'};
            const result = await adminArticleApi.adminList(params);

            expect(mockedApi.get).toHaveBeenCalledWith('/admin/articles', params);
            expect(result).toEqual(mockListResponse);
        });

        it('should call GET /admin/articles without params', async () => {
            mockedApi.get.mockResolvedValueOnce(mockListResponse);

            await adminArticleApi.adminList();

            expect(mockedApi.get).toHaveBeenCalledWith('/admin/articles', undefined);
        });
    });

    describe('adminArticleApi.get', () => {
        it('should call GET /admin/articles/:id', async () => {
            mockedApi.get.mockResolvedValueOnce(mockArticle);

            const result = await adminArticleApi.get('article-001');

            expect(mockedApi.get).toHaveBeenCalledWith('/admin/articles/article-001');
            expect(result).toEqual(mockArticle);
        });
    });

    describe('adminArticleApi.create', () => {
        it('should call POST /admin/articles with create data', async () => {
            const createData: CreateArticleRequest = {
                title: 'New Article',
                slug: 'new-article',
                content: 'Article content',
                summary: 'Article summary',
                state: 'draft',
                category_id: 1,
                tags: ['new'],
                featured: false,
            };
            mockedApi.post.mockResolvedValueOnce({...mockArticle, ...createData, id: 'article-002'});

            const result = await adminArticleApi.create(createData);

            expect(mockedApi.post).toHaveBeenCalledWith('/admin/articles', createData);
            expect(result.id).toBe('article-002');
        });

        it('should send minimal create request', async () => {
            const minimalData: CreateArticleRequest = {
                title: 'Minimal Article',
                content: 'Content',
            };
            mockedApi.post.mockResolvedValueOnce({...mockArticle, id: 'article-003'});

            await adminArticleApi.create(minimalData);

            expect(mockedApi.post).toHaveBeenCalledWith('/admin/articles', minimalData);
        });
    });

    describe('adminArticleApi.update', () => {
        it('should call PUT /admin/articles/:id with update data', async () => {
            const updateData: UpdateArticleRequest = {
                title: 'Updated Title',
                state: 'published',
            };
            mockedApi.put.mockResolvedValueOnce({...mockArticle, ...updateData});

            const result = await adminArticleApi.update('article-001', updateData);

            expect(mockedApi.put).toHaveBeenCalledWith('/admin/articles/article-001', updateData);
            expect(result.title).toBe('Updated Title');
        });
    });

    describe('adminArticleApi.delete', () => {
        it('should call DELETE /admin/articles/:id', async () => {
            mockedApi.del.mockResolvedValueOnce(undefined);

            await adminArticleApi.delete('article-001');

            expect(mockedApi.del).toHaveBeenCalledWith('/admin/articles/article-001');
        });
    });

    describe('adminArticleApi.updateState', () => {
        it('should call PATCH /admin/articles/:id/state with state', async () => {
            mockedApi.patch.mockResolvedValueOnce(undefined);

            await adminArticleApi.updateState('article-001', 'published');

            expect(mockedApi.patch).toHaveBeenCalledWith('/admin/articles/article-001/state', {state: 'published'});
        });

        it('should send draft state', async () => {
            mockedApi.patch.mockResolvedValueOnce(undefined);

            await adminArticleApi.updateState('article-001', 'draft');

            expect(mockedApi.patch).toHaveBeenCalledWith('/admin/articles/article-001/state', {state: 'draft'});
        });
    });

    // ========================================================================
    // Type structure verification
    // ========================================================================

    describe('Article type structure', () => {
        it('should have all required fields', () => {
            const article: Article = mockArticle;
            expect(article.id).toBeDefined();
            expect(article.title).toBeDefined();
            expect(article.slug).toBeDefined();
            expect(article.content).toBeDefined();
            expect(article.state).toBeDefined();
            expect(article.user_id).toBeDefined();
            expect(article.view_count).toBeDefined();
            expect(article.comment_count).toBeDefined();
            expect(article.featured).toBeDefined();
            expect(article.create_time).toBeDefined();
        });

        it('should have optional fields as undefined', () => {
            const minimal: Article = {
                id: '1',
                title: 'Test',
                slug: 'test',
                content: '',
                state: 'draft',
                user_id: 'u1',
                view_count: 0,
                comment_count: 0,
                featured: false,
                create_time: '2026-01-01T00:00:00Z',
            };
            expect(minimal.summary).toBeUndefined();
            expect(minimal.category_id).toBeUndefined();
            expect(minimal.media_id).toBeUndefined();
            expect(minimal.thumbnail).toBeUndefined();
            expect(minimal.tags).toBeUndefined();
            expect(minimal.published_at).toBeUndefined();
            expect(minimal.update_time).toBeUndefined();
            expect(minimal.media).toBeUndefined();
        });

        it('should support media brief object', () => {
            const article: Article = mockArticle;
            expect(article.media).toBeDefined();
            expect(article.media?.id).toBe('media-001');
            expect(article.media?.title).toBe('Test Video');
            expect(article.media?.duration).toBe(300);
            expect(article.media?.short_token).toBe('abc123');
        });
    });
});
