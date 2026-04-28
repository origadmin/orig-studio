import {Spinner} from "@/components/ui/spinner"
import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
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
} from '@/components/ui/alert-dialog';
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, RotateCcw} from 'lucide-react';
import {categoryApi, Category} from '@/lib/api/category';
import {TablePagination} from '@/components/common/TablePagination';

const Categories: React.FC = () => {
    const {t} = useTranslation();
    const [searchParams, setSearchParams] = useState({keyword: '', page: 1, page_size: 20});
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [total, setTotal] = useState(0);
    
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [showDeleteDialog, setShowDeleteDialog] = useState(false);
    
    const [currentCategory, setCurrentCategory] = useState<Category | null>(null);
    const [formData, setFormData] = useState<Partial<Category>>({
        name: '',
        slug: '',
        description: '',
        parent_id: undefined,
        order: 0,
    });

    useEffect(() => {
        loadCategories();
    }, [searchParams.page]);

    const loadCategories = async (params = searchParams) => {
        try {
            setLoading(true);
            const response = await categoryApi.getAll({page: params.page, page_size: params.page_size});
            const categoryList = Array.isArray(response?.items) ? response.items : [];
            setCategories(categoryList);
            if (response?.total !== undefined) {
                setTotal(response.total);
            }
        } catch (error) {
            console.error('Failed to fetch categories:', error);
        } finally {
            setLoading(false);
        }
    };

    const resetForm = () => {
        setFormData({
            name: '',
            slug: '',
            description: '',
            parent_id: undefined,
            order: 0,
        });
    };

    const handleCreate = async () => {
        try {
            await categoryApi.create(formData as Partial<Category>);
            await loadCategories();
            setShowCreateDialog(false);
            resetForm();
        } catch (err) {
            console.error('Failed to create category:', err);
        }
    };

    const handleUpdate = async () => {
        if (!currentCategory) return;

        try {
            await categoryApi.update(currentCategory.id, formData as Partial<Category>);
            await loadCategories();
            setShowEditDialog(false);
            resetForm();
            setCurrentCategory(null);
        } catch (err) {
            console.error('Failed to update category:', err);
        }
    };

    const handleDelete = async () => {
        if (!currentCategory) return;

        try {
            await categoryApi.delete(currentCategory.id);
            await loadCategories();
            setShowDeleteDialog(false);
            setCurrentCategory(null);
        } catch (err) {
            console.error('Failed to delete category:', err);
        }
    };

    const openCreateDialog = () => {
        resetForm();
        setShowCreateDialog(true);
    };

    const openEditDialog = (category: Category) => {
        setCurrentCategory(category);
        setFormData({
            name: category.name,
            slug: category.slug,
            description: category.description || '',
            parent_id: category.parent_id ?? undefined,
            order: category.order || 0,
        });
        setShowEditDialog(true);
    };

    const openDeleteDialog = (category: Category) => {
        setCurrentCategory(category);
        setShowDeleteDialog(true);
    };

    const handleView = (category: Category) => {
        window.open(`/categories/${category.slug}`, '_blank');
    };

    const activeCount = categories.filter(c => c.status === 1 || String(c.status) === 'active' || String(c.status) === 'Enabled').length;
    const totalMedia = categories.reduce((sum, c) => sum + (c.media_count || 0), 0);

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.categories')}</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage your content categories
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
                                        placeholder={t('admin.search') || t('admin.categories') + '...'}
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
                                        loadCategories(newParams);
                                    }}
                                >
                                    <RotateCcw className="h-4 w-4 mr-2"/>
                                    Reset
                                </Button>
                                <Button
                                    variant="default"
                                    size="sm"
                                    onClick={() => loadCategories()}
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
                        <div className="text-2xl font-bold text-info dark:text-blue-400">{categories.length}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalCategories')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-success dark:text-green-400">{activeCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.activeCategories')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-cyan-600 dark:text-cyan-400">{totalMedia}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalMedia')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-cyan-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-warning dark:text-amber-400">{Math.min(categories.length, 5)}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.topCategories')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-amber-500 w-full opacity-10"/>
                </Card>
            </div>

            {/* 分类表格 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.categoryList')}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button size="sm" onClick={openCreateDialog}>
                                <Plus className="mr-2 h-4 w-4"/>
                                {t('admin.newCategory')}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="py-12 text-center">
                            <Spinner className="mx-auto" />
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>ID</TableHead>
                                    <TableHead>{t('admin.name')}</TableHead>
                                    <TableHead>Slug</TableHead>
                                    <TableHead>{t('admin.description')}</TableHead>
                                    <TableHead className="text-right">{t('admin.mediaCount')}</TableHead>
                                    <TableHead>{t('admin.order')}</TableHead>
                                    <TableHead>{t('admin.status')}</TableHead>
                                    <TableHead className="text-right">{t('admin.actions')}</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {categories.length > 0 ? categories.map((category) => (
                                    <TableRow key={category.id}>
                                        <TableCell className="font-medium">{category.id}</TableCell>
                                        <TableCell>{category.name}</TableCell>
                                        <TableCell>
                                            <code className="text-xs bg-muted px-2 py-1 rounded">{category.slug}</code>
                                        </TableCell>
                                        <TableCell className="text-muted-foreground max-w-[200px] truncate">
                                            {category.description || '-'}
                                        </TableCell>
                                        <TableCell className="text-right">{category.media_count || 0}</TableCell>
                                        <TableCell>{category.order || 0}</TableCell>
                                        <TableCell>
                                            <Badge
                                                variant={(category.status === 1 || String(category.status) === 'active' || String(category.status) === 'Enabled') ? 'default' : 'secondary'}>
                                                {(category.status === 1 || String(category.status) === 'active' || String(category.status) === 'Enabled') ? t('admin.enabled') : t('admin.disabled')}
                                            </Badge>
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
                                                    <DropdownMenuItem onClick={() => handleView(category)}>
                                                        <Eye className="mr-2 h-4 w-4"/>
                                                        {t('admin.view')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => openEditDialog(category)}>
                                                        <Edit className="mr-2 h-4 w-4"/>
                                                        {t('admin.edit')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem 
                                                        className="text-destructive focus:text-destructive" 
                                                        onClick={() => openDeleteDialog(category)}
                                                    >
                                                        <Trash2 className="mr-2 h-4 w-4"/>
                                                        {t('admin.delete')}
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow key="empty">
                                        <TableCell colSpan={8} className="text-center py-8">
                                            {t('admin.noCategoriesFound') || "No categories found"}
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

            {/* Create Category Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.newCategory') || "New Category"}</DialogTitle>
                        <DialogDescription>{t('admin.createNewCategory') || "Create a new category"}</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.name') || "Name"}</label>
                            <Input
                                value={formData.name}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.slug') || "Slug"}</label>
                            <Input
                                value={formData.slug}
                                onChange={(e) => setFormData({...formData, slug: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.description') || "Description"}</label>
                            <Textarea
                                value={formData.description}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.order') || "Order"}</label>
                            <Input
                                type="number"
                                value={formData.order}
                                onChange={(e) => setFormData({...formData, order: parseInt(e.target.value) || 0})}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                            {t('admin.cancel') || "Cancel"}
                        </Button>
                        <Button onClick={handleCreate}>
                            {t('admin.create') || "Create"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Edit Category Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.editCategory') || "Edit Category"}</DialogTitle>
                        <DialogDescription>{t('admin.editCategoryDesc') || "Edit category details"}</DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.name') || "Name"}</label>
                            <Input
                                value={formData.name}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.slug') || "Slug"}</label>
                            <Input
                                value={formData.slug}
                                onChange={(e) => setFormData({...formData, slug: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.description') || "Description"}</label>
                            <Textarea
                                value={formData.description}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-medium mb-1">{t('admin.order') || "Order"}</label>
                            <Input
                                type="number"
                                value={formData.order}
                                onChange={(e) => setFormData({...formData, order: parseInt(e.target.value) || 0})}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                            {t('admin.cancel') || "Cancel"}
                        </Button>
                        <Button onClick={handleUpdate}>
                            {t('admin.save') || "Save"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Delete Category Alert Dialog */}
            <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>{t('admin.deleteCategory') || "Delete Category"}</AlertDialogTitle>
                        <AlertDialogDescription>
                            {t('admin.deleteCategoryConfirm') || "Are you sure you want to delete this category? This action cannot be undone."}
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>{t('admin.cancel') || "Cancel"}</AlertDialogCancel>
                        <AlertDialogAction className="bg-red-600 hover:bg-red-700" onClick={handleDelete}>
                            {t('admin.delete') || "Delete"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
};

export default Categories;