import React, {useState, useEffect} from 'react';
import {Link} from '@tanstack/react-router';
import {Users, UserPlus, Loader2} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Spinner} from '@/components/ui/spinner';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDate} from '@/lib/format';
import {subscriptionApi} from '@/lib/api/subscription';
import ErrorPage from '@/components/common/ErrorPage';
import {PAGINATION_CONFIG} from '@/config/pagination';

const SubscriptionsPage = () => {
    const {t} = useTranslation();
    const [subscriptions, setSubscriptions] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);

    useEffect(() => {
        fetchData(1);
    }, []);

    const fetchData = async (pageNum: number) => {
        try {
            setLoading(true);
            setError(null);
            const response = await subscriptionApi.getSubscriptions({page: pageNum, page_size: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE});
            setSubscriptions(prev => pageNum === 1 ? response.items : [...prev, ...response.items]);
            setHasMore(response.items.length === PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
        } catch (err) {
            setError('Failed to fetch data');
            console.error('Failed to fetch data:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleLoadMore = () => {
        if (!loading && hasMore) {
            const nextPage = page + 1;
            setPage(nextPage);
            fetchData(nextPage);
        }
    };

    if (loading && page === 1) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Users size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('subscriptions.title')}</h1>
            </div>

            {subscriptions.length === 0 ? (
                <div className="text-center py-16 text-muted-foreground">
                    <Users size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>{t('subscriptions.noSubscriptions')}</p>
                </div>
            ) : (
                <div className="space-y-4">
                    {subscriptions.map((user) => (
                        <div key={user.id}
                             className="flex items-center justify-between p-4 bg-white dark:bg-gray-800 rounded-lg shadow-sm">
                            <Link to={user.username ? `/@${user.username}` : `/u/${user.user_id}`} className="flex items-center gap-3">
                                <Avatar className="h-12 w-12">
                                    <AvatarImage src={user.avatar}/>
                                    <AvatarFallback>{user.username?.[0] || 'U'}</AvatarFallback>
                                </Avatar>
                                <div>
                                    <p className="font-medium text-gray-900 dark:text-white">{user.username}</p>
                                    <p className="text-xs text-gray-500 dark:text-muted-foreground">
                                        {t('subscriptions.subscribedAt', {date: formatDate(user.subscribed_at)})}
                                    </p>
                                </div>
                            </Link>
                            <Button variant="outline" className="rounded-full">
                                <UserPlus className="w-4 h-4 mr-2"/>
                                {t('common.subscribed')}
                            </Button>
                        </div>
                    ))}
                    {hasMore && (
                        <div className="flex justify-center mt-8">
                            <Button
                                variant="outline"
                                onClick={handleLoadMore}
                                disabled={loading}
                            >
                                {loading ? (
                                    <>
                                        <Loader2 className="w-4 h-4 mr-2 animate-spin"/>
                                        {t('common.loading')}
                                    </>
                                ) : (
                                    t('common.loadMore')
                                )}
                            </Button>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default SubscriptionsPage;
