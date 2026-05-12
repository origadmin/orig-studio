import React, {useState, useRef, useCallback} from 'react';
import {
    Upload, X, File, Video, CheckCircle,
    AlertCircle,
} from 'lucide-react';
import {Button} from '@/components/ui/button';
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
    startedAt?: number;
    completedAt?: number;
}

export interface UploadComponentProps {
    onSuccess?: () => void;
    onCancel?: () => void;
}

export function UploadComponent({onSuccess, onCancel}: UploadComponentProps) {
    const {t} = useTranslation();
    const fileInputRef = useRef<HTMLInputElement>(null);

    const [files, setFiles] = useState<UploadFileItem[]>([]);
    const [selectedFileId, setSelectedFileId] = useState<string | null>(null);
    const [isDragging, setIsDragging] = useState(false);
    const activeTasksRef = useRef<Map<string, UploadTask>>(new Map());

    const updateFile = useCallback((id: string, updates: Partial<UploadFileItem>) => {
        setFiles((prev) =>
            prev.map((f) => (f.id === id ? {...f, ...updates} : f)),
        );
    }, []);

    const performUpload = useCallback(async (fileItem: UploadFileItem) => {
        if (fileItem.status === 'success' || ['uploading', 'initiating', 'completing'].includes(fileItem.status)) return;

        if (!shouldUseChunkedUpload(fileItem.file.size)) {
            updateFile(fileItem.id, {status: 'uploading', progress: 0});
            try {
                await mediaApi.upload(fileItem.file, {
                    title: fileItem.file.name.replace(/\.[^.]+$/, ''),
                }, (percent) => {
                    updateFile(fileItem.id, {progress: percent});
                });
                updateFile(fileItem.id, {status: 'success', progress: 100});
                if (onSuccess) onSuccess();
            } catch (err) {
                const msg = err instanceof Error ? err.message : 'Upload failed';
                updateFile(fileItem.id, {status: 'error', error: msg});
            }
            return;
        }

        const callbacks: UploadCallbacks = {
            onProgress: (taskId, progress, speed) => updateFile(taskId, {progress, speed}),
            onStatusChange: (taskId, status) => updateFile(taskId, {status}),
            onSuccess: (taskId) => {
                updateFile(taskId, {status: 'success', progress: 100, completedAt: Date.now()});
                activeTasksRef.current.delete(taskId);
                if (onSuccess) onSuccess();
            },
            onError: (taskId, error) => {
                updateFile(taskId, {status: 'error', error});
                activeTasksRef.current.delete(taskId);
            },
            onUploadId: (taskId, uploadId) => {
                updateFile(taskId, {uploadId});
            },
        };

        const task: UploadTask = {
            id: fileItem.id,
            file: fileItem.file,
            progress: fileItem.progress,
            status: 'waiting',
            parts: [],
            uploadId: fileItem.uploadId,
        };

        activeTasksRef.current.set(fileItem.id, task);

        startMultipartUpload(task, callbacks).catch(() => {
        });
    }, [updateFile, onSuccess]);

    const handleFiles = (newFiles: File[]) => {
        const validTypes = ['video/', 'image/', 'audio/'];
        const valid = newFiles.filter((f) =>
            validTypes.some((t) => f.type.startsWith(t)),
        );

        const newItems: UploadFileItem[] = valid.map((f) => ({
            id: Math.random().toString(36).substr(2, 9),
            file: f,
            preview: f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined,
            progress: 0,
            status: 'waiting',
        }));

        setFiles((prev) => [...prev, ...newItems]);
        if (!selectedFileId && newItems.length > 0) {
            setSelectedFileId(newItems[0].id);
        }

        newItems.forEach(item => performUpload(item));
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

    const getStatusIcon = (status: UploadStatus) => {
        switch (status) {
            case 'success':
                return <CheckCircle className="w-5 h-5 text-success"/>;
            case 'error':
                return <AlertCircle className="w-5 h-5 text-destructive"/>;
            case 'uploading':
                return <div
                    className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"/>;
            default:
                return null;
        }
    };

    return (
        <div className="w-full space-y-6">
            <div
                className={`border-2 border-dashed rounded-2xl p-8 text-center transition-colors cursor-pointer ${
                    isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-input'
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
                <Upload className="w-10 h-10 text-info mx-auto mb-3"/>
                <p className="text-sm font-medium text-gray-700">{t('upload.dropzoneTitle')}</p>
                <p className="text-xs text-muted-foreground mt-1">{t('upload.supportedFormats')}</p>
            </div>

            <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2">
                <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wider sticky top-0 bg-white/95 backdrop-blur py-2 z-10">
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
                                        <Video className="w-6 h-6 text-info"/> :
                                        <File className="w-6 h-6 text-muted-foreground"/>}
                                </div>
                            )}
                        </div>
                        <div className="flex-1 min-w-0">
                            <div className="flex justify-between items-start">
                                <p className="font-medium text-sm text-gray-900 truncate">{fileItem.file.name}</p>
                                <span
                                    className="text-[10px] text-muted-foreground">{formatFileSize(fileItem.file.size)}</span>
                            </div>

                            {['uploading', 'initiating', 'completing'].includes(fileItem.status) ? (
                                <div className="mt-1.5 space-y-1">
                                    <div className="h-1 bg-muted rounded-full overflow-hidden">
                                        <div className="h-full bg-blue-600 transition-all"
                                             style={{width: `${fileItem.progress}%`}}/>
                                    </div>
                                    <div className="flex justify-between text-[10px] text-gray-500">
                                        <span>{fileItem.status === 'completing' ? 'merging...' : `${fileItem.progress}%`}</span>
                                        {fileItem.speed && <span>{formatFileSize(fileItem.speed)}/s</span>}
                                    </div>
                                </div>
                            ) : (
                                <p className={`text-[10px] mt-1 ${fileItem.status === 'success' ? 'text-success' : fileItem.status === 'error' ? 'text-destructive' : 'text-muted-foreground'}`}>
                                    {t(`upload.status${fileItem.status.charAt(0).toUpperCase()}${fileItem.status.slice(1)}`)}
                                </p>
                            )}
                        </div>
                        <div className="flex items-center gap-1">
                            {getStatusIcon(fileItem.status)}
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-muted-foreground hover:text-destructive"
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
                {files.length === 0 && (
                    <div className="text-center py-8 text-muted-foreground text-sm">No files selected</div>
                )}
            </div>

            {onCancel && (
                <div className="flex justify-end pt-2 border-t">
                    <Button variant="ghost" className="text-muted-foreground" onClick={onCancel}>
                        {t('common.cancel')}
                    </Button>
                </div>
            )}
        </div>
    );
}