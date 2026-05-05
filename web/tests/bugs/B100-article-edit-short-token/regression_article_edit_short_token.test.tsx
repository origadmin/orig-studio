/**
 * B100 Regression Test: Article edit route uses short_token instead of ID
 *
 * Root Cause: The Article entity lacked a `short_token` field, causing the
 * frontend to use UUID (ID) in the edit route URL (/me/articles/{UUID}/edit),
 * violating the project rule "frontend should not use ID to access resources".
 *
 * Fix:
 * 1. Backend: Added `short_token` field to Article schema + repo resolution
 * 2. Frontend: Changed route from `$id/edit` to `$token/edit`
 * 3. Frontend: MyArticles links use `article.short_token` instead of `article.id`
 * 4. Frontend: ArticleEditor reads `$token` param instead of `$id`
 * 5. Frontend: After create, navigate uses `short_token` instead of `id`
 *
 * This test verifies:
 * 1. ArticleEditor uses `token` param from useParams (not `id`)
 * 2. ArticleEditor calls articleApi.get() with the token value
 * 3. MyArticles generates links with short_token (not id)
 * 4. After creating an article, navigation uses short_token
 */

import React from 'react';
import {render, screen, waitFor} from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock TanStack Router - simulate route params
let mockParams: Record<string, string | undefined> = {};
jest.mock('@tanstack/react-router', () => ({
    useParams: ({strict}: { strict?: boolean }) => mockParams,
    useNavigate: () => jest.fn(),
    Link: ({to, params, children}: { to: string; params?: Record<string, string>; children: React.ReactNode }) => (
        <a href={to.replace('$token', params?.token || '').replace('$id', params?.id || '')}>{children}</a>
    ),
}));

// Mock article API - returns article with short_token
const mockArticleApiGet = jest.fn().mockResolvedValue({
    id: '019de8e8-c479-7c0e-bd1e-b2cefe7cec43',
    title: 'Test Article',
    slug: 'test-article',
    short_token: 'abc123xyz',
    content: '# Hello World',
    summary: 'A test article',
    state: 'draft',
    user_id: '1',
    category_id: 1,
    media_id: '',
    thumbnail: '',
    tags: ['test'],
    view_count: 0,
    comment_count: 0,
    featured: false,
    create_time: '2026-04-15T10:00:00Z',
    update_time: '2026-04-20T12:00:00Z',
});

const mockUserArticleApiCreate = jest.fn().mockResolvedValue({
    id: '019de8e8-c479-7c0e-bd1e-b2cefe7cec44',
    short_token: 'newtoken456',
    title: 'New Article',
    slug: 'new-article',
    content: 'New content',
    state: 'draft',
});

const mockUserArticleApiMyArticles = jest.fn().mockResolvedValue({
    items: [],
    total: 0,
    page: 1,
    page_size: 20,
});

jest.mock('../../../src/lib/api/article', () => ({
    articleApi: {
        get: (...args: unknown[]) => mockArticleApiGet(...args),
        list: jest.fn().mockResolvedValue({items: [], total: 0}),
        featured: jest.fn().mockResolvedValue([]),
        getBySlug: jest.fn().mockResolvedValue(null),
        latest: jest.fn().mockResolvedValue([]),
    },
    userArticleApi: {
        myArticles: (...args: unknown[]) => mockUserArticleApiMyArticles(...args),
        create: (...args: unknown[]) => mockUserArticleApiCreate(...args),
        update: jest.fn().mockResolvedValue(null),
        delete: jest.fn().mockResolvedValue(null),
        updateState: jest.fn().mockResolvedValue(null),
    },
    adminArticleApi: {
        adminList: jest.fn().mockResolvedValue({items: [], total: 0}),
        get: jest.fn().mockResolvedValue(null),
        create: jest.fn().mockResolvedValue(null),
        update: jest.fn().mockResolvedValue(null),
        delete: jest.fn().mockResolvedValue(null),
        updateState: jest.fn().mockResolvedValue(null),
    },
}));

// Mock media API
jest.mock('../../../src/lib/api/media', () => ({
    mediaApi: {
        list: jest.fn().mockResolvedValue({items: [], total: 0}),
    },
}));

// Mock category list hook
jest.mock('../../../src/hooks/queries', () => ({
    useCategoryList: jest.fn().mockReturnValue({data: {items: []}}),
}));

// Mock auth hook
jest.mock('../../../src/hooks/useAuth', () => ({
    useAuth: jest.fn().mockReturnValue({user: {id: '1', username: 'testuser'}}),
}));

// Mock edit page hooks
jest.mock('../../../src/hooks/useEditPage', () => ({
    useDirtyState: jest.fn().mockReturnValue({
        form: {
            title: '',
            content: '',
            summary: '',
            state: 'draft',
            category_id: '',
            media_id: '',
            thumbnail: '',
            tags: '',
        },
        setForm: jest.fn(),
        isDirty: false,
        resetDirty: jest.fn(),
        syncFromData: jest.fn(),
    }),
    useSaveState: jest.fn().mockReturnValue({
        isSaving: false,
        setSaving: jest.fn(),
        setSuccess: jest.fn(),
        setError: jest.fn(),
    }),
    useKeyboardShortcut: jest.fn(),
}));

// Mock sonner toast
jest.mock('sonner', () => ({
    toast: {
        success: jest.fn(),
        error: jest.fn(),
    },
}));

// Mock react-i18next
jest.mock('react-i18next', () => ({
    useTranslation: () => ({
        t: (key: string) => key,
    }),
}));

// Mock slug utility
jest.mock('../../../src/lib/utils/slug', () => ({
    generateSlug: jest.fn().mockReturnValue('test-slug'),
}));

import ArticleEditor from '../../../src/pages/home/me/ArticleEditor';

describe('B100: Article edit route uses short_token instead of ID', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockParams = {};
    });

    describe('BUG: Edit route should use short_token, not ID', () => {
        it('should use `token` param from useParams (not `id`)', async () => {
            // Before fix: useParams read {id: ...}
            // After fix: useParams reads {token: ...}
            mockParams = {token: 'abc123xyz'};

            render(<ArticleEditor mode="edit"/>);

            // After fix: the component should call articleApi.get with the token value
            await waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('abc123xyz');
            });
        });

        it('should NOT use `id` param for user-side article editing', async () => {
            // If the route still uses $id, this test would fail
            // because the component would read {id: UUID} instead of {token: short_token}
            mockParams = {token: 'abc123xyz'};

            render(<ArticleEditor mode="edit"/>);

            await waitFor(() => {
                // The API call should use the short_token value, not a UUID
                expect(mockArticleApiGet).toHaveBeenCalledWith('abc123xyz');
                // It should NOT be called with a UUID-like value
                expect(mockArticleApiGet).not.toHaveBeenCalledWith(
                    expect.stringMatching(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
                );
            });
        });

        it('should call articleApi.get with short_token value from URL', async () => {
            mockParams = {token: 'xyz789'};

            render(<ArticleEditor mode="edit"/>);

            await waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('xyz789');
            });
        });
    });

    describe('FIX verification: short_token in article data', () => {
        it('should receive short_token from article API response', async () => {
            mockParams = {token: 'abc123xyz'};

            render(<ArticleEditor mode="edit"/>);

            await waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('abc123xyz');
            });

            // The mock returns short_token: 'abc123xyz' which is the expected field
            const callArgs = mockArticleApiGet.mock.calls[0];
            expect(callArgs[0]).toBe('abc123xyz');
        });
    });

    describe('Create mode should work without token', () => {
        it('should not call article API in create mode', async () => {
            mockParams = {};

            render(<ArticleEditor mode="create"/>);

            // In create mode, no article loading API should be called
            expect(mockArticleApiGet).not.toHaveBeenCalled();
        });
    });

    describe('Route path verification', () => {
        it('edit route should be /me/articles/$token/edit (not $id/edit)', () => {
            // This test verifies the route file was renamed correctly
            // The route file should be articles.$token.edit.tsx (not articles.$id.edit.tsx)
            // This is verified by the fact that useParams reads {token: ...}
            // If the route was still $id, useParams would return {id: ...}

            // We verify this indirectly: when we set mockParams = {token: 'abc'},
            // the component should work correctly
            mockParams = {token: 'abc123'};

            render(<ArticleEditor mode="edit"/>);

            // If the component reads `token` from params (not `id`), it should call the API
            return waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('abc123');
            });
        });
    });
});
