import React, {useState, useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
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
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye} from 'lucide-react';
import {categoryApi} from '@/lib/api/category';

const Categories: React.FC = () => {
    const {t} = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [categories, setCategories] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchCategories = async () => {
            try {
                setLoading(true);
                const response = await categoryApi.getAll();
                setCategories(response || []);
            } catch (error) {
                console.error('Failed to fetch categories:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchCategories();
    }, []);

    const filteredCategories = categories.filter(cat =>
        cat.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (cat.description && cat.description.toLowerCase().includes(searchTerm.toLowerCase()))
    );

    const activeCount = categories.filter(c => c.status === 'active' || c.status === 'Enabled').length;
    const totalMedia = categories.reduce((sum, c) => sum + (c.media_count || 0), 0);

    return (
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{categories.length}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalCategories')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{activeCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.activeCategories')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{totalMedia}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalMedia')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-blue-600">{Math.min(categories.length, 5)}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.topCategories')}</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                    <Input
                        placeholder={t('admin.search') || t('admin.categories') + '...'}
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10"
                    />
                </div>
                <Button>
                    <Plus className="mr-2 h-4 w-4"/>
                    {t('admin.newCategory')}
                </Button>
            </div>

            {/* 分类表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>{t('admin.categoryList')}</CardTitle>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="py-12 text-center">
                            <div
                                className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full mx-auto"/>
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
                                {filteredCategories.length > 0 ? filteredCategories.map((category) => (
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
                                                variant={(category.status === 'active' || category.status === 'Enabled') ? 'default' : 'secondary'}>
                                                {(category.status === 'active' || category.status === 'Enabled') ? t('admin.enabled') : t('admin.disabled')}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button variant="ghost" size="sm">
                                                        <MoreHorizontal className="h-4 w-4"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuItem>
                                                        <Eye className="mr-2 h-4 w-4"/>
                                                        {t('admin.view')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem>
                                                        <Edit className="mr-2 h-4 w-4"/>
                                                        {t('admin.edit')}
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem className="text-red-600">
                                                        <Trash2 className="mr-2 h-4 w-4"/>
                                                        {t('admin.delete')}
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                )) : (
                                    <TableRow>
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
        </div>
    );
};

export default Categories;
