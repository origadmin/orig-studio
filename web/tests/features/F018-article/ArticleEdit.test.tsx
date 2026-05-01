/**
 * F018 ArticleEdit Component Tests
 *
 * Tests for the admin ArticleEdit page component.
 * Covers: create mode, edit mode, form filling, save, dirty tracking.
 */

import React from 'react';
import {render, screen, waitFor, fireEvent, act} from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock TanStack Router
jest.mock('@tanstack/react-router', () => ({
    useParams: ({strict}: { strict?: boolean }) => {
        return (global as any).__mockParams || {id: undefined};
    },
    useNavigate: () => jest.fn(),
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
        get: jest.fn().mockResolvedValue({
            id: 'article-001',
            title: 'Existing Article',
            slug: 'existing-article',
            content: '# Hello\n\nWorld',
            summary: 'An existing article',
            state: 'draft',
            user_id: 'user-001',
            category_id: 1,
            media_id: 'media-001',
            thumbnail: '/thumbs/existing.jpg',
            tags: ['golang', 'test'],
            view_count: 100,
            comment_count: 5,
            featured: false,
            create_time: '2026-04-15T10:00:00Z',
            update_time: '2026-04-20T12:00:00Z',
            media: {
                id: 'media-001',
                title: 'Test Video',
                thumbnail: '/thumbs/video.jpg',
                duration: 300,
                type: 'video',
                short_token: 'abc123',
            },
        }),
        create: jest.fn().mockResolvedValue({id: 'new-id'}),
        update: jest.fn().mockResolvedValue({}),
        delete: jest.fn().mockResolvedValue(undefined),
        updateState: jest.fn().mockResolvedValue(undefined),
    },
}));

// Mock media API - must return resolved promises
jest.mock('../../../src/lib/api/media', () => ({
    adminMediaApi: {
        list: jest.fn().mockResolvedValue({items: [], total: 0}),
    },
}));

// Mock category queries
jest.mock('../../../src/hooks/queries', () => ({
    useCategoryList: () => ({
        data: {
            items: [
                {id: 1, name: 'Technology'},
                {id: 2, name: 'Design'},
            ],
        },
    }),
}));

// Mock format utility
jest.mock('../../../src/lib/format', () => ({
    formatDateTime: jest.fn((date: string) => '2026-04-15 10:00:00'),
}));

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

// Mock toast
jest.mock('sonner', () => ({
    toast: {
        success: jest.fn(),
        error: jest.fn(),
    },
}));

// Mock EditPageHeader - renders badges to enable testing
jest.mock('../../../src/components/common/EditPageHeader', () => ({
    EditPageHeader: ({title, isDirty, onSave, onBack, badges}: {
        title: string;
        isDirty: boolean;
        onSave: () => void;
        onBack: () => void;
        badges?: Array<{ type: string; variant: string; label: string }>;
    }) => (
        <div data-testid="edit-page-header">
            <span data-testid="header-title">{title}</span>
            <span data-testid="header-dirty">{isDirty ? 'dirty' : 'clean'}</span>
            {badges && badges.map((badge, i) => (
                <span key={i} data-testid={`header-badge-${i}`}>{badge.label}</span>
            ))}
            <button onClick={onSave} data-testid="save-btn">Save</button>
            <button onClick={onBack} data-testid="back-btn">Back</button>
        </div>
    ),
}));

// Mock DeleteConfirmDialog
jest.mock('../../../src/components/common/DeleteConfirmDialog', () => ({
    DeleteConfirmDialog: ({open, title}: { open: boolean; title: string }) => (
        open ? <div data-testid="delete-dialog">{title}</div> : null
    ),
}));

// Mock useEditPage hooks
jest.mock('../../../src/hooks/useEditPage', () => ({
    useDirtyState: (initialData: any) => {
        const [form, setForm] = React.useState(initialData);
        const initialRef = React.useRef(initialData);
        const isDirty = JSON.stringify(form) !== JSON.stringify(initialRef.current);
        const resetDirty = () => {
            initialRef.current = {...form};
        };
        const syncFromData = (data: any) => {
            setForm(data);
            initialRef.current = {...data};
        };
        return {form, setForm, isDirty, resetDirty, syncFromData};
    },
    useSaveState: () => {
        const [saveState, setSaveState] = React.useState<'idle' | 'saving' | 'success' | 'error'>('idle');
        return {
            saveState,
            isSaving: saveState === 'saving',
            setSaving: () => setSaveState('saving'),
            setSuccess: () => setSaveState('success'),
            setError: () => setSaveState('error'),
            reset: () => setSaveState('idle'),
        };
    },
    useKeyboardShortcut: jest.fn(),
}));

import {adminArticleApi, type Article} from '../../../src/lib/api/article';
import {adminMediaApi} from '../../../src/lib/api/media';
import ArticleEditPage from '../../../src/pages/admin/ArticleEdit';

const mockedAdminArticleApi = adminArticleApi as jest.Mocked<typeof adminArticleApi>;
const mockedAdminMediaApi = adminMediaApi as jest.Mocked<typeof adminMediaApi>;

// Test fixtures
const mockArticle: Article = {
    id: 'article-001',
    title: 'Existing Article',
    slug: 'existing-article',
    content: '# Hello\n\nWorld',
    summary: 'An existing article',
    state: 'draft',
    user_id: 'user-001',
    category_id: 1,
    media_id: 'media-001',
    thumbnail: '/thumbs/existing.jpg',
    tags: ['golang', 'test'],
    view_count: 100,
    comment_count: 5,
    featured: false,
    create_time: '2026-04-15T10:00:00Z',
    update_time: '2026-04-20T12:00:00Z',
    media: {
        id: 'media-001',
        title: 'Test Video',
        thumbnail: '/thumbs/video.jpg',
        duration: 300,
        type: 'video',
        short_token: 'abc123',
    },
};

describe('F018: ArticleEdit Component', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        (global as any).__mockParams = {id: undefined};

        // Ensure default mock implementations return resolved promises
        mockedAdminArticleApi.get.mockResolvedValue(mockArticle);
        mockedAdminArticleApi.create.mockResolvedValue({id: 'new-id', ...mockArticle});
        mockedAdminArticleApi.update.mockResolvedValue(mockArticle);
        mockedAdminArticleApi.delete.mockResolvedValue(undefined);
        mockedAdminMediaApi.list.mockResolvedValue({items: [], total: 0});
    });

    afterEach(() => {
        delete (global as any).__mockParams;
    });

    // ========================================================================
    // Create Mode
    // ========================================================================

    describe('Create Mode', () => {
        it('should render create mode with "Create Article" title', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByTestId('header-title')).toHaveTextContent('Create Article');
            });
        });

        it('should render form fields in create mode', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Title')).toBeInTheDocument();
            });
            expect(screen.getByLabelText('Slug')).toBeInTheDocument();
            expect(screen.getByLabelText('Content (Markdown)')).toBeInTheDocument();
            expect(screen.getByLabelText('Summary')).toBeInTheDocument();
            expect(screen.getByLabelText('Tags (comma-separated)')).toBeInTheDocument();
        });

        it('should show Draft badge in create mode via header badges', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByTestId('header-badge-0')).toHaveTextContent('Draft');
            });
        });

        it('should show "No video associated" when no media selected', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByText('No video associated')).toBeInTheDocument();
            });
        });

        it('should show "Select Video" button', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByText('Select Video')).toBeInTheDocument();
            });
        });

        it('should start with clean dirty state', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByTestId('header-dirty')).toHaveTextContent('clean');
            });
        });

        it('should track dirty state when form is modified', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Title')).toBeInTheDocument();
            });

            const titleInput = screen.getByLabelText('Title');
            await act(async () => {
                fireEvent.change(titleInput, {target: {value: 'New Title'}});
            });

            expect(screen.getByTestId('header-dirty')).toHaveTextContent('dirty');
        });
    });

    // ========================================================================
    // Edit Mode
    // ========================================================================

    describe('Edit Mode', () => {
        beforeEach(() => {
            (global as any).__mockParams = {id: 'article-001'};
        });

        it('should render edit mode with article title', async () => {
            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByTestId('header-title')).toHaveTextContent('Existing Article');
            });
        });

        it('should load article data and populate form', async () => {
            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Title')).toHaveValue('Existing Article');
            });
        });

        it('should show metadata card in edit mode', async () => {
            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByText('Metadata')).toBeInTheDocument();
            });
        });

        it('should show associated video when article has media', async () => {
            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByText('Test Video')).toBeInTheDocument();
            });
        });

        it('should show Draft badge in edit mode via header badges', async () => {
            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByTestId('header-badge-0')).toHaveTextContent('Draft');
            });
        });

        it('should show Quick Actions section', async () => {
            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByText('Quick Actions')).toBeInTheDocument();
            });
        });
    });

    // ========================================================================
    // Form Interaction
    // ========================================================================

    describe('Form Interaction', () => {
        it('should update title field', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Title')).toBeInTheDocument();
            });

            const titleInput = screen.getByLabelText('Title');
            await act(async () => {
                fireEvent.change(titleInput, {target: {value: 'My New Article'}});
            });

            expect(titleInput).toHaveValue('My New Article');
        });

        it('should update content field', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Content (Markdown)')).toBeInTheDocument();
            });

            const contentTextarea = screen.getByLabelText('Content (Markdown)');
            await act(async () => {
                fireEvent.change(contentTextarea, {target: {value: '# Hello World'}});
            });

            expect(contentTextarea).toHaveValue('# Hello World');
        });

        it('should update tags and show tag badges', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Tags (comma-separated)')).toBeInTheDocument();
            });

            const tagsInput = screen.getByLabelText('Tags (comma-separated)');
            await act(async () => {
                fireEvent.change(tagsInput, {target: {value: 'golang, react, testing'}});
            });

            expect(screen.getByText('golang')).toBeInTheDocument();
            expect(screen.getByText('react')).toBeInTheDocument();
            expect(screen.getByText('testing')).toBeInTheDocument();
        });

        it('should switch between Content and Publish Settings tabs', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByText('Content')).toBeInTheDocument();
            });

            const publishTab = screen.getByText('Publish Settings');
            await act(async () => {
                fireEvent.click(publishTab);
            });

            expect(screen.getByText('State')).toBeInTheDocument();
            expect(screen.getByText('Category')).toBeInTheDocument();
        });

        it('should show Featured checkbox in Publish Settings tab', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByText('Publish Settings')).toBeInTheDocument();
            });

            await act(async () => {
                fireEvent.click(screen.getByText('Publish Settings'));
            });

            expect(screen.getByLabelText('Featured Article')).toBeInTheDocument();
        });
    });

    // ========================================================================
    // Save Operation
    // ========================================================================

    describe('Save Operation', () => {
        it('should call adminArticleApi.create in create mode', async () => {
            render(<ArticleEditPage mode="create"/>);

            await waitFor(() => {
                expect(screen.getByLabelText('Title')).toBeInTheDocument();
            });

            // Fill in required fields
            await act(async () => {
                fireEvent.change(screen.getByLabelText('Title'), {target: {value: 'New Article'}});
            });
            await act(async () => {
                fireEvent.change(screen.getByLabelText('Content (Markdown)'), {target: {value: 'Content'}});
            });

            // Click save
            await act(async () => {
                fireEvent.click(screen.getByTestId('save-btn'));
            });

            await waitFor(() => {
                expect(mockedAdminArticleApi.create).toHaveBeenCalled();
            });
        });
    });

    // ========================================================================
    // Edit Mode Loading State
    // ========================================================================

    describe('Edit Mode Loading State', () => {
        it('should show loading spinner while fetching article', () => {
            (global as any).__mockParams = {id: 'article-001'};
            mockedAdminArticleApi.get.mockReturnValue(new Promise(() => {
            }));

            render(<ArticleEditPage mode="edit"/>);

            expect(document.querySelector('.animate-spin, [data-slot="spinner"]')).toBeTruthy();
        });
    });

    // ========================================================================
    // Edit Mode Error State
    // ========================================================================

    describe('Edit Mode Error State', () => {
        it('should show error when article fails to load', async () => {
            (global as any).__mockParams = {id: 'article-001'};
            mockedAdminArticleApi.get.mockRejectedValueOnce(new Error('Not found'));

            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByText('Failed to load article')).toBeInTheDocument();
            });
        });

        it('should show Back to List button on error', async () => {
            (global as any).__mockParams = {id: 'article-001'};
            mockedAdminArticleApi.get.mockRejectedValueOnce(new Error('Not found'));

            render(<ArticleEditPage mode="edit"/>);

            await waitFor(() => {
                expect(screen.getByText('Back to List')).toBeInTheDocument();
            });
        });
    });
});
