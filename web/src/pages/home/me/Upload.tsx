/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Upload Page - 支持分片上传与断点续传
 */

import React, {useState, useRef, useCallback} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {
    Upload, X, File, Image, Video, CheckCircle,
    AlertCircle, Pause, Play, RotateCcw,
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {mediaApi} from '@/lib/api/media';
import {formatFileSize, formatRelativeTime} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {
    startMultipartUpload,
    cancelUpload,
    shouldUseChunkedUpload,
    type UploadTask,
    type UploadCallbacks,
    type UploadStatus,
} from '@/lib/upload';

interface UploadFileItem {
    id: string;
    file: File;
    preview?: string;
    progress: number;
    status: UploadStatus;
    error?: string;
    uploadId?: string;
    speed?: number;
    parts?: { part_number: number; etag: string; size: number }[];
    startedAt?: number;
    completedAt?: number;
}

const categories = [
    '技术', '编程', '运维', '数据科学', '云计算',
    '前端', '职业', '音乐', '游戏', '娱乐',
];

const UploadPage = () => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [files, setFiles] = useState<UploadFileItem[]>([]);
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [category, setCategory] = useState('');
    const [tags, setTags] = useState<string[]>([]);
    const [tagInput, setTagInput] = useState('');
    const [isDragging, setIsDragging] = useState(false);

    const updateFile = useCallback((id: string, updates: Partial<UploadFileItem>) => {
        setFiles((prev) =>
            prev.map((f) => (f.id === id ? {...f, ...updates} : f)),
        );
    }, []);

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(true);
    };

    const handleDragLeave = () => setIsDragging(false);

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
        handleFiles(Array.from(e.dataTransfer.files));
    };

    const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.files) handleFiles(Array.from(e.target.files));
    };

    const handleFiles = (newFiles: File[]) => {
        const validTypes = ['video/', 'image/', 'audio/'];
        const valid = newFiles.filter((f) =>
            validTypes.some((t) => f.type.startsWith(t)),
        );
        setFiles((prev) => [
            ...prev,
            ...valid.map((f) => ({
                id: Math.random().toString(36).substr(2, 9),
                file: f,
                preview: f.type.startsWith('image/')
                    ? URL.createObjectURL(f)
                    : undefined,
                progress: 0,
                status: 'waiting' as UploadStatus,
            })),
        ]);
    };

    const removeFile = (id: string) => {
        setFiles((prev) => prev.filter((f) => f.id !== id));
        cancelUpload(id);
    };

    const addTag = () => {
        const val = tagInput.trim();
        if (val && !tags.includes(val)) {
            setTags([...tags, val]);
            setTagInput('');
        }
    };

    const removeTag = (tag: string) => setTags(tags.filter((t) => t !== tag));

    // Upload handler using chunked upload for large files
    const handleUpload = async () => {
        if (!title || files.length === 0) return;

        const metadata = {
            title,
            description,
            category_id: category ? parseInt(category) : undefined,
            tags,
        };

        for (const fileItem of files) {
            if (fileItem.status === 'success' || fileItem.status === 'uploading') continue;

            // Small file: use simple upload
            if (!shouldUseChunkedUpload(fileItem.file.size)) {
                updateFile(fileItem.id, {status: 'uploading', progress: 0});
                try {
                    await mediaApi.upload(fileItem.file, metadata, (percent) => {
                        updateFile(fileItem.id, {progress: percent});
                    });
                    updateFile(fileItem.id, {status: 'success', progress: 100});
                } catch (err) {
                    const msg = err instanceof Error ? err.message : 'Upload failed';
                    updateFile(fileItem.id, {status: 'error', error: msg});
                }
                continue;
            }

            // Large file: use multipart upload
            const callbacks: UploadCallbacks = {
                onProgress: (taskId, progress, speed) => {
                    updateFile(fileItem.id, {progress, speed});
                },
                onStatusChange: (taskId, status) => {
                    updateFile(fileItem.id, {status});
                },
                onSuccess: (taskId) => {
                    updateFile(fileItem.id, {
                        status: 'success',
                        progress: 100,
                        completedAt: Date.now(),
                    });
                },
                onError: (taskId, error) => {
                    updateFile(fileItem.id, {status: 'error', error});
                },
            };

            const task: UploadTask = {
                id: fileItem.id,
                file: fileItem.file,
                progress: 0,
                status: 'waiting',
                parts: fileItem.parts || [],
                uploadId: fileItem.uploadId,
                title: metadata.title,
                description: metadata.description,
                categoryId: metadata.category_id,
                tags: metadata.tags,
            };

            // Fire and forget - each upload runs independently
            startMultipartUpload(task, callbacks).catch(() => {
                // Error already handled in callbacks
            });
        }
    };

    const handleCancelUpload = (id: string) => {
        cancelUpload(id);
        updateFile(id, {status: 'aborted', progress: 0});
    };

    const getFileIcon = (type: string) => {
        if (type.startsWith('video/')) return <Video className="w-8 h-8 text-blue-500"/>;
        if (type.startsWith('image/')) return <Image className="w-8 h-8 text-green-500"/>;
        return <File className="w-8 h-8 text-slate-500"/>;
    };

    const getStatusIcon = (status: UploadStatus) => {
        switch (status) {
            case 'success':
                return <CheckCircle className="w-5 h-5 text-green-500"/>;
            case 'error':
                return <AlertCircle className="w-5 h-5 text-red-500"/>;
            case 'aborted':
                return <RotateCcw className="w-5 h-5 text-orange-500"/>;
            default:
                return null;
        }
    };

    const getStatusText = (status: UploadStatus) => {
        switch (status) {
            case 'waiting':
                return t('upload.statusWaiting');
            case 'initiating':
                return t('upload.statusInitiating');
            case 'uploading':
                return t('upload.statusUploading');
            case 'paused':
                return t('upload.statusPaused');
            case 'completing':
                return t('upload.statusCompleting');
            case 'success':
                return t('upload.statusSuccess');
            case 'error':
                return t('upload.statusError');
            case 'aborted':
                return t('upload.statusAborted');
            default:
                return status;
        }
    };

    const uploading = files.some(
        (f) => ['uploading', 'initiating', 'completing'].includes(f.status),
    );
    const allDone = files.length > 0 && files.every((f) => f.status === 'success');
    const canSubmit = !!title && files.length > 0 && !uploading;

    return (
        <div className="max-w-4xl mx-auto space-y-8">
            {/* Header */}
            <div className="text-center">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
                    {t('upload.title')}
                </h1>
                <p className="text-gray-500 dark:text-gray-400">
                    {t('upload.description')}
                </p>
            </div>

            {/* Drop zone */}
            <div
                className={`border-2 border-dashed rounded-2xl p-12 text-center transition-colors ${
                    isDragging
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/30'
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                }`}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
            >
                <input
                    ref={fileInputRef}
                    type="file"
                    multiple
                    accept="video/*,image/*,audio/*"
                    onChange={handleFileSelect}
                    className="hidden"
                />
                <div className="flex flex-col items-center">
                    <div
                        className="w-16 h-16 bg-blue-100 dark:bg-blue-900/40 rounded-full flex items-center justify-center mb-4">
                        <Upload className="w-8 h-8 text-blue-600 dark:text-blue-400"/>
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                        {t('upload.dropzoneTitle')}
                    </h3>
                    <p className="text-gray-500 dark:text-gray-400 mb-4">
                        {t('upload.dropzoneDesc')}
                    </p>
                    <Button onClick={() => fileInputRef.current?.click()}>
                        {t('upload.selectFiles')}
                    </Button>
                    <p className="text-xs text-gray-400 dark:text-gray-500 mt-4">
                        {t('upload.supportedFormats')}
                    </p>
                </div>
            </div>

            {/* File list */}
            {files.length > 0 && (
                <Card className="dark:bg-gray-800">
                    <CardHeader>
                        <CardTitle className="dark:text-white">
                            {t('upload.selectedFiles', {count: files.length})}
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        {files.map((fileItem) => (
                            <div
                                key={fileItem.id}
                                className="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl"
                            >
                                {/* Preview / Icon */}
                                <div
                                    className="w-16 h-16 bg-gray-200 dark:bg-gray-600 rounded-lg overflow-hidden shrink-0">
                                    {fileItem.preview ? (
                                        <img
                                            src={fileItem.preview}
                                            alt=""
                                            className="w-full h-full object-cover"
                                        />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center">
                                            {getFileIcon(fileItem.file.type)}
                                        </div>
                                    )}
                                </div>

                                {/* File info */}
                                <div className="flex-1 min-w-0">
                                    <p className="font-medium text-gray-900 dark:text-white truncate">
                                        {fileItem.file.name}
                                    </p>
                                    <p className="text-sm text-gray-500 dark:text-gray-400">
                                        {formatFileSize(fileItem.file.size)}
                                        {fileItem.speed != null && fileItem.status === 'uploading' && (
                                            <span className="ml-2">
                                                · {formatFileSize(fileItem.speed)}/s
                                            </span>
                                        )}
                                    </p>

                                    {/* Progress bar */}
                                    {['uploading', 'initiating', 'completing'].includes(fileItem.status) && (
                                        <div
                                            className="mt-2 h-2 bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden">
                                            <div
                                                className={`h-full transition-all duration-300 ${
                                                    fileItem.status === 'completing'
                                                        ? 'bg-yellow-500'
                                                        : 'bg-blue-600'
                                                }`}
                                                style={{width: `${fileItem.progress}%`}}
                                            />
                                        </div>
                                    )}

                                    {/* Status / Error */}
                                    {fileItem.status !== 'waiting' && (
                                        <p className={`text-xs mt-1 ${
                                            fileItem.status === 'success'
                                                ? 'text-green-500'
                                                : fileItem.status === 'error'
                                                    ? 'text-red-500'
                                                    : 'text-gray-400'
                                        }`}>
                                            {getStatusText(fileItem.status)}
                                            {fileItem.error && `: ${fileItem.error}`}
                                        </p>
                                    )}
                                </div>

                                {/* Actions */}
                                <div className="flex items-center gap-2">
                                    {getStatusIcon(fileItem.status)}
                                    {fileItem.status === 'pending' && (
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => removeFile(fileItem.id)}
                                        >
                                            <X className="w-4 h-4"/>
                                        </Button>
                                    )}
                                    {fileItem.status === 'uploading' && (
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => handleCancelUpload(fileItem.id)}
                                        >
                                            <Pause className="w-4 h-4"/>
                                        </Button>
                                    )}
                                    {fileItem.status === 'aborted' && (
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => removeFile(fileItem.id)}
                                        >
                                            <RotateCcw className="w-4 h-4"/>
                                        </Button>
                                    )}
                                </div>
                            </div>
                        ))}
                    </CardContent>
                </Card>
            )}

            {/* Details form */}
            <Card className="dark:bg-gray-800">
                <CardHeader>
                    <CardTitle className="dark:text-white">{t('upload.contentDetails')}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-6">
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                            {t('upload.titleLabel')}
                        </label>
                        <Input
                            placeholder={t('upload.titlePlaceholder')}
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                            {t('upload.descLabel')}
                        </label>
                        <Textarea
                            placeholder={t('upload.descPlaceholder')}
                            rows={4}
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                        />
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                            {t('upload.categoryLabel')}
                        </label>
                        <select
                            className="w-full h-10 px-3 rounded-md border border-input bg-background text-sm dark:bg-gray-700 dark:text-white"
                            value={category}
                            onChange={(e) => setCategory(e.target.value)}
                        >
                            <option value="">{t('upload.selectCategory')}</option>
                            {categories.map((cat) => (
                                <option key={cat} value={cat}>{cat}</option>
                            ))}
                        </select>
                    </div>
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">
                            {t('upload.tagLabel')}
                        </label>
                        <div className="flex gap-2 mb-2">
                            <Input
                                placeholder={t('upload.addTagPlaceholder')}
                                value={tagInput}
                                onChange={(e) => setTagInput(e.target.value)}
                                onKeyDown={(e) => e.key === 'Enter' && addTag()}
                            />
                            <Button variant="outline" onClick={addTag}>
                                {t('upload.addTag')}
                            </Button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            {tags.map((tag) => (
                                <Badge key={tag} variant="secondary" className="flex items-center gap-1">
                                    {tag}
                                    <button onClick={() => removeTag(tag)} className="hover:text-red-500">
                                        <X className="w-3 h-3"/>
                                    </button>
                                </Badge>
                            ))}
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Actions */}
            <div className="flex justify-end gap-4">
                <Button variant="outline" onClick={() => navigate({to: '/'})}>
                    {t('common.cancel')}
                </Button>
                <Button onClick={handleUpload} disabled={!canSubmit}>
                    {allDone
                        ? t('upload.allDone')
                        : uploading
                            ? t('upload.uploading')
                            : t('upload.upload')}
                </Button>
            </div>
        </div>
    );
};

export default UploadPage;
