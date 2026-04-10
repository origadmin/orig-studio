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
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, RotateCcw} from 'lucide-react';
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
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.categories')}</h2>
                                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5">
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
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-9 w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-9 px-3"
                                    onClick={() => {
                                        setSearchTerm('');
                                    }}
                                >
                                    <RotateCcw className="h-4 w-4 mr-2"/>
                                    Reset
                                </Button>
                                <Button
                                    variant="default"
                                    size="sm"
                                    className="h-9 px-4"
                                    onClick={() => {
                                        // 这里可以添加搜索逻辑
                                    }}
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
                        <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{categories.length}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalCategories')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-blue-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600 dark:text-green-400">{activeCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.activeCategories')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-green-500 w-full opacity-10"/>
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
                        <div className="text-2xl font-bold text-amber-600 dark:text-amber-400">{Math.min(categories.length, 5)}</div>
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
                            <Button>
                                <Plus className="mr-2 h-4 w-4"/>
                                {t('admin.newCategory')}
                            </Button>
                        </div>
                    </div>
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
