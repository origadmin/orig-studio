import React, {useState, useEffect} from 'react';
import {Link} from '@tanstack/react-router';
import {Users, UserPlus, Search, Loader2} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {formatDate} from '@/lib/format';
import {subscriptionApi} from '@/lib/api/subscription';
import ErrorPage from '@/components/common/ErrorPage';

const SubscriptionsPage = () => {
    const {t} = useTranslation();
    const [activeTab, setActiveTab] = useState<'subscriptions' | 'followers'>('subscriptions');
    const [subscriptions, setSubscriptions] = useState<any[]>([]);
    const [followers, setFollowers] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);

    useEffect(() => {
        fetchData();
    }, [activeTab]);

    const fetchData = async () => {
        try {
            setLoading(true);
            setError(null);

            let response;
            if (activeTab === 'subscriptions') {
                response = await subscriptionApi.getSubscriptions({page, page_size: 20});
                setSubscriptions(prev => page === 1 ? response.items : [...prev, ...response.items]);
                setHasMore(response.items.length === 20);
            } else {
                response = await subscriptionApi.getFollowers({page, page_size: 20});
                setFollowers(prev => page === 1 ? response.items : [...prev, ...response.items]);
                setHasMore(response.items.length === 20);
            }
        } catch (err) {
            setError('Failed to fetch data');
            console.error('Failed to fetch data:', err);
        } finally {
            setLoading(false);
        }
    };

    const handleLoadMore = () => {
        if (!loading && hasMore) {
            setPage(prev => prev + 1);
        }
    };

    const renderUserList = () => {
        const list = activeTab === 'subscriptions' ? subscriptions : followers;

        if (loading && page === 1) {
            return (
                <div className="flex items-center justify-center min-h-[400px]">
                    <div
                        className="animate-spin w-8 h-8 border-4 border-emerald-600 border-t-transparent rounded-full"/>
                </div>
            );
        }

        if (error) {
            return <ErrorPage message={error}/>;
        }

        if (list.length === 0) {
            return (
                <div className="text-center py-16 text-gray-400">
                    <Users size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>{activeTab === 'subscriptions' ? t('subscriptions.noSubscriptions') : t('subscriptions.noFollowers')}</p>
                </div>
            );
        }

        return (
            <div className="space-y-4">
                {list.map((user) => (
                    <div key={user.id}
                         className="flex items-center justify-between p-4 bg-white dark:bg-gray-800 rounded-lg shadow-sm">
                        <Link to={`/u/${user.user_id}`} className="flex items-center gap-3">
                            <Avatar className="h-12 w-12">
                                <AvatarImage src={user.avatar}/>
                                <AvatarFallback>{user.username?.[0] || 'U'}</AvatarFallback>
                            </Avatar>
                            <div>
                                <p className="font-medium text-gray-900 dark:text-white">{user.username}</p>
                                <p className="text-xs text-gray-500 dark:text-gray-400">
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
        );
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Users size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('subscriptions.title')}</h1>
            </div>

            <div className="flex border-b dark:border-gray-700">
                <button
                    className={`px-4 py-3 font-medium ${activeTab === 'subscriptions'
                        ? 'border-b-2 border-emerald-600 text-emerald-600'
                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'}
                    `}
                    onClick={() => setActiveTab('subscriptions')}
                >
                    {t('subscriptions.subscriptions')}
                </button>
                <button
                    className={`px-4 py-3 font-medium ${activeTab === 'followers'
                        ? 'border-b-2 border-emerald-600 text-emerald-600'
                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'}
                    `}
                    onClick={() => setActiveTab('followers')}
                >
                    {t('subscriptions.followers')}
                </button>
            </div>

            {renderUserList()}
        </div>
    );
};

export default SubscriptionsPage;
