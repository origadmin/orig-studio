import React, {useState} from 'react';
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

// 模拟数据
const mockCategories = [
    {id: 1, name: '科技', slug: 'tech', description: 'Technology videos', mediaCount: 156, order: 1, status: 'active'},
    {id: 2, name: '音乐', slug: 'music', description: 'Music videos', mediaCount: 89, order: 2, status: 'active'},
    {id: 3, name: '游戏', slug: 'gaming', description: 'Gaming content', mediaCount: 234, order: 3, status: 'active'},
    {
        id: 4,
        name: '教育',
        slug: 'education',
        description: 'Educational videos',
        mediaCount: 67,
        order: 4,
        status: 'active'
    },
    {
        id: 5,
        name: '娱乐',
        slug: 'entertainment',
        description: 'Entertainment',
        mediaCount: 312,
        order: 5,
        status: 'active'
    },
    {id: 6, name: '体育', slug: 'sports', description: 'Sports videos', mediaCount: 45, order: 6, status: 'inactive'},
];

const Categories: React.FC = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [categories] = useState(mockCategories);

    const filteredCategories = categories.filter(cat =>
        cat.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        cat.description.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const activeCount = categories.filter(c => c.status === 'active').length;
    const totalMedia = categories.reduce((sum, c) => sum + c.mediaCount, 0);

    return (
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{categories.length}</div>
                        <p className="text-sm text-muted-foreground">总分类数</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{activeCount}</div>
                        <p className="text-sm text-muted-foreground">活跃分类</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{totalMedia}</div>
                        <p className="text-sm text-muted-foreground">媒体总数</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-blue-600">5</div>
                        <p className="text-sm text-muted-foreground">一级分类</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="relative flex-1 max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                    <Input
                        placeholder="搜索分类..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="pl-10"
                    />
                </div>
                <Button>
                    <Plus className="mr-2 h-4 w-4"/>
                    新建分类
                </Button>
            </div>

            {/* 分类表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>分类列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>名称</TableHead>
                                <TableHead>Slug</TableHead>
                                <TableHead>描述</TableHead>
                                <TableHead className="text-right">媒体数</TableHead>
                                <TableHead>排序</TableHead>
                                <TableHead>状态</TableHead>
                                <TableHead className="text-right">操作</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredCategories.map((category) => (
                                <TableRow key={category.id}>
                                    <TableCell className="font-medium">{category.id}</TableCell>
                                    <TableCell>{category.name}</TableCell>
                                    <TableCell>
                                        <code className="text-xs bg-muted px-2 py-1 rounded">{category.slug}</code>
                                    </TableCell>
                                    <TableCell className="text-muted-foreground max-w-[200px] truncate">
                                        {category.description}
                                    </TableCell>
                                    <TableCell className="text-right">{category.mediaCount}</TableCell>
                                    <TableCell>{category.order}</TableCell>
                                    <TableCell>
                                        <Badge variant={category.status === 'active' ? 'default' : 'secondary'}>
                                            {category.status === 'active' ? '启用' : '禁用'}
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
                                                    查看
                                                </DropdownMenuItem>
                                                <DropdownMenuItem>
                                                    <Edit className="mr-2 h-4 w-4"/>
                                                    编辑
                                                </DropdownMenuItem>
                                                <DropdownMenuItem className="text-red-600">
                                                    <Trash2 className="mr-2 h-4 w-4"/>
                                                    删除
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

export default Categories;
