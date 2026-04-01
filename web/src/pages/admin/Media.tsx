/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 管理端 - 媒体管理页面
 */

import {useState, useEffect} from 'react';
import {Play, Eye, MoreVertical, Trash2, Edit, Search, Upload, Image as ImageIcon, Video} from 'lucide-react';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
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
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
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
import {mediaApi, type Media} from '@/lib/api/media';
import {UploadComponent} from '@/components/upload/UploadComponent';
import {formatFileSize} from '@/lib/format';

export default function MediaPage() {
    const [mediaList, setMediaList] = useState<Media[]>([]);
    const [loading, setLoading] = useState(false);

    // 弹窗状态
    const [uploadDialogOpen, setUploadDialogOpen] = useState(false);

    // 编辑状态
    const [editDialogOpen, setEditDialogOpen] = useState(false);
    const [editingMedia, setEditingMedia] = useState<Media | null>(null);
    const [editForm, setEditForm] = useState({
        title: '',
        description: '',
        status: 'draft',
        category: '',
        tags: '',
    });

    // 删除状态
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const [deletingMedia, setDeletingMedia] = useState<Media | null>(null);

    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');

    const loadMedia = async () => {
        setLoading(true);
        try {
            const res = await mediaApi.adminList({page: 1, page_size: 50});
            // 确保如果后端返回数据中有 list 则使用 list，否则直接使用数据数组或空数组
            setMediaList(res?.list || (Array.isArray(res) ? res : []));
        } catch (err) {
            console.error("Failed to load media", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadMedia();
    }, []);

    const filteredMedia = mediaList.filter(item => {
        const matchesSearch = item.title?.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || item.state === statusFilter;
        return matchesSearch && matchesStatus;
    });

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

    const formatViews = (count: number) => {
        if (!count) return '0';
        if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`;
        if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
        return count.toString();
    };

    const handleEditClick = (media: Media) => {
        setEditingMedia(media);
        setEditForm({
            title: media.title || '',
            description: media.description || '',
            status: media.state || 'draft',
            category: media.edges?.category?.name || '',
            tags: media.tags?.join(', ') || '',
        });
        setEditDialogOpen(true);
    };

    const handleSaveEdit = async () => {
        if (!editingMedia) return;
        try {
            await mediaApi.update(editingMedia.id, {
                title: editForm.title,
                description: editForm.description,
                state: editForm.status,
                // category 需要映射 ID，为了简单此时如果后端接受名/传空先略过或者只做演示
                tags: editForm.tags.split(',').map(s => s.trim()).filter(Boolean),
            });
            setEditDialogOpen(false);
            loadMedia();
        } catch (err) {
            console.error("Failed to update media", err);
        }
    };

    const handleDeleteClick = (media: Media) => {
        setDeletingMedia(media);
        setDeleteDialogOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (!deletingMedia) return;
        try {
            await mediaApi.delete(deletingMedia.id);
            setDeleteDialogOpen(false);
            loadMedia();
        } catch (err) {
            console.error("Failed to delete media", err);
        }
    };

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-2xl font-bold text-slate-900">媒体管理</h2>
                    <p className="text-slate-500 text-sm mt-1">在这里集中管理所有的视频与图片资源</p>
                </div>
                <Button className="bg-blue-600 hover:bg-blue-700" onClick={() => setUploadDialogOpen(true)}>
                    <Upload className="w-4 h-4 mr-2"/>
                    上传媒体
                </Button>
            </div>

            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400"/>
                    <Input
                        placeholder="搜索媒体..."
                        className="pl-10"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                    <SelectTrigger className="w-[180px]">
                        <SelectValue placeholder="所有状态"/>
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">所有状态</SelectItem>
                        <SelectItem value="active">已发布 (Active)</SelectItem>
                        <SelectItem value="draft">草稿 (Draft)</SelectItem>
                        <SelectItem value="deleted">已删除 (Deleted)</SelectItem>
                    </SelectContent>
                </Select>
            </div>

            {/* Stats Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">媒体总数</p>
                                <p className="text-2xl font-bold">{mediaList.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <Video className="w-6 h-6 text-blue-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">视频</p>
                                <p className="text-2xl font-bold">{mediaList.filter(m => m.type === 'video').length}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <Play className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">图片</p>
                                <p className="text-2xl font-bold">{mediaList.filter(m => m.type === 'image' || !m.type).length}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <ImageIcon className="w-6 h-6 text-green-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">总播放量</p>
                                <p className="text-2xl font-bold">{formatViews(mediaList.reduce((acc, m) => acc + (m.view_count || 0), 0))}</p>
                            </div>
                            <div className="w-12 h-12 bg-orange-100 rounded-xl flex items-center justify-center">
                                <Eye className="w-6 h-6 text-orange-600"/>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Media Table */}
            <Card>
                <CardHeader>
                    <CardTitle>所有媒体</CardTitle>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="py-12 text-center text-slate-400">正在加载数据...</div>
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
                                    <TableHead>作者</TableHead>
                                    <TableHead>日期</TableHead>
                                    <TableHead className="text-right">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredMedia.length > 0 ? filteredMedia.map((media) => (
                                    <TableRow key={media.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-3">
                                                <div
                                                    className="w-16 h-10 bg-slate-100 rounded overflow-hidden shrink-0 flex items-center justify-center">
                                                    {media.thumbnail ? (
                                                        <img
                                                            src={media.thumbnail.startsWith('http') ? media.thumbnail : `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:9090'}${media.thumbnail.startsWith('/') ? '' : '/'}${media.thumbnail}`}
                                                            alt="" className="w-full h-full object-cover"/>
                                                    ) : (
                                                        media.type === 'video' ?
                                                            <Video className="w-4 h-4 text-slate-400"/> :
                                                            <ImageIcon className="w-4 h-4 text-slate-400"/>
                                                    )}
                                                </div>
                                                <span className="font-medium line-clamp-1 max-w-[200px]"
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
                                        <TableCell
                                            className="text-sm text-slate-500">{media.edges?.user?.[0]?.nickname || media.edges?.user?.[0]?.username || '-'}</TableCell>
                                        <TableCell
                                            className="text-sm text-slate-500">{new Date(media.created_at).toLocaleDateString()}</TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button variant="ghost" size="sm">
                                                        <MoreVertical className="w-4 h-4"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuItem
                                                        onClick={() => window.open(`/watch?v=${media.friendly_token || media.id}`, '_blank')}>
                                                        <Eye className="w-4 h-4 mr-2"/>
                                                        查看
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => handleEditClick(media)}>
                                                        <Edit className="w-4 h-4 mr-2"/>
                                                        编辑
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem className="text-red-600"
                                                                      onClick={() => handleDeleteClick(media)}>
                                                        <Trash2 className="w-4 h-4 mr-2"/>
                                                        删除
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow>
                                        <TableCell colSpan={9} className="h-24 text-center text-slate-400">
                                            没有找到匹配的媒体数据
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            {/* 上传模态框 */}
            <Dialog open={uploadDialogOpen} onOpenChange={setUploadDialogOpen}>
                <DialogContent className="max-w-4xl">
                    <DialogHeader>
                        <DialogTitle>上传媒体文件</DialogTitle>
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

            {/* 编辑模态框 */}
            <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>编辑媒体信息</DialogTitle>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label htmlFor="title">标题</Label>
                            <Input
                                id="title"
                                value={editForm.title}
                                onChange={(e) => setEditForm({...editForm, title: e.target.value})}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="status">状态</Label>
                            <Select value={editForm.status}
                                    onValueChange={(val) => setEditForm({...editForm, status: val})}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择状态"/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="draft">草稿 (Draft)</SelectItem>
                                    <SelectItem value="active">已发布 (Active)</SelectItem>
                                    <SelectItem value="deleted">已删除 (Deleted)</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="tags">标签 (逗号分隔)</Label>
                            <Input
                                id="tags"
                                value={editForm.tags}
                                onChange={(e) => setEditForm({...editForm, tags: e.target.value})}
                                placeholder="如：编程, 运维"
                            />
                        </div>
                    </div>
                    <div className="flex justify-end gap-3">
                        <Button variant="outline" onClick={() => setEditDialogOpen(false)}>取消</Button>
                        <Button onClick={handleSaveEdit}>保存修改</Button>
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
        </div>
    );
}