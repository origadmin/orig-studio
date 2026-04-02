/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React from 'react';
import {useUpload} from '../../hooks/useUpload';

export const UploadList: React.FC = () => {
    const {tasks, removeTask, cancelTask, clearCompleted} = useUpload();

    if (tasks.length === 0) {
        return null;
    }

    const formatSize = (bytes: number) => {
        if (bytes === 0) return '0B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    const formatSpeed = (bytesPerSec: number) => {
        return formatSize(bytesPerSec) + '/s';
    };

    const getStatusColor = (status: string) => {
        switch (status) {
            case 'success':
                return 'text-green-500';
            case 'error':
                return 'text-red-500';
            case 'aborted':
                return 'text-gray-400';
            case 'completing':
                return 'text-blue-500 font-bold';
            default:
                return 'text-blue-500';
        }
    };

    return (
        <div className="mt-8 bg-white shadow rounded-lg overflow-hidden border border-gray-200">
            <div className="px-6 py-4 bg-gray-50 border-b border-gray-200 flex justify-between items-center">
                <h3 className="text-lg font-medium text-gray-800">
                    Upload Tasks ({tasks.length})
                </h3>
                <button
                    onClick={clearCompleted}
                    className="text-sm text-blue-600 hover:text-blue-800"
                >
                    Clear Completed
                </button>
            </div>
            <ul className="divide-y divide-gray-200">
                {tasks.map((task) => (
                    <li key={task.id} className="px-6 py-4 hover:bg-gray-50 transition-colors">
                        <div className="flex items-center justify-between mb-2">
                            <div className="flex-1 min-w-0 pr-4">
                                <p className="text-sm font-medium text-gray-900 truncate">
                                    {task.title || task.file.name}
                                </p>
                                <p className="text-xs text-gray-500">
                                    {formatSize(task.file.size)} • {task.status.toUpperCase()}
                                    {task.speed && task.status === 'uploading' && ` • ${formatSpeed(task.speed)}`}
                                </p>
                            </div>
                            <div className="flex-shrink-0 flex items-center space-x-2">
                                {task.status !== 'success' && task.status !== 'error' && task.status !== 'aborted' && (
                                    <button
                                        onClick={() => cancelTask(task.id)}
                                        className="p-1 hover:bg-gray-200 rounded text-gray-400"
                                        title="Cancel upload"
                                    >
                                        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                                  d="M6 18L18 6M6 6l12 12"/>
                                        </svg>
                                    </button>
                                )}
                                <button
                                    onClick={() => removeTask(task.id)}
                                    className="p-1 hover:bg-gray-200 rounded text-gray-400"
                                    title="Remove from list"
                                >
                                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                                              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                                    </svg>
                                </button>
                            </div>
                        </div>

                        <div className="w-full bg-gray-200 rounded-full h-2.5 mb-1 relative overflow-hidden">
                            <div
                                className={`h-full rounded-full transition-all duration-300 ${
                                    task.status === 'success' ? 'bg-green-500' :
                                        task.status === 'error' ? 'bg-red-500' :
                                            task.status === 'aborted' ? 'bg-gray-400' :
                                                'bg-blue-600'
                                }`}
                                style={{width: `${task.progress}%`}}
                            ></div>
                        </div>

                        {task.error && (
                            <p className="text-xs text-red-500 mt-1 italic">
                                Error: {task.error}
                            </p>
                        )}

                        {task.status === 'completing' && (
                            <p className="text-xs text-blue-500 mt-1 animate-pulse">
                                Merging parts on server... please wait
                            </p>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    );
};
