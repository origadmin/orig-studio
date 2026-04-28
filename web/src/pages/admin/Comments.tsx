import {Spinner} from "@/components/ui/spinner"
import React, {useState, useEffect} from 'react';
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
    DropdownMenuLabel,
    DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {MoreHorizontal, Search, Eye, Trash2, MessageCircle, ThumbsUp, Flag, Ban, Loader2, Filter, RotateCcw} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {adminCommentApi} from '@/lib/api/comment';
import ErrorPage from '@/components/error/ErrorPage';
import {TablePagination} from '@/components/common/TablePagination';

interface Comment {
    id: number;
    user: { name: string; avatar: string; username: string };
    media: { title: string; id: number };
    content: string;
    likes: number;
    replies: number;
    status: string;
    isSpam: boolean;
    createdAt: string;
}

const Comments: React.FC = () => {
    const {t} = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [comments, setComments] = useState<Comment[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [pageSize] = useState(20);
    const [total, setTotal] = useState(0);

    // Fetch comments from API
    useEffect(() => {
        const fetchComments = async () => {
            try {
                setLoading(true);
                const response = await adminCommentApi.list({page, page_size: pageSize});
                const commentList = Array.isArray((response as any)?.items) ? (response as any).items : [];
                if ((response as any)?.total !== undefined) {
                    setTotal((response as any).total);
                }
                const mappedComments = commentList.map((comment: any) => ({
                    id: comment.id,
                    user: {
                        name: comment.username || 'Unknown User',
                        avatar: comment.avatar || '',
                        username: comment.username || 'unknown'
                    },
                    media: {
                        title: comment.media?.title || 'Unknown Media',
                        id: comment.media_id || 0
                    },
                    content: comment.content || comment.body || comment.text || '',
                    likes: comment.like_count || 0,
                    replies: comment.reply_count || 0,
                    status: comment.status || 'approved',
                    isSpam: comment.is_spam || false,
                    createdAt: comment.create_time || comment.created_at || new Date().toISOString()
                }));
                setComments(mappedComments);
            } catch (err: any) {
                // 捕获错误并显示友好的消息
                setError(err.message || t('common.error'));
                console.error('Failed to fetch comments:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchComments();
    }, [t, page]);

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
        if (isSpam) return <Badge variant="secondary">{t('admin.spam')}</Badge>;
        const variants: Record<string, 'default' | 'secondary' | 'outline'> = {
            approved: 'default',
            pending: 'outline',
            reported: 'secondary',
        };
        const labels: Record<string, string> = {
            approved: t('admin.approved'),
            pending: t('admin.pending'),
            reported: t('admin.reported'),
        };
        return <Badge variant={variants[status] || 'outline'}>{labels[status] || status}</Badge>;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    // 不再返回错误页面，而是在页面内部显示错误消息

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* 操作栏 */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* 页面标题 */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">{t('admin.comments')}</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage your user comments
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
                                        placeholder={t('admin.search') || t('admin.comments') + '...'}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={statusFilter} onValueChange={setStatusFilter}>
                                    <SelectTrigger className="w-[140px] h-8 rounded-btn-sm focus-visible:ring-1 focus-visible:ring-ring focus-visible:ring-offset-0">
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
                                        <SelectItem value="approved">{t('admin.approved')}</SelectItem>
                                        <SelectItem value="pending">{t('admin.pending')}</SelectItem>
                                        <SelectItem value="reported">{t('admin.reported')}</SelectItem>
                                    </SelectContent>
                                </Select>
                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
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
                                        onClick={() => {
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

            {/* 错误消息 */}
            {error && (
                <Card className="bg-red-50 dark:bg-red-950/30 border-red-200 dark:border-red-800">
                    <CardContent className="p-4">
                        <div className="flex items-center gap-2">
                            <Ban className="h-5 w-5 text-destructive dark:text-red-400"/>
                            <p className="text-red-700 dark:text-red-300">{error}</p>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="flex items-center gap-2">
                            <MessageCircle className="h-5 w-5 text-info"/>
                            <div>
                                <div className="text-2xl font-bold text-info dark:text-blue-400">{totalComments}</div>
                                <p className="text-sm text-muted-foreground">{t('admin.totalComments')}</p>
                            </div>
                        </div>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-info w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-success dark:text-green-400">{approvedCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.approved')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-success w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{pendingCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.pending')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-warning w-full opacity-10"/>
                </Card>
                <Card className="relative overflow-hidden shadow-sm border-none ring-1 ring-slate-200 dark:ring-slate-800">
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-destructive dark:text-red-400">{reportedCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.spam')}</p>
                    </CardContent>
                    <div className="absolute bottom-0 left-0 h-1 bg-destructive w-full opacity-10"/>
                </Card>
            </div>

            {/* 评论表格 */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>{t('admin.commentList')}</CardTitle>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    {loading ? (
                        <div className="flex items-center justify-center min-h-[400px]">
                            <Spinner />
                        </div>
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>ID</TableHead>
                                    <TableHead>{t('admin.user')}</TableHead>
                                    <TableHead>{t('admin.commentContent')}</TableHead>
                                    <TableHead>{t('admin.belongVideo')}</TableHead>
                                    <TableHead className="text-center">{t('admin.likes')}</TableHead>
                                    <TableHead className="text-center">{t('admin.replies')}</TableHead>
                                    <TableHead>{t('admin.status')}</TableHead>
                                    <TableHead>{t('admin.publishTime')}</TableHead>
                                    <TableHead className="text-right">{t('admin.actions')}</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {filteredComments.length > 0 ? (
                                    filteredComments.map((comment) => (
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
                                                        <Button 
                                                            variant="ghost" 
                                                            size="icon" 
                                                            className="h-6 w-6" 
                                                            title="More Actions"
                                                        >
                                                            <MoreHorizontal className="h-3 w-3"/>
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                        <DropdownMenuSeparator/>
                                                        <DropdownMenuItem>
                                                            <Eye className="mr-2 h-4 w-4"/>
                                                            {t('admin.view')}
                                                        </DropdownMenuItem>
                                                        {comment.status === 'pending' && (
                                                            <>
                                                                <DropdownMenuItem>
                                                                    <MessageCircle className="mr-2 h-4 w-4"/>
                                                                    {t('admin.approve')}
                                                                </DropdownMenuItem>
                                                                <DropdownMenuItem className="text-destructive focus:text-destructive">
                                                                    <Ban className="mr-2 h-4 w-4"/>
                                                                    {t('admin.reject')}
                                                                </DropdownMenuItem>
                                                            </>
                                                        )}
                                                        {comment.isSpam && (
                                                            <DropdownMenuItem>
                                                                <Ban className="mr-2 h-4 w-4"/>
                                                                {t('admin.banUser')}
                                                            </DropdownMenuItem>
                                                        )}
                                                        <DropdownMenuItem className="text-destructive focus:text-destructive">
                                                            <Trash2 className="mr-2 h-4 w-4"/>
                                                            {t('admin.delete')}
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                ) : (
                                    <TableRow key="empty">
                                        <TableCell colSpan={9} className="text-center py-8">
                                            <p className="text-muted-foreground">{t('admin.noComments') || 'No comments found'}</p>
                                        </TableCell>
                                    </TableRow>
                                )}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>

            <TablePagination
                page={page}
                pageSize={pageSize}
                total={total}
                onPageChange={setPage}
            />
        </div>
    );
};

export default Comments;
