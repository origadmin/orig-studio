/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Upload Page
 */

import React, {useState, useRef} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {Upload, X, File, Image, Video, CheckCircle, AlertCircle} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {mediaApi} from '@/lib/api/media';
import {formatFileSize} from '@/lib/format';
import {useTranslation} from 'react-i18next';

interface UploadFile {
    id: string;
    file: File;
    preview?: string;
    progress: number;
    status: 'pending' | 'uploading' | 'success' | 'error';
}

const categories = [
    "技术", "编程", "运维", "数据科学", "云计算",
    "前端", "职业", "音乐", "游戏", "娱乐",
];

const UploadPage = () => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const [files, setFiles] = useState<UploadFile[]>([]);
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [category, setCategory] = useState('');
    const [tags, setTags] = useState<string[]>([]);
    const [tagInput, setTagInput] = useState('');
    const [isDragging, setIsDragging] = useState(false);

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
        const valid = newFiles.filter(f =>
            validTypes.some(t => f.type.startsWith(t))
        );
        setFiles(prev => [
            ...prev,
            ...valid.map(f => ({
                id: Math.random().toString(36).substr(2, 9),
                file: f,
                preview: f.type.startsWith('image/') ? URL.createObjectURL(f) : undefined,
                progress: 0,
                status: 'pending' as const,
            })),
        ]);
    };

    const removeFile = (id: string) => setFiles(prev => prev.filter(f => f.id !== id));

    const addTag = () => {
        const val = tagInput.trim();
        if (val && !tags.includes(val)) {
            setTags([...tags, val]);
            setTagInput('');
        }
    };

    const removeTag = (tag: string) => setTags(tags.filter(t => t !== tag));

    const handleUpload = async () => {
        if (!title || files.length === 0) return;
        setFiles(prev => prev.map(f => ({...f, status: 'uploading'})));

        for (const file of files) {
            try {
                await mediaApi.upload(
                    file.file,
                    {
                        title,
                        description,
                        category_id: category ? parseInt(category) : undefined,
                        tags,
                        privacy: 1, // public
                    },
                    (percent) => {
                        setFiles(prev =>
                            prev.map(f => f.id === file.id ? {...f, progress: percent} : f)
                        );
                    },
                );
                setFiles(prev =>
                    prev.map(f => f.id === file.id ? {...f, status: 'success'} : f)
                );
            } catch (err) {
                console.error('Upload failed:', err);
                setFiles(prev =>
                    prev.map(f => f.id === file.id ? {...f, status: 'error'} : f)
                );
            }
        }

        // All uploads done, navigate after short delay
        setTimeout(() => navigate({to: '/'}), 1500);
    };

    const getFileIcon = (type: string) => {
        if (type.startsWith('video/')) return <Video className="w-8 h-8 text-blue-500"/>;
        if (type.startsWith('image/')) return <Image className="w-8 h-8 text-green-500"/>;
        return <File className="w-8 h-8 text-slate-500"/>;
    };

    const uploading = files.some(f => f.status === 'uploading');
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
                        {files.map(file => (
                            <div key={file.id}
                                 className="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                                <div
                                    className="w-16 h-16 bg-gray-200 dark:bg-gray-600 rounded-lg overflow-hidden shrink-0">
                                    {file.preview ? (
                                        <img src={file.preview} alt="" className="w-full h-full object-cover"/>
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center">
                                            {getFileIcon(file.file.type)}
                                        </div>
                                    )}
                                </div>
                                <div className="flex-1 min-w-0">
                                    <p className="font-medium text-gray-900 dark:text-white truncate">{file.file.name}</p>
                                    <p className="text-sm text-gray-500 dark:text-gray-400">{formatFileSize(file.file.size)}</p>
                                    {file.status === 'uploading' && (
                                        <div
                                            className="mt-2 h-2 bg-gray-200 dark:bg-gray-600 rounded-full overflow-hidden">
                                            <div className="h-full bg-blue-600 transition-all"
                                                 style={{width: `${file.progress}%`}}/>
                                        </div>
                                    )}
                                </div>
                                <div className="flex items-center gap-2">
                                    {file.status === 'success' && <CheckCircle className="w-5 h-5 text-green-500"/>}
                                    {file.status === 'error' && <AlertCircle className="w-5 h-5 text-red-500"/>}
                                    {file.status === 'pending' && (
                                        <Button variant="ghost" size="sm" onClick={() => removeFile(file.id)}>
                                            <X className="w-4 h-4"/>
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
                        <label
                            className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('upload.titleLabel')}</label>
                        <Input placeholder={t('upload.titlePlaceholder')} value={title}
                               onChange={(e) => setTitle(e.target.value)}/>
                    </div>
                    <div className="space-y-2">
                        <label
                            className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('upload.descLabel')}</label>
                        <Textarea placeholder={t('upload.descPlaceholder')} rows={4} value={description}
                                  onChange={(e) => setDescription(e.target.value)}/>
                    </div>
                    <div className="space-y-2">
                        <label
                            className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('upload.categoryLabel')}</label>
                        <select
                            className="w-full h-10 px-3 rounded-md border border-input bg-background text-sm dark:bg-gray-700 dark:text-white"
                            value={category}
                            onChange={(e) => setCategory(e.target.value)}
                        >
                            <option value="">{t('upload.selectCategory')}</option>
                            {categories.map(cat => (
                                <option key={cat} value={cat}>{cat}</option>
                            ))}
                        </select>
                    </div>
                    <div className="space-y-2">
                        <label
                            className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('upload.tagLabel')}</label>
                        <div className="flex gap-2 mb-2">
                            <Input placeholder={t('upload.addTagPlaceholder')} value={tagInput}
                                onChange={(e) => setTagInput(e.target.value)}
                                   onKeyDown={(e) => e.key === 'Enter' && addTag()}/>
                            <Button variant="outline" onClick={addTag}>{t('upload.addTag')}</Button>
                        </div>
                        <div className="flex flex-wrap gap-2">
                            {tags.map(tag => (
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
                    {uploading ? t('upload.uploading') : t('upload.upload')}
                </Button>
            </div>
        </div>
    );
};

export default UploadPage;
