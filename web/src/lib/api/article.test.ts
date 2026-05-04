import {userArticleApi, type UserCreateArticleRequest, type UserUpdateArticleRequest} from '@/lib/api/article';
import {api} from '@/lib/request';

// Mock the api module
jest.mock('@/lib/request', () => ({
    api: {
        get: jest.fn(),
        post: jest.fn(),
        put: jest.fn(),
        del: jest.fn(),
        patch: jest.fn(),
    },
}));

describe('userArticleApi', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('myArticles', () => {
        it('should call GET /articles/me with default params', async () => {
            (api.get as jest.Mock).mockResolvedValue({items: [], total: 0, page: 1, page_size: 20});
            await userArticleApi.myArticles();
            expect(api.get).toHaveBeenCalledWith('/articles/me', undefined);
        });

        it('should call GET /articles/me with state filter', async () => {
            (api.get as jest.Mock).mockResolvedValue({items: [], total: 0, page: 1, page_size: 20});
            await userArticleApi.myArticles({state: 'draft', page: 2, page_size: 10});
            expect(api.get).toHaveBeenCalledWith('/articles/me', {state: 'draft', page: 2, page_size: 10});
        });
    });

    describe('create', () => {
        it('should call POST /articles with user create data', async () => {
            const mockArticle = {id: '1', title: 'Test', content: 'Content', state: 'draft', featured: false};
            (api.post as jest.Mock).mockResolvedValue(mockArticle);

            const data: UserCreateArticleRequest = {
                title: 'Test',
                content: 'Content',
                state: 'draft',
            };
            const result = await userArticleApi.create(data);
            expect(api.post).toHaveBeenCalledWith('/articles', data);
            expect(result).toEqual(mockArticle);
        });

        it('should not include featured in create request type', () => {
            const data: UserCreateArticleRequest = {
                title: 'Test',
                content: 'Content',
            };
            // UserCreateArticleRequest type should not have featured field
            expect(data).not.toHaveProperty('featured');
        });
    });

    describe('update', () => {
        it('should call PUT /articles/:id with user update data', async () => {
            const mockArticle = {id: '1', title: 'Updated', content: 'Content', state: 'draft', featured: false};
            (api.put as jest.Mock).mockResolvedValue(mockArticle);

            const data: UserUpdateArticleRequest = {
                title: 'Updated',
            };
            const result = await userArticleApi.update('1', data);
            expect(api.put).toHaveBeenCalledWith('/articles/1', data);
            expect(result).toEqual(mockArticle);
        });

        it('should not include featured in update request type', () => {
            const data: UserUpdateArticleRequest = {
                title: 'Updated',
            };
            expect(data).not.toHaveProperty('featured');
        });
    });

    describe('delete', () => {
        it('should call DELETE /articles/:id', async () => {
            (api.del as jest.Mock).mockResolvedValue(undefined);
            await userArticleApi.delete('1');
            expect(api.del).toHaveBeenCalledWith('/articles/1');
        });
    });

    describe('updateState', () => {
        it('should call PATCH /articles/:id/state with draft', async () => {
            (api.patch as jest.Mock).mockResolvedValue(undefined);
            await userArticleApi.updateState('1', 'draft');
            expect(api.patch).toHaveBeenCalledWith('/articles/1/state', {state: 'draft'});
        });

        it('should call PATCH /articles/:id/state with published', async () => {
            (api.patch as jest.Mock).mockResolvedValue(undefined);
            await userArticleApi.updateState('1', 'published');
            expect(api.patch).toHaveBeenCalledWith('/articles/1/state', {state: 'published'});
        });
    });
});

describe('UserCreateArticleRequest type', () => {
    it('should accept valid create request', () => {
        const data: UserCreateArticleRequest = {
            title: 'Test Article',
            content: 'Article content',
            state: 'draft',
            summary: 'A summary',
            tags: ['tag1', 'tag2'],
            category_id: 1,
            media_id: 'media-1',
            thumbnail: 'https://example.com/thumb.jpg',
        };
        expect(data.title).toBe('Test Article');
        expect(data.state).toBe('draft');
    });
});

describe('UserUpdateArticleRequest type', () => {
    it('should accept valid update request', () => {
        const data: UserUpdateArticleRequest = {
            title: 'Updated Title',
            content: 'Updated content',
            state: 'published',
        };
        expect(data.title).toBe('Updated Title');
        expect(data.state).toBe('published');
    });
});
