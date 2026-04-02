/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React, {useCallback, useState} from 'react';
import {useUpload} from '../../hooks/useUpload';

export interface UploadZoneProps {
    onFilesAdded?: (files: File[]) => void;
    accept?: string;
    maxSize?: number; // bytes
}

export const UploadZone: React.FC<UploadZoneProps> = ({
                                                          onFilesAdded,
                                                          accept = 'video/*,image/*',
                                                          maxSize = 500 * 1024 * 1024
                                                      }) => {
    const [isDragActive, setIsDragActive] = useState(false);
    const {addTask} = useUpload();

    const handleDragEnter = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragActive(true);
    }, []);

    const handleDragLeave = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragActive(false);
    }, []);

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setIsDragActive(false);

        const files = Array.from(e.dataTransfer.files);
        if (files.length > 0) {
            files.forEach(file => {
                if (file.size <= maxSize) {
                    addTask(file);
                } else {
                    alert(`File ${file.name} is too large (max ${maxSize / (1024 * 1024)}MB)`);
                }
            });
            if (onFilesAdded) onFilesAdded(files);
        }
    }, [addTask, maxSize, onFilesAdded]);

    const handleFileInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        const files = Array.from(e.target.files || []);
        if (files.length > 0) {
            files.forEach(file => {
                if (file.size <= maxSize) {
                    addTask(file);
                } else {
                    alert(`File ${file.name} is too large (max ${maxSize / (1024 * 1024)}MB)`);
                }
            });
            if (onFilesAdded) onFilesAdded(files);
        }
    }, [addTask, maxSize, onFilesAdded]);

    return (
        <div
            onDragEnter={handleDragEnter}
            onDragOver={handleDragEnter}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`border-2 border-dashed rounded-lg p-10 flex flex-col items-center justify-center transition-colors cursor-pointer ${
                isDragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
            }`}
            onClick={() => document.getElementById('file-input')?.click()}
        >
            <input
                id="file-input"
                type="file"
                multiple
                accept={accept}
                onChange={handleFileInput}
                className="hidden"
            />
            <div className="text-4xl mb-4 text-gray-400">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12" fill="none" viewBox="0 0 24 24"
                     stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                          d="7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
                </svg>
            </div>
            <p className="text-lg font-medium text-gray-600">
                Drag & drop files here or click to browse
            </p>
            <p className="text-sm text-gray-400 mt-2">
                Supports videos and images up to {maxSize / (1024 * 1024)}MB
            </p>
        </div>
    );
};
