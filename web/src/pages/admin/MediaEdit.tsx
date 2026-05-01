import {Spinner} from "@/components/ui/spinner"
import {useState, useEffect, useMemo, useCallback} from 'react';
import {useParams, useNavigate} from '@tanstack/react-router';
import {useAdminMediaDetail, useUpdateMedia, useDeleteMedia, useCategoryList} from '@/hooks/queries';
import {adminMediaApi, encodingApi, type EncodeProfile} from '@/lib/api/media';
import {api, API_BASE_URL} from '@/lib/request';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Label} from '@/components/ui/label';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Badge} from '@/components/ui/badge';
import {Separator} from '@/components/ui/separator';
import {EditPageHeader, type HeaderBadgeConfig, type EncodingStatusConfig} from '@/components/common/EditPageHeader';
import {DeleteConfirmDialog} from '@/components/common/DeleteConfirmDialog';
import {useDirtyState, useSaveState, useKeyboardShortcut} from '@/hooks/useEditPage';
import {ArrowLeft, RefreshCw, Play, Eye, ThumbsUp, MessageSquare, Download, AlertTriangle, CheckCircle, Clock, XCircle, Image, Film} from 'lucide-react';
import {formatDateTime} from '@/lib/format';
import {toast} from 'sonner';
import type {Media} from '@/lib/api/media';

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
        // Try parsing as numeric string first (e.g., "1", "2")
        const num = Number(value);
        if (!isNaN(num) && num >= 0) return num;
        // Then try enum name (e.g., "PRIVACY_PUBLIC")
        const mapped = PRIVACY_NAME_TO_VALUE[value];
        if (mapped !== undefined) return mapped;
    }
    return 1; // Default to PUBLIC
}

interface EncodingTask {
    id: string;
    media_id: string;
    profile_id: number;
    status: string;
    output_path: string;
    error_message: string;
    chunk: boolean;
    create_time: string;
    update_time: string;
}

interface MediaStats {
    view_count: number;
    like_count: number;
    dislike_count: number;
    comment_count: number;
    favorite_count: number;
    encoding_status: string;
}

/**
 * Map Media state to Badge variant and label
 */
const STATE_BADGE_MAP: Record<string, { variant: HeaderBadgeConfig['variant']; label: string }> = {
    active: {variant: 'default', label: '已发布'},
    draft: {variant: 'secondary', label: '草稿'},
    deleted: {variant: 'destructive', label: '已删除'},
};

/**
 * Map Media to HeaderBadgeConfig[]
 */
function mapMediaToHeaderBadges(media: Media): HeaderBadgeConfig[] {
    const badges: HeaderBadgeConfig[] = [];

    // Type Badge
    badges.push({
        type: 'media-type',
        variant: 'outline',
        label: media.type,
        ariaLabel: `媒体类型: ${media.type}`,
    });

    // State Badge
    const stateConfig = STATE_BADGE_MAP[media.state] || {variant: 'outline' as const, label: media.state};
    badges.push({
        type: 'state',
        variant: stateConfig.variant,
        label: stateConfig.label,
        ariaLabel: `状态: ${stateConfig.label}`,
    });

    // Featured Badge (conditional)
    if (media.featured) {
        badges.push({
            type: 'featured',
            variant: 'outline',
            label: '推荐',
            ariaLabel: '推荐内容',
            className: 'text-warning border-amber-300',
        });
    }

    return badges;
}

/**
 * Map encoding_status string to EncodingStatusConfig
 */
function mapEncodingStatus(status: string | undefined): EncodingStatusConfig | undefined {
    const validStatuses = ['success', 'processing', 'pending', 'failed'];
    if (!status || !validStatuses.includes(status)) return undefined;
    return {status: status as EncodingStatusConfig['status']};
}

/**
 * Resolve a potentially relative URL to a full URL.
 * Backend returns relative paths like "uploads/xxx.jpg" for thumbnails/posters.
 * We need to prepend the API base URL to make them loadable.
 */
function resolveMediaUrl(url: string | undefined): string | undefined {
    if (!url) return undefined;
    // Already absolute URL (http://, https://, data:, blob:)
    if (/^(https?:|data:|blob:)/i.test(url)) return url;
    // Relative path: prepend API base URL
    const base = API_BASE_URL || '';
    return `${base}/${url.replace(/^\//, '')}`;
}

export default function MediaEditPage() {
    const {id} = useParams({strict: false}) as {id: string};
    const navigate = useNavigate();
    const {data: media, isLoading, error} = useAdminMediaDetail(id);
    const updateMutation = useUpdateMedia();
    const deleteMutation = useDeleteMedia();
    const {data: categoriesData} = useCategoryList();

    // Form state with dirty tracking
    const {form, setForm, isDirty, resetDirty, syncFromData} = useDirtyState({
        title: '',
        description: '',
        state: 'draft',
        category_id: '' as string | number,
        tags: '',
        privacy: 1,
        featured: false,
        enable_comments: true,
        allow_download: false,
        listable: false,
    });

    const [stats, setStats] = useState<MediaStats | null>(null);
    const [tasks, setTasks] = useState<EncodingTask[]>([]);
    const [profiles, setProfiles] = useState<Map<number, EncodeProfile>>(new Map());
    const [activeTab, setActiveTab] = useState<'metadata' | 'publish' | 'encoding' | 'stats'>('metadata');
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [thumbnailError, setThumbnailError] = useState(false);
    const [regenThumbnailConfirmOpen, setRegenThumbnailConfirmOpen] = useState(false);
    const [regenSpriteConfirmOpen, setRegenSpriteConfirmOpen] = useState(false);
    const [isRegenerating, setIsRegenerating] = useState(false);

    // Save state management
    const {saveState, isSaving, setSaving, setSuccess, setError} = useSaveState();

    // Initialize form from media data
    useEffect(() => {
        if (media) {
            syncFromData({
                title: media.title || '',
                description: media.description || '',
                state: media.state || 'draft',
                category_id: media.category_id ?? '',
                tags: media.tags?.join(', ') || '',
                privacy: normalizePrivacy(media.privacy),
                featured: media.featured || false,
                enable_comments: media.enable_comments ?? true,
                allow_download: media.allow_download ?? false,
                listable: media.listable ?? false,
            });
        }
    }, [media, syncFromData]);

    // Fetch stats, tasks, and profiles
    useEffect(() => {
        if (id) {
            adminMediaApi.getStats(id).then(setStats).catch(() => {});
            adminMediaApi.getTasks(id).then((res: any) => setTasks(res?.tasks || res?.items || [])).catch(() => {});
        }
    }, [id]);

    // Fetch encode profiles for profile name resolution
    useEffect(() => {
        encodingApi.profiles.list().then((res: any) => {
            const profileList: EncodeProfile[] = res?.profiles || res || [];
            const map = new Map<number, EncodeProfile>();
            profileList.forEach(p => map.set(p.id, p));
            setProfiles(map);
        }).catch(() => {});
    }, []);

    // Save handler
    const handleSave = useCallback(async () => {
        if (!id || isSaving) return;
        setSaving();
        try {
            await updateMutation.mutateAsync({
                id,
                data: {
                    title: form.title,
                    description: form.description,
                    state: form.state,
                    category_id: form.category_id !== '' && form.category_id !== undefined ? Number(form.category_id) : undefined,
                    tags: form.tags.split(',').map(s => s.trim()).filter(Boolean),
                    privacy: form.privacy,
                    featured: form.featured,
                    enable_comments: form.enable_comments,
                    allow_download: form.allow_download,
                    listable: form.listable,
                } as any,
            });
            resetDirty();
            setSuccess();
            toast.success('保存成功');
        } catch (err: any) {
            setError();
            toast.error(`保存失败: ${err?.message || '未知错误'}`);
            console.error('Failed to save', err);
        }
    }, [id, isSaving, form, updateMutation, setSaving, setSuccess, setError, resetDirty]);

    // Delete handler
    const handleDelete = useCallback(async () => {
        if (!id) return;
        setIsDeleting(true);
        try {
            await deleteMutation.mutateAsync(id);
            setDeleteDialogOpen(false);
            toast.success('媒体已删除');
            navigate({to: '/admin/media'});
        } catch (err: any) {
            setIsDeleting(false);
            toast.error(`删除失败: ${err?.message || '未知错误'}`);
            console.error('Failed to delete', err);
        }
    }, [id, deleteMutation, navigate]);

    // Preview handler
    const handlePreview = useCallback(() => {
        if (media?.short_token) {
            window.open(`/watch?v=${media.short_token}`, '_blank', 'noopener,noreferrer');
        }
    }, [media?.short_token]);

    // Back handler
    const handleBack = useCallback(() => {
        navigate({to: '/admin/media'});
    }, [navigate]);

    // Keyboard shortcut: Ctrl+S / Cmd+S
    useKeyboardShortcut('ctrl+s', handleSave, {enabled: !isSaving});

    // Retry encoding task
    const handleRetryTask = async (taskId: string) => {
        if (!id) return;
        try {
            await adminMediaApi.retryTask(id, taskId);
            const res = await adminMediaApi.getTasks(id);
            setTasks((res as any)?.tasks || (res as any)?.items || []);
        } catch (err) {
            console.error('Failed to retry task', err);
        }
    };

    const handleRegenerateThumbnail = async () => {
        if (!id) return;
        setIsRegenerating(true);
        try {
            await api.post(`/admin/medias/${id}/regenerate-thumbnail`, {});
            toast.success('缩略图重新生成已调度，请稍后刷新页面查看结果');
            // Refresh tasks to reflect any status changes
            const res = await adminMediaApi.getTasks(id);
            setTasks((res as any)?.tasks || (res as any)?.items || []);
        } catch (err: any) {
            const errMsg = err?.response?.data?.error || err?.response?.data?.message || err?.message || '未知错误';
            toast.error(`重新生成缩略图失败: ${errMsg}`);
            console.error('Failed to regenerate thumbnail', err);
        } finally {
            setIsRegenerating(false);
            setRegenThumbnailConfirmOpen(false);
        }
    };

    const handleRegenerateSprite = async () => {
        if (!id) return;
        setIsRegenerating(true);
        try {
            await api.post(`/admin/medias/${id}/regenerate-sprite`, {});
            toast.success('雪碧图重新生成已调度，后台处理中请稍候');
            // Refresh tasks to reflect any status changes
            const res = await adminMediaApi.getTasks(id);
            setTasks((res as any)?.tasks || (res as any)?.items || []);
        } catch (err: any) {
            const errMsg = err?.response?.data?.error || err?.response?.data?.message || err?.message || '未知错误';
            if (errMsg.includes('already processing') || errMsg.includes('already in progress')) {
                toast.warning('雪碧图正在生成中，请稍后再试');
            } else {
                toast.error(`重新生成雪碧图失败: ${errMsg}`);
            }
            console.error('Failed to regenerate sprite', err);
        } finally {
            setIsRegenerating(false);
            setRegenSpriteConfirmOpen(false);
        }
    };

    // Compute header badges and encoding status from media
    const headerBadges = useMemo(() => media ? mapMediaToHeaderBadges(media) : [], [media]);
    const encodingConfig = useMemo(() => media ? mapEncodingStatus(media.encoding_status) : undefined, [media]);

    const encodingStatusBadge = (status: string | undefined): "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info" => {
        switch (status) {
            case 'success': return 'success';
            case 'processing': return 'info';
            case 'pending': return 'warning';
            case 'failed': return 'destructive';
            default: return 'secondary';
        }
    };

    const encodingStatusLabel = (status: string | undefined) => {
        switch (status) {
            case 'success': return '完成';
            case 'processing': return '转码中';
            case 'pending': return '排队中';
            case 'failed': return '失败';
            case 'partial': return '部分完成';
            default: return status || '--';
        }
    };

    // Resolve profile_id to a human-readable profile name
    const getProfileName = (profileId: number): string => {
        const profile = profiles.get(profileId);
        if (profile) {
            return profile.name || `Profile #${profileId}`;
        }
        return `Profile #${profileId}`;
    };

    // Get profile resolution info for display
    const getProfileInfo = (profileId: number): string => {
        const profile = profiles.get(profileId);
        if (profile) {
            const parts: string[] = [];
            if (profile.resolution) parts.push(profile.resolution);
            if (profile.extension) parts.push(profile.extension.toUpperCase());
            return parts.length > 0 ? parts.join(' / ') : '';
        }
        return '';
    };

    // Compute task summary counts
    const taskSummary = useMemo(() => {
        const counts = {success: 0, processing: 0, pending: 0, failed: 0, partial: 0, total: tasks.length};
        tasks.forEach(t => {
            if (t.status in counts) {
                counts[t.status as keyof typeof counts]++;
            }
        });
        return counts;
    }, [tasks]);

    // Format task summary text
    const taskSummaryText = useMemo(() => {
        if (tasks.length === 0) return '';
        const parts: string[] = [];
        if (taskSummary.success > 0) parts.push(`${taskSummary.success} 完成`);
        if (taskSummary.processing > 0) parts.push(`${taskSummary.processing} 转码中`);
        if (taskSummary.pending > 0) parts.push(`${taskSummary.pending} 排队中`);
        if (taskSummary.failed > 0) parts.push(`${taskSummary.failed} 失败`);
        if (taskSummary.partial > 0) parts.push(`${taskSummary.partial} 部分完成`);
        return parts.join('，');
    }, [taskSummary, tasks.length]);

    const reviewStatusBadge = (status: string | undefined): "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info" => {
        switch (status) {
            case 'approved': return 'success';
            case 'pending': return 'warning';
            case 'rejected': return 'destructive';
            default: return 'secondary';
        }
    };

    const reviewStatusLabel = (status: string | undefined) => {
        switch (status) {
            case 'approved': return '已通过';
            case 'pending': return '待审核';
            case 'rejected': return '已拒绝';
            default: return '未审核';
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Spinner />
            </div>
        );
    }

    if (error || !media) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
                <AlertTriangle className="w-12 h-12 text-destructive"/>
                <p className="text-lg text-muted-foreground">无法加载媒体信息</p>
                <Button variant="outline" onClick={() => navigate({to: '/admin/media'})}>
                    <ArrowLeft className="w-4 h-4 mr-2"/>返回列表
                </Button>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            <EditPageHeader
                title={media.title || '未命名媒体'}
                isDirty={isDirty}
                isSaving={isSaving}
                saveState={saveState}
                onBack={handleBack}
                onSave={handleSave}
                onPreview={media.short_token ? handlePreview : undefined}
                onDelete={() => setDeleteDialogOpen(true)}
                badges={headerBadges}
                encodingStatus={encodingConfig}
            />

            <div className="max-w-7xl mx-auto px-6 py-6">
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <div className="lg:col-span-2 space-y-6">
                        <div className="flex gap-1 border-b">
                            {(['metadata', 'publish', 'encoding', 'stats'] as const).map(tab => (
                                <button key={tab}
                                        className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
                                            activeTab === tab
                                                ? 'border-primary text-primary'
                                                : 'border-transparent text-muted-foreground hover:text-foreground'
                                        }`}
                                        onClick={() => setActiveTab(tab)}>
                                    {{metadata: '元数据', publish: '发布设置', encoding: '编码任务', stats: '统计信息'}[tab]}
                                </button>
                            ))}
                        </div>

                        {activeTab === 'metadata' && (
                            <div className="space-y-6 bg-card rounded-lg border p-6">
                                <div className="space-y-2">
                                    <Label htmlFor="title">标题</Label>
                                    <Input id="title" value={form.title}
                                           onChange={e => setForm({...form, title: e.target.value})}/>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="description">描述</Label>
                                    <textarea id="description"
                                              className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                                              value={form.description}
                                              onChange={e => setForm({...form, description: e.target.value})}
                                              placeholder="输入媒体描述..."/>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="tags">标签 (逗号分隔)</Label>
                                    <Input id="tags" value={form.tags}
                                           onChange={e => setForm({...form, tags: e.target.value})}
                                           placeholder="如：编程, 教程, 运维"/>
                                    {form.tags && (
                                        <div className="flex flex-wrap gap-1 mt-2">
                                            {form.tags.split(',').map((tag, i) => tag.trim() && (
                                                <Badge key={i} variant="secondary" className="text-xs">{tag.trim()}</Badge>
                                            ))}
                                        </div>
                                    )}
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label>媒体类型</Label>
                                        <Input value={media.type || ''} disabled className="bg-muted"/>
                                    </div>
                                    <div className="space-y-2">
                                        <Label>时长</Label>
                                        <Input value={media.duration ? `${Math.floor(media.duration / 60)}:${String(Math.floor(media.duration % 60)).padStart(2, '0')}` : 'N/A'} disabled className="bg-muted"/>
                                    </div>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label>分辨率</Label>
                                        <Input value={media.width && media.height ? `${media.width}x${media.height}` : 'N/A'} disabled className="bg-muted"/>
                                    </div>
                                    <div className="space-y-2">
                                        <Label>文件大小</Label>
                                        <Input value={media.size || 'N/A'} disabled className="bg-muted"/>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <Label>文件URL</Label>
                                    <Input value={media.url || ''} disabled className="bg-muted text-xs"/>
                                </div>
                                <div className="space-y-2">
                                    <Label>MD5</Label>
                                    <Input value={media.md5sum || 'N/A'} disabled className="bg-muted font-mono text-xs"/>
                                </div>
                            </div>
                        )}

                        {activeTab === 'publish' && (
                            <div className="space-y-6 bg-card rounded-lg border p-6">
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label>状态</Label>
                                        <Select value={form.state} onValueChange={val => setForm({...form, state: val})}>
                                            <SelectTrigger><SelectValue/></SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="draft">草稿</SelectItem>
                                                <SelectItem value="active">已发布</SelectItem>
                                                <SelectItem value="deleted">已删除</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                    <div className="space-y-2">
                                        <Label>隐私级别</Label>
                                        <Select value={String(form.privacy)} onValueChange={val => setForm({...form, privacy: Number(val)})}>
                                            <SelectTrigger><SelectValue/></SelectTrigger>
                                            <SelectContent>
                                                <SelectItem value="1">公开</SelectItem>
                                                <SelectItem value="3">未列出</SelectItem>
                                                <SelectItem value="2">私密</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <Label>分类</Label>
                                    <Select value={form.category_id !== '' && form.category_id !== undefined ? String(form.category_id) : '_none_'} onValueChange={val => setForm({...form, category_id: val === '_none_' ? '' : val})}>
                                        <SelectTrigger><SelectValue placeholder="选择分类"/></SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="_none_">无分类</SelectItem>
                                            {(Array.isArray(categoriesData?.items) ? categoriesData.items : Array.isArray(categoriesData) ? categoriesData : []).map((cat: any) => (
                                                <SelectItem key={cat.id} value={String(cat.id)}>
                                                    {cat.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <Separator/>
                                <div className="grid grid-cols-2 gap-6">
                                    <div className="flex items-center gap-3">
                                        <input type="checkbox" id="featured" checked={form.featured}
                                               onChange={e => setForm({...form, featured: e.target.checked})}
                                               className="h-4 w-4 rounded border-input text-primary focus:ring-primary"/>
                                        <div>
                                            <Label htmlFor="featured" className="cursor-pointer">推荐内容</Label>
                                            <p className="text-xs text-muted-foreground">在首页推荐区域展示</p>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <input type="checkbox" id="listable" checked={form.listable}
                                               onChange={e => setForm({...form, listable: e.target.checked})}
                                               className="h-4 w-4 rounded border-input text-primary focus:ring-primary"/>
                                        <div>
                                            <Label htmlFor="listable" className="cursor-pointer">允许列表展示</Label>
                                            <p className="text-xs text-muted-foreground">在视频列表中公开展示</p>
                                        </div>
                                    </div>
                                </div>
                                <div className="grid grid-cols-2 gap-6">
                                    <div className="flex items-center gap-3">
                                        <input type="checkbox" id="enable_comments" checked={form.enable_comments}
                                               onChange={e => setForm({...form, enable_comments: e.target.checked})}
                                               className="h-4 w-4 rounded border-input text-primary focus:ring-primary"/>
                                        <div>
                                            <Label htmlFor="enable_comments" className="cursor-pointer">允许评论</Label>
                                            <p className="text-xs text-muted-foreground">用户可以发表评论</p>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <input type="checkbox" id="allow_download" checked={form.allow_download}
                                               onChange={e => setForm({...form, allow_download: e.target.checked})}
                                               className="h-4 w-4 rounded border-input text-primary focus:ring-primary"/>
                                        <div>
                                            <Label htmlFor="allow_download" className="cursor-pointer">允许下载</Label>
                                            <p className="text-xs text-muted-foreground">用户可以下载原始文件</p>
                                        </div>
                                    </div>
                                </div>
                                {media.reported_times !== undefined && media.reported_times > 0 && (
                                    <>
                                        <Separator/>
                                        <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive">
                                            <AlertTriangle className="w-4 h-4"/>
                                            <span className="text-sm font-medium">该媒体已被举报 {media.reported_times} 次</span>
                                        </div>
                                    </>
                                )}
                            </div>
                        )}

                        {activeTab === 'encoding' && (
                            <div className="space-y-4 bg-card rounded-lg border p-6">
                                <div className="flex items-center justify-between">
                                    <div>
                                        <h3 className="font-medium">编码任务</h3>
                                        {taskSummaryText && (
                                            <p className="text-xs text-muted-foreground mt-1">
                                                共 {tasks.length} 个任务: {taskSummaryText}
                                            </p>
                                        )}
                                    </div>
                                    {media.type === 'video' && (
                                        <div className="flex gap-2">
                                            <Button variant="outline" size="sm"
                                                    onClick={() => setRegenThumbnailConfirmOpen(true)}
                                                    disabled={isRegenerating}
                                                    title="从视频中截取一帧作为缩略图，替换当前缩略图">
                                                <Image className="w-3 h-3 mr-1"/>重新生成缩略图
                                            </Button>
                                            <Button variant="outline" size="sm"
                                                    onClick={() => setRegenSpriteConfirmOpen(true)}
                                                    disabled={isRegenerating}
                                                    title="重新生成视频预览用的雪碧图（进度条缩略图），用于播放器进度条预览">
                                                <Film className="w-3 h-3 mr-1"/>重新生成雪碧图
                                            </Button>
                                        </div>
                                    )}
                                </div>
                                {tasks.length === 0 ? (
                                    <p className="text-sm text-muted-foreground py-8 text-center">暂无编码任务</p>
                                ) : (
                                    <div className="space-y-3">
                                        {tasks.map(task => (
                                            <div key={task.id} className="p-3 rounded-md border space-y-2">
                                                <div className="flex items-center justify-between">
                                                    <div className="flex items-center gap-3">
                                                        <Badge variant={encodingStatusBadge(task.status)} className="text-[10px] px-1.5 py-0 h-4 shrink-0">
                                                            {encodingStatusLabel(task.status)}
                                                        </Badge>
                                                        <div>
                                                            <p className="text-sm font-medium">{getProfileName(task.profile_id)}</p>
                                                            {getProfileInfo(task.profile_id) && (
                                                                <p className="text-xs text-muted-foreground">{getProfileInfo(task.profile_id)}</p>
                                                            )}
                                                        </div>
                                                        {task.chunk && (
                                                            <Badge variant="outline" className="text-[10px] px-1 py-0 h-4">分段</Badge>
                                                        )}
                                                    </div>
                                                    {task.status === 'failed' && (
                                                        <Button variant="outline" size="sm"
                                                                onClick={() => handleRetryTask(task.id)}>
                                                            <RefreshCw className="w-3 h-3 mr-1"/>重试
                                                        </Button>
                                                    )}
                                                </div>
                                                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                                                    {task.create_time && (
                                                        <span>创建: {formatDateTime(task.create_time)}</span>
                                                    )}
                                                    {task.update_time && task.update_time !== task.create_time && (
                                                        <span>更新: {formatDateTime(task.update_time)}</span>
                                                    )}
                                                </div>
                                                {task.status === 'failed' && task.error_message && (
                                                    <div className="flex items-start gap-2 p-2 rounded bg-destructive/10 text-destructive text-xs">
                                                        <AlertTriangle className="w-3 h-3 shrink-0 mt-0.5"/>
                                                        <span className="break-all">{task.error_message}</span>
                                                    </div>
                                                )}
                                                {task.status === 'success' && task.output_path && (
                                                    <div className="text-xs text-muted-foreground truncate" title={task.output_path}>
                                                        输出: {task.output_path}
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                )}

                                {/* Regenerate Thumbnail Confirmation */}
                                <DeleteConfirmDialog
                                    open={regenThumbnailConfirmOpen}
                                    onOpenChange={setRegenThumbnailConfirmOpen}
                                    title="重新生成缩略图"
                                    isDeleting={isRegenerating}
                                    onConfirm={handleRegenerateThumbnail}
                                    confirmLabel="确认重新生成"
                                    description="将从视频中截取一帧作为新的缩略图，替换当前的缩略图。此操作不可撤销，确认继续？"
                                />

                                {/* Regenerate Sprite Confirmation */}
                                <DeleteConfirmDialog
                                    open={regenSpriteConfirmOpen}
                                    onOpenChange={setRegenSpriteConfirmOpen}
                                    title="重新生成雪碧图"
                                    isDeleting={isRegenerating}
                                    onConfirm={handleRegenerateSprite}
                                    confirmLabel="确认重新生成"
                                    description="将重新生成视频预览用的雪碧图（播放器进度条缩略图）。此操作需要一定时间，确认继续？"
                                />
                            </div>
                        )}

                        {activeTab === 'stats' && stats && (
                            <div className="space-y-4 bg-card rounded-lg border p-6">
                                <h3 className="font-medium">统计信息</h3>
                                <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                                    <div className="p-4 rounded-md border">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <Eye className="w-4 h-4"/><span className="text-xs">播放量</span>
                                        </div>
                                        <p className="text-2xl font-bold mt-1">{stats.view_count.toLocaleString()}</p>
                                    </div>
                                    <div className="p-4 rounded-md border">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <ThumbsUp className="w-4 h-4"/><span className="text-xs">点赞</span>
                                        </div>
                                        <p className="text-2xl font-bold mt-1">{stats.like_count.toLocaleString()}</p>
                                    </div>
                                    <div className="p-4 rounded-md border">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <MessageSquare className="w-4 h-4"/><span className="text-xs">评论</span>
                                        </div>
                                        <p className="text-2xl font-bold mt-1">{stats.comment_count.toLocaleString()}</p>
                                    </div>
                                    <div className="p-4 rounded-md border">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <Download className="w-4 h-4"/><span className="text-xs">收藏</span>
                                        </div>
                                        <p className="text-2xl font-bold mt-1">{stats.favorite_count.toLocaleString()}</p>
                                    </div>
                                    <div className="p-4 rounded-md border">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <AlertTriangle className="w-4 h-4"/><span className="text-xs">踩</span>
                                        </div>
                                        <p className="text-2xl font-bold mt-1">{stats.dislike_count.toLocaleString()}</p>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>

                    <div className="space-y-6">
                        <div className="bg-card rounded-lg border p-4">
                            <h3 className="font-medium mb-3">预览</h3>
                            {media.thumbnail && !thumbnailError ? (
                                <img src={resolveMediaUrl(media.thumbnail)} alt={media.title}
                                     className="w-full aspect-video object-cover rounded-md"
                                     onError={() => setThumbnailError(true)}/>
                            ) : (
                                <div className="w-full aspect-video bg-muted rounded-md flex items-center justify-center">
                                    <Play className="w-8 h-8 text-muted-foreground"/>
                                </div>
                            )}
                        </div>

                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <h3 className="font-medium">基本信息</h3>
                            <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
                                <span className="text-muted-foreground">ID</span>
                                <span className="font-mono text-xs text-right break-all">{media.id}</span>
                                <span className="text-muted-foreground">Short Token</span>
                                <span className="font-mono text-xs text-right break-all">{media.short_token || 'N/A'}</span>
                                <span className="text-muted-foreground">创建时间</span>
                                <span className="text-xs text-right whitespace-nowrap">{formatDateTime(media.create_time)}</span>
                                <span className="text-muted-foreground">更新时间</span>
                                <span className="text-xs text-right whitespace-nowrap">{formatDateTime(media.update_time)}</span>
                                <span className="text-muted-foreground">编码状态</span>
                                <div className="flex justify-end">
                                    <Badge variant={encodingStatusBadge(media.encoding_status)} className="text-[10px] px-1.5 py-0 h-4">
                                        {encodingStatusLabel(media.encoding_status)}
                                    </Badge>
                                </div>
                                <span className="text-muted-foreground">审核状态</span>
                                <div className="flex justify-end">
                                    <Badge variant={reviewStatusBadge(media.review_status)} className="text-[10px] px-1.5 py-0 h-4">
                                        {reviewStatusLabel(media.review_status)}
                                    </Badge>
                                </div>
                            </div>
                        </div>

                        <div className="bg-card rounded-lg border p-4 space-y-3">
                            <h3 className="font-medium">快捷操作</h3>
                            <div className="space-y-2">
                                {media.state !== 'active' && (
                                    <Button variant="outline" size="sm" className="w-full justify-start"
                                            onClick={async () => {
                                                try {
                                                    await updateMutation.mutateAsync({id, data: {state: 'active'}});
                                                    toast.success('已发布');
                                                } catch (err: any) {
                                                    toast.error(`操作失败: ${err?.message || '未知错误'}`);
                                                }
                                            }}>
                                        <CheckCircle className="w-4 h-4 mr-2"/>发布
                                    </Button>
                                )}
                                {media.state === 'active' && (
                                    <Button variant="outline" size="sm" className="w-full justify-start"
                                            onClick={async () => {
                                                try {
                                                    await updateMutation.mutateAsync({id, data: {state: 'draft'}});
                                                    toast.success('已转为草稿');
                                                } catch (err: any) {
                                                    toast.error(`操作失败: ${err?.message || '未知错误'}`);
                                                }
                                            }}>
                                        <Clock className="w-4 h-4 mr-2"/>转为草稿
                                    </Button>
                                )}
                                <Button variant="outline" size="sm" className="w-full justify-start"
                                        onClick={async () => {
                                            try {
                                                await updateMutation.mutateAsync({id, data: {featured: !media.featured}});
                                                toast.success(media.featured ? '已取消推荐' : '已设为推荐');
                                            } catch (err: any) {
                                                toast.error(`操作失败: ${err?.message || '未知错误'}`);
                                            }
                                        }}>
                                    {media.featured ? <XCircle className="w-4 h-4 mr-2"/> : <CheckCircle className="w-4 h-4 mr-2"/>}
                                    {media.featured ? '取消推荐' : '设为推荐'}
                                </Button>
                                <Button variant="outline" size="sm" className="w-full justify-start"
                                        onClick={async () => {
                                            try {
                                                await updateMutation.mutateAsync({id, data: {listable: !media.listable}});
                                                toast.success(media.listable ? '已从列表隐藏' : '已上线展示');
                                            } catch (err: any) {
                                                toast.error(`操作失败: ${err?.message || '未知错误'}`);
                                            }
                                        }}>
                                    {media.listable ? <XCircle className="w-4 h-4 mr-2"/> : <Eye className="w-4 h-4 mr-2"/>}
                                    {media.listable ? '从列表隐藏' : '上线展示'}
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <DeleteConfirmDialog
                open={deleteDialogOpen}
                onOpenChange={setDeleteDialogOpen}
                title={media.title || '未命名媒体'}
                isDeleting={isDeleting}
                onConfirm={handleDelete}
            />
        </div>
    );
}
