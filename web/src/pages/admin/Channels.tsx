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
import {MoreHorizontal, Plus, Search, Edit, Trash2, Eye, UserPlus, Users} from 'lucide-react';

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

    const formatNumber = (num: number) => {
        if (num >= 10000) return (num / 10000).toFixed(1) + t('common.wan');
        return num.toString();
    };

    return (
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <Users className="h-5 w-5 text-blue-600"/>
                            <div>
                                <div className="text-2xl font-bold">{channels.length}</div>
                                <p className="text-sm text-muted-foreground">{t('admin.channelTotal')}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold">{formatNumber(totalSubscribers)}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.totalSubscribers')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{verifiedCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.verifiedChannels')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600">{pendingCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.pending')}</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="flex gap-2 flex-1">
                    <div className="relative flex-1 max-w-md">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder={t('admin.search') || t('admin.channels') + '...'}
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="pl-10"
                        />
                    </div>
                    <select
                        className="px-3 py-2 border rounded-md bg-background"
                        value={statusFilter}
                        onChange={(e) => setStatusFilter(e.target.value)}
                    >
                        <option value="all">{t('admin.allStatus')}</option>
                        <option value="verified">{t('common.verified')}</option>
                        <option value="active">{t('admin.normal')}</option>
                        <option value="pending">{t('admin.pending')}</option>
                    </select>
                </div>
                <Button>
                    <Plus className="mr-2 h-4 w-4"/>
                    {t('admin.newChannel')}
                </Button>
            </div>

            {/* 频道表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>{t('admin.channelList')}</CardTitle>
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
