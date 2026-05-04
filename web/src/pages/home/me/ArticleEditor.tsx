/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * User-side Article Editor (simplified 2-column layout)
 */

import {useState, useEffect, useCallback} from 'react';
import {useNavigate, useParams} from '@tanstack/react-router';
import {articleApi, userArticleApi, type Article, type UserCreateArticleRequest, type UserUpdateArticleRequest, type MediaBrief} from '@/lib/api/article';
import {mediaApi, type Media} from '@/lib/api/media';
import {useCategoryList} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {useDirtyState, useSaveState, useKeyboardShortcut} from '@/hooks/useEditPage';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Label} from '@/components/ui/label';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Badge} from '@/components/ui/badge';
import {Spinner} from '@/components/ui/spinner';
import {toast} from 'sonner';
import {useTranslation} from 'react-i18next';
import {
    ArrowLeft,
    AlertTriangle,
    Film,
    X,
    Search,
    FileText,
} from 'lucide-react';
import {generateSlug} from '@/lib/utils/slug';

// ============================================================================
// User Media Selector Dialog (only shows user's own videos)
// ============================================================================

interface UserMediaSelectorDialogProps {
    open: boolean;
    onClose: () => void;
    onSelect: (media: Media) => void;
}

function UserMediaSelectorDialog({open, onClose, onSelect}: UserMediaSelectorDialogProps) {
    const {user} = useAuth();
    const [medias, setMedias] = useState<Media[]>([]);
    const [loading, setLoading] = useState(false);
    const [search, setSearch] = useState('');
    const [selectedId, setSelectedId] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const pageSize = 12;

    useEffect(() => {
        if (!open) return;
        setLoading(true);
        // Only fetch user's own videos
        mediaApi.list({
            page,
            page_size: pageSize,
            type: 'video',
            user_id: user?.id || undefined,
            keyword: search || undefined,
        })
            .then(res => {
                const items = res?.items || [];
                setMedias(items);
                setTotal(res?.total || 0);
            })
            .catch(() => {
                setMedias([]);
            })
            .finally(() => setLoading(false));
    }, [open, page, search, user?.id]);

    const handleSelect = () => {
        const media = medias.find(m => m.id === selectedId);
        if (media) {
            onSelect(media);
            onClose();
        }
    };

    if (!open) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
            <div className="bg-background rounded-lg shadow-xl w-full max-w-3xl max-h-[80vh] flex flex-col">
                <div className="flex items-center justify-between p-4 border-b">
                    <h3 className="font-semibold text-lg">Select Video</h3>
                    <Button variant="ghost" size="icon" onClick={onClose}>
                        <X className="w-4 h-4"/>
                    </Button>
                </div>

                <div className="p-4 border-b">
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder="Search your videos..."
                            value={search}
                            onChange={e => {
                                setSearch(e.target.value);
                                setPage(1);
                            }}
                            className="pl-10"
                        />
                    </div>
                </div>

                <div className="flex-1 overflow-y-auto p-4">
                    {loading ? (
                        <div className="flex items-center justify-center py-12">
                            <Spinner/>
                        </div>
                    ) : medias.length === 0 ? (
                        <div className="text-center py-12 text-muted-foreground">
                            No videos found
                        </div>
                    ) : (
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                            {medias.map(media => (
                                <button
                                    key={media.id}
                                    className={`relative rounded-lg border-2 overflow-hidden text-left transition-colors ${
                                        selectedId === media.id
                                            ? 'border-primary'
                                            : 'border-transparent hover:border-muted-foreground/30'
                                    }`}
                                    onClick={() => setSelectedId(media.id)}
                                >
                                    <div className="aspect-video bg-muted relative">
                                        {media.thumbnail ? (
                                            <img
                                                src={media.thumbnail}
                                                alt={media.title}
                                                className="w-full h-full object-cover"
                                                loading="lazy"
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <Film className="w-8 h-8 text-muted-foreground"/>
                                            </div>
                                        )}
                                        {media.duration > 0 && (
                                            <span className="absolute bottom-1 right-1 bg-black/80 text-white text-[10px] px-1.5 py-0.5 rounded">
                                                {Math.floor(media.duration / 60)}:{String(Math.floor(media.duration % 60)).padStart(2, '0')}
                                            </span>
                                        )}
                                    </div>
                                    <div className="p-2">
                                        <p className="text-xs font-medium truncate">{media.title}</p>
                                    </div>
                                </button>
                            ))}
                        </div>
                    )}
                </div>

                <div className="flex items-center justify-between p-4 border-t">
                    <span className="text-sm text-muted-foreground">
                        {total} video(s) found
                    </span>
                    <div className="flex gap-2">
                        <Button variant="outline" onClick={onClose}>Cancel</Button>
                        <Button onClick={handleSelect} disabled={!selectedId}>
                            Confirm
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}

// ============================================================================
// UserArticleEditor
// ============================================================================

export default function UserArticleEditor({mode}: { mode: 'create' | 'edit' }) {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const {data: categoriesData} = useCategoryList();

    // Get article ID from URL params (same pattern as admin ArticleEdit.tsx)
    const {id: routeId} = useParams({strict: false}) as {id?: string};
    const articleId = mode === 'edit' ? routeId : undefined;

    // Article data for edit mode
    const [article, setArticle] = useState<Article | null>(null);
    const [loading, setLoading] = useState(mode === 'edit');
    const [loadError, setLoadError] = useState<string | null>(null);

    // Form state with dirty tracking
    const {form, setForm, isDirty, resetDirty, syncFromData} = useDirtyState({
        title: '',
        content: '',
        summary: '',
        state: 'draft' as string,
        category_id: '' as string | number,
        media_id: '',
        thumbnail: '',
        tags: '',
    });

    // Selected media info (for display in sidebar)
    const [selectedMedia, setSelectedMedia] = useState<MediaBrief | null>(null);
    const [mediaSelectorOpen, setMediaSelectorOpen] = useState(false);

    // Save state management
    const {isSaving, setSaving, setSuccess, setError} = useSaveState();

    // Load article data in edit mode
    useEffect(() => {
        if (mode !== 'edit' || !articleId) return;
        setLoading(true);
        setLoadError(null);
        articleApi.get(articleId)
            .then(data => {
                setArticle(data);
                setSelectedMedia(data.media || null);
                syncFromData({
                    title: data.title || '',
                    content: data.content || '',
                    summary: data.summary || '',
                    state: data.state || 'draft',
                    category_id: data.category_id ?? '',
                    media_id: data.media_id || '',
                    thumbnail: data.thumbnail || '',
                    tags: data.tags?.join(', ') || '',
                });
            })
            .catch(err => {
                setLoadError(t('articleEditor.loadFailed'));
                console.error('Error loading article:', err);
            })
            .finally(() => setLoading(false));
    }, [mode, articleId, syncFromData, t]);

    // Auto-generate slug from title (display only, not editable)
    const displaySlug = form.title ? generateSlug(form.title) : '';

    // Save handler
    const handleSave = useCallback(async () => {
        if (isSaving) return;
        if (!form.title || !form.content) {
            toast.error(t('articleEditor.requiredFields'));
            return;
        }
        setSaving();
        try {
            const tagsArray = form.tags.split(',').map(s => s.trim()).filter(Boolean);
            const categoryId = form.category_id !== '' && form.category_id !== undefined
                ? Number(form.category_id) : undefined;

            if (mode === 'create') {
                const data: UserCreateArticleRequest = {
                    title: form.title,
                    content: form.content,
                    summary: form.summary || undefined,
                    state: (form.state as 'draft' | 'published') || 'draft',
                    category_id: categoryId,
                    media_id: form.media_id || undefined,
                    thumbnail: form.thumbnail || undefined,
                    tags: tagsArray.length > 0 ? tagsArray : undefined,
                };
                const created = await userArticleApi.create(data);
                resetDirty();
                setSuccess();
                toast.success(t('articleEditor.saveSuccess'));
                // Navigate to edit page with new ID
                if (created?.id) {
                    navigate({to: '/me/articles/$id/edit', params: {id: created.id}});
                }
            } else if (articleId) {
                const data: UserUpdateArticleRequest = {
                    title: form.title,
                    content: form.content,
                    summary: form.summary,
                    state: (form.state as 'draft' | 'published'),
                    category_id: categoryId,
                    media_id: form.media_id || undefined,
                    thumbnail: form.thumbnail,
                    tags: tagsArray,
                };
                await userArticleApi.update(articleId, data);
                resetDirty();
                setSuccess();
                toast.success(t('articleEditor.saveSuccess'));
            }
        } catch (err: any) {
            setError();
            const status = err?.status || err?.response?.status;
            const msg = err?.message || err?.response?.data?.message || '';
            if (status === 403 || msg.toLowerCase().includes('forbidden') || msg.toLowerCase().includes('disabled')) {
                toast.error(t('articleEditor.featureDisabled') || 'Article feature is not available', {
                    description: t('articleEditor.featureDisabledDesc') || 'This feature has been disabled by the administrator.',
                });
            } else {
                toast.error(t('articleEditor.saveFailed'), {
                    description: msg || '',
                });
            }
            console.error('Failed to save', err);
        }
    }, [mode, articleId, isSaving, form, resetDirty, setSaving, setSuccess, setError, navigate, toast, t]);

    // Back handler
    const handleBack = useCallback(() => {
        navigate({to: '/me/articles'});
    }, [navigate]);

    // Media selector handler
    const handleMediaSelect = useCallback((media: Media) => {
        setForm(prev => ({
            ...prev,
            media_id: media.id,
            thumbnail: media.thumbnail || prev.thumbnail,
        }));
        setSelectedMedia({
            id: media.id,
            title: media.title,
            thumbnail: media.thumbnail,
            duration: media.duration,
            type: media.type,
            short_token: media.short_token,
        });
    }, [setForm]);

    // Clear media handler
    const handleClearMedia = useCallback(() => {
        setForm(prev => ({
            ...prev,
            media_id: '',
        }));
        setSelectedMedia(null);
    }, [setForm]);

    // Keyboard shortcut: Ctrl+S / Cmd+S
    useKeyboardShortcut('ctrl+s', handleSave, {enabled: !isSaving});

    // Categories for select
    const categories = categoriesData?.items || categoriesData || [];

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Spinner/>
            </div>
        );
    }

    if (loadError) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
                <AlertTriangle className="w-12 h-12 text-destructive"/>
                <p className="text-lg text-muted-foreground">{loadError}</p>
                <Button variant="outline" onClick={handleBack}>
                    <ArrowLeft className="w-4 h-4 mr-2"/>Back to My Articles
                </Button>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            {/* Header */}
            <div className="sticky top-14 z-40 bg-background border-b">
                <div className="flex items-center justify-between px-6 py-3">
                    <div className="flex items-center gap-3">
                        <Button variant="ghost" size="sm" onClick={handleBack}>
                            <ArrowLeft className="w-4 h-4 mr-1"/>
                            {t('common.back') || 'Back'}
                        </Button>
                        <h1 className="text-lg font-semibold">
                            {mode === 'create' ? t('articleEditor.createTitle') : t('articleEditor.editTitle')}
                        </h1>
                        {mode === 'edit' && article && (
                            <Badge variant={article.state === 'published' ? 'default' : 'secondary'} className="text-xs">
                                {article.state}
                            </Badge>
                        )}
                    </div>
                    <Button onClick={handleSave} disabled={isSaving}>
                        {isSaving ? t('articleEditor.saving') : t('articleEditor.save')}
                    </Button>
                </div>
            </div>

            {/* Main content: 2-column layout */}
            <div className="max-w-7xl mx-auto px-6 py-6">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    {/* Left 2/3: Content editing */}
                    <div className="lg:col-span-2 space-y-6">
                        <div className="bg-card rounded-lg border p-6 space-y-6">
                            {/* Title */}
                            <div className="space-y-2">
                                <Label htmlFor="title">{t('articleEditor.titlePlaceholder')}</Label>
                                <Input
                                    id="title"
                                    value={form.title}
                                    onChange={e => setForm({...form, title: e.target.value})}
                                    placeholder={t('articleEditor.titlePlaceholder')}
                                />
                            </div>

                            {/* Slug (read-only display) */}
                            {displaySlug && (
                                <div className="space-y-1">
                                    <Label className="text-xs text-muted-foreground">
                                        {t('articleEditor.slugLabel')}
                                    </Label>
                                    <p className="text-sm text-muted-foreground bg-muted/50 px-3 py-2 rounded-md font-mono">
                                        {displaySlug}
                                    </p>
                                    <p className="text-xs text-muted-foreground">
                                        {t('articleEditor.slugAutoGenerated')}
                                    </p>
                                </div>
                            )}

                            {/* Content */}
                            <div className="space-y-2">
                                <Label htmlFor="content">{t('articleEditor.contentPlaceholder')}</Label>
                                <textarea
                                    id="content"
                                    value={form.content}
                                    onChange={e => setForm({...form, content: e.target.value})}
                                    placeholder={t('articleEditor.contentPlaceholder')}
                                    className="w-full min-h-[400px] bg-background border rounded-md px-3 py-2 text-sm font-mono resize-y focus:outline-none focus:ring-2 focus:ring-ring"
                                />
                            </div>

                            {/* Summary */}
                            <div className="space-y-2">
                                <Label htmlFor="summary">{t('articleEditor.summaryPlaceholder')}</Label>
                                <textarea
                                    id="summary"
                                    value={form.summary}
                                    onChange={e => setForm({...form, summary: e.target.value})}
                                    placeholder={t('articleEditor.summaryPlaceholder')}
                                    className="w-full min-h-[80px] bg-background border rounded-md px-3 py-2 text-sm resize-y focus:outline-none focus:ring-2 focus:ring-ring"
                                />
                            </div>

                            {/* Tags */}
                            <div className="space-y-2">
                                <Label htmlFor="tags">{t('articleEditor.tagsPlaceholder')}</Label>
                                <Input
                                    id="tags"
                                    value={form.tags}
                                    onChange={e => setForm({...form, tags: e.target.value})}
                                    placeholder={t('articleEditor.tagsPlaceholder')}
                                />
                            </div>
                        </div>
                    </div>

                    {/* Right 1/3: Video + Settings */}
                    <div className="space-y-6">
                        {/* Associated Video */}
                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <Label className="text-sm font-medium">{t('articleEditor.videoLabel')}</Label>
                            {selectedMedia ? (
                                <div className="space-y-3">
                                    <div className="aspect-video bg-muted rounded-md overflow-hidden relative">
                                        {selectedMedia.thumbnail ? (
                                            <img
                                                src={selectedMedia.thumbnail}
                                                alt={selectedMedia.title}
                                                className="w-full h-full object-cover"
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <Film className="w-8 h-8 text-muted-foreground"/>
                                            </div>
                                        )}
                                    </div>
                                    <div className="flex items-center justify-between">
                                        <p className="text-sm font-medium truncate flex-1">{selectedMedia.title}</p>
                                        <Button variant="ghost" size="sm" onClick={handleClearMedia}>
                                            <X className="w-3.5 h-3.5"/>
                                        </Button>
                                    </div>
                                </div>
                            ) : (
                                <div className="aspect-video bg-muted rounded-md flex flex-col items-center justify-center gap-2">
                                    <FileText className="w-8 h-8 text-muted-foreground"/>
                                    <p className="text-xs text-muted-foreground">{t('articleEditor.noVideo')}</p>
                                </div>
                            )}
                            <Button
                                variant="outline"
                                size="sm"
                                className="w-full"
                                onClick={() => setMediaSelectorOpen(true)}
                            >
                                {selectedMedia ? t('articleEditor.changeVideo') : t('articleEditor.selectVideo')}
                            </Button>
                        </div>

                        {/* Publish Settings */}
                        <div className="bg-card rounded-lg border p-4 space-y-4">
                            <Label className="text-sm font-medium">{t('articleEditor.publishSettings')}</Label>

                            {/* State - only draft/published */}
                            <div className="space-y-2">
                                <Label className="text-xs text-muted-foreground">{t('articleEditor.stateLabel')}</Label>
                                <Select
                                    value={form.state}
                                    onValueChange={val => setForm({...form, state: val})}
                                >
                                    <SelectTrigger>
                                        <SelectValue/>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="draft">{t('articleEditor.stateDraft')}</SelectItem>
                                        <SelectItem value="published">{t('articleEditor.statePublished')}</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Category */}
                            <div className="space-y-2">
                                <Label className="text-xs text-muted-foreground">{t('articleEditor.categoryLabel')}</Label>
                                <Select
                                    value={form.category_id ? String(form.category_id) : '__none__'}
                                    onValueChange={val => setForm({...form, category_id: val === '__none__' ? '' : Number(val)})}
                                >
                                    <SelectTrigger>
                                        <SelectValue placeholder={t('articleEditor.selectCategory')}/>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="__none__">{t('articleEditor.selectCategory')}</SelectItem>
                                        {Array.isArray(categories) && categories.map((cat: any) => (
                                            <SelectItem key={cat.id} value={String(cat.id)}>
                                                {cat.name}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>

                            {/* Thumbnail URL */}
                            <div className="space-y-2">
                                <Label className="text-xs text-muted-foreground">{t('articleEditor.thumbnailLabel')}</Label>
                                <Input
                                    value={form.thumbnail}
                                    onChange={e => setForm({...form, thumbnail: e.target.value})}
                                    placeholder={t('articleEditor.thumbnailPlaceholder')}
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Media Selector Dialog */}
            <UserMediaSelectorDialog
                open={mediaSelectorOpen}
                onClose={() => setMediaSelectorOpen(false)}
                onSelect={handleMediaSelect}
            />
        </div>
    );
}
