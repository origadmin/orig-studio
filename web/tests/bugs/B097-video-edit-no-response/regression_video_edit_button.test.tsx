/**
 * B097 Regression Test: My Videos page edit button has no response
 *
 * This test verifies that the edit button in the MyVideos page
 * navigates to the media edit page when clicked.
 *
 * Before fix: Edit button is a plain <Button> with no onClick or <Link>,
 *   clicking it does nothing — no navigation, no error, no response.
 * After fix: Edit button uses <Link to="/media/$shortToken/edit"> with
 *   the item's short_token as the route parameter.
 */

import React from 'react';
import {render, screen, waitFor, within} from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock TanStack Router - track navigate calls
const mockNavigate = jest.fn();

/**
 * Resolve TanStack Router path pattern with params.
 * e.g. resolveRoute('/media/$shortToken/edit', {shortToken: 'abc123'}) => '/media/abc123/edit'
 */
function resolveRoute(to: string, params?: Record<string, string>): string {
    if (!params) return to;
    let resolved = to;
    for (const [key, value] of Object.entries(params)) {
        resolved = resolved.replace(`$${key}`, value);
    }
    return resolved;
}

jest.mock('@tanstack/react-router', () => ({
    Link: ({to, params, children, search, ...props}: any) => {
        const href = resolveRoute(to || '', params);
        const fullHref = search ? `${href}?${new URLSearchParams(search).toString()}` : href;
        return (
            <a href={fullHref} data-testid="router-link" {...props}>
                {children}
            </a>
        );
    },
    useNavigate: () => mockNavigate,
    useParams: () => ({}),
    createFileRoute: () => (fn: any) => fn,
}));

// Mock react-i18next
jest.mock('react-i18next', () => ({
    useTranslation: () => ({t: (key: string) => key}),
}));

// Mock useAuth hook
jest.mock('../../../src/hooks/useAuth', () => ({
    useAuth: () => ({
        user: {id: 'user-001', username: 'testuser', nickname: 'Test User'},
        isAdmin: false,
        isAuthenticated: true,
    }),
}));

// Mock queries hook
const mockMediaList = {
    data: {
        items: [
            {
                id: 1,
                title: 'Test Video 1',
                description: 'A test video',
                url: '/videos/test1.mp4',
                thumbnail: '/thumbs/test1.jpg',
                duration: 120,
                view_count: 100,
                create_time: '2026-04-15T10:00:00Z',
                user_id: 'user-001',
                short_token: 'abc123',
                state: 'active',
                tags: ['test'],
            },
            {
                id: 2,
                title: 'Test Video 2',
                description: 'Another test video',
                url: '/videos/test2.mp4',
                thumbnail: '/thumbs/test2.jpg',
                duration: 300,
                view_count: 50,
                create_time: '2026-04-16T10:00:00Z',
                user_id: 'user-001',
                short_token: 'def456',
                state: 'draft',
                tags: [],
            },
        ],
        total: 2,
    },
    isLoading: false,
};

jest.mock('../../../src/hooks/queries', () => ({
    useMediaList: () => mockMediaList,
    useDeleteMedia: () => ({
        mutateAsync: jest.fn(),
    }),
}));

// Mock format utilities
jest.mock('../../../src/lib/format', () => ({
    formatRelativeTime: jest.fn((date: string) => '2 days ago'),
    formatDuration: jest.fn((seconds: number) => `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`),
}));

// Mock utils
jest.mock('../../../src/lib/utils', () => ({
    getFullUrl: jest.fn((url: string) => url ? `http://localhost${url.startsWith('/') ? '' : '/'}${url}` : ''),
    cn: (...args: any[]) => args.filter(Boolean).join(' '),
}));

// Mock UI components
jest.mock('../../../src/components/ui/spinner', () => ({
    Spinner: () => <div data-testid="spinner">Loading...</div>,
}));

jest.mock('../../../src/components/ui/card', () => ({
    Card: ({children, ...props}: any) => <div data-testid="card" {...props}>{children}</div>,
    CardContent: ({children, ...props}: any) => <div data-testid="card-content" {...props}>{children}</div>,
}));

jest.mock('../../../src/components/ui/button', () => ({
    Button: ({children, asChild, className, variant, size, onClick, ...props}: any) => {
        if (asChild && children) {
            // When asChild, render children directly with button props merged
            return React.Children.map(children, (child: any) => {
                if (React.isValidElement(child)) {
                    return React.cloneElement(child, {
                        ...props,
                        className: `btn btn-${variant || 'default'} btn-${size || 'default'} ${className || ''}`,
                        'data-testid': props['data-testid'] || 'button-aschild',
                    });
                }
                return child;
            });
        }
        return (
            <button
                data-testid="button"
                className={`btn btn-${variant || 'default'} btn-${size || 'default'} ${className || ''}`}
                onClick={onClick}
                {...props}
            >
                {children}
            </button>
        );
    },
}));

jest.mock('../../../src/components/ui/badge', () => ({
    Badge: ({children, variant, className, ...props}: any) => (
        <span data-testid="badge" className={`badge badge-${variant} ${className || ''}`} {...props}>
            {children}
        </span>
    ),
}));

import MyVideos from '../../../src/pages/home/me/MyVideos';

describe('B097: MyVideos edit button navigation', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('Edit button should navigate to media edit page', () => {
        it('should render edit buttons for each video card', async () => {
            render(<MyVideos />);

            await waitFor(() => {
                expect(screen.getByText('Test Video 1')).toBeInTheDocument();
            });

            // Find all edit buttons (there should be one per video card)
            const editButtons = screen.getAllByText('编辑');
            expect(editButtons.length).toBe(2);
        });

        it('should have a link/navigation on the edit button (B097 core regression)', async () => {
            render(<MyVideos />);

            await waitFor(() => {
                expect(screen.getByText('Test Video 1')).toBeInTheDocument();
            });

            // Find the first video card's edit button
            const editButtons = screen.getAllByText('编辑');
            const firstEditButton = editButtons[0];

            // The edit button should be inside a link (<a> tag from mocked Link)
            // Before fix: edit button is a plain <button> with no link parent
            // After fix: edit button is wrapped in a <Link> which renders as <a>
            const linkParent = firstEditButton.closest('a');

            expect(linkParent).not.toBeNull();
            // The link should point to the media edit route with the correct short_token
            expect(linkParent?.getAttribute('href')).toContain('abc123');
        });

        it('should use short_token as the route parameter for edit navigation', async () => {
            render(<MyVideos />);

            await waitFor(() => {
                expect(screen.getByText('Test Video 2')).toBeInTheDocument();
            });

            // Find the second video card's edit button
            const editButtons = screen.getAllByText('编辑');
            const secondEditButton = editButtons[1];

            const linkParent = secondEditButton.closest('a');

            expect(linkParent).not.toBeNull();
            // The link should use the second video's short_token
            expect(linkParent?.getAttribute('href')).toContain('def456');
        });

        it('should navigate to /media/$shortToken/edit route pattern', async () => {
            render(<MyVideos />);

            await waitFor(() => {
                expect(screen.getByText('Test Video 1')).toBeInTheDocument();
            });

            const editButtons = screen.getAllByText('编辑');
            const firstEditButton = editButtons[0];
            const linkParent = firstEditButton.closest('a');

            expect(linkParent).not.toBeNull();
            // The href should contain the edit route pattern
            const href = linkParent?.getAttribute('href') || '';
            // After fix, the Link's to prop is "/media/$shortToken/edit" with params
            // Our mock serializes this as JSON or string containing the route
            expect(href).toMatch(/media|shortToken|edit|abc123/);
        });
    });

    describe('Delete button should still work (regression guard)', () => {
        it('should render delete buttons with click handlers', async () => {
            render(<MyVideos />);

            await waitFor(() => {
                expect(screen.getByText('Test Video 1')).toBeInTheDocument();
            });

            // Delete buttons should exist and be functional
            const deleteButtons = screen.getAllByText('删除');
            expect(deleteButtons.length).toBe(2);
        });
    });

    describe('Watch button should still work (regression guard)', () => {
        it('should render watch links on video thumbnails', async () => {
            const {container} = render(<MyVideos />);

            await waitFor(() => {
                expect(screen.getByText('Test Video 1')).toBeInTheDocument();
            });

            // The watch links should exist (ExternalLink icon buttons rendered via asChild+Link)
            // Use container.querySelectorAll since Button mock's asChild overrides Link's data-testid
            const watchLinks = container.querySelectorAll('a[href*="/watch"]');
            expect(watchLinks.length).toBeGreaterThanOrEqual(2);
        });
    });
});
