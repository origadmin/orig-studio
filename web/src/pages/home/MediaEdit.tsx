import {useState, useEffect, useMemo, useCallback} from 'react';
import {useParams, useNavigate} from '@tanstack/react-router';
import {usePublicMediaDetail, useUpdatePublicMedia, useDeleteMedia, useCategoryList} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {EditPageHeader, type HeaderBadgeConfig, type EncodingStatusConfig} from '@/components/common/EditPageHeader';
import {DeleteConfirmDialog} from '@/components/common/DeleteConfirmDialog';
import {MediaEditForm, type MediaEditFormState} from '@/components/common/MediaEditForm';
import {useDirtyState, useSaveState, useKeyboardShortcut} from '@/hooks/useEditPage';
import {Spinner} from '@/components/ui/spinner';
import {Button} from '@/components/ui/button';
import {AlertTriangle, ArrowLeft, Play} from 'lucide-react';
import {toast} from 'sonner';
import {API_BASE_URL} from '@/lib/request';

/**
 * Normalize privacy value from backend to a numeric enum value.
 * Backend (protojson) may return either:
 *   - A string enum name like "PRIVACY_PUBLIC", "PRIVACY_PRIVATE", "PRIVACY_UNLISTED"
 *   - A numeric value like 1, 2, 3
 * Frontend Select uses numeric string values ("1", "2", "3") for consistency.
 */
const PRIVACY_NAME_TO_VALUE: Record<string, number> = {
    PRIVACY_UNSPECIFIED: 0,
    PRIVACY_PUBLIC: 1,
    PRIVACY_PRIVATE: 2,
    PRIVACY_UNLISTED: 3,
    PRIVACY_PAID: 4,
    PRIVACY_SUBSCRIBERS_ONLY: 5,
};

function normalizePrivacy(value: unknown): number {
    if (typeof value === 'number') return value;
    if (typeof value === 'string') {
        const num = Number(value);
        if (!isNaN(num) && num >= 0) return num;
        const mapped = PRIVACY_NAME_TO_VALUE[value];
        if (mapped !== undefined) return mapped;
    }
    return 1;
}

const STATE_BADGE_MAP: Record<string, { variant: HeaderBadgeConfig['variant']; label: string }> = {
    active: {variant: 'default', label: 'Published'},
    draft: {variant: 'secondary', label: 'Draft'},
    deleted: {variant: 'destructive', label: 'Deleted'},
};

function mapMediaToHeaderBadges(media: any, isAdmin: boolean): HeaderBadgeConfig[] {
    const badges: HeaderBadgeConfig[] = [];

    badges.push({
        type: 'media-type',
        variant: 'outline',
        label: media.type,
        ariaLabel: `Media type: ${media.type}`,
    });

    const stateConfig = STATE_BADGE_MAP[media.state] || {variant: 'outline' as const, label: media.state};
    badges.push({
        type: 'state',
        variant: stateConfig.variant,
        label: stateConfig.label,
        ariaLabel: `Status: ${stateConfig.label}`,
    });

    if (isAdmin && media.featured) {
        badges.push({
            type: 'featured',
            variant: 'outline',
            label: 'Featured',
            ariaLabel: 'Featured content',
            className: 'text-warning border-amber-300',
        });
    }

    return badges;
}

function mapEncodingStatus(status: string | undefined): EncodingStatusConfig | undefined {
    const validStatuses = ['success', 'processing', 'pending', 'failed'];
    if (!status || !validStatuses.includes(status)) return undefined;
    return {status: status as EncodingStatusConfig['status']};
}

function resolveMediaUrl(url: string | undefined): string | undefined {
    if (!url) return undefined;
    if (/^(https?:|data:|blob:)/i.test(url)) return url;
    const base = API_BASE_URL || '';
    return `${base}/${url.replace(/^\//, '')}`;
}

export default function MediaEditPage() {
    const {shortToken} = useParams({strict: false}) as { shortToken: string };
    const navigate = useNavigate();
    const {user, isAdmin} = useAuth();
    const {data: media, isLoading, error} = usePublicMediaDetail(shortToken);
    const updateMutation = useUpdatePublicMedia();
    const deleteMutation = useDeleteMedia();
    const {data: categoriesData} = useCategoryList();

    const {form, setForm, isDirty, resetDirty} = useDirtyState<MediaEditFormState>({
        title: '',
        description: '',
        category_id: '' as string | number,
        tags: '',
        privacy: 1,
        state: 'draft',
        enable_comments: true,
        allow_download: false,
        featured: false,
        listable: false,
    });

    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const {saveState, isSaving, setSaving, setSuccess, setError} = useSaveState();

    useEffect(() => {
        if (media) {
            setForm({
                title: media.title || '',
                description: media.description || '',
                category_id: media.category_id ?? '',
                tags: media.tags?.join(', ') || '',
                privacy: normalizePrivacy(media.privacy),
                state: media.state || 'draft',
                enable_comments: media.enable_comments ?? true,
                allow_download: media.allow_download ?? false,
                featured: media.featured || false,
                listable: media.listable ?? false,
            });
        }
    }, [media, setForm]);

    const handleSave = useCallback(async () => {
        if (!shortToken || isSaving) return;
        setSaving();
        try {
            await updateMutation.mutateAsync({
                shortToken,
                data: {
                    title: form.title,
                    description: form.description,
                    category_id: form.category_id !== '' && form.category_id !== undefined ? Number(form.category_id) : undefined,
                    tags: form.tags.split(',').map(s => s.trim()).filter(Boolean),
                    privacy: form.privacy,
                    state: isAdmin ? form.state : undefined,
                    enable_comments: form.enable_comments,
                    allow_download: form.allow_download,
                    featured: isAdmin ? form.featured : undefined,
                    listable: isAdmin ? form.listable : undefined,
                },
            });
            resetDirty();
            setSuccess();
            toast.success('Saved successfully');
        } catch (err: any) {
            setError();
            toast.error(`Save failed: ${err?.message || 'Unknown error'}`);
        }
    }, [shortToken, isSaving, form, isAdmin, updateMutation, setSaving, setSuccess, setError, resetDirty]);

    const handleDelete = useCallback(async () => {
        if (!media?.id) return;
        setIsDeleting(true);
        try {
            await deleteMutation.mutateAsync(media.id);
            setDeleteDialogOpen(false);
            toast.success('Media deleted');
            navigate({to: '/'});
        } catch (err: any) {
            setIsDeleting(false);
            toast.error(`Delete failed: ${err?.message || 'Unknown error'}`);
        }
    }, [media?.id, deleteMutation, navigate]);

    const handlePreview = useCallback(() => {
        if (shortToken) {
            window.open(`/watch?v=${shortToken}`, '_blank', 'noopener,noreferrer');
        }
    }, [shortToken]);

    const handleBack = useCallback(() => {
        navigate({to: '/watch', search: {v: shortToken}});
    }, [navigate, shortToken]);

    useKeyboardShortcut('ctrl+s', handleSave, {enabled: !isSaving});

    const headerBadges = useMemo(() => media ? mapMediaToHeaderBadges(media, isAdmin) : [], [media, isAdmin]);
    const encodingConfig = useMemo(() => media ? mapEncodingStatus(media.encoding_status) : undefined, [media]);

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Spinner/>
            </div>
        );
    }

    if (error || !media) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
                <AlertTriangle className="w-12 h-12 text-destructive"/>
                <p className="text-lg text-muted-foreground">Unable to load media info</p>
                <Button variant="outline" onClick={() => navigate({to: '/'})}>
                    <ArrowLeft className="w-4 h-4 mr-2"/>Back to Home
                </Button>
            </div>
        );
    }

    const isOwner = user && String(user.id) === String(media.user_id);
    if (!isOwner && !isAdmin) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
                <AlertTriangle className="w-12 h-12 text-destructive"/>
                <p className="text-lg text-muted-foreground">You do not have permission to edit this media</p>
                <Button variant="outline" onClick={() => navigate({to: '/watch', search: {v: shortToken}})}>
                    <ArrowLeft className="w-4 h-4 mr-2"/>Back to Video
                </Button>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            <EditPageHeader
                title={media.title || 'Untitled Media'}
                isDirty={isDirty}
                isSaving={isSaving}
                saveState={saveState}
                onBack={handleBack}
                onSave={handleSave}
                onPreview={handlePreview}
                onDelete={() => setDeleteDialogOpen(true)}
                badges={headerBadges}
                encodingStatus={encodingConfig}
            />

            <div className="max-w-7xl mx-auto px-4 sm:px-6 py-6">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <div className="lg:col-span-2">
                        <div className="bg-card rounded-lg border p-6">
                            <MediaEditForm
                                form={form}
                                setForm={setForm}
                                media={media}
                                categories={categoriesData}
                                isAdmin={isAdmin}
                                showAdminOnlyFields={false}
                            />
                        </div>
                    </div>

                    <div className="space-y-6">
                        <div className="bg-card rounded-lg border p-4">
                            <h3 className="font-medium mb-3">Preview</h3>
                            {media.thumbnail ? (
                                <img
                                    src={resolveMediaUrl(media.thumbnail)}
                                    alt={media.title}
                                    className="w-full aspect-video object-cover rounded-md"
                                    onError={(e) => {
                                        (e.target as HTMLImageElement).style.display = 'none';
                                    }}
                                />
                            ) : (
                                <div className="w-full aspect-video bg-muted rounded-md flex items-center justify-center">
                                    <Play className="w-8 h-8 text-muted-foreground"/>
                                </div>
                            )}
                        </div>

                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <h3 className="font-medium">Info</h3>
                            <div className="space-y-2 text-sm">
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Duration</span>
                                    <span className="text-xs">
                                        {media.duration ? `${Math.floor(media.duration / 60)}:${String(Math.floor(media.duration % 60)).padStart(2, '0')}` : 'N/A'}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Resolution</span>
                                    <span className="text-xs">
                                        {media.width && media.height ? `${media.width}x${media.height}` : 'N/A'}
                                    </span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-muted-foreground">Encoding</span>
                                    <span className="text-xs">{media.encoding_status}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <DeleteConfirmDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                title={media.title || 'Untitled Media'}
                isDeleting={isDeleting}
                onConfirm={handleDelete}
            />
        </div>
    );
}
