import {renderHook, act} from '@testing-library/react';
import {useUpload} from './useUpload';

// Mock upload module
jest.mock('../lib/upload/multipart', () => ({
    startMultipartUpload: jest.fn(),
    cancelUpload: jest.fn(),
}));

import {
    startMultipartUpload,
    cancelUpload,
} from '../lib/upload/multipart';

describe('useUpload Hook', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('task management', () => {
        it('should initialize with empty tasks', () => {
            const {result} = renderHook(() => useUpload());
            expect(result.current.tasks).toEqual([]);
        });

        it('should add a new task', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            expect(result.current.tasks).toHaveLength(1);
            expect(result.current.tasks[0].file).toBe(testFile);
            expect(result.current.tasks[0].status).toBe('waiting');
            expect(result.current.tasks[0].progress).toBe(0);
            expect(startMultipartUpload).toHaveBeenCalled();
        });

        it('should add task with metadata', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});
            const metadata = {
                title: 'Custom Title',
                description: 'Custom Description',
                categoryId: 1,
                tags: ['tag1', 'tag2'],
            };

            act(() => {
                result.current.addTask(testFile, metadata);
            });

            expect(result.current.tasks[0].title).toBe('Custom Title');
            expect(result.current.tasks[0].description).toBe('Custom Description');
            expect(result.current.tasks[0].categoryId).toBe(1);
            expect(result.current.tasks[0].tags).toEqual(['tag1', 'tag2']);
        });

        it('should use file name as default title', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'my-video.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            expect(result.current.tasks[0].title).toBe('my-video.mp4');
        });

        it('should remove a task', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            const taskId = result.current.tasks[0].id;

            act(() => {
                result.current.removeTask(taskId);
            });

            expect(result.current.tasks).toHaveLength(0);
            expect(cancelUpload).toHaveBeenCalledWith(taskId);
        });

        it('should cancel a task', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            const taskId = result.current.tasks[0].id;

            act(() => {
                result.current.cancelTask(taskId);
            });

            expect(result.current.tasks[0].status).toBe('aborted');
            expect(cancelUpload).toHaveBeenCalledWith(taskId);
        });

        it('should clear completed tasks', () => {
            const {result} = renderHook(() => useUpload());
            const testFile1 = new File(['test content 1'], 'test1.mp4', {type: 'video/mp4'});
            const testFile2 = new File(['test content 2'], 'test2.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile1);
                result.current.addTask(testFile2);
            });

            // Mock the startMultipartUpload callback to mark first task as success
            const mockStartMultipartUpload = startMultipartUpload as jest.Mock;
            const firstCall = mockStartMultipartUpload.mock.calls[0];
            const callbacks = firstCall[1];

            act(() => {
                callbacks.onSuccess(result.current.tasks[0].id);
            });

            act(() => {
                result.current.clearCompleted();
            });

            expect(result.current.tasks).toHaveLength(1);
            expect(result.current.tasks[0].title).toBe('test2.mp4');
        });
    });

    describe('upload callbacks', () => {
        it('should update progress', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            const taskId = result.current.tasks[0].id;

            // Get the callbacks from the mock
            const mockStartMultipartUpload = startMultipartUpload as jest.Mock;
            const callbacks = mockStartMultipartUpload.mock.calls[0][1];

            act(() => {
                callbacks.onProgress(taskId, 50, 1024000);
            });

            expect(result.current.tasks[0].progress).toBe(50);
            expect(result.current.tasks[0].speed).toBe(1024000);
        });

        it('should update status', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            const taskId = result.current.tasks[0].id;

            const mockStartMultipartUpload = startMultipartUpload as jest.Mock;
            const callbacks = mockStartMultipartUpload.mock.calls[0][1];

            act(() => {
                callbacks.onStatusChange(taskId, 'uploading');
            });

            expect(result.current.tasks[0].status).toBe('uploading');
        });

        it('should mark task as success', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            const taskId = result.current.tasks[0].id;

            const mockStartMultipartUpload = startMultipartUpload as jest.Mock;
            const callbacks = mockStartMultipartUpload.mock.calls[0][1];

            act(() => {
                callbacks.onSuccess(taskId);
            });

            expect(result.current.tasks[0].status).toBe('success');
            expect(result.current.tasks[0].progress).toBe(100);
            expect(result.current.tasks[0].completedAt).toBeDefined();
        });

        it('should mark task as error', () => {
            const {result} = renderHook(() => useUpload());
            const testFile = new File(['test content'], 'test.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile);
            });

            const taskId = result.current.tasks[0].id;

            const mockStartMultipartUpload = startMultipartUpload as jest.Mock;
            const callbacks = mockStartMultipartUpload.mock.calls[0][1];

            const testError = 'Upload failed';

            act(() => {
                callbacks.onError(taskId, testError);
            });

            expect(result.current.tasks[0].status).toBe('error');
            expect(result.current.tasks[0].error).toBe(testError);
        });
    });

    describe('multiple tasks', () => {
        it('should handle multiple tasks', () => {
            const {result} = renderHook(() => useUpload());
            const testFile1 = new File(['test content 1'], 'test1.mp4', {type: 'video/mp4'});
            const testFile2 = new File(['test content 2'], 'test2.mp4', {type: 'video/mp4'});
            const testFile3 = new File(['test content 3'], 'test3.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile1);
                result.current.addTask(testFile2);
                result.current.addTask(testFile3);
            });

            expect(result.current.tasks).toHaveLength(3);
            expect(result.current.tasks[0].title).toBe('test1.mp4');
            expect(result.current.tasks[1].title).toBe('test2.mp4');
            expect(result.current.tasks[2].title).toBe('test3.mp4');
        });

        it('should update only specific task', () => {
            const {result} = renderHook(() => useUpload());
            const testFile1 = new File(['test content 1'], 'test1.mp4', {type: 'video/mp4'});
            const testFile2 = new File(['test content 2'], 'test2.mp4', {type: 'video/mp4'});

            act(() => {
                result.current.addTask(testFile1);
                result.current.addTask(testFile2);
            });

            const taskId1 = result.current.tasks[0].id;
            const taskId2 = result.current.tasks[1].id;

            const mockStartMultipartUpload = startMultipartUpload as jest.Mock;

            // Update first task
            const callbacks1 = mockStartMultipartUpload.mock.calls[0][1];
            act(() => {
                callbacks1.onProgress(taskId1, 50, 1024000);
            });

            // Update second task
            const callbacks2 = mockStartMultipartUpload.mock.calls[1][1];
            act(() => {
                callbacks2.onProgress(taskId2, 75, 2048000);
            });

            expect(result.current.tasks[0].progress).toBe(50);
            expect(result.current.tasks[1].progress).toBe(75);
        });
    });
});
