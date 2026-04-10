import React, {useState} from 'react';
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
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, Hash, Filter, RotateCcw} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';

// 模拟数据
const mockTags = [
    {id: 1, name: '人工智能', slug: 'ai', mediaCount: 234, trending: true, status: 'active'},
    {id: 2, name: '机器学习', slug: 'machine-learning', mediaCount: 156, trending: true, status: 'active'},
    {id: 3, name: 'Python', slug: 'python', mediaCount: 189, trending: false, status: 'active'},
    {id: 4, name: 'JavaScript', slug: 'javascript', mediaCount: 312, trending: false, status: 'active'},
    {id: 5, name: '区块链', slug: 'blockchain', mediaCount: 89, trending: false, status: 'active'},
    {id: 6, name: '云计算', slug: 'cloud-computing', mediaCount: 67, trending: true, status: 'active'},
    {id: 7, name: '大数据', slug: 'big-data', mediaCount: 45, trending: false, status: 'inactive'},
    {id: 8, name: '物联网', slug: 'iot', mediaCount: 23, trending: false, status: 'active'},
    {id: 9, name: '5G', slug: '5g', mediaCount: 78, trending: false, status: 'active'},
    {id: 10, name: '网络安全', slug: 'cybersecurity', mediaCount: 56, trending: false, status: 'active'},
];

const Tags: React.FC = () => {
    const {t} = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [trendingFilter, setTrendingFilter] = useState('all');
    const [tags] = useState(mockTags);

    const filteredTags = tags.filter(tag => {
        const matchesSearch = tag.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            tag.slug.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesTrending = trendingFilter === 'all' ||
            (trendingFilter === 'trending' && tag.trending) ||
            (trendingFilter === 'normal' && !tag.trending);
        return matchesSearch && matchesTrending;
    });

    const totalTags = tags.length;
    const activeTags = tags.filter(t => t.status === 'active').length;
    const trendingTags = tags.filter(t => t.trending).length;
    const totalMedia = tags.reduce((sum, t) => sum + t.mediaCount, 0);

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
                                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5">
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
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-9 w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={trendingFilter} onValueChange={setTrendingFilter}>
                                    <SelectTrigger className="w-[140px] h-9 focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {trendingFilter === 'all' ? (
                                                <span className="text-muted-foreground">Trending</span>
                                            ) : (
                                                <SelectValue placeholder="Trending"/>
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                        <SelectItem value="trending">{t('admin.trending')}</SelectItem>
                                        <SelectItem value="normal">{t('admin.normalTag')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-9 px-3"
                                        onClick={() => {
                                            setSearchTerm('');
                                            setTrendingFilter('all');
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
                        <div className="text-2xl font-bold text-green-600 dark:text-green-400">{activeTags}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.activeTags')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-green-500 w-full opacity-10"/>
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
                        <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">{totalMedia}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.relatedMedia')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-blue-500 w-full opacity-10"/>
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
                            <Button>
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
                                <TableHead>{t('admin.trendingCol')}</TableHead>
                                <TableHead>{t('admin.status')}</TableHead>
                                <TableHead className="text-right">{t('admin.actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredTags.map((tag) => (
                                <TableRow key={tag.id}>
                                    <TableCell className="font-medium">{tag.id}</TableCell>
                                    <TableCell>
                                        <span className="font-medium">{tag.name}</span>
                                    </TableCell>
                                    <TableCell>
                                        <code className="text-xs bg-muted px-2 py-1 rounded">{tag.slug}</code>
                                    </TableCell>
                                    <TableCell className="text-right">{tag.mediaCount}</TableCell>
                                    <TableCell>
                                        {tag.trending ? (
                                            <Badge variant="default"
                                                   className="bg-orange-500">{t('admin.trending')}</Badge>
                                        ) : (
                                            <span className="text-muted-foreground">-</span>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={tag.status === 'active' ? 'secondary' : 'outline'}>
                                            {tag.status === 'active' ? t('admin.enabled') : t('admin.disabled')}
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
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
};

export default Tags;
