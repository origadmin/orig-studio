/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 媒体管理页面
 */

import React, {useState, useEffect} from 'react';
import {useLocation, useNavigate, useRouterState, Link} from '@tanstack/react-router';
import {
    Play,
    Eye,
    MoreVertical,
    Trash2,
    Edit,
    Search,
    Upload,
    Image as ImageIcon,
    Video,
    ExternalLink,
    RotateCcw,
    Loader2,
    Filter
} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuLabel,
    DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import {Label} from "@/components/ui/label";
import {mediaApi, encodingApi, adminMediaApi, type Media, type MediaVariantSummary} from '@/lib/api/media';
import {API_BASE_URL} from '@/lib/request';
import {useAdminMediaList, useDeleteMedia} from '@/hooks/queries';
import {UploadComponent} from '@/components/upload/UploadComponent';
import {formatFileSize, formatDate} from '@/lib/format';
import {TablePagination} from '@/components/common/TablePagination';
import {PAGINATION} from '@/config/pagination';

export default function MediaPage() {
    const location = useLocation();
    const navigate = useNavigate();
    const urlSearch = new URLSearchParams(location.search).get("q");

    // 弹窗状态
    const [uploadDialogOpen, setUploadDialogOpen] = useState(false);

    // 编辑状态
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [deletingMedia, setDeletingMedia] = useState<Media | null>(null);

    // 转码详情弹窗
    const [variantDetailOpen, setVariantDetailOpen] = useState(false);
    const [variantData, setVariantData] = useState<MediaVariantSummary | null>(null);
    const [retryingAllId, setRetryingAllId] = useState<string | number | null>(null);

    const [searchParams, setSearchParams] = useState({keyword: urlSearch || '', state: '', page: 1, page_size: PAGINATION.DEFAULT_PAGE_SIZE});

    const [total, setTotal] = useState(0);

    // React Query Hooks
    const {data: mediaData, isLoading: loading, refetch: loadMedia} = useAdminMediaList(searchParams);
    const deleteMutation = useDeleteMedia();

    const mediaList = mediaData?.items || (Array.isArray(mediaData) ? mediaData : []) as Media[];

    React.useEffect(() => {
        if (mediaData?.total !== undefined) {
            setTotal(mediaData.total);
        }
    }, [mediaData?.total]);

    const formatDuration = (seconds: number) => {
        if (!seconds) return '-';
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;
        if (hours > 0) {
            return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
        }
        return `${minutes}:${String(secs).padStart(2, '0')}`;
    };

    const formatViews = (count: number | undefined | null) => {
        if (count === undefined || count === null || count === 0) return '0';
        if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
        if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
        return count.toString();
    };

    const handleEditClick = (media: Media) => {
        navigate({to: '/admin/media/$id', params: {id: String(media.id)}});
    };

    const handleDeleteClick = (media: Media) => {
        setDeletingMedia(media);
        setDeleteDialogOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (!deletingMedia) return;
        try {
            await deleteMutation.mutateAsync(String(deletingMedia.id));
            setDeleteDialogOpen(false);
            loadMedia();
        } catch (err) {
            console.error("Failed to delete media", err);
        }
    };

    // Show transcoding variant details for a media
    const handleShowVariants = async (media: Media) => {
        try {
            const data = await adminMediaApi.getVariants(media.id);
            setVariantData(data as unknown as MediaVariantSummary);
            setVariantDetailOpen(true);
        } catch (err: any) {
            console.error("Failed to fetch variants:", err.message);
        }
    };

    // Retry all failed tasks for a media (from media management page)
    const handleRetryAllFailed = async (mediaId: string | number) => {
        setRetryingAllId(mediaId);
        try {
            await encodingApi.retryAllFailed(String(mediaId));
            // Refresh variant detail if open, or just refresh the list
            if (variantData?.media_id === mediaId) {
                handleShowVariants({id: mediaId} as Media);
            }
        } catch (err: any) {
            console.error("Retry all failed:", err.message);
        } finally {
            setRetryingAllId(null);
        }
    };

    const encStatusBadge = (status?: string): "default" | "secondary" | "destructive" | "outline" | "success" | "warning" | "info" => {
        switch (status) {
            case "success":
                return "success";
            case "processing":
                return "info";
            case "pending":
                return "warning";
            case "partial":
                return "warning";
            case "failed":
                return "destructive";
            default:
                return "secondary";
        }
    };

    const encStatusLabel = (status?: string) => {
        switch (status) {
            case "success":
                return "完成";
            case "processing":
                return "转码中";
            case "pending":
                return "排队中";
            case "partial":
                return "部分完成";
            case "failed":
                return "失败";
            default:
                return status || "--";
        }
    };

    // Helper: resolve preview image URL
    const resolvePreview = (path?: string) => {
        if (!path) return "";
        if (path.startsWith("http")) return path;
        const base = API_BASE_URL;
        return `${base}${path.startsWith("/") ? "" : "/"}${path}`;
    };

    return (
            <div className="space-y-4 p-4 md:p-6">
                {/* ═══ Header ════════════════════════════════ */}
                <Card className="overflow-hidden">
                    <CardContent className="p-6">
                        <div className="flex flex-col gap-4">
                            {/* 页面标题 */}
                            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                                <div>
                                    <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">媒体管理</h2>
                                    <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">在这里集中管理所有的视频与图片资源</p>
                                </div>
                            </div>

                            {/* 分隔线 */}
                            <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                            {/* 搜索和筛选 */}
                            <div className="flex flex-col lg:flex-row gap-4">
                                <div className="flex-1 min-w-[120px] max-w-[400px]">
                                    <div className="relative w-full">
                                        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                        <Input
                                            placeholder="搜索媒体..."
                                            value={searchParams.keyword}
                                            onChange={(e) => setSearchParams({...searchParams, keyword: e.target.value})}
                                            className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                        />
                                    </div>
                                </div>
                                <div className="flex flex-wrap items-center gap-2">
                                    <Select value={searchParams.state || 'all'} onValueChange={(v) => setSearchParams({...searchParams, state: v === 'all' ? '' : v})}>
                                        <SelectTrigger className="w-[140px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                            <div className="flex items-center gap-2">
                                                <Filter className="h-4 w-4"/>
                                                {!searchParams.state ? (
                                                    <span className="text-muted-foreground">Status</span>
                                                ) : (
                                                    <SelectValue placeholder="Status"/>
                                                )}
                                            </div>
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                            <SelectItem value="active">已发布 (Active)</SelectItem>
                                            <SelectItem value="draft">草稿 (Draft)</SelectItem>
                                            <SelectItem value="deleted">已删除 (Deleted)</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => {
                                                const newParams = {keyword: '', state: '', page: 1, page_size: PAGINATION.DEFAULT_PAGE_SIZE};
                                                setSearchParams(newParams);
                                                loadMedia();
                                            }}
                                        >
                                            <RotateCcw className="h-4 w-4 mr-2"/>
                                            Reset
                                        </Button>
                                        <Button
                                            variant="default"
                                            size="sm"
                                            onClick={() => loadMedia()}
                                        >
                                            <Search className="h-4 w-4 mr-2"/>
                                            Search
                                        </Button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </CardContent>
                </Card>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">媒体总数</p>
                                <p className="text-2xl font-bold text-info dark:text-blue-400">{mediaList.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <Video className="w-6 h-6 text-info"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">视频</p>
                                <p className="text-2xl font-bold text-purple-600 dark:text-purple-400">{mediaList.filter(m => m.type === 'video').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <Play className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-purple-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">图片</p>
                                <p className="text-2xl font-bold text-success dark:text-green-400">{mediaList.filter(m => m.type === 'image' || !m.type).length}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <ImageIcon className="w-6 h-6 text-success"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">总播放量</p>
                                <p className="text-2xl font-bold text-orange-600 dark:text-orange-400">{formatViews(mediaList.reduce((acc, m) => acc + (m.view_count || 0), 0))}</p>
                            </div>
                            <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
                                <Eye className="w-6 h-6 text-orange-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-orange-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
            </div>

            {/* Media Table */}
            <Card>
                <CardHeader className="flex flex-row items-center justify-between gap-4">
                    <div>
                        <CardTitle>所有媒体</CardTitle>
                        <CardDescription>
                            {mediaList.length} 条媒体记录
                        </CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                        <Button size="sm" onClick={() => setUploadDialogOpen(true)}>
                            <Upload className="w-4 h-4 mr-2"/>
                            上传媒体
                        </Button>
                    </div>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="py-12 text-center text-muted-foreground">正在加载数据...</div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>名称</TableHead>
                                    <TableHead>类型</TableHead>
                                    <TableHead>大小</TableHead>
                                    <TableHead>时长</TableHead>
                                    <TableHead>播放量</TableHead>
                                    <TableHead>状态</TableHead>
                                    <TableHead>转码</TableHead>
                                    <TableHead>作者</TableHead>
                                    <TableHead>日期</TableHead>
                                    <TableHead className="text-right">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {mediaList.length > 0 ? mediaList.map((media) => (
                                    <TableRow key={media.id} id={`media-row-${media.id}`}>
                                        <TableCell>
                                            <div className="flex items-center gap-3">
                                                <div
                                                    className="w-16 h-10 bg-slate-100 rounded overflow-hidden shrink-0 flex items-center justify-center relative group">
                                                    {media.thumbnail ? (
                                                        <img
                                                            src={media.thumbnail.startsWith('http') ? media.thumbnail : `${API_BASE_URL}${media.thumbnail.startsWith('/') ? '' : '/'}${media.thumbnail}`}
                                                            alt="" className="w-full h-full object-cover transition-opacity duration-300"
                                                            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}/>
                                                    ) : (
                                                        media.type === 'video' ?
                                                            <Video className="w-4 h-4 text-muted-foreground"/> :
                                                            <ImageIcon className="w-4 h-4 text-muted-foreground"/>
                                                    )}
                                                    {(media.preview_file_path || media.preview_file) && (
                                                        <img
                                                            src={(media.preview_file_path || media.preview_file).startsWith('http') ? (media.preview_file_path || media.preview_file) : `${API_BASE_URL}${(media.preview_file_path || media.preview_file).startsWith('/') ? '' : '/'}${media.preview_file_path || media.preview_file}`}
                                                            alt="preview" 
                                                            className="w-full h-full object-cover absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300"
                                                            onError={(e) => { (e.target as HTMLImageElement).remove(); }}/>
                                                    )}
                                                </div>
                                                <span className="font-medium line-clamp-2"
                                                      title={media.title}>
                                                    {media.title || '未命名媒体'}
                                                </span>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="outline">
                                                {media.type === 'video' ? <Video className="w-3 h-3 mr-1"/> :
                                                    <ImageIcon className="w-3 h-3 mr-1"/>}
                                                {media.type || 'unknown'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-sm text-slate-500">
                                            {media.size ? formatFileSize(parseInt(media.size)) : '-'}
                                        </TableCell>
                                        <TableCell
                                            className="text-sm text-slate-500">{formatDuration(media.duration)}</TableCell>
                                        <TableCell
                                            className="text-sm text-slate-500">{formatViews(media.view_count)}</TableCell>
                                        <TableCell>
                                            <Badge variant={media.state === 'active' ? 'default' : 'secondary'}>
                                                {media.state || 'draft'}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            {media.encoding_status ? (
                                                <div className="flex items-center gap-1.5">
                                                    <Badge variant={encStatusBadge(media.encoding_status)}
                                                           className="text-[10px] px-1.5 py-0 h-4">
                                                        {encStatusLabel(media.encoding_status)}
                                                    </Badge>
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        className="h-6 w-6 p-0 text-[11px]"
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleShowVariants(media);
                                                        }}
                                                        title="查看转码详情"
                                                    >
                                                        <ExternalLink className="w-3 h-3"/>
                                                    </Button>
                                                </div>
                                            ) : (
                                                <span className="text-xs text-muted-foreground">--</span>
                                            )}
                                        </TableCell>
                                        <TableCell
                                            className="text-sm text-slate-500">{media.edges?.user?.[0]?.nickname || media.edges?.user?.[0]?.username || '-'}</TableCell>
                                        <TableCell
                                            className="text-sm text-slate-500">{formatDate(media.create_time || media.created_at)}</TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button 
                                                        variant="ghost" 
                                                        size="icon" 
                                                        className="h-6 w-6" 
                                                        title="More Actions"
                                                    >
                                                        <MoreVertical className="h-3 w-3"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem
                                                        onClick={() => window.open(`/watch?v=${media.short_token || media.id}`, '_blank')}>
                                                        <Eye className="w-4 h-4 mr-2"/>
                                                        查看
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem asChild>
                                                        <Link to="/admin/media/$id" params={{id: String(media.id)}}>
                                                            <Edit className="w-4 h-4 mr-2"/>
                                                            编辑
                                                        </Link>
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem 
                                                        className="text-destructive focus:text-destructive"
                                                                      onClick={() => handleDeleteClick(media)}>
                                                        <Trash2 className="w-4 h-4 mr-2"/>
                                                        删除
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow key="empty">
                                        <TableCell colSpan={10} className="h-24 text-center text-muted-foreground">
                                            没有找到匹配的媒体数据
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            <TablePagination
                page={searchParams.page}
                pageSize={searchParams.page_size}
                total={total}
                onPageChange={(p) => setSearchParams({...searchParams, page: p})}
            />

            {/* 上传模态框 */}
            <Dialog open={uploadDialogOpen} onOpenChange={setUploadDialogOpen}>
                <DialogContent className="max-w-4xl">
                    <DialogHeader>
                        <DialogTitle>上传媒体文件</DialogTitle>
                        <DialogDescription>
                            上传视频或音频文件到媒体库
                        </DialogDescription>
                    </DialogHeader>
                    <div className="py-4">
                        <UploadComponent
                            onSuccess={() => {
                                setUploadDialogOpen(false);
                                loadMedia();
                            }}
                            onCancel={() => setUploadDialogOpen(false)}
                        />
                    </div>
                </DialogContent>
            </Dialog>

            {/* 删除确认框 */}
            <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认删除？</AlertDialogTitle>
                        <AlertDialogDescription>
                            您将永久删除资源 "{deletingMedia?.title}"。此操作无法撤销，数据将从服务器移除。
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>手滑了，取消</AlertDialogCancel>
                        <AlertDialogAction onClick={handleConfirmDelete} className="bg-red-600 hover:bg-red-700">
                            是的，我要删除
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            {/* 转码详情弹窗 — 媒体管理页的聚合视图 */}
            <Dialog open={variantDetailOpen} onOpenChange={setVariantDetailOpen}>
                <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            转码概况
                            {variantData?.encoding_status && (
                                <Badge variant={encStatusBadge(variantData.encoding_status)} className="text-xs">
                                    {encStatusLabel(variantData.encoding_status)}
                                </Badge>
                            )}
                        </DialogTitle>
                        <DialogDescription>
                            查看媒体文件的转码状态和详细信息
                        </DialogDescription>
                    </DialogHeader>

                    {variantData && (
                        <div className="space-y-4 py-2">
                            {/* Summary stats — all encoding statuses */}
                            <div className="grid grid-cols-5 gap-2 text-center">
                                <div className="rounded-lg bg-yellow-50 dark:bg-yellow-950/20 p-3">
                                    <p className="text-lg font-bold text-yellow-600">{variantData.video_pending_count ?? 0}</p>
                                    <p className="text-[11px] text-muted-foreground">排队</p>
                                </div>
                                <div className="rounded-lg bg-blue-50 dark:bg-blue-950/20 p-3">
                                    <p className="text-lg font-bold text-info">{variantData.video_processing_count ?? 0}</p>
                                    <p className="text-[11px] text-muted-foreground">转码中</p>
                                </div>
                                <div className="rounded-lg bg-green-50 dark:bg-green-950/20 p-3">
                                    <p className="text-lg font-bold text-success">{variantData.video_success_count}</p>
                                    <p className="text-[11px] text-muted-foreground">成功</p>
                                </div>
                                <div className="rounded-lg bg-red-50 dark:bg-red-950/20 p-3">
                                    <p className="text-lg font-bold text-destructive">{variantData.video_failed_count}</p>
                                    <p className="text-[11px] text-muted-foreground">失败</p>
                                </div>
                                <div className="rounded-lg bg-slate-100 dark:bg-slate-800 p-3">
                                    <p className="text-lg font-bold text-slate-700">{variantData.video_total_count}</p>
                                    <p className="text-[11px] text-muted-foreground">总计</p>
                                </div>
                            </div>

                            {/* HLS / Preview paths */}
                            {(variantData.hls_file || variantData.preview_file) && (
                                <div className="text-xs space-y-1 bg-muted/50 rounded-md p-3">
                                    {variantData.hls_file && (
                                        <div className="flex items-center gap-1.5">
                                            <span className="font-medium text-muted-foreground">HLS:</span>
                                            <code
                                                className="text-green-700 dark:text-green-400">{variantData.hls_file}</code>
                                        </div>
                                    )}
                                    {variantData.preview_file && (
                                        <div className="flex items-center gap-1.5">
                                            <span className="font-medium text-muted-foreground">Preview:</span>
                                            <img
                                                src={resolvePreview(variantData.preview_file)}
                                                alt="preview"
                                                className="h-12 rounded border"
                                            />
                                        </div>
                                    )}
                                </div>
                            )}

                            {/* Variant list */}
                            {variantData.variants && variantData.variants.length > 0 && (
                                <div className="space-y-1.5">
                                    <p className="text-sm font-medium flex items-center gap-2">
                                        各清晰度任务
                                        {variantData.video_failed_count > 0 && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className="h-6 text-[10px] ml-auto"
                                                disabled={retryingAllId === variantData.media_id}
                                                onClick={() => handleRetryAllFailed(variantData.media_id)}
                                            >
                                                {retryingAllId === variantData.media_id ? (
                                                    <Loader2 className="w-3 h-3 animate-spin mr-1"/>
                                                ) : (
                                                    <RotateCcw className="w-3 h-3 mr-1"/>
                                                )}
                                                重试全部失败
                                            </Button>
                                        )}
                                    </p>
                                    {variantData.variants.map((v) => (
                                        <div
                                            key={v.task_id}
                                            className={`flex items-center justify-between rounded-md px-3 py-2 text-xs ${
                                                v.status === "failed" ? "bg-red-50 dark:bg-red-950/20" :
                                                    v.status === "success" ? "bg-green-50 dark:bg-green-950/20" :
                                                        "bg-muted/30"
                                            }`}
                                        >
                                            <div className="flex items-center gap-2 min-w-0">
                                                <span className="font-mono font-medium truncate">{v.profile_name}</span>
                                                {v.resolution && (
                                                    <span
                                                        className="text-muted-foreground hidden sm:inline">{v.resolution}</span>
                                                )}
                                            </div>
                                            <div className="flex items-center gap-2 shrink-0 ml-2">
                                                <Badge variant={encStatusBadge(v.status)} className="text-[10px] px-1.5 py-0 h-4">
                                                    {encStatusLabel(v.status)}
                                                </Badge>
                                                {v.output_path && v.status === "success" && (
                                                    <code
                                                        className="text-[10px] text-green-700 dark:text-green-400 max-w-[150px] truncate block">{v.output_path}</code>
                                                )}
                                                {v.error_message && (
                                                    <span className="text-destructive max-w-[200px] truncate block"
                                                          title={v.error_message}>{v.error_message}</span>
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}

                            {/* Link to TranscodingStatus page for full task view */}
                            <div className="pt-2 border-t">
                                <a
                                    href={`/admin/transcoding/status?media_id=${variantData.media_id}`}
                                    target="_blank"
                                    rel="noreferrer"
                                    className="inline-flex items-center gap-1.5 text-xs text-info hover:text-blue-800 hover:underline"
                                >
                                    在转码任务页面查看完整任务列表
                                    <ExternalLink className="w-3 h-3"/>
                                </a>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}