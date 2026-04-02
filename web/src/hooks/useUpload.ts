/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import {useState, useCallback, useEffect, useRef} from 'react';
import {
    UploadTask,
    UploadStatus,
    UploadCallbacks,
    startMultipartUpload,
    cancelUpload,
} from '../lib/upload/multipart';

export interface UploadState {
    tasks: UploadTask[];
    addTask: (file: File, metadata?: Partial<UploadTask>) => void;
    removeTask: (taskId: string) => void;
    cancelTask: (taskId: string) => void;
    clearCompleted: () => void;
}

export function useUpload(): UploadState {
    const [tasks, setTasks] = useState<UploadTask[]>([]);
    const tasksRef = useRef<UploadTask[]>([]);

    // Sync state with ref for callback closure access
    useEffect(() => {
        tasksRef.current = tasks;
    }, [tasks]);

    const updateTask = useCallback((taskId: string, updates: Partial<UploadTask>) => {
        setTasks((prev) =>
            prev.map((t) => (t.id === taskId ? {...t, ...updates} : t))
        );
    }, []);

    const callbacks: UploadCallbacks = {
        onProgress: (taskId, progress, speed) => {
            updateTask(taskId, {progress, speed});
        },
        onStatusChange: (taskId, status) => {
            updateTask(taskId, {status});
        },
        onSuccess: (taskId) => {
            updateTask(taskId, {status: 'success', progress: 100, completedAt: Date.now()});
        },
        onError: (taskId, error) => {
            updateTask(taskId, {status: 'error', error});
        },
    };

    const addTask = useCallback((file: File, metadata?: Partial<UploadTask>) => {
        const taskId = `${Date.now()}-${file.name}`;
        const newTask: UploadTask = {
            id: taskId,
            file,
            progress: 0,
            status: 'waiting',
            title: metadata?.title || file.name,
            description: metadata?.description,
            categoryId: metadata?.categoryId,
            tags: metadata?.tags,
            ...metadata,
        };

        setTasks((prev) => [...prev, newTask]);

        // Automatically start upload
        startMultipartUpload(newTask, callbacks);
    }, [callbacks]);

    const removeTask = useCallback((taskId: string) => {
        cancelUpload(taskId);
        setTasks((prev) => prev.filter((t) => t.id !== taskId));
    }, []);

    const cancelTask = useCallback((taskId: string) => {
        cancelUpload(taskId);
        updateTask(taskId, {status: 'aborted'});
    }, [updateTask]);

    const clearCompleted = useCallback(() => {
        setTasks((prev) => prev.filter((t) => t.status !== 'success'));
    }, []);

    return {
        tasks,
        addTask,
        removeTask,
        cancelTask,
        clearCompleted,
    };
}
