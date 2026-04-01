/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Upload Page - 支持分片上传与独立文件信息管理
 */

import React, {useState, useRef, useCallback, useEffect} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {
    Upload, X, File, Image, Video, CheckCircle,
    AlertCircle, Pause, Play, RotateCcw, Edit2, ChevronRight
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {mediaApi} from '@/lib/api/media';
import {formatFileSize} from '@/lib/format';
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
    // 独立元数据
    title: string;
    description: string;
    category: string;
    tags: string[];
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
    const [selectedFileId, setSelectedFileId] = useState<string | null>(null);
    const [tagInput, setTagInput] = useState('');
    const [isDragging, setIsDragging] = useState(false);

    const updateFile = useCallback((id: string, updates: Partial<UploadFileItem>) => {
        setFiles((prev) =>
            prev.map((f) => (f.id === id ? {...f, ...updates} : f)),
        );
    }, []);

    const handleFiles = (newFiles: File[]) => {
        const validTypes = ['video/', 'image/', 'audio/'];
        const valid = newFiles.filter((f) =>
            validTypes.some((t) => f.type.startsWith(t)),
        );

        const newItems = valid.map((f) => ({
            id: Math.random().toString(36).substr(2, 9),
            file: f,
            preview: f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined,
            progress: 0,
            status: 'waiting' as UploadStatus,
            title: f.name.replace(/\.[^.]+$/, ''), // 默认标题为文件名
            description: '',
            category: '',
            tags: [],
        }));

        setFiles((prev) => [...prev, ...newItems]);
        if (!selectedFileId && newItems.length > 0) {
            setSelectedFileId(newItems[0].id);
        }
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setIsDragging(false);
        handleFiles(Array.from(e.dataTransfer.files));
    };

    const removeFile = (id: string) => {
        setFiles((prev) => {
            const filtered = prev.filter((f) => f.id !== id);
            if (selectedFileId === id) {
                setSelectedFileId(filtered.length > 0 ? filtered[0].id : null);
            }
            return filtered;
        });
        cancelUpload(id);
    };

    const selectedFile = files.find(f => f.id === selectedFileId);

    const addTag = () => {
        if (!selectedFile) return;
        const val = tagInput.trim();
        if (val && !selectedFile.tags.includes(val)) {
            updateFile(selectedFile.id, {tags: [...selectedFile.tags, val]});
            setTagInput('');
        }
    };

    const removeTag = (tag: string) => {
        if (!selectedFile) return;
        updateFile(selectedFile.id, {tags: selectedFile.tags.filter(t => t !== tag)});
    };

    const handleUpload = async () => {
        for (const fileItem of files) {
            if (fileItem.status === 'success' || ['uploading', 'initiating', 'completing'].includes(fileItem.status)) continue;

            const metadata = {
                title: fileItem.title,
                description: fileItem.description,
                category_id: fileItem.category ? categories.indexOf(fileItem.category) + 1 : undefined,
                tags: fileItem.tags,
            };

            // Small file: simple upload
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

            // Large file: multipart upload
            const callbacks: UploadCallbacks = {
                onProgress: (taskId, progress, speed) => {
                    updateFile(taskId, {progress, speed});
                },
                onStatusChange: (taskId, status) => {
                    updateFile(taskId, {status});
                },
                onSuccess: (taskId) => {
                    updateFile(taskId, {
                        status: 'success',
                        progress: 100,
                        completedAt: Date.now(),
                    });
                },
                onError: (taskId, error) => {
                    updateFile(taskId, {status: 'error', error});
                },
            };

            const task: UploadTask = {
                id: fileItem.id,
                file: fileItem.file,
                progress: fileItem.progress,
                status: 'waiting',
                parts: [],
                uploadId: fileItem.uploadId,
                title: fileItem.title,
                description: fileItem.description,
                categoryId: metadata.category_id,
                tags: fileItem.tags,
            };

            startMultipartUpload(task, callbacks).catch(() => {
            });
        }
    };

    const uploading = files.some(f => ['uploading', 'initiating', 'completing'].includes(f.status));
    const allDone = files.length > 0 && files.every(f => f.status === 'success');
    // 只要有一个文件没填标题且不是上传成功状态，就不能提交
    const canSubmit = files.length > 0 && files.every(f => f.status === 'success' || !!f.title) && !uploading;

    const getStatusIcon = (status: UploadStatus) => {
        switch (status) {
            case 'success':
                return <CheckCircle className="w-5 h-5 text-green-500"/>;
            case 'error':
                return <AlertCircle className="w-5 h-5 text-red-500"/>;
            case 'aborted':
                return <RotateCcw className="w-5 h-5 text-orange-500"/>;
            case 'uploading':
                return <div
                    className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"/>;
            default:
                return null;
        }
    };

    return (
        <div className="max-w-6xl mx-auto p-6 space-y-8">
            <div className="text-center">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">{t('upload.title')}</h1>
                <p className="text-gray-500 dark:text-gray-400">{t('upload.description')}</p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
                {/* Left: File List and Dropzone */}
                <div className="lg:col-span-7 space-y-6">
                    <div
                        className={`border-2 border-dashed rounded-2xl p-8 text-center transition-colors cursor-pointer ${
                            isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'
                        }`}
                        onDragOver={(e) => {
                            e.preventDefault();
                            setIsDragging(true);
                        }}
                        onDragLeave={() => setIsDragging(false)}
                        onDrop={handleDrop}
                        onClick={() => fileInputRef.current?.click()}
                    >
                        <input
                            ref={fileInputRef}
                            type="file"
                            multiple
                            accept="video/*,image/*,audio/*"
                            onChange={(e) => {
                                if (e.target.files) handleFiles(Array.from(e.target.files));
                            }}
                            className="hidden"
                        />
                        <Upload className="w-10 h-10 text-blue-600 mx-auto mb-3"/>
                        <p className="text-sm font-medium text-gray-700">{t('upload.dropzoneTitle')}</p>
                        <p className="text-xs text-gray-400 mt-1">{t('upload.supportedFormats')}</p>
                    </div>

                    <div className="space-y-3">
                        <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider">
                            {t('upload.selectedFiles', {count: files.length})}
                        </h3>
                        {files.map((fileItem) => (
                            <div
                                key={fileItem.id}
                                onClick={() => setSelectedFileId(fileItem.id)}
                                className={`flex items-center gap-4 p-3 rounded-xl transition-all border cursor-pointer ${
                                    selectedFileId === fileItem.id
                                        ? 'border-blue-500 bg-blue-50 shadow-sm'
                                        : 'border-transparent bg-white hover:bg-gray-50'
                                }`}
                            >
                                <div className="w-12 h-12 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
                                    {fileItem.preview ? (
                                        <img src={fileItem.preview} alt="" className="w-full h-full object-cover"/>
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center">
                                            {fileItem.file.type.startsWith('video/') ?
                                                <Video className="w-6 h-6 text-blue-500"/> :
                                                <File className="w-6 h-6 text-gray-400"/>}
                                        </div>
                                    )}
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex justify-between items-start">
                                        <p className="font-medium text-sm text-gray-900 truncate">{fileItem.title || fileItem.file.name}</p>
                                        <span
                                            className="text-[10px] text-gray-400">{formatFileSize(fileItem.file.size)}</span>
                                    </div>

                                    {['uploading', 'initiating', 'completing'].includes(fileItem.status) ? (
                                        <div className="mt-1.5 space-y-1">
                                            <div className="h-1 bg-gray-200 rounded-full overflow-hidden">
                                                <div className="h-full bg-blue-600 transition-all"
                                                     style={{width: `${fileItem.progress}%`}}/>
                                            </div>
                                            <div className="flex justify-between text-[10px] text-gray-500">
                                                <span>{fileItem.status === 'completing' ? '合并中...' : `${fileItem.progress}%`}</span>
                                                {fileItem.speed && <span>{formatFileSize(fileItem.speed)}/s</span>}
                                            </div>
                                        </div>
                                    ) : (
                                        <p className={`text-[10px] mt-1 ${fileItem.status === 'success' ? 'text-green-500' : fileItem.status === 'error' ? 'text-red-500' : 'text-gray-400'}`}>
                                            {fileItem.status === 'waiting' && !fileItem.title ? '⚠️ 请填写标题' : t(`upload.status${fileItem.status.charAt(0).toUpperCase()}${fileItem.status.slice(1)}`)}
                                        </p>
                                    )}
                                </div>
                                <div className="flex items-center gap-1">
                                    {getStatusIcon(fileItem.status)}
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-8 w-8 text-gray-400 hover:text-red-500"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            removeFile(fileItem.id);
                                        }}
                                    >
                                        <X className="w-4 h-4"/>
                                    </Button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Right: Metadata Form */}
                <div className="lg:col-span-5">
                    {selectedFile ? (
                        <Card className="sticky top-6 shadow-md border-gray-200">
                            <CardHeader className="pb-3 border-b">
                                <CardTitle className="text-lg flex items-center gap-2">
                                    <Edit2 className="w-4 h-4 text-blue-600"/>
                                    编辑信息
                                </CardTitle>
                            </CardHeader>
                            <CardContent className="pt-6 space-y-5">
                                <div className="space-y-1.5">
                                    <label
                                        className="text-xs font-bold text-gray-500 uppercase">{t('upload.titleLabel')}</label>
                                    <Input
                                        value={selectedFile.title}
                                        onChange={(e) => updateFile(selectedFile.id, {title: e.target.value})}
                                        placeholder="为媒体起一个吸引人的标题"
                                        disabled={selectedFile?.status === 'success'}
                                    />
                                </div>

                                <div className="space-y-1.5">
                                    <label
                                        className="text-xs font-bold text-gray-500 uppercase">{t('upload.descLabel')}</label>
                                    <Textarea
                                        value={selectedFile.description}
                                        onChange={(e) => updateFile(selectedFile.id, {description: e.target.value})}
                                        placeholder="介绍一下这个媒体的内容..."
                                        rows={3}
                                    />
                                </div>

                                <div className="space-y-1.5">
                                    <label
                                        className="text-xs font-bold text-gray-500 uppercase">{t('upload.categoryLabel')}</label>
                                    <select
                                        className="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
                                        value={selectedFile.category}
                                        onChange={(e) => updateFile(selectedFile.id, {category: e.target.value})}
                                    >
                                        <option value="">选择分类</option>
                                        {categories.map(cat => <option key={cat} value={cat}>{cat}</option>)}
                                    </select>
                                </div>

                                <div className="space-y-1.5">
                                    <label
                                        className="text-xs font-bold text-gray-500 uppercase">{t('upload.tagLabel')}</label>
                                    <div className="flex gap-2">
                                        <Input
                                            value={tagInput}
                                            onChange={(e) => setTagInput(e.target.value)}
                                            onKeyDown={(e) => e.key === 'Enter' && addTag()}
                                            placeholder="输入标签按回车"
                                        />
                                        <Button variant="outline" size="sm" onClick={addTag}>添加</Button>
                                    </div>
                                    <div className="flex flex-wrap gap-1.5 mt-2">
                                        {selectedFile.tags.map(tag => (
                                            <Badge key={tag} variant="secondary"
                                                   className="px-2 py-0.5 text-[10px] flex items-center gap-1">
                                                {tag}
                                                <X className="w-3 h-3 cursor-pointer hover:text-red-500"
                                                   onClick={() => removeTag(tag)}/>
                                            </Badge>
                                        ))}
                                    </div>
                                </div>

                                <div className="pt-4 border-t mt-6 flex flex-col gap-3">
                                    <Button
                                        className="w-full bg-blue-600 hover:bg-blue-700"
                                        onClick={handleUpload}
                                        disabled={!canSubmit || uploading}
                                    >
                                        {uploading ? '上传中...' : files.length > 1 ? '上传全部文件' : '开始上传'}
                                    </Button>
                                    <Button variant="ghost" className="w-full text-gray-400"
                                            onClick={() => navigate({to: '/'})}>
                                        取消
                                    </Button>
                                </div>
                            </CardContent>
                        </Card>
                    ) : (
                        <div
                            className="h-full flex flex-col items-center justify-center text-gray-400 border-2 border-dashed rounded-2xl p-12">
                            <File className="w-12 h-12 mb-4 opacity-20"/>
                            <p className="text-sm italic">选择左侧文件以编辑详情</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default UploadPage;
