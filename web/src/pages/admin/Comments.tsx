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
} from '@/components/ui/dropdown-menu';
import {MoreHorizontal, Search, Eye, Trash2, MessageCircle, ThumbsUp, Flag, Ban, Loader2} from 'lucide-react';
import {commentApi} from '@/lib/api/comment';

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

    // Fetch comments from API
    useEffect(() => {
        const fetchComments = async () => {
            try {
                setLoading(true);
                const response = await commentApi.getAll({});
                // Map API response to our comment interface
                const mappedComments = (response || []).map((comment: any) => ({
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
                    content: comment.body || comment.text || '',
                    likes: comment.like_count || 0,
                    replies: comment.reply_count || 0,
                    status: comment.status || 'approved',
                    isSpam: comment.is_spam || false,
                    createdAt: comment.created_at || new Date().toISOString()
                }));
                setComments(mappedComments);
            } catch (err) {
                setError(t('common.error'));
                console.error('Failed to fetch comments:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchComments();
    }, [t]);

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
        if (isSpam) return <Badge variant="destructive">{t('admin.spam')}</Badge>;
        const variants: Record<string, 'default' | 'secondary' | 'outline'> = {
            approved: 'default',
            pending: 'outline',
            reported: 'destructive',
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
                <div className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-20 text-gray-400">
                <p className="text-lg mb-1">{t('common.loading')}</p>
                <p className="text-sm">{error}</p>
            </div>
        );
    }

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
                                <p className="text-sm text-muted-foreground">{t('admin.totalComments')}</p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-green-600">{approvedCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.approved')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-yellow-600">{pendingCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.pending')}</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardContent className="pt-6">
                        <div className="text-2xl font-bold text-red-600">{reportedCount}</div>
                        <p className="text-sm text-muted-foreground">{t('admin.spam')}</p>
                    </CardContent>
                </Card>
            </div>

            {/* 操作栏 */}
            <div className="flex flex-col sm:flex-row gap-4 justify-between">
                <div className="flex gap-2 flex-1">
                    <div className="relative flex-1 max-w-md">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                        <Input
                            placeholder={t('admin.search') || t('admin.comments') + '...'}
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
                        <option value="approved">{t('admin.approved')}</option>
                        <option value="pending">{t('admin.pending')}</option>
                        <option value="reported">{t('admin.reported')}</option>
                    </select>
                </div>
            </div>

            {/* 评论表格 */}
            <Card>
                <CardHeader>
                    <CardTitle>{t('admin.commentList')}</CardTitle>
                </CardHeader>
                <CardContent>
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
                                                    {t('admin.view')}
                                                </DropdownMenuItem>
                                                {comment.status === 'pending' && (
                                                    <>
                                                        <DropdownMenuItem>
                                                            <MessageCircle className="mr-2 h-4 w-4"/>
                                                            {t('admin.approve')}
                                                        </DropdownMenuItem>
                                                        <DropdownMenuItem className="text-red-600">
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

export default Comments;
