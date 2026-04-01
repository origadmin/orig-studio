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
import {MoreHorizontal, Search, Eye, Trash2, MessageCircle, ThumbsUp, Flag, Ban} from 'lucide-react';

// 模拟数据
const mockComments = [
    {
        id: 1,
        user: {name: '用户A', avatar: '', username: 'user_a'},
        media: {title: 'Python 入门教程', id: 1},
        content: '讲解得非常清楚，感谢up主！',
        likes: 156,
        replies: 23,
        status: 'approved',
        isSpam: false,
        createdAt: '2024-05-20 14:30'
    },
    {
        id: 2,
        user: {name: '用户B', avatar: '', username: 'user_b'},
        media: {title: 'React 高级教程', id: 2},
        content: '这个视频太好了，收藏了！',
        likes: 89,
        replies: 12,
        status: 'approved',
        isSpam: false,
        createdAt: '2024-05-20 13:15'
    },
    {
        id: 3,
        user: {name: '用户C', avatar: '', username: 'user_c'},
        media: {title: '机器学习实战', id: 3},
        content: '请问这个代码在哪里下载？',
        likes: 34,
        replies: 5,
        status: 'pending',
        isSpam: false,
        createdAt: '2024-05-20 12:00'
    },
    {
        id: 4,
        user: {name: '垃圾用户', avatar: '', username: 'spam_user'},
        media: {title: 'Go 语言教程', id: 4},
        content: '点击这里获取免费礼物！http://spam.com',
        likes: 0,
        replies: 0,
        status: 'reported',
        isSpam: true,
        createdAt: '2024-05-20 11:45'
    },
    {
        id: 5,
        user: {name: '用户D', avatar: '', username: 'user_d'},
        media: {title: 'Docker 教程', id: 5},
        content: '终于找到讲得这么细的了',
        likes: 67,
        replies: 8,
        status: 'approved',
        isSpam: false,
        createdAt: '2024-05-20 10:30'
    },
    {
        id: 6,
        user: {name: '用户E', avatar: '', username: 'user_e'},
        media: {title: 'Kubernetes 入门', id: 6},
        content: '支持up主，期待更多内容',
        likes: 45,
        replies: 3,
        status: 'approved',
        isSpam: false,
        createdAt: '2024-05-20 09:15'
    },
];

const Comments: React.FC = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [comments] = useState(mockComments);

    const filteredComments = comments.filter(comment => {
        const matchesSearch = comment.content.toLowerCase().includes(searchTerm.toLowerCase()) ||
            comment.user.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            comment.media.title.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || comment.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const totalComments = comments.length;
    const approvedCount = comments.filter(c => c.status === 'approved').length;
    const pendingCount = comments.filter(c => c.status === 'pending').length;
    const reportedCount = comments.filter(c => c.status === 'reported' || c.isSpam).length;

    const getStatusBadge = (status: string, isSpam: boolean) => {
        if (isSpam) return <Badge variant="destructive">垃圾信息</Badge>;
        const variants: Record<string, 'default' | 'secondary' | 'outline'> = {
            approved: 'default',
            pending: 'outline',
            reported: 'destructive',
        };
        const labels: Record<string, string> = {
            approved: '已通过',
            pending: '待审核',
            reported: '已举报',
        };
        return <Badge variant={variants[status] || 'outline'}>{labels[status] || status}</Badge>;
    };

    return (
        <div className="space-y-6">
            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <MessageCircle className="h-5 w-5 text-blue-600"/>
                            <div>
                                <div className="text-2xl font-bold">{totalComments}</div>
                                <p className="text-sm text-muted-foreground">评论总数</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{approvedCount}</div>
                        <p className="text-sm text-muted-foreground">已通过</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600">{pendingCount}</div>
                        <p className="text-sm text-muted-foreground">待审核</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-red-600">{reportedCount}</div>
                        <p className="text-sm text-muted-foreground">举报/垃圾</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="flex gap-2 flex-1">
                    <div className="relative flex-1 max-w-md">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder="搜索评论..."
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
                        <option value="all">全部状态</option>
                        <option value="approved">已通过</option>
                        <option value="pending">待审核</option>
                        <option value="reported">已举报</option>
                    </select>
                </div>
            </div>

            {/* 评论表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>评论列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>ID</TableHead>
                                <TableHead>用户</TableHead>
                                <TableHead>评论内容</TableHead>
                                <TableHead>所属视频</TableHead>
                                <TableHead className="text-center">点赞</TableHead>
                                <TableHead className="text-center">回复</TableHead>
                                <TableHead>状态</TableHead>
                                <TableHead>发布时间</TableHead>
                                <TableHead className="text-right">操作</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredComments.map((comment) => (
                                <TableRow key={comment.id}>
                                    <TableCell className="font-medium">{comment.id}</TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            <Avatar className="h-6 w-6">
                                                <AvatarFallback
                                                    className="text-xs">{comment.user.name[0]}</AvatarFallback>
                                            </Avatar>
                                            <div>
                                                <div className="text-sm font-medium">{comment.user.name}</div>
                                                <div
                                                    className="text-xs text-muted-foreground">@{comment.user.username}</div>
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell className="max-w-[250px]">
                                        <p className="truncate">{comment.content}</p>
                                    </TableCell>
                                    <TableCell>
                                        <span className="text-sm">{comment.media.title}</span>
                                    </TableCell>
                                    <TableCell className="text-center">
                                        <div className="flex items-center justify-center gap-1">
                                            <ThumbsUp className="h-3 w-3 text-muted-foreground"/>
                                            {comment.likes}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-center">
                                        <div className="flex items-center justify-center gap-1">
                                            <MessageCircle className="h-3 w-3 text-muted-foreground"/>
                                            {comment.replies}
                                        </div>
                                    </TableCell>
                                    <TableCell>{getStatusBadge(comment.status, comment.isSpam)}</TableCell>
                                    <TableCell className="text-muted-foreground text-sm">{comment.createdAt}</TableCell>
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
                                                {comment.status === 'pending' && (
                                                    <>
                                                        <DropdownMenuItem>
                                                            <MessageCircle className="mr-2 h-4 w-4"/>
                                                            通过
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem className="text-red-600">
                                                            <Ban className="mr-2 h-4 w-4"/>
                                                            拒绝
                                                        </DropdownMenuItem>
                                                    </>
                                                )}
                                                {comment.isSpam && (
                                                    <DropdownMenuItem>
                                                        <Ban className="mr-2 h-4 w-4"/>
                                                        封禁用户
                                                    </DropdownMenuItem>
                                                )}
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

export default Comments;
