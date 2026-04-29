import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
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
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, Hash, Filter, RotateCcw} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {tagApi, Tag, CreateTagRequest, UpdateTagRequest} from '@/lib/api/admin-tags';
import {TablePagination} from '@/components/common/TablePagination';

const Tags: React.FC = () => {
    const {t} = useTranslation();
    const [searchParams, setSearchParams] = useState({keyword: '', page: 1, page_size: 20});
    const [tags, setTags] = useState<Tag[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [total, setTotal] = useState(0);
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    const [currentTag, setCurrentTag] = useState<Tag | null>(null);
    const [formData, setFormData] = useState<Partial<CreateTagRequest & UpdateTagRequest>>({
        name: '',
        slug: '',
        description: '',
        color: '',
        status: 'active',
    });

    // 加载标签数据
    useEffect(() => {
        loadTags();
    }, [searchParams.page]);

    const loadTags = async (params = searchParams) => {
        setLoading(true);
        setError(null);
        try {
            const apiParams: any = {page: params.page, page_size: params.page_size};
            if (params.keyword) {
                apiParams.keyword = params.keyword;
            }
            const response = await tagApi.list(apiParams);
            const tagList = Array.isArray(response?.items) ? response.items : [];
            setTags(tagList);
            if (response?.total !== undefined) {
                setTotal(response.total);
            }
        } catch (err) {
            setError('Failed to load tags');
            console.error('Error loading tags:', err);
        } finally {
            setLoading(false);
        }
    };

    const resetForm = () => {
        setFormData({
            name: '',
            slug: '',
            description: '',
            color: '',
            status: 'active',
        });
    };

    const handleCreate = async () => {
        try {
            await tagApi.create(formData as CreateTagRequest);
            await loadTags();
            setShowCreateDialog(false);
            resetForm();
        } catch (err) {
            console.error('Failed to create tag:', err);
        }
    };

    const handleUpdate = async () => {
        if (!currentTag) return;

        try {
            await tagApi.update(currentTag.id, formData as UpdateTagRequest);
            await loadTags();
            setShowEditDialog(false);
            resetForm();
            setCurrentTag(null);
        } catch (err) {
            console.error('Failed to update tag:', err);
        }
    };

    const handleDelete = async () => {
        if (!currentTag) return;

        try {
            await tagApi.delete(currentTag.id);
            await loadTags();
            setShowDeleteDialog(false);
            setCurrentTag(null);
        } catch (err) {
            console.error('Failed to delete tag:', err);
        }
    };

    const openCreateDialog = () => {
        resetForm();
        setShowCreateDialog(true);
    };

    const openEditDialog = (tag: Tag) => {
        setCurrentTag(tag);
        setFormData({
            name: tag.name,
            slug: tag.slug,
            description: tag.description || '',
            color: tag.color || '',
            status: tag.status,
        });
        setShowEditDialog(true);
    };

    const openDeleteDialog = (tag: Tag) => {
        setCurrentTag(tag);
        setShowDeleteDialog(true);
    };

    const handleView = (tag: Tag) => {
        window.open(`/tags/${tag.slug}`, '_blank');
    };

    const totalTags = tags.length;
    const activeTags = tags.length; // 假设所有标签都是active
    const trendingTags = 0; // API返回数据中没有trending字段
    const totalMedia = tags.reduce((sum, t) => sum + (t.count || 0), 0);

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.tags')}</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage your content tags
                                </p>
                            </div>
                        </div>

                        {/* 分隔线 */}
                        <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                        {/* 搜索和筛选 */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1 min-w-[120px] max-w-[400px]">
                                <div className="relative w-full">
                                    <Search
                                        className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                    <Input
                                        placeholder={t('admin.search') || t('admin.tags') + '...'}
                                        value={searchParams.keyword}
                                        onChange={(e) => setSearchParams({...searchParams, keyword: e.target.value})}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => {
                                        const newParams = {keyword: '', page: 1, page_size: 20};
                                        setSearchParams(newParams);
                                        loadTags(newParams);
                                    }}
                                >
                                    <RotateCcw className="h-4 w-4 mr-2"/>
                                    Reset
                                </Button>
                                <Button
                                    variant="default"
                                    size="sm"
                                    onClick={() => loadTags()}
                                >
                                    <Search className="h-4 w-4 mr-2"/>
                                    Search
                                </Button>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <Hash className="h-5 w-5 text-purple-600"/>
                            <div>
                                <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">{totalTags}</div>
                                <p className="text-sm text-muted-foreground">{t('admin.tagTotal')}</p>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-purple-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-success dark:text-green-400">{activeTags}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.activeTags')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-orange-500 dark:text-orange-400">{trendingTags}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.trendingTags')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-orange-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-info dark:text-blue-400">{totalMedia}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.relatedMedia')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                </Card>
            </div>

            {/* 标签表格 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.tagList')}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button size="sm" onClick={openCreateDialog}>
                                <Plus className="mr-2 h-4 w-4"/>
                                {t('admin.newTag')}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>{t('admin.tagName')}</TableHead>
                                <TableHead>Slug</TableHead>
                                <TableHead className="text-right">{t('admin.mediaCount')}</TableHead>
                                <TableHead>Created</TableHead>
                                <TableHead className="text-right">{t('admin.actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center py-8">
                                        <div className="animate-pulse">Loading tags...</div>
                                    </TableCell>
                                </TableRow>
                            ) : error ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center py-8">
                                        <div className="text-destructive">{error}</div>
                                        <Button 
                                            variant="outline" 
                                            size="sm" 
                                            className="mt-2"
                                            onClick={() => window.location.reload()}
                                        >
                                            Retry
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ) : tags.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={6} className="text-center py-8">
                                        No tags found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                tags.map((tag) => (
                                    <TableRow key={tag.id}>
                                        <TableCell className="font-medium">{tag.id}</TableCell>
                                        <TableCell>
                                            <span className="font-medium">{tag.name}</span>
                                        </TableCell>
                                        <TableCell>
                                            <code className="text-xs bg-muted px-2 py-1 rounded">{tag.slug}</code>
                                        </TableCell>
                                        <TableCell className="text-right">{tag.count || 0}</TableCell>
                                        <TableCell>
                                            <span className="text-sm text-muted-foreground">
                                                {new Date(tag.create_time).toLocaleDateString()}
                                            </span>
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button 
                                                        variant="ghost" 
                                                        size="icon" 
                                                        className="h-6 w-6" 
                                                        title="More Actions"
                                                    >
                                                        <MoreHorizontal className="h-3 w-3"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem onClick={() => handleView(tag)}>
                                                        <Eye className="mr-2 h-4 w-4"/>
                                                        {t('admin.view')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => openEditDialog(tag)}>
                                                        <Edit className="mr-2 h-4 w-4"/>
                                                        {t('admin.edit')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem 
                                                        className="text-destructive focus:text-destructive" 
                                                        onClick={() => openDeleteDialog(tag)}
                                                    >
                                                        <Trash2 className="mr-2 h-4 w-4"/>
                                                        {t('admin.delete')}
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <TablePagination
                page={searchParams.page}
                pageSize={searchParams.page_size}
                total={total}
                onPageChange={(p) => setSearchParams({...searchParams, page: p})}
            />

            {/* Create Tag Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.newTag') || 'New Tag'}</DialogTitle>
                        <DialogDescription>
                            Create a new tag for organizing your content
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Name *
                            </h4>
                            <Input
                                placeholder="Enter tag name"
                                value={formData.name || ''}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Slug *
                            </h4>
                            <Input
                                placeholder="Enter tag slug"
                                value={formData.slug || ''}
                                onChange={(e) => setFormData({...formData, slug: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Description
                            </h4>
                            <Textarea
                                placeholder="Enter tag description"
                                value={formData.description || ''}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Color
                            </h4>
                            <Input
                                placeholder="#000000"
                                value={formData.color || ''}
                                onChange={(e) => setFormData({...formData, color: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Status
                            </h4>
                            <Select
                                value={formData.status || 'active'}
                                onValueChange={(value) => setFormData({...formData, status: value})}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select status"/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="active">Active</SelectItem>
                                    <SelectItem value="inactive">Inactive</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button onClick={handleCreate}>
                            {t('common.save') || 'Save'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Edit Tag Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.editTag') || 'Edit Tag'}</DialogTitle>
                        <DialogDescription>
                            Update the tag information
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Name *
                            </h4>
                            <Input
                                placeholder="Enter tag name"
                                value={formData.name || ''}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Slug *
                            </h4>
                            <Input
                                placeholder="Enter tag slug"
                                value={formData.slug || ''}
                                onChange={(e) => setFormData({...formData, slug: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Description
                            </h4>
                            <Textarea
                                placeholder="Enter tag description"
                                value={formData.description || ''}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Color
                            </h4>
                            <Input
                                placeholder="#000000"
                                value={formData.color || ''}
                                onChange={(e) => setFormData({...formData, color: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                Status
                            </h4>
                            <Select
                                value={formData.status || 'active'}
                                onValueChange={(value) => setFormData({...formData, status: value})}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder="Select status"/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="active">Active</SelectItem>
                                    <SelectItem value="inactive">Inactive</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                            {t('common.cancel') || 'Cancel'}
                        </Button>
                        <Button onClick={handleUpdate}>
                            {t('common.save') || 'Save'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Delete Tag Dialog */}
            <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t('admin.deleteTag') || 'Delete Tag'}</AlertDialogTitle>
                        <AlertDialogDescription>
                            Are you sure you want to delete this tag? This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel onClick={() => setShowDeleteDialog(false)}>
                            {t('common.cancel') || 'Cancel'}
                        </AlertDialogCancel>
                        <AlertDialogAction 
                            onClick={handleDelete}
                            className="bg-red-600 hover:bg-red-700"
                        >
                            {t('common.delete') || 'Delete'}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
};

export default Tags;
