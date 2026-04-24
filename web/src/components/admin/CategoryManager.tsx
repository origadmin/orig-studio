import React, {useState, useEffect} from 'react';
import {
    Plus,
    Edit,
    Trash2,
    ChevronUp,
    ChevronDown,
    ChevronRight,
    ChevronDown as ChevronDownIcon,
    X
} from 'lucide-react';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from '@/components/ui/table';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger
} from '@/components/ui/dialog';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Skeleton} from '@/components/ui/skeleton';
import {useTranslation} from 'react-i18next';
import {categoryApi, type Category} from '@/lib/api/category';
import ErrorPage from '@/components/common/ErrorPage';

const CategoryManager: React.FC = () => {
    const {t} = useTranslation();
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showCreateDialog, setShowCreateDialog] = useState(false);
    const [showEditDialog, setShowEditDialog] = useState(false);
    const [currentCategory, setCurrentCategory] = useState<Category | null>(null);
    const [formData, setFormData] = useState<Partial<Category>>({
        name: '',
        slug: '',
        description: '',
        parent_id: undefined,
        order: 0,
    });
    const [expanded, setExpanded] = useState<Record<string, boolean>>({});

    useEffect(() => {
        fetchCategories();
    }, []);

    const fetchCategories = async () => {
        try {
            setLoading(true);
            setError(null);
            const response = await categoryApi.getAll();
            setCategories((response as any)?.items || response || []);
        } catch (err) {
            setError('Failed to fetch categories');
            console.error('Failed to fetch categories:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async () => {
        try {
            await categoryApi.create(formData);
            await fetchCategories();
            setShowCreateDialog(false);
            resetForm();
        } catch (err) {
            console.error('Failed to create category:', err);
        }
    };

    const handleUpdate = async () => {
        if (!currentCategory) return;

        try {
            await categoryApi.update(currentCategory.id, formData);
            await fetchCategories();
            setShowEditDialog(false);
            resetForm();
            setCurrentCategory(null);
        } catch (err) {
            console.error('Failed to update category:', err);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this category?')) return;

        try {
            await categoryApi.delete(id);
            await fetchCategories();
        } catch (err) {
            console.error('Failed to delete category:', err);
        }
    };

    const handleOrderChange = async (id: string, direction: 'up' | 'down') => {
        const category = categories.find(c => c.id === id);
        if (!category) return;

        const siblings = categories.filter(c => c.parent_id === category.parent_id);
        const index = siblings.findIndex(c => c.id === id);

        if (direction === 'up' && index > 0) {
            const targetCategory = siblings[index - 1];
            await categoryApi.update(id, {order: targetCategory.order});
            await categoryApi.update(targetCategory.id, {order: category.order});
        } else if (direction === 'down' && index < siblings.length - 1) {
            const targetCategory = siblings[index + 1];
            await categoryApi.update(id, {order: targetCategory.order});
            await categoryApi.update(targetCategory.id, {order: category.order});
        }

        await fetchCategories();
    };

    const handleExpand = (id: string) => {
        setExpanded(prev => ({
            ...prev,
            [id]: !prev[id],
        }));
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

    const openEditDialog = (category: Category) => {
        setCurrentCategory(category);
        setFormData({
            name: category.name,
            slug: category.slug,
            description: category.description || '',
            parent_id: category.parent_id,
            order: category.order,
        });
        setShowEditDialog(true);
    };

    const openCreateDialog = () => {
        resetForm();
        setShowCreateDialog(true);
    };

    const getParentName = (parentId: string | undefined) => {
        if (!parentId) return t('admin.noParent');
        const parent = categories.find(c => c.id === parentId);
        return parent ? parent.name : t('admin.parentNotFound');
    };

    const renderCategoryTree = (parentId: string | undefined = undefined, level: number = 0) => {
        const children = categories
            .filter(c => c.parent_id === parentId)
            .sort((a, b) => a.order - b.order);

        return children.map(category => (
            <React.Fragment key={category.id}>
                <TableRow>
                    <TableCell style={{paddingLeft: `${level * 20}px`}}>
                        <div className="flex items-center gap-2">
                            {categories.some(c => c.parent_id === category.id) && (
                                <button
                                    onClick={() => handleExpand(category.id)}
                                    className="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded"
                                >
                                    {expanded[category.id] ? (
                                        <ChevronDownIcon className="w-4 h-4"/>
                                    ) : (
                                        <ChevronRight className="w-4 h-4"/>
                                    )}
                                </button>
                            )}
                            <span className="font-medium">{category.name}</span>
                        </div>
                    </TableCell>
                    <TableCell>{category.slug}</TableCell>
                    <TableCell>{getParentName(category.parent_id)}</TableCell>
                    <TableCell>{category.order}</TableCell>
                    <TableCell>
                        <div className="flex items-center gap-2">
                            <Button
                                size="sm"
                                onClick={() => handleOrderChange(category.id, 'up')}
                                disabled={!
                                    categories
                                        .filter(c => c.parent_id === category.parent_id)
                                        .some(c => c.order < category.order)
                                }
                            >
                                <ChevronUp className="w-4 h-4"/>
                            </Button>
                            <Button
                                size="sm"
                                onClick={() => handleOrderChange(category.id, 'down')}
                                disabled={!
                                    categories
                                        .filter(c => c.parent_id === category.parent_id)
                                        .some(c => c.order > category.order)
                                }
                            >
                                <ChevronDown className="w-4 h-4"/>
                            </Button>
                            <Button size="sm" variant="outline" onClick={() => openEditDialog(category)}>
                                <Edit className="w-4 h-4"/>
                            </Button>
                            <Button
                                size="sm"
                                variant="outline"
                                className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20"
                                onClick={() => handleDelete(category.id)}
                            >
                                <Trash2 className="w-4 h-4"/>
                            </Button>
                        </div>
                    </TableCell>
                </TableRow>
                {expanded[category.id] && renderCategoryTree(category.id, level + 1)}
            </React.Fragment>
        ));
    };

    if (loading) {
        return (
            <div className="space-y-4">
                <Card>
                    <CardHeader>
                        <CardTitle>Categories</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <div className="flex items-center justify-between">
                                <Skeleton className="h-8 w-32"/>
                            </div>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Name</TableHead>
                                        <TableHead>Slug</TableHead>
                                        <TableHead>Parent</TableHead>
                                        <TableHead>Order</TableHead>
                                        <TableHead>Actions</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {Array.from({length: 5}).map((_, i) => (
                                        <TableRow key={i}>
                                            <TableCell>
                                                <Skeleton className="h-4 w-64"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-32"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-32"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-16"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-8 w-32"/>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle>{t('admin.categories')}</CardTitle>
                    <CardDescription>
                        {t('admin.categoryList')}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center justify-between mb-6">
                        <Button onClick={openCreateDialog}>
                            <Plus className="w-4 h-4 mr-2"/>
                            {t('admin.newCategory')}
                        </Button>
                    </div>

                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>{t('admin.name')}</TableHead>
                                <TableHead>{t('admin.slug')}</TableHead>
                                <TableHead>{t('admin.parent')}</TableHead>
                                <TableHead>{t('admin.order')}</TableHead>
                                <TableHead>{t('admin.actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {renderCategoryTree()}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            {/* Create Category Dialog */}
            <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.newCategory')}</DialogTitle>
                        <DialogDescription>
                            {t('admin.createCategoryDesc')}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.name')} *
                            </h4>
                            <Input
                                placeholder={t('admin.enterCategoryName')}
                                value={formData.name || ''}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.slug')} *
                            </h4>
                            <Input
                                placeholder={t('admin.enterCategorySlug')}
                                value={formData.slug || ''}
                                onChange={(e) => setFormData({...formData, slug: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.description')}
                            </h4>
                            <Textarea
                                placeholder={t('admin.enterCategoryDescription')}
                                value={formData.description || ''}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.parent')}
                            </h4>
                            <Select
                                value={formData.parent_id || ''}
                                onValueChange={(value) => setFormData({...formData, parent_id: value || undefined})}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder={t('admin.selectParentCategory')}/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="">{t('admin.noParent')}</SelectItem>
                                    {categories.map(category => (
                                        <SelectItem key={category.id} value={category.id}>
                                            {category.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.order')}
                            </h4>
                            <Input
                                type="number"
                                value={formData.order || 0}
                                onChange={(e) => setFormData({...formData, order: parseInt(e.target.value) || 0})}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                            {t('common.cancel')}
                        </Button>
                        <Button onClick={handleCreate}>
                            {t('common.save')}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Edit Category Dialog */}
            <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('admin.editCategory')}</DialogTitle>
                        <DialogDescription>
                            {t('admin.editCategoryDesc')}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.name')} *
                            </h4>
                            <Input
                                placeholder={t('admin.enterCategoryName')}
                                value={formData.name || ''}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.slug')} *
                            </h4>
                            <Input
                                placeholder={t('admin.enterCategorySlug')}
                                value={formData.slug || ''}
                                onChange={(e) => setFormData({...formData, slug: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.description')}
                            </h4>
                            <Textarea
                                placeholder={t('admin.enterCategoryDescription')}
                                value={formData.description || ''}
                                onChange={(e) => setFormData({...formData, description: e.target.value})}
                            />
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.parent')}
                            </h4>
                            <Select
                                value={formData.parent_id || ''}
                                onValueChange={(value) => setFormData({...formData, parent_id: value || undefined})}
                            >
                                <SelectTrigger>
                                    <SelectValue placeholder={t('admin.selectParentCategory')}/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="">{t('admin.noParent')}</SelectItem>
                                    {categories
                                        .filter(category => !currentCategory || category.id !== currentCategory.id)
                                        .map(category => (
                                            <SelectItem key={category.id} value={category.id}>
                                                {category.name}
                                            </SelectItem>
                                        ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('admin.order')}
                            </h4>
                            <Input
                                type="number"
                                value={formData.order || 0}
                                onChange={(e) => setFormData({...formData, order: parseInt(e.target.value) || 0})}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowEditDialog(false)}>
                            {t('common.cancel')}
                        </Button>
                        <Button onClick={handleUpdate}>
                            {t('common.save')}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default CategoryManager;
