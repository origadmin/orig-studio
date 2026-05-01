/**
 * B087 Regression Test: Tag create API missing listings_thumbnail field
 *
 * This test verifies that the CreateTagRequest type and tagApi.create()
 * function handle the tag creation correctly, with or without listings_thumbnail.
 *
 * Before fix: Backend Ent schema requires listings_thumbnail (not Optional),
 *   causing "missing required field Tag.listings_thumbnail" error.
 * After fix: Backend schema makes listings_thumbnail Optional with Default(""),
 *   so frontend does not need to send this field.
 */

import {tagApi, CreateTagRequest} from '../../../src/lib/api/admin-tags';

// Mock request module
jest.mock('../../../src/lib/request', () => ({
    api: {
        post: jest.fn(),
        get: jest.fn(),
        put: jest.fn(),
        del: jest.fn(),
    },
}));

import {api} from '../../../src/lib/request';

const mockedApi = api as jest.Mocked<typeof api>;

describe('B087: Tag create - listings_thumbnail field', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('CreateTagRequest type', () => {
        it('should not require listings_thumbnail field', () => {
            // Before fix: Backend requires listings_thumbnail but frontend doesn't send it
            // After fix: Backend makes listings_thumbnail Optional, so this request is valid
            const request: CreateTagRequest = {
                name: 'test-tag-b087',
                slug: 'test-tag-b087',
                status: 'active',
            };

            // CreateTagRequest should be valid without listings_thumbnail
            expect(request.name).toBe('test-tag-b087');
            expect(request.slug).toBe('test-tag-b087');
            expect(request.status).toBe('active');
            // listings_thumbnail is not part of CreateTagRequest interface
            // This is correct because it's an optional field with a backend default
        });

        it('should allow optional description and color', () => {
            const request: CreateTagRequest = {
                name: 'test-tag-with-extras',
                slug: 'test-tag-with-extras',
                description: 'A test tag',
                color: '#FF0000',
                status: 'active',
            };

            expect(request.description).toBe('A test tag');
            expect(request.color).toBe('#FF0000');
        });
    });

    describe('tagApi.create', () => {
        it('should send POST /admin/tags with minimal fields', async () => {
            const mockResponse = {
                id: '1',
                name: 'test-tag-b087',
                slug: 'test-tag-b087',
                count: 0,
                status: 'active',
                create_time: '2026-05-01T00:00:00Z',
                update_time: '2026-05-01T00:00:00Z',
            };

            mockedApi.post.mockResolvedValueOnce(mockResponse);

            const request: CreateTagRequest = {
                name: 'test-tag-b087',
                slug: 'test-tag-b087',
                status: 'active',
            };

            const result = await tagApi.create(request);

            expect(mockedApi.post).toHaveBeenCalledWith('/admin/tags', request);
            expect(result).toEqual(mockResponse);
        });

        it('should successfully create tag without listings_thumbnail', async () => {
            // This is the core regression test for B087
            // Before fix: Backend returns "missing required field Tag.listings_thumbnail"
            // After fix: Backend accepts the request because listings_thumbnail is Optional
            const mockResponse = {
                id: '2',
                name: 'minimal-tag',
                slug: 'minimal-tag',
                count: 0,
                status: 'active',
                create_time: '2026-05-01T00:00:00Z',
                update_time: '2026-05-01T00:00:00Z',
            };

            mockedApi.post.mockResolvedValueOnce(mockResponse);

            const result = await tagApi.create({
                name: 'minimal-tag',
                slug: 'minimal-tag',
                status: 'active',
            });

            expect(result.id).toBe('2');
            expect(result.name).toBe('minimal-tag');
            // The request payload should NOT include listings_thumbnail
            expect(mockedApi.post).toHaveBeenCalledWith(
                '/admin/tags',
                expect.not.objectContaining({listings_thumbnail: expect.anything()})
            );
        });
    });
});
