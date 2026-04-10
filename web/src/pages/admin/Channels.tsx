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
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, UserPlus, Users, Filter, Loader2, RotateCcw} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';

// 模拟数据
const mockChannels = [
    {
        id: 1,
        name: '科技频道',
        slug: 'tech-channel',
        description: '最新科技资讯和评测',
        owner: {name: '张三', avatar: ''},
        subscriberCount: 15600,
        mediaCount: 89,
        category: '科技',
        status: 'verified',
        createdAt: '2024-01-15'
    },
    {
        id: 2,
        name: '音乐现场',
        slug: 'music-live',
        description: '高质量音乐现场视频',
        owner: {name: '李四', avatar: ''},
        subscriberCount: 8900,
        mediaCount: 45,
        category: '音乐',
        status: 'verified',
        createdAt: '2024-02-20'
    },
    {
        id: 3,
        name: '游戏实况',
        slug: 'gaming-live',
        description: '游戏直播和实况录像',
        owner: {name: '王五', avatar: ''},
        subscriberCount: 23400,
        mediaCount: 156,
        category: '游戏',
        status: 'active',
        createdAt: '2024-03-10'
    },
    {
        id: 4,
        name: '教育课堂',
        slug: 'edu-class',
        description: '中小学在线教育',
        owner: {name: '赵六', avatar: ''},
        subscriberCount: 5600,
        mediaCount: 234,
        category: '教育',
        status: 'active',
        createdAt: '2024-04-05'
    },
    {
        id: 5,
        name: '影视解说',
        slug: 'movie-talk',
        description: '电影电视剧解说',
        owner: {name: '钱七', avatar: ''},
        subscriberCount: 12300,
        mediaCount: 67,
        category: '娱乐',
        status: 'pending',
        createdAt: '2024-05-12'
    },
];

const Channels: React.FC = () => {
    const {t} = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [channels] = useState(mockChannels);

    const filteredChannels = channels.filter(channel => {
        const matchesSearch = channel.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            channel.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || channel.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const totalSubscribers = channels.reduce((sum, c) => sum + c.subscriberCount, 0);
    const verifiedCount = channels.filter(c => c.status === 'verified').length;
    const pendingCount = channels.filter(c => c.status === 'pending').length;

    const getStatusBadge = (status: string) => {
        const variants: Record<string, 'default' | 'secondary' | 'destructive' | 'outline'> = {
            verified: 'default',
            active: 'secondary',
            pending: 'outline',
            banned: 'destructive',
        };
        const labels: Record<string, string> = {
            verified: t('common.verified'),
            active: t('admin.normal'),
            pending: t('admin.pending'),
            banned: t('admin.banned'),
        };
        return <Badge variant={variants[status] || 'outline'}>{labels[status] || status}</Badge>;
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
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.channels')}</h2>
                                <p className="text-sm text-slate-500 dark:text-slate-400 mt-1.5">
                                    Manage your content channels
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
                                        placeholder={t('admin.search') || t('admin.channels') + '...'}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-9 w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={statusFilter} onValueChange={setStatusFilter}>
                                    <SelectTrigger className="w-[140px] h-9 focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            {statusFilter === 'all' ? (
                                                <span className="text-muted-foreground">Status</span>
                                            ) : (
                                                <SelectValue placeholder="Status"/>
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all" className="justify-center text-center font-medium opacity-70">--- All ---</SelectItem>
                                        <SelectItem value="verified">{t('common.verified')}</SelectItem>
                                        <SelectItem value="active">{t('admin.normal')}</SelectItem>
                                        <SelectItem value="pending">{t('admin.pending')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-9 px-3"
                                        onClick={() => {
                                            setSearchTerm('');
                                            setStatusFilter('all');
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
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.channelTotal')}</p>
                                <p className="text-2xl font-bold text-blue-600">{channels.length}</p>
                            </div>
                            <div className="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
                                <Users className="w-6 h-6 text-blue-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-blue-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.totalSubscribers')}</p>
                                <p className="text-2xl font-bold text-purple-600">{formatNumber(totalSubscribers)}</p>
                            </div>
                            <div className="w-12 h-12 bg-purple-100 rounded-xl flex items-center justify-center">
                                <UserPlus className="w-6 h-6 text-purple-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-purple-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.verifiedChannels')}</p>
                                <p className="text-2xl font-bold text-green-600">{verifiedCount}</p>
                            </div>
                            <div className="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
                                <Eye className="w-6 h-6 text-green-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-green-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <p className="text-sm text-slate-500">{t('admin.pending')}</p>
                                <p className="text-2xl font-bold text-yellow-600">{pendingCount}</p>
                            </div>
                            <div className="w-12 h-12 bg-yellow-100 rounded-xl flex items-center justify-center">
                                <Loader2 className="w-6 h-6 text-yellow-600"/>
                            </div>
                        </div>
                        <div className="absolute bottom-0 left-0 h-1 bg-yellow-500 w-full opacity-10"/>
                    </CardContent>
                </Card>
            </div>

            {/* 频道表格 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.channelList')}</CardTitle>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button>
                                <Plus className="mr-2 h-4 w-4"/>
                                {t('admin.newChannel')}
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>{t('admin.channel')}</TableHead>
                                <TableHead>{t('admin.owner')}</TableHead>
                                <TableHead className="text-right">{t('admin.subscriberCount')}</TableHead>
                                <TableHead className="text-right">{t('admin.videoCount')}</TableHead>
                                <TableHead>{t('admin.category')}</TableHead>
                                <TableHead>{t('admin.status')}</TableHead>
                                <TableHead>{t('admin.createdAt')}</TableHead>
                                <TableHead className="text-right">{t('admin.actions')}</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredChannels.map((channel) => (
                                <TableRow key={channel.id}>
                                    <TableCell>
                                        <div className="flex items-center gap-3">
                                            <Avatar className="h-10 w-10">
                                                <AvatarFallback>{channel.name[0]}</AvatarFallback>
                                            </Avatar>
                                            <div>
                                                <div className="font-medium">{channel.name}</div>
                                                <div className="text-xs text-muted-foreground">{channel.slug}</div>
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            <Avatar className="h-6 w-6">
                                                <AvatarFallback
                                                    className="text-xs">{channel.owner.name[0]}</AvatarFallback>
                                            </Avatar>
                                            {channel.owner.name}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right font-medium">
                                        {formatNumber(channel.subscriberCount)}
                                    </TableCell>
                                    <TableCell className="text-right">{channel.mediaCount}</TableCell>
                                    <TableCell>
                                        <Badge variant="outline">{channel.category}</Badge>
                                    </TableCell>
                                    <TableCell>{getStatusBadge(channel.status)}</TableCell>
                                    <TableCell className="text-muted-foreground">{channel.createdAt}</TableCell>
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
                                                {channel.status === 'pending' && (
                                                    <DropdownMenuItem>
                                                        <UserPlus className="mr-2 h-4 w-4"/>
                                                        {t('admin.verify')}
                                                    </DropdownMenuItem>
                                                )}
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

export default Channels;
