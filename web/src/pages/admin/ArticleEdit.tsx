/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Admin - Article Create/Edit Page (Video Website Style)
 */

import {useState, useEffect, useMemo, useCallback} from 'react';
import {useParams, useNavigate} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {adminArticleApi, type Article, type CreateArticleRequest, type UpdateArticleRequest, type MediaBrief} from '@/lib/api/article';
import {adminMediaApi, type Media} from '@/lib/api/media';
import {useCategoryList} from '@/hooks/queries';
import {api, API_BASE_URL} from '@/lib/request';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Label} from '@/components/ui/label';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Badge} from '@/components/ui/badge';
import {Separator} from '@/components/ui/separator';
import {EditPageHeader, type HeaderBadgeConfig} from '@/components/common/EditPageHeader';
import {DeleteConfirmDialog} from '@/components/common/DeleteConfirmDialog';
import {useDirtyState, useSaveState, useKeyboardShortcut} from '@/hooks/useEditPage';
import {Spinner} from '@/components/ui/spinner';
import {
    ArrowLeft, AlertTriangle, Film, Play, X,
    CheckCircle, Clock, Eye, Image as ImageIcon,
    Search, ExternalLink
} from 'lucide-react';
import {formatDateTime} from '@/lib/format';
import {generateSlug} from '@/lib/utils/slug';
import {toast} from 'sonner';

/**
 * Resolve a potentially relative URL to a full URL.
 */
function resolveMediaUrl(url: string | undefined): string | undefined {
    if (!url) return undefined;
    if (/^(https?:|data:|blob:)/i.test(url)) return url;
    const base = API_BASE_URL || '';
    return `${base}/${url.replace(/^\//, '')}`;
}

/**
 * Map Article state to Badge variant and label
 */
function getStateBadgeConfig(state: string, t: (key: string) => string): { variant: HeaderBadgeConfig['variant']; label: string } {
    const map: Record<string, { variant: HeaderBadgeConfig['variant']; label: string }> = {
        draft: {variant: 'secondary', label: t('admin.draft')},
        published: {variant: 'default', label: t('admin.published')},
        archived: {variant: 'destructive', label: t('admin.archived')},
    };
    return map[state] || {variant: 'outline' as const, label: state};
}

function mapArticleToHeaderBadges(article: Article, t: (key: string) => string): HeaderBadgeConfig[] {
    const badges: HeaderBadgeConfig[] = [];

    // State Badge
    const stateConfig = getStateBadgeConfig(article.state, t);
    badges.push({
        type: 'state',
        variant: stateConfig.variant,
        label: stateConfig.label,
        ariaLabel: `State: ${stateConfig.label}`,
    });

    // Featured Badge
    if (article.featured) {
        badges.push({
            type: 'featured',
            variant: 'outline',
            label: t('admin.featured'),
            ariaLabel: t('admin.featuredArticle'),
            className: 'text-warning border-amber-300',
        });
    }

    // Media Badge
    if (article.media_id) {
        badges.push({
            type: 'media-type',
            variant: 'outline',
            label: t('admin.video'),
            ariaLabel: t('admin.hasAssociatedVideo'),
        });
    }

    return badges;
}

// ============================================================================
// Media Selector Dialog
// ============================================================================

interface MediaSelectorDialogProps {
    open: boolean;
    onClose: () => void;
    onSelect: (media: Media) => void;
}

function MediaSelectorDialog({open, onClose, onSelect}: MediaSelectorDialogProps) {
    const {t} = useTranslation();
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
        adminMediaApi.list({page, page_size: pageSize, type: 'video', keyword: search || undefined})
            .then(res => {
                const items = res?.items || [];
                setMedias(items);
                setTotal(res?.total || 0);
            })
            .catch(() => {
                setMedias([]);
            })
            .finally(() => setLoading(false));
    }, [open, page, search]);

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
                    <h3 className="font-semibold text-lg">{t('admin.selectVideo')}</h3>
                    <Button variant="ghost" size="icon" onClick={onClose}>
                        <X className="w-4 h-4"/>
                    </Button>
                </div>

                <div className="p-4 border-b">
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder={t('admin.searchVideos')}
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
                            {t('admin.noVideosFound')}
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
                                                src={resolveMediaUrl(media.thumbnail)}
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
                        {t('admin.videosFound', {count: total})}
                    </span>
                    <div className="flex gap-2">
                        <Button variant="outline" onClick={onClose}>{t('admin.cancel')}</Button>
                        <Button onClick={handleSelect} disabled={!selectedId}>
                            {t('admin.confirm')}
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}

// ============================================================================
// ArticleEditPage
// ============================================================================

interface ArticleEditPageProps {
    mode: 'create' | 'edit';
}

export default function ArticleEditPage({mode}: ArticleEditPageProps) {
    const {t} = useTranslation();
    const {id} = useParams({strict: false}) as {id?: string};
    const navigate = useNavigate();
    const {data: categoriesData} = useCategoryList();

    // Article data for edit mode
    const [article, setArticle] = useState<Article | null>(null);
    const [loading, setLoading] = useState(mode === 'edit');
    const [loadError, setLoadError] = useState<string | null>(null);

    // Form state with dirty tracking
    const {form, setForm, isDirty, resetDirty, syncFromData} = useDirtyState({
        title: '',
        slug: '',
        content: '',
        summary: '',
        state: 'draft',
        category_id: '' as string | number,
        media_id: '',
        thumbnail: '',
        tags: '',
        featured: false,
    });

    // Selected media info (for display in sidebar)
    const [selectedMedia, setSelectedMedia] = useState<MediaBrief | null>(null);
    const [thumbnailError, setThumbnailError] = useState(false);
    const [activeTab, setActiveTab] = useState<'content' | 'publish'>('content');
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [mediaSelectorOpen, setMediaSelectorOpen] = useState(false);

    // Save state management
    const {saveState, isSaving, setSaving, setSuccess, setError} = useSaveState();

    // Track whether slug was manually edited
    const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);

    // Load article data in edit mode
    useEffect(() => {
        if (mode !== 'edit' || !id) return;
        setLoading(true);
        setLoadError(null);
        adminArticleApi.get(id)
            .then(data => {
                setArticle(data);
                setSelectedMedia(data.media || null);
                syncFromData({
                    title: data.title || '',
                    slug: data.slug || '',
                    content: data.content || '',
                    summary: data.summary || '',
                    state: data.state || 'draft',
                    category_id: data.category_id ?? '',
                    media_id: data.media_id || '',
                    thumbnail: data.thumbnail || '',
                    tags: data.tags?.join(', ') || '',
                    featured: data.featured || false,
                });
            })
            .catch(err => {
                setLoadError(t('admin.failedToLoadArticle'));
                console.error('Error loading article:', err);
            })
            .finally(() => setLoading(false));
    }, [mode, id, syncFromData]);

    // Auto-generate slug from title
    useEffect(() => {
        if (!slugManuallyEdited && form.title && !form.slug) {
            setForm(prev => ({...prev, slug: generateSlug(form.title)}));
        }
    }, [form.title, slugManuallyEdited, form.slug, setForm]);

    // Save handler
    const handleSave = useCallback(async () => {
        if (isSaving) return;
        setSaving();
        try {
            const tagsArray = form.tags.split(',').map(s => s.trim()).filter(Boolean);
            const categoryId = form.category_id !== '' && form.category_id !== undefined
                ? Number(form.category_id) : undefined;

            if (mode === 'create') {
                const data: CreateArticleRequest = {
                    title: form.title,
                    slug: form.slug || generateSlug(form.title),
                    content: form.content,
                    summary: form.summary || undefined,
                    state: form.state || 'draft',
                    category_id: categoryId,
                    media_id: form.media_id || undefined,
                    thumbnail: form.thumbnail || undefined,
                    tags: tagsArray.length > 0 ? tagsArray : undefined,
                    featured: form.featured,
                };
                const created = await adminArticleApi.create(data);
                resetDirty();
                setSuccess();
                toast.success(t('admin.articleCreated'));
                // Navigate to edit page with new ID
                if (created?.id) {
                    navigate({to: '/admin/articles/$id/edit', params: {id: created.id}});
                }
            } else if (id) {
                const data: UpdateArticleRequest = {
                    title: form.title,
                    slug: form.slug,
                    content: form.content,
                    summary: form.summary,
                    state: form.state,
                    category_id: categoryId,
                    media_id: form.media_id || undefined,
                    thumbnail: form.thumbnail,
                    tags: tagsArray,
                    featured: form.featured,
                };
                await adminArticleApi.update(id, data);
                resetDirty();
                setSuccess();
                toast.success(t('admin.articleSaved'));
            }
        } catch (err: any) {
            setError();
            toast.error(`${t('admin.saveFailed')}: ${err?.message || t('admin.unknownError')}`);
            console.error('Failed to save', err);
        }
    }, [mode, id, isSaving, form, resetDirty, setSaving, setSuccess, setError, navigate]);

    // Delete handler
    const handleDelete = useCallback(async () => {
        if (!id) return;
        setIsDeleting(true);
        try {
            await adminArticleApi.delete(id);
            setDeleteDialogOpen(false);
            toast.success(t('admin.articleDeleted'));
            navigate({to: '/admin/articles'});
        } catch (err: any) {
            setIsDeleting(false);
            toast.error(`${t('admin.deleteFailed')}: ${err?.message || t('admin.unknownError')}`);
            console.error('Failed to delete', err);
        }
    }, [id, navigate]);

    // Preview handler
    const handlePreview = useCallback(() => {
        if (article?.slug) {
            window.open(`/articles/${article.slug}`, '_blank', 'noopener,noreferrer');
        }
    }, [article?.slug]);

    // Back handler
    const handleBack = useCallback(() => {
        navigate({to: '/admin/articles'});
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

    // Compute header badges
    const headerBadges = useMemo(() => {
        if (mode === 'create') {
            return [{type: 'state' as const, variant: 'secondary' as const, label: t('admin.draft'), ariaLabel: `State: ${t('admin.draft')}`}];
        }
        if (article) return mapArticleToHeaderBadges(article, t);
        return [];
    }, [mode, article, t]);

    // Resolve thumbnail for display
    const displayThumbnail = resolveMediaUrl(form.thumbnail || selectedMedia?.thumbnail);

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
                    <ArrowLeft className="w-4 h-4 mr-2"/>{t('admin.backToList')}
                </Button>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            <EditPageHeader
                title={mode === 'create' ? t('admin.createArticle') : (article?.title || t('admin.untitledArticle'))}
                isDirty={isDirty}
                isSaving={isSaving}
                saveState={saveState}
                onBack={handleBack}
                onSave={handleSave}
                onPreview={mode === 'edit' && article?.slug ? handlePreview : undefined}
                onDelete={mode === 'edit' ? () => setDeleteDialogOpen(true) : () => {}}
                badges={headerBadges}
            />

            <div className="max-w-7xl mx-auto px-6 py-6">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    {/* Left 2/3: Content + Publish tabs */}
                    <div className="lg:col-span-2 space-y-6">
                        <div className="flex gap-1 border-b">
                            {(['content', 'publish'] as const).map(tab => (
                                <button key={tab}
                                        className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                                            activeTab === tab
                                                ? 'border-primary text-primary'
                                                : 'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        onClick={() => setActiveTab(tab)}>
                                    {{content: t('admin.content'), publish: t('admin.publishSettings')}[tab]}
                                </button>
                            ))}
                        </div>

                        {activeTab === 'content' && (
                            <div className="space-y-6 bg-card rounded-lg border p-6">
                                <div className="space-y-2">
                                    <Label htmlFor="title">{t('admin.title')}</Label>
                                    <Input id="title" value={form.title}
                                           onChange={e => {
                                               setForm({...form, title: e.target.value});
                                               if (!slugManuallyEdited) {
                                                   setForm(prev => ({...prev, slug: generateSlug(e.target.value)}));
                                               }
                                           }}
                                           placeholder={t('admin.articleTitle')}/>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="slug">{t('admin.slug')}</Label>
                                    <Input id="slug" value={form.slug}
                                           onChange={e => {
                                               setForm({...form, slug: e.target.value});
                                               setSlugManuallyEdited(true);
                                           }}
                                           placeholder="article-url-slug"/>
                                    <p className="text-xs text-muted-foreground">
                                        {t('admin.autoGeneratedFromTitle')}
                                    </p>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="content">{t('admin.contentMarkdown')}</Label>
                                    <textarea id="content"
                                              className="flex min-h-[400px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 font-mono"
                                              value={form.content}
                                              onChange={e => setForm({...form, content: e.target.value})}
                                              placeholder={t('admin.writeContent')}/>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="summary">{t('admin.summary')}</Label>
                                    <textarea id="summary"
                                              className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                                              value={form.summary}
                                              onChange={e => setForm({...form, summary: e.target.value})}
                                              placeholder={t('admin.briefSummary')}/>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="tags">{t('admin.tagsCommaSeparated')}</Label>
                                    <Input id="tags" value={form.tags}
                                           onChange={e => setForm({...form, tags: e.target.value})}
                                           placeholder={t('admin.tagsPlaceholder')}/>
                                    {form.tags && (
                                        <div className="flex flex-wrap gap-1 mt-2">
                                            {form.tags.split(',').map((tag, i) => tag.trim() && (
                                                <Badge key={i} variant="secondary" className="text-xs">{tag.trim()}</Badge>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}

                        {activeTab === 'publish' && (
                            <div className="space-y-6 bg-card rounded-lg border p-6">
                                <div className="space-y-2">
                                    <Label>{t('admin.state')}</Label>
                                    <Select value={form.state} onValueChange={val => setForm({...form, state: val})}>
                                        <SelectTrigger><SelectValue/></SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="draft">{t('admin.draft')}</SelectItem>
                                            <SelectItem value="published">{t('admin.published')}</SelectItem>
                                            <SelectItem value="archived">{t('admin.archived')}</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-2">
                                    <Label>{t('admin.category')}</Label>
                                    <Select value={form.category_id !== '' && form.category_id !== undefined ? String(form.category_id) : '_none_'}
                                            onValueChange={val => setForm({...form, category_id: val === '_none_' ? '' : val})}>
                                        <SelectTrigger><SelectValue placeholder={t('admin.selectCategory')}/></SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="_none_">{t('admin.noCategory')}</SelectItem>
                                            {(Array.isArray(categoriesData?.items) ? categoriesData.items : Array.isArray(categoriesData) ? categoriesData : []).map((cat: any) => (
                                                <SelectItem key={cat.id} value={String(cat.id)}>
                                                    {cat.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <Separator/>
                                <div className="flex items-center gap-3">
                                    <input type="checkbox" id="featured" checked={form.featured}
                                           onChange={e => setForm({...form, featured: e.target.checked})}
                                           className="h-4 w-4 rounded border-input text-primary focus:ring-primary"/>
                                    <div>
                                        <Label htmlFor="featured" className="cursor-pointer">{t('admin.featuredArticle')}</Label>
                                        <p className="text-xs text-muted-foreground">{t('admin.displayInFeatured')}</p>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="thumbnail">{t('admin.thumbnailUrl')}</Label>
                                    <Input id="thumbnail" value={form.thumbnail}
                                           onChange={e => setForm({...form, thumbnail: e.target.value})}
                                           placeholder={t('admin.customThumbnailUrl')}/>
                                    <p className="text-xs text-muted-foreground">
                                        {t('admin.videoThumbnailDefault')}
                                    </p>
                                </div>
                            </div>
                        )}
                    </div>

                    {/* Right 1/3: Video panel + Metadata */}
                    <div className="space-y-6">
                        {/* Video Preview Panel */}
                        <div className="bg-card rounded-lg border p-4">
                            <h3 className="font-medium mb-3">{t('admin.associatedVideo')}</h3>
                            {form.media_id && selectedMedia ? (
                                <div className="space-y-3">
                                    <div className="relative aspect-video bg-muted rounded-md overflow-hidden">
                                        {displayThumbnail && !thumbnailError ? (
                                            <img src={displayThumbnail} alt={selectedMedia.title}
                                                 className="w-full h-full object-cover"
                                                 onError={() => setThumbnailError(true)}/>
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center">
                                                <Play className="w-8 h-8 text-muted-foreground"/>
                                            </div>
                                        )}
                                        {selectedMedia.duration > 0 && (
                                            <span className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-0.5 rounded">
                                                {Math.floor(selectedMedia.duration / 60)}:{String(Math.floor(selectedMedia.duration % 60)).padStart(2, '0')}
                                            </span>
                                        )}
                                    </div>
                                    <div>
                                        <p className="text-sm font-medium truncate">{selectedMedia.title}</p>
                                        <p className="text-xs text-muted-foreground">{selectedMedia.type}</p>
                                    </div>
                                    <div className="flex gap-2">
                                        {selectedMedia.short_token && (
                                            <Button variant="outline" size="sm" className="flex-1"
                                                    onClick={() => window.open(`/watch?v=${selectedMedia.short_token}`, '_blank')}>
                                                <ExternalLink className="w-3 h-3 mr-1"/>{t('admin.preview')}
                                            </Button>
                                        )}
                                        <Button variant="outline" size="sm" className="flex-1"
                                                onClick={() => setMediaSelectorOpen(true)}>
                                            {t('admin.change')}
                                        </Button>
                                        <Button variant="ghost" size="sm"
                                                onClick={handleClearMedia}
                                                title={t('admin.removeVideoAssociation')}>
                                            <X className="w-3 h-3"/>
                                        </Button>
                                    </div>
                                </div>
                            ) : (
                                <div className="space-y-3">
                                    <div className="w-full aspect-video bg-muted rounded-md flex flex-col items-center justify-center gap-2">
                                        <Film className="w-10 h-10 text-muted-foreground"/>
                                        <p className="text-sm text-muted-foreground">{t('admin.noVideoAssociated')}</p>
                                    </div>
                                    <Button variant="outline" className="w-full"
                                            onClick={() => setMediaSelectorOpen(true)}>
                                        <Film className="w-4 h-4 mr-2"/>{t('admin.selectVideoBtn')}
                                    </Button>
                                </div>
                            )}
                        </div>

                        {/* Metadata Card */}
                        {mode === 'edit' && article && (
                            <div className="bg-card rounded-lg border p-4 space-y-3">
                                <h3 className="font-medium">{t('admin.metadata')}</h3>
                                <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
                                    <span className="text-muted-foreground">{t('admin.id')}</span>
                                    <span className="font-mono text-xs text-right break-all">{article.id}</span>
                                    <span className="text-muted-foreground">{t('admin.slug')}</span>
                                    <span className="text-xs text-right break-all">{article.slug || t('admin.na')}</span>
                                    <span className="text-muted-foreground">{t('admin.views')}</span>
                                    <span className="text-xs text-right">{article.view_count}</span>
                                    <span className="text-muted-foreground">{t('admin.comments')}</span>
                                    <span className="text-xs text-right">{article.comment_count}</span>
                                    <span className="text-muted-foreground">{t('admin.created')}</span>
                                    <span className="text-xs text-right whitespace-nowrap">{formatDateTime(article.create_time)}</span>
                                    <span className="text-muted-foreground">{t('admin.updated')}</span>
                                    <span className="text-xs text-right whitespace-nowrap">{formatDateTime(article.update_time)}</span>
                                </div>
                            </div>
                        )}

                        {/* Quick Actions */}
                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <h3 className="font-medium">{t('admin.quickActions')}</h3>
                            <div className="space-y-2">
                                {form.state !== 'published' && (
                                    <Button variant="outline" size="sm" className="w-full justify-start"
                                            onClick={() => setForm({...form, state: 'published'})}>
                                        <CheckCircle className="w-4 h-4 mr-2"/>{t('admin.publish')}
                                    </Button>
                                )}
                                {form.state === 'published' && (
                                    <Button variant="outline" size="sm" className="w-full justify-start"
                                            onClick={() => setForm({...form, state: 'draft'})}>
                                        <Clock className="w-4 h-4 mr-2"/>{t('admin.revertToDraft')}
                                    </Button>
                                )}
                                <Button variant="outline" size="sm" className="w-full justify-start"
                                        onClick={() => setForm({...form, featured: !form.featured})}>
                                    {form.featured ? <X className="w-4 h-4 mr-2"/> : <CheckCircle className="w-4 h-4 mr-2"/>}
                                    {form.featured ? t('admin.removeFeatured') : t('admin.setFeatured')}
                                </Button>
                                {!form.media_id && (
                                    <Button variant="outline" size="sm" className="w-full justify-start"
                                            onClick={() => setMediaSelectorOpen(true)}>
                                        <Film className="w-4 h-4 mr-2"/>{t('admin.associateVideo')}
                                    </Button>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Media Selector Dialog */}
            <MediaSelectorDialog
                open={mediaSelectorOpen}
                onClose={() => setMediaSelectorOpen(false)}
                onSelect={handleMediaSelect}
            />

            {/* Delete Confirmation Dialog */}
            {mode === 'edit' && (
                <DeleteConfirmDialog
                    open={deleteDialogOpen}
                    onOpenChange={setDeleteDialogOpen}
                    title={article?.title || t('admin.untitledArticle')}
                    isDeleting={isDeleting}
                    onConfirm={handleDelete}
                />
            )}
        </div>
    );
}
