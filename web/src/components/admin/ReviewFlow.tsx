import React, {useState, useEffect} from 'react';
import {Check, X, Clock, User, File, Search, Filter, Trash2} from 'lucide-react';
import {Card, CardContent, CardHeader, CardTitle, CardDescription} from '@/components/ui/card';
import {Button} from '@/components/ui/button';
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from '@/components/ui/table';
import {Input} from '@/components/ui/input';
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from '@/components/ui/select';
import {Skeleton} from '@/components/ui/skeleton';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger
} from '@/components/ui/dialog';
import {Textarea} from '@/components/ui/textarea';
import {Badge} from '@/components/ui/badge';
import {formatDate, formatViews} from '@/lib/format';
import {useTranslation} from 'react-i18next';
import {reviewApi, type ReviewItem} from '@/lib/api/review';
import ErrorPage from '@/components/common/ErrorPage';

const ReviewFlow: React.FC = () => {
    const {t} = useTranslation();
    const [activeTab, setActiveTab] = useState<'pending' | 'history'>('pending');
    const [reviewItems, setReviewItems] = useState<ReviewItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(10);
    const [total, setTotal] = useState(0);
    const [search, setSearch] = useState('');
    const [typeFilter, setTypeFilter] = useState('');
    const [statusFilter, setStatusFilter] = useState('');
    const [selectedItems, setSelectedItems] = useState<string[]>([]);
    const [showBatchDialog, setShowBatchDialog] = useState(false);
    const [batchStatus, setBatchStatus] = useState<'approved' | 'rejected'>('approved');
    const [batchReason, setBatchReason] = useState('');
    const [showSingleDialog, setShowSingleDialog] = useState(false);
    const [currentItem, setCurrentItem] = useState<ReviewItem | null>(null);
    const [singleStatus, setSingleStatus] = useState<'approved' | 'rejected'>('approved');
    const [singleReason, setSingleReason] = useState('');

    useEffect(() => {
        fetchReviewItems();
    }, [activeTab, page, pageSize, search, typeFilter, statusFilter]);

    const fetchReviewItems = async () => {
        try {
            setLoading(true);
            setError(null);
            let response;
            if (activeTab === 'pending') {
                response = await reviewApi.getPending({
                    page,
                    page_size: pageSize,
                    type: typeFilter,
                });
            } else {
                response = await reviewApi.getHistory({
                    page,
                    page_size: pageSize,
                    type: typeFilter,
                    status: statusFilter,
                });
            }
            setReviewItems(response.list || []);
            setTotal(response.total || 0);
        } catch (err) {
            setError('Failed to fetch review items');
            console.error('Failed to fetch review items:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleReview = async (id: string, status: string, reason?: string) => {
        try {
            await reviewApi.review(id, {status, reason});
            // 重新获取审核列表
            await fetchReviewItems();
        } catch (err) {
            console.error('Failed to review item:', err);
        }
    };

    const handleBatchReview = async () => {
        if (selectedItems.length === 0) return;

        try {
            await reviewApi.batchReview({
                ids: selectedItems,
                status: batchStatus,
                reason: batchReason,
            });
            // 重新获取审核列表
            await fetchReviewItems();
            setShowBatchDialog(false);
            setSelectedItems([]);
            setBatchStatus('approved');
            setBatchReason('');
        } catch (err) {
            console.error('Failed to batch review items:', err);
        }
    };

    const handleSelectItem = (id: string) => {
        setSelectedItems(prev => {
            if (prev.includes(id)) {
                return prev.filter(itemId => itemId !== id);
            } else {
                return [...prev, id];
            }
        });
    };

    const handleSelectAll = () => {
        if (selectedItems.length === reviewItems.length) {
            setSelectedItems([]);
        } else {
            setSelectedItems(reviewItems.map(item => item.id));
        }
    };

    const openSingleDialog = (item: ReviewItem) => {
        setCurrentItem(item);
        setSingleStatus('approved');
        setSingleReason('');
        setShowSingleDialog(true);
    };

    const handleSingleReview = () => {
        if (!currentItem) return;
        handleReview(currentItem.id, singleStatus, singleReason);
        setShowSingleDialog(false);
        setCurrentItem(null);
    };

    const getStatusBadge = (status: string) => {
        switch (status) {
            case 'pending':
                return <Badge variant="secondary"
                              className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">{t('review.pending')}</Badge>;
            case 'approved':
                return <Badge variant="secondary"
                              className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">{t('review.approved')}</Badge>;
            case 'rejected':
                return <Badge variant="secondary"
                              className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">{t('review.rejected')}</Badge>;
            default:
                return <Badge variant="secondary">{status}</Badge>;
        }
    };

    if (loading) {
        return (
            <div className="space-y-4">
                <Card>
                    <CardHeader>
                        <CardTitle>Review Flow</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <Skeleton className="h-8 w-32"/>
                                    <Skeleton className="h-8 w-32"/>
                                </div>
                                <div className="flex items-center gap-2">
                                    <Skeleton className="h-8 w-40"/>
                                    <Skeleton className="h-8 w-40"/>
                                </div>
                            </div>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="w-[40px]"></TableHead>
                                        <TableHead>Media</TableHead>
                                        <TableHead>User</TableHead>
                                        <TableHead>Status</TableHead>
                                        <TableHead>Created</TableHead>
                                        <TableHead>Actions</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {Array.from({length: 5}).map((_, i) => (
                                        <TableRow key={i}>
                                            <TableCell>
                                                <Skeleton className="h-4 w-4"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-64"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-32"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-24"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-4 w-32"/>
                                            </TableCell>
                                            <TableCell>
                                                <Skeleton className="h-8 w-32"/>
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    </CardContent>
                </Card>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    return (
        <div className="space-y-6">
            {/* Tabs */}
            <div className="flex items-center gap-2">
                <Button
                    variant={activeTab === 'pending' ? 'default' : 'outline'}
                    onClick={() => setActiveTab('pending')}
                >
                    <Clock className="w-4 h-4 mr-2"/>
                    {t('review.pending')}
                </Button>
                <Button
                    variant={activeTab === 'history' ? 'default' : 'outline'}
                    onClick={() => setActiveTab('history')}
                >
                    <Check className="w-4 h-4 mr-2"/>
                    {t('review.history')}
                </Button>
            </div>

            {/* Filter and Search */}
            <Card>
                <CardContent className="pt-6">
                    <div className="flex flex-wrap items-center gap-4">
                        <div className="flex-1">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"/>
                                <Input
                                    placeholder={t('common.search')}
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                        <div className="w-40">
                            <Select value={typeFilter} onValueChange={setTypeFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder={t('common.type')}/>
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="">{t('common.all')}</SelectItem>
                                    <SelectItem value="video">Video</SelectItem>
                                    <SelectItem value="image">Image</SelectItem>
                                    <SelectItem value="audio">Audio</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        {activeTab === 'history' && (
                            <div className="w-40">
                                <Select value={statusFilter} onValueChange={setStatusFilter}>
                                    <SelectTrigger>
                                        <SelectValue placeholder={t('common.status')}/>
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="">{t('common.all')}</SelectItem>
                                        <SelectItem value="approved">{t('review.approved')}</SelectItem>
                                        <SelectItem value="rejected">{t('review.rejected')}</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>
                        )}
                    </div>
                </CardContent>
            </Card>

            {/* Batch Actions */}
            {selectedItems.length > 0 && (
                <Card>
                    <CardContent className="flex items-center justify-between">
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                            {t('review.selectedItems', {count: selectedItems.length})}
                        </p>
                        <div className="flex items-center gap-2">
                            <Button variant="outline" onClick={() => setSelectedItems([])}>
                                <Trash2 className="w-4 h-4 mr-2"/>
                                {t('common.clear')}
                            </Button>
                            <Button onClick={() => setShowBatchDialog(true)}>
                                {t('review.batchReview')}
                            </Button>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Review List */}
            <Card>
                <CardHeader>
                    <CardTitle>{activeTab === 'pending' ? t('review.pendingItems') : t('review.historyItems')}</CardTitle>
                    <CardDescription>
                        {t('review.totalItems', {total})}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[40px]">
                                    {activeTab === 'pending' && (
                                        <input
                                            type="checkbox"
                                            checked={selectedItems.length === reviewItems.length && reviewItems.length > 0}
                                            onChange={handleSelectAll}
                                        />
                                    )}
                                </TableHead>
                                <TableHead>Media</TableHead>
                                <TableHead>User</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Created</TableHead>
                                {activeTab === 'history' && <TableHead>Reviewed By</TableHead>}
                                <TableHead>Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {reviewItems.length === 0 ? (
                                <TableRow>
                                    <TableCell colSpan={activeTab === 'history' ? 6 : 5} className="text-center">
                                        {t('review.noItems')}
                                    </TableCell>
                                </TableRow>
                            ) : (
                                reviewItems.map((item) => (
                                    <TableRow key={item.id}>
                                        <TableCell>
                                            {activeTab === 'pending' && (
                                                <input
                                                    type="checkbox"
                                                    checked={selectedItems.includes(item.id)}
                                                    onChange={() => handleSelectItem(item.id)}
                                                />
                                            )}
                                        </TableCell>
                                        <TableCell>
                                            <div className="space-y-1">
                                                <p className="font-medium text-gray-900 dark:text-white">{item.media_title}</p>
                                                <p className="text-xs text-gray-500 dark:text-gray-400">{item.media_type}</p>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            <div className="space-y-1">
                                                <p className="font-medium text-gray-900 dark:text-white">{item.username}</p>
                                                <p className="text-xs text-gray-500 dark:text-gray-400">ID: {item.user_id}</p>
                                            </div>
                                        </TableCell>
                                        <TableCell>
                                            {getStatusBadge(item.status)}
                                        </TableCell>
                                        <TableCell>
                                            <p className="text-sm text-gray-600 dark:text-gray-400">{formatDate(item.created_at)}</p>
                                        </TableCell>
                                        {activeTab === 'history' && (
                                            <TableCell>
                                                <p className="text-sm text-gray-600 dark:text-gray-400">
                                                    {item.reviewer_name || 'N/A'}
                                                </p>
                                            </TableCell>
                                        )}
                                        <TableCell>
                                            {activeTab === 'pending' ? (
                                                <div className="flex items-center gap-2">
                                                    <Button
                                                        size="sm"
                                                        className="bg-green-600 hover:bg-green-700"
                                                        onClick={() => handleReview(item.id, 'approved')}
                                                    >
                                                        <Check className="w-4 h-4"/>
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        className="bg-red-600 hover:bg-red-700"
                                                        onClick={() => handleReview(item.id, 'rejected')}
                                                    >
                                                        <X className="w-4 h-4"/>
                                                    </Button>
                                                    <Button
                                                        size="sm"
                                                        variant="outline"
                                                        onClick={() => openSingleDialog(item)}
                                                    >
                                                        {t('common.details')}
                                                    </Button>
                                                </div>
                                            ) : (
                                                <Button
                                                    size="sm"
                                                    variant="outline"
                                                    onClick={() => openSingleDialog(item)}
                                                >
                                                    {t('common.details')}
                                                </Button>
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))
                            )}
                        </TableBody>
                    </Table>

                    {/* Pagination */}
                    <div className="flex items-center justify-between mt-6">
                        <p className="text-sm text-gray-600 dark:text-gray-400">
                            {t('common.pagination', {
                                start: (page - 1) * pageSize + 1,
                                end: Math.min(page * pageSize, total),
                                total,
                            })}
                        </p>
                        <div className="flex items-center gap-2">
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setPage(prev => Math.max(prev - 1, 1))}
                                disabled={page === 1}
                            >
                                {t('common.previous')}
                            </Button>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setPage(prev => prev + 1)}
                                disabled={page * pageSize >= total}
                            >
                                {t('common.next')}
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Batch Review Dialog */}
            <Dialog open={showBatchDialog} onOpenChange={setShowBatchDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('review.batchReview')}</DialogTitle>
                        <DialogDescription>
                            {t('review.batchReviewDescription', {count: selectedItems.length})}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div>
                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                {t('common.status')}
                            </h4>
                            <div className="flex items-center gap-2">
                                <Button
                                    variant={batchStatus === 'approved' ? 'default' : 'outline'}
                                    onClick={() => setBatchStatus('approved')}
                                >
                                    <Check className="w-4 h-4 mr-2"/>
                                    {t('review.approved')}
                                </Button>
                                <Button
                                    variant={batchStatus === 'rejected' ? 'default' : 'outline'}
                                    onClick={() => setBatchStatus('rejected')}
                                >
                                    <X className="w-4 h-4 mr-2"/>
                                    {t('review.rejected')}
                                </Button>
                            </div>
                        </div>
                        {batchStatus === 'rejected' && (
                            <div>
                                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                    {t('review.reason')}
                                </h4>
                                <Textarea
                                    placeholder={t('review.reasonPlaceholder')}
                                    value={batchReason}
                                    onChange={(e) => setBatchReason(e.target.value)}
                                />
                            </div>
                        )}
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowBatchDialog(false)}>
                            {t('common.cancel')}
                        </Button>
                        <Button onClick={handleBatchReview}>
                            {t('review.submit')}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Single Review Dialog */}
            <Dialog open={showSingleDialog} onOpenChange={setShowSingleDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('review.reviewItem')}</DialogTitle>
                        <DialogDescription>
                            {currentItem?.media_title}
                        </DialogDescription>
                    </DialogHeader>
                    {currentItem && (
                        <div className="space-y-4 py-4">
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                    {t('common.media')}
                                </h4>
                                <p className="text-sm text-gray-900 dark:text-white">{currentItem.media_title}</p>
                                <p className="text-xs text-gray-500 dark:text-gray-400">{currentItem.media_type}</p>
                            </div>
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                    {t('common.user')}
                                </h4>
                                <p className="text-sm text-gray-900 dark:text-white">{currentItem.username}</p>
                                <p className="text-xs text-gray-500 dark:text-gray-400">ID: {currentItem.user_id}</p>
                            </div>
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                    {t('common.status')}
                                </h4>
                                {getStatusBadge(currentItem.status)}
                            </div>
                            <div className="space-y-2">
                                <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                    {t('common.createdAt')}
                                </h4>
                                <p className="text-sm text-gray-600 dark:text-gray-400">{formatDate(currentItem.created_at)}</p>
                            </div>
                            {currentItem.reason && (
                                <div className="space-y-2">
                                    <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                                        {t('review.reason')}
                                    </h4>
                                    <p className="text-sm text-gray-600 dark:text-gray-400">{currentItem.reason}</p>
                                </div>
                            )}
                            {activeTab === 'pending' && (
                                <div className="space-y-4">
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                            {t('review.action')}
                                        </h4>
                                        <div className="flex items-center gap-2">
                                            <Button
                                                variant={singleStatus === 'approved' ? 'default' : 'outline'}
                                                onClick={() => setSingleStatus('approved')}
                                            >
                                                <Check className="w-4 h-4 mr-2"/>
                                                {t('review.approved')}
                                            </Button>
                                            <Button
                                                variant={singleStatus === 'rejected' ? 'default' : 'outline'}
                                                onClick={() => setSingleStatus('rejected')}
                                            >
                                                <X className="w-4 h-4 mr-2"/>
                                                {t('review.rejected')}
                                            </Button>
                                        </div>
                                    </div>
                                    {singleStatus === 'rejected' && (
                                        <div>
                                            <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                                                {t('review.reason')}
                                            </h4>
                                            <Textarea
                                                placeholder={t('review.reasonPlaceholder')}
                                                value={singleReason}
                                                onChange={(e) => setSingleReason(e.target.value)}
                                            />
                                        </div>
                                    )}
                                </div>
                            )}
                        </div>
                    )}
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setShowSingleDialog(false)}>
                            {t('common.close')}
                        </Button>
                        {activeTab === 'pending' && (
                            <Button onClick={handleSingleReview}>
                                {t('review.submit')}
                            </Button>
                        )}
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default ReviewFlow;
