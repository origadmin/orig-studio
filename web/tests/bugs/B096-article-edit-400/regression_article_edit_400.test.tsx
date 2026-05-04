/**
 * B096 Regression Test: Article edit page infinite spinner (400 Bad Request)
 *
 * Root Cause: The route component articles.$id.edit.tsx renders <Page mode="edit"/>
 * but does NOT pass the `id` route param to ArticleEditor. The component expects
 * `articleId` as a prop but receives `undefined`, causing the useEffect that loads
 * article data to return early, leaving the component in infinite loading state.
 *
 * Fix: ArticleEditor uses useParams() to get the `id` from the URL, and uses
 * articleApi.get(id) instead of userArticleApi.myArticles() + find.
 *
 * This test verifies:
 * 1. ArticleEditor correctly obtains articleId from useParams when in edit mode
 * 2. Article data is loaded via articleApi.get(id) when articleId is present
 * 3. The component does NOT call userArticleApi.myArticles() for single article load
 */

import React from 'react';
import {render, screen, waitFor} from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock TanStack Router - simulate route params
let mockParams: Record<string, string | undefined> = {};
jest.mock('@tanstack/react-router', () => ({
    useParams: ({strict}: { strict?: boolean }) => mockParams,
    useNavigate: () => jest.fn(),
}));

// Mock article API - all methods return resolved promises by default
const mockArticleApiGet = jest.fn().mockResolvedValue({
    id: 'article-001',
    title: 'Test Article',
    slug: 'test-article',
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
        create: jest.fn().mockResolvedValue(null),
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

// Mock media API - returns resolved promise
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

describe('B096: Article edit page - route param propagation', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockParams = {};
    });

    describe('BUG reproduction: edit mode without articleId', () => {
        it('should not call any article API when articleId is undefined', async () => {
            // Before fix: useParams was not used, articleId was always undefined
            // After fix: useParams provides the id, but if URL has no id, articleId is undefined
            mockParams = {};

            render(<ArticleEditor mode="edit"/>);

            // Wait for effects to run
            await waitFor(() => {
                // No article API should be called because articleId is undefined
                expect(mockArticleApiGet).not.toHaveBeenCalled();
                expect(mockUserArticleApiMyArticles).not.toHaveBeenCalled();
            });
        });
    });

    describe('FIX verification: edit mode with articleId from useParams', () => {
        it('should call articleApi.get(id) when id param is available in URL', async () => {
            mockParams = {id: 'article-001'};

            render(<ArticleEditor mode="edit"/>);

            // After fix: the component should call articleApi.get with the id from useParams
            await waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('article-001');
            });
        });

        it('should NOT call userArticleApi.myArticles for single article loading', async () => {
            mockParams = {id: 'article-001'};

            render(<ArticleEditor mode="edit"/>);

            await waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('article-001');
            });

            // After fix: myArticles should NOT be called for loading a single article
            // Before fix: the component used myArticles().then(find by id)
            expect(mockUserArticleApiMyArticles).not.toHaveBeenCalled();
        });

        it('should call articleApi.get with correct id for different articles', async () => {
            mockParams = {id: 'article-999'};

            render(<ArticleEditor mode="edit"/>);

            await waitFor(() => {
                expect(mockArticleApiGet).toHaveBeenCalledWith('article-999');
            });
        });
    });

    describe('Create mode should work without articleId', () => {
        it('should not call article API in create mode', async () => {
            mockParams = {};

            render(<ArticleEditor mode="create"/>);

            // In create mode, no article loading API should be called
            expect(mockArticleApiGet).not.toHaveBeenCalled();
            expect(mockUserArticleApiMyArticles).not.toHaveBeenCalled();
        });
    });
});
