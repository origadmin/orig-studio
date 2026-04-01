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
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, Hash} from 'lucide-react';

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
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <Hash className="h-5 w-5 text-purple-600"/>
                            <div>
                                <div className="text-2xl font-bold">{totalTags}</div>
                                <p className="text-sm text-muted-foreground">标签总数</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{activeTags}</div>
                        <p className="text-sm text-muted-foreground">活跃标签</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-orange-500">{trendingTags}</div>
                        <p className="text-sm text-muted-foreground">热门标签</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{totalMedia}</div>
                        <p className="text-sm text-muted-foreground">关联媒体</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="flex gap-2 flex-1">
                    <div className="relative flex-1 max-w-md">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder="搜索标签..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="pl-10"
                        />
                    </div>
                    <select
                        className="px-3 py-2 border rounded-md bg-background"
                        value={trendingFilter}
                        onChange={(e) => setTrendingFilter(e.target.value)}
                    >
                        <option value="all">全部</option>
                        <option value="trending">热门</option>
                        <option value="normal">普通</option>
                    </select>
                </div>
                <Button>
                    <Plus className="mr-2 h-4 w-4"/>
                    新建标签
                </Button>
            </div>

            {/* 标签表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>标签列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>标签名称</TableHead>
                                <TableHead>Slug</TableHead>
                                <TableHead className="text-right">媒体数</TableHead>
                                <TableHead>热门</TableHead>
                                <TableHead>状态</TableHead>
                                <TableHead className="text-right">操作</TableHead>
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
                                            <Badge variant="default" className="bg-orange-500">热门</Badge>
                                        ) : (
                                            <span className="text-muted-foreground">-</span>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant={tag.status === 'active' ? 'secondary' : 'outline'}>
                                            {tag.status === 'active' ? '启用' : '禁用'}
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

export default Tags;
