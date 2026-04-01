import React, {useState} from 'react';
import {Card, CardContent, CardHeader, CardTitle} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
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
import {MoreHorizontal, Search, Edit, Trash2, Eye, PlayCircle, Lock, Globe, User} from 'lucide-react';

// 模拟数据
const mockPlaylists = [
    {
        id: 1,
        name: 'Python 全栈教程',
        slug: 'python-fullstack',
        description: '从入门到精通 Python',
        owner: {name: '张三', username: 'zhangsan'},
        mediaCount: 45,
        viewCount: 12300,
        visibility: 'public',
        status: 'active',
        createdAt: '2024-04-15'
    },
    {
        id: 2,
        name: 'React 进阶之路',
        slug: 'react-advanced',
        description: 'React 高级模式和最佳实践',
        owner: {name: '李四', username: 'lisi'},
        mediaCount: 32,
        viewCount: 8900,
        visibility: 'public',
        status: 'active',
        createdAt: '2024-04-20'
    },
    {
        id: 3,
        name: '我的收藏',
        slug: 'my-favorites',
        description: '我喜欢的视频',
        owner: {name: '王五', username: 'wangwu'},
        mediaCount: 89,
        viewCount: 5600,
        visibility: 'private',
        status: 'active',
        createdAt: '2024-05-01'
    },
    {
        id: 4,
        name: 'Go 语言实战',
        slug: 'go-practical',
        description: 'Go 项目实战教程',
        owner: {name: '赵六', username: 'zhaoliu'},
        mediaCount: 28,
        viewCount: 4500,
        visibility: 'public',
        status: 'active',
        createdAt: '2024-05-10'
    },
    {
        id: 5,
        name: '算法与数据结构',
        slug: 'algorithms',
        description: '面试必看算法题',
        owner: {name: '钱七', username: 'qianqi'},
        mediaCount: 67,
        viewCount: 15600,
        visibility: 'unlisted',
        status: 'active',
        createdAt: '2024-05-15'
    },
    {
        id: 6,
        name: '待整理',
        slug: 'to-organize',
        description: '还没整理的视频',
        owner: {name: '孙八', username: 'sunba'},
        mediaCount: 12,
        viewCount: 0,
        visibility: 'private',
        status: 'draft',
        createdAt: '2024-05-18'
    },
];

const Playlists: React.FC = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [visibilityFilter, setVisibilityFilter] = useState('all');
    const [playlists] = useState(mockPlaylists);

    const filteredPlaylists = playlists.filter(playlist => {
        const matchesSearch = playlist.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            playlist.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesVisibility = visibilityFilter === 'all' || playlist.visibility === visibilityFilter;
        return matchesSearch && matchesVisibility;
    });

    const totalPlaylists = playlists.length;
    const publicCount = playlists.filter(p => p.visibility === 'public').length;
    const privateCount = playlists.filter(p => p.visibility === 'private').length;
    const totalViews = playlists.reduce((sum, p) => sum + p.viewCount, 0);

    const getVisibilityBadge = (visibility: string) => {
        const configs = {
            public: {icon: Globe, label: '公开', variant: 'default' as const},
            private: {icon: Lock, label: '私有', variant: 'secondary' as const},
            unlisted: {icon: Eye, label: '未列出', variant: 'outline' as const},
        };
        const config = configs[visibility] || configs.public;
        const Icon = config.icon;
        return (
            <Badge variant={config.variant}>
                <Icon className="mr-1 h-3 w-3"/>
                {config.label}
            </Badge>
        );
    };

    const formatNumber = (num: number) => {
        if (num >= 10000) return (num / 10000).toFixed(1) + '万';
        return num.toString();
    };

    return (
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <PlayCircle className="h-5 w-5 text-indigo-600"/>
                            <div>
                                <div className="text-2xl font-bold">{totalPlaylists}</div>
                                <p className="text-sm text-muted-foreground">播放列表总数</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{publicCount}</div>
                        <p className="text-sm text-muted-foreground">公开列表</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600">{privateCount}</div>
                        <p className="text-sm text-muted-foreground">私有列表</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{formatNumber(totalViews)}</div>
                        <p className="text-sm text-muted-foreground">总播放量</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="flex gap-2 flex-1">
                    <div className="relative flex-1 max-w-md">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder="搜索播放列表..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="pl-10"
                        />
                    </div>
                    <select
                        className="px-3 py-2 border rounded-md bg-background"
                        value={visibilityFilter}
                        onChange={(e) => setVisibilityFilter(e.target.value)}
                    >
                        <option value="all">全部可见性</option>
                        <option value="public">公开</option>
                        <option value="private">私有</option>
                        <option value="unlisted">未列出</option>
                    </select>
                </div>
                <Button>
                    <PlayCircle className="mr-2 h-4 w-4"/>
                    新建播放列表
                </Button>
            </div>

            {/* 播放列表表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>播放列表列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>名称</TableHead>
                                <TableHead>创建者</TableHead>
                                <TableHead className="text-right">视频数</TableHead>
                                <TableHead className="text-right">播放量</TableHead>
                                <TableHead>可见性</TableHead>
                                <TableHead>创建时间</TableHead>
                                <TableHead className="text-right">操作</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredPlaylists.map((playlist) => (
                                <TableRow key={playlist.id}>
                                    <TableCell className="font-medium">{playlist.id}</TableCell>
                                    <TableCell>
                                        <div>
                                            <div className="font-medium">{playlist.name}</div>
                                            <div className="text-xs text-muted-foreground">{playlist.slug}</div>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            <Avatar className="h-6 w-6">
                                                <AvatarFallback className="text-xs">
                                                    <User className="h-3 w-3"/>
                                                </AvatarFallback>
                                            </Avatar>
                                            {playlist.owner.name}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">{playlist.mediaCount}</TableCell>
                                    <TableCell className="text-right">{formatNumber(playlist.viewCount)}</TableCell>
                                    <TableCell>{getVisibilityBadge(playlist.visibility)}</TableCell>
                                    <TableCell className="text-muted-foreground">{playlist.createdAt}</TableCell>
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

export default Playlists;
