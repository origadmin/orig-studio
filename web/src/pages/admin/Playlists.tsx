import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
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
import {MoreHorizontal, Search, Edit, Trash2, Eye, PlayCircle, Lock, Globe, User, Filter, RotateCcw} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';

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
    const {t} = useTranslation();
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
            public: {icon: Globe, label: t('admin.pub'), variant: 'default' as const},
            private: {icon: Lock, label: t('admin.priv'), variant: 'secondary' as const},
            unlisted: {icon: Eye, label: t('admin.unlisted'), variant: 'outline' as const},
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

    const formatNumber = (num: number | undefined | null) => {
        if (num === undefined || num === null) return '0';
        if (num >= 10000) return (num / 10000).toFixed(1) + t('common.wan');
        return num.toString();
    };

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.playlists')}</h2>
                                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5">
                                    Manage your playlists
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
                                        placeholder={t('admin.search') || t('admin.playlists') + '...'}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-9 w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={visibilityFilter} onValueChange={setVisibilityFilter}>
                                    <SelectTrigger className="w-[140px] h-9 focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {visibilityFilter === 'all' ? (
                                                <span className="text-muted-foreground">Visibility</span>
                                            ) : (
                                                <SelectValue placeholder="Visibility"/>
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                        <SelectItem value="public">{t('admin.pub')}</SelectItem>
                                        <SelectItem value="private">{t('admin.priv')}</SelectItem>
                                        <SelectItem value="unlisted">{t('admin.unlisted')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-9 px-3"
                                        onClick={() => {
                                            setSearchTerm('');
                                            setVisibilityFilter('all');
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
                            <PlayCircle className="h-5 w-5 text-indigo-600"/>
                            <div>
                                <div className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">{totalPlaylists}</div>
                                <p className="text-sm text-muted-foreground">{t('admin.playlistTotal')}</p>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-indigo-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600 dark:text-green-400">{publicCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.publicLists')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-green-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{privateCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.privateLists')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-yellow-500 w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-cyan-600 dark:text-cyan-400">{formatNumber(totalViews)}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalViews')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-cyan-500 w-full opacity-10"/>
                </Card>
            </div>

            {/* 播放列表表格 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.playlistList')}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button>
                                <PlayCircle className="mr-2 h-4 w-4"/>
                                {t('admin.newPlaylist')}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>{t('admin.name')}</TableHead>
                                <TableHead>{t('admin.creator')}</TableHead>
                                <TableHead className="text-right">{t('admin.videoCount')}</TableHead>
                                <TableHead className="text-right">{t('admin.viewCount')}</TableHead>
                                <TableHead>{t('admin.visibility')}</TableHead>
                                <TableHead>{t('admin.createdAt')}</TableHead>
                                <TableHead className="text-right">{t('admin.actions')}</TableHead>
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

export default Playlists;
