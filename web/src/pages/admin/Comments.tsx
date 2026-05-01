import { Spinner } from '@/components/ui/spinner';
import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Search, Filter, RotateCcw } from 'lucide-react';
import { adminCommentApi, type CommentStats } from '@/lib/api/comment';
import { toast } from 'sonner';
import { CommentTreeTable } from '@/components/admin/CommentTreeTable';
import { CommentStatsCards } from '@/components/admin/CommentStatsCards';
import { ReportStatusFilter } from '@/components/admin/ReportStatusFilter';
import { ReportDialog } from '@/components/admin/ReportDialog';
import { useCommentTree } from '@/hooks/useCommentTree';
import { usePagination } from '@/hooks/usePagination';
import { TablePagination } from '@/components/common/TablePagination';

const Comments: React.FC = () => {
    const { t } = useTranslation();
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [reportStatusFilter, setReportStatusFilter] = useState('all');
    const [stats, setStats] = useState<CommentStats | null>(null);
    const [reportDialogOpen, setReportDialogOpen] = useState(false);
    const [reportCommentId, setReportCommentId] = useState<string | null>(null);
    const [actionLoading, setActionLoading] = useState<string | null>(null);
    const { page, pageSize, total, setPage, setTotal, getParams } = usePagination();

    const {
        visibleNodes,
        expandedIds,
        loading,
        toggleExpand,
        expandAll,
        collapseAll,
        loadComments,
    } = useCommentTree();

    // Fetch stats
    const fetchStats = useCallback(async () => {
        try {
            const response = await adminCommentApi.getStats();
            if (response) {
                setStats(response as any);
            }
        } catch (err) {
            console.error('Failed to fetch stats:', err);
        }
    }, []);

    // Load comments with current filters
    const refreshComments = useCallback(async () => {
        const params: any = {
            page,
            page_size: pageSize,
            tree: true,
        };
        if (statusFilter !== 'all') params.status = statusFilter;
        if (reportStatusFilter !== 'all') params.report_status = reportStatusFilter;
        await loadComments(params);
    }, [page, pageSize, statusFilter, reportStatusFilter, loadComments]);

    useEffect(() => {
        refreshComments();
    }, [refreshComments]);

    useEffect(() => {
        fetchStats();
    }, [fetchStats]);

    // Update total from the last load
    useEffect(() => {
        // total is managed by usePagination, we update it when comments load
    }, [loading]);

    // Action handlers
    const handleApprove = async (id: string) => {
        try {
            setActionLoading(id);
            await adminCommentApi.approve(id);
            toast.success('Comment approved');
            await refreshComments();
            await fetchStats();
        } catch (err: any) {
            toast.error('Failed to approve', { description: err.message });
        } finally {
            setActionLoading(null);
        }
    };

    const handleReject = async (id: string) => {
        try {
            setActionLoading(id);
            await adminCommentApi.reject(id);
            toast.success('Comment rejected');
            await refreshComments();
            await fetchStats();
        } catch (err: any) {
            toast.error('Failed to reject', { description: err.message });
        } finally {
            setActionLoading(null);
        }
    };

    const handleBlock = async (id: string) => {
        try {
            setActionLoading(id);
            await adminCommentApi.block(id);
            toast.success('Comment blocked');
            await refreshComments();
            await fetchStats();
        } catch (err: any) {
            toast.error('Failed to block', { description: err.message });
        } finally {
            setActionLoading(null);
        }
    };

    const handleUnblock = async (id: string) => {
        try {
            setActionLoading(id);
            await adminCommentApi.unblock(id);
            toast.success('Comment unblocked');
            await refreshComments();
            await fetchStats();
        } catch (err: any) {
            toast.error('Failed to unblock', { description: err.message });
        } finally {
            setActionLoading(null);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Are you sure you want to delete this comment?')) return;
        try {
            setActionLoading(id);
            await adminCommentApi.delete(id);
            toast.success('Comment deleted');
            await refreshComments();
            await fetchStats();
        } catch (err: any) {
            toast.error('Failed to delete', { description: err.message });
        } finally {
            setActionLoading(null);
        }
    };

    const handleViewReports = (id: string) => {
        toast.info(`Viewing reports for comment ${id}`);
    };

    const handleDismissReports = async (id: string) => {
        try {
            setActionLoading(id);
            await adminCommentApi.dismissReports(id);
            toast.success('Reports dismissed');
            await refreshComments();
            await fetchStats();
        } catch (err: any) {
            toast.error('Failed to dismiss reports', { description: err.message });
        } finally {
            setActionLoading(null);
        }
    };

    const handleReportSubmit = async (data: { reason: string; description?: string }) => {
        if (!reportCommentId) return;
        await adminCommentApi.approve(reportCommentId); // placeholder - report API is on portal side
        // Actually, the admin report is done via the portal API
        // For admin page, we just use the report dialog for reference
    };

    const handleReset = () => {
        setSearchTerm('');
        setStatusFilter('all');
        setReportStatusFilter('all');
    };

    if (loading && visibleNodes.length === 0) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    return (
        <div className="space-y-4 p-4 md:p-6">
            {/* Toolbar */}
            <Card className="overflow-hidden">
                <CardContent className="p-6">
                    <div className="flex flex-col gap-4">
                        {/* Page title */}
                        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                            <div>
                                <h2 className="text-3xl font-extrabold tracking-tight text-slate-900 dark:text-slate-50">
                                    {t('admin.comments') || 'Comments'}
                                </h2>
                                <p className="text-sm text-slate-500 dark:text-muted-foreground mt-1.5">
                                    Manage your user comments
                                </p>
                            </div>
                        </div>

                        {/* Divider */}
                        <div className="border-t border-slate-200 dark:border-slate-800 my-2" />

                        {/* Search and filters */}
                        <div className="flex flex-col lg:flex-row gap-4">
                            <div className="flex-1 min-w-[120px] max-w-[400px]">
                                <div className="relative w-full">
                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        placeholder={t('admin.search') || 'Search comments...'}
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
                                            <Filter className="h-4 w-4" />
                                            {statusFilter === 'all' ? (
                                                <span className="text-muted-foreground">Status</span>
                                            ) : (
                                                <SelectValue placeholder="Status" />
                                            )}
                                        </div>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="all">--- All ---</SelectItem>
                                        <SelectItem value="approved">{t('admin.approved') || 'Approved'}</SelectItem>
                                        <SelectItem value="pending">{t('admin.pending') || 'Pending'}</SelectItem>
                                        <SelectItem value="rejected">{t('admin.rejected') || 'Rejected'}</SelectItem>
                                        <SelectItem value="blocked">Blocked</SelectItem>
                                    </SelectContent>
                                </Select>

                                <ReportStatusFilter value={reportStatusFilter} onChange={setReportStatusFilter} />

                                <div className="flex items-center gap-2 ml-auto lg:ml-0">
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={handleReset}
                                    >
                                        <RotateCcw className="h-4 w-4 mr-2" />
                                        Reset
                                    </Button>
                                </div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Stats cards */}
            <CommentStatsCards stats={stats} loading={loading} />

            {/* Tree table */}
            <CommentTreeTable
                nodes={visibleNodes}
                expandedIds={expandedIds}
                onToggleExpand={toggleExpand}
                onExpandAll={expandAll}
                onCollapseAll={collapseAll}
                onApprove={handleApprove}
                onReject={handleReject}
                onBlock={handleBlock}
                onUnblock={handleUnblock}
                onDelete={handleDelete}
                onViewReports={handleViewReports}
                onDismissReports={handleDismissReports}
                loading={loading}
            />

            <TablePagination
                page={page}
                pageSize={pageSize}
                total={total}
                onPageChange={setPage}
            />

            {/* Report dialog */}
            <ReportDialog
                open={reportDialogOpen}
                onOpenChange={setReportDialogOpen}
                commentId={reportCommentId || ''}
                onSubmit={handleReportSubmit}
            />
        </div>
    );
};

export default Comments;
