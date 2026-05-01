/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Admin - Article Management Page
 */

import {useState, useEffect} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {Search, Plus, FileText, MoreVertical, Trash2, Edit, Eye, Filter, Film} from 'lucide-react';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {Badge} from '@/components/ui/badge';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
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
import {adminArticleApi, Article} from '@/lib/api/article';
import {formatDateTime} from '@/lib/format';
import {extractList} from '@/lib/extract';
import {TablePagination} from '@/components/common/TablePagination';
import {usePagination} from '@/hooks/usePagination';
import {toast} from 'sonner';

export default function ArticlePage() {
    const navigate = useNavigate();
    const [articles, setArticles] = useState<Article[]>([]);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const {page, pageSize, total, setPage, setTotal, getParams} = usePagination();

    // Load article data
    useEffect(() => {
        const loadArticles = async () => {
            setLoading(true);
            setError(null);
            try {
                const response = await adminArticleApi.adminList(getParams());
                const articleList = extractList<Article>(response);
                setArticles(articleList);
                if ((response as any)?.total !== undefined) {
                    setTotal((response as any).total);
                }
            } catch (err) {
                setError('Failed to load articles');
                console.error('Error loading articles:', err);
            } finally {
                setLoading(false);
            }
        };

        loadArticles();
    }, [page]);

    const filteredArticles = articles.filter(item => {
        const matchesSearch = item.title.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || item.state === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const formatViews = (count: number | undefined | null) => {
        if (count === undefined || count === null) return '0';
        if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
        return count.toString();
    };

    const handleDelete = async (id: string) => {
        try {
            await adminArticleApi.delete(id);
            setArticles(prev => prev.filter(a => a.id !== id));
            toast.success('Article deleted');
        } catch (err: any) {
            toast.error(`Delete failed: ${err?.message || 'Unknown error'}`);
        }
    };

    const stateBadgeVariant = (state: string): "default" | "secondary" | "destructive" | "outline" => {
        switch (state) {
            case 'published': return 'default';
            case 'draft': return 'secondary';
            case 'archived': return 'destructive';
            default: return 'outline';
        }
    };

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* Action bar */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* Page title */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">Article Management</h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage articles, pages, and static content
                                </p>
                            </div>
                        </div>

                        {/* Separator */}
                        <div className="border-t border-slate-200 dark:border-slate-800 my-2"/>

                        {/* Search and filter */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1 min-w-0">
                                <div className="relative w-full">
                                    <Search
                                        className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"/>
                                    <Input
                                        placeholder="Search articles..."
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="pl-10 h-8 rounded-btn-sm w-full focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                                    />
                                </div>
                            </div>
                            <div className="flex flex-wrap items-center gap-2">
                                <Select value={statusFilter} onValueChange={setStatusFilter}>
                                    <SelectTrigger className="w-[140px] h-8 rounded-btn-sm focus:ring-2 focus:ring-primary focus:border-transparent">
                                        <div className="flex items-center gap-2">
                                            <Filter className="h-4 w-4"/>
                                            <SelectValue placeholder="All Status"/>
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all">All Status</SelectItem>
                                        <SelectItem value="published">Published</SelectItem>
                                        <SelectItem value="draft">Draft</SelectItem>
                                        <SelectItem value="archived">Archived</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Article Table */}
            <Card>
                <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle>All Articles</CardTitle>
                            <CardDescription>Manage your articles and pages</CardDescription>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button size="sm" onClick={() => navigate({to: '/admin/articles/new'})}>
                                <Plus className="w-4 h-4 mr-2"/>
                                Create Article
                            </Button>
                        </div>
                    </div>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Title</TableHead>
                                <TableHead>Video</TableHead>
                                <TableHead>State</TableHead>
                                <TableHead>Views</TableHead>
                                <TableHead>Date</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {loading ? (
                                <TableRow key="loading">
                                    <TableCell colSpan={6} className="text-center py-8">
                                        <div className="animate-pulse">Loading articles...</div>
                                    </TableCell>
                                </TableRow>
                            ) : error ? (
                                <TableRow key="error">
                                    <TableCell colSpan={6} className="text-center py-8">
                                        <div className="text-destructive">{error}</div>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            className="mt-2"
                                            onClick={() => window.location.reload()}
                                        >
                                            Retry
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ) : filteredArticles.length === 0 ? (
                                <TableRow key="empty">
                                    <TableCell colSpan={6} className="text-center py-8">
                                        No articles found
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredArticles.map((item) => (
                                    <TableRow key={item.id}>
                                        <TableCell>
                                            <div className="flex items-center gap-3">
                                                <div className="w-10 h-10 bg-slate-100 dark:bg-slate-800 rounded-lg flex items-center justify-center">
                                                    <FileText className="w-5 h-5 text-slate-500"/>
                                                </div>
                                                <div>
                                                    <span className="font-medium">{item.title}</span>
                                                    {item.featured && (
                                                        <Badge variant="outline" className="ml-2 text-[10px] px-1 py-0 text-warning border-amber-300">
                                                            Featured
                                                        </Badge>
                                                    )}
                                                </div>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            {item.media_id ? (
                                                <Badge variant="outline" className="gap-1">
                                                    <Film className="w-3 h-3"/>
                                                    Video
                                                </Badge>
                                            ) : (
                                                <span className="text-muted-foreground text-sm">--</span>
                                            )}
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant={stateBadgeVariant(item.state)}>
                                                {item.state}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-sm text-slate-500">{formatViews(item.view_count)}</TableCell>
                                        <TableCell className="text-sm text-slate-500">{formatDateTime(item.create_time)}</TableCell>
                                        <TableCell className="text-right">
                                            <DropdownMenu>
                                                <DropdownMenuTrigger asChild>
                                                    <Button
                                                        variant="ghost"
                                                        size="icon"
                                                        className="h-6 w-6"
                                                        title="More Actions"
                                                    >
                                                        <MoreVertical className="h-3 w-3"/>
                                                    </Button>
                                                </DropdownMenuTrigger>
                                                <DropdownMenuContent align="end">
                                                    <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                                    <DropdownMenuSeparator/>
                                                    <DropdownMenuItem onClick={() => {
                                                        if (item.slug) {
                                                            window.open(`/articles/${item.slug}`, '_blank');
                                                        }
                                                    }}>
                                                        <Eye className="w-4 h-4 mr-2"/>
                                                        View
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem onClick={() => navigate({to: '/admin/articles/$id/edit', params: {id: item.id}})}>
                                                        <Edit className="w-4 h-4 mr-2"/>
                                                        Edit
                                                    </DropdownMenuItem>
                                                    <DropdownMenuItem className="text-destructive focus:text-destructive"
                                                        onClick={() => handleDelete(item.id)}>
                                                        <Trash2 className="w-4 h-4 mr-2"/>
                                                        Delete
                                                    </DropdownMenuItem>
                                                </DropdownMenuContent>
                                            </DropdownMenu>
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>
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
}
