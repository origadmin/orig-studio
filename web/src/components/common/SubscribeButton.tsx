import React, {useState, useEffect} from 'react';
import {Button} from '@/components/ui/button';
import {UserPlus} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {subscriptionApi} from '../../lib/api';

interface SubscribeButtonProps {
    userId: string;
    initialSubscriberCount?: number;
    className?: string;
}

const SubscribeButton: React.FC<SubscribeButtonProps> = ({
                                                             userId,
                                                             initialSubscriberCount = 0,
                                                             className = ''
                                                         }) => {
    const {t} = useTranslation();
    const [isSubscribed, setIsSubscribed] = useState(false);
    const [subscriberCount, setSubscriberCount] = useState(initialSubscriberCount);
    const [loading, setLoading] = useState(false);
    const [initialLoading, setInitialLoading] = useState(true);

    useEffect(() => {
        const fetchStatus = async () => {
            try {
                const response = await subscriptionApi.getStatus(userId);
                setIsSubscribed(response.is_subscribed);
                if (response.subscriber_count !== undefined) {
                    setSubscriberCount(response.subscriber_count);
                }
            } catch (err) {
                console.error('Failed to fetch subscription status:', err);
            } finally {
                setInitialLoading(false);
            }
        };
        fetchStatus();
    }, [userId]);

    const handleSubscribe = async () => {
        try {
            setLoading(true);
            if (isSubscribed) {
                await subscriptionApi.unsubscribe(userId);
                setIsSubscribed(false);
                setSubscriberCount(prev => Math.max(0, prev - 1));
            } else {
                await subscriptionApi.subscribe(userId);
                setIsSubscribed(true);
                setSubscriberCount(prev => prev + 1);
            }
        } catch (err) {
            console.error('Failed to toggle subscription:', err);
        } finally {
            setLoading(false);
        }
    };

    if (initialLoading) {
        return (
            <Button disabled className={className}>
                <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mr-2"/>
                {t('common.loading')}
            </Button>
        );
    }

    return (
        <Button
            onClick={handleSubscribe}
            disabled={loading}
            className={`${className} ${isSubscribed ? 'bg-gray-800 hover:bg-gray-700' : 'bg-red-600 hover:bg-red-700'}`}
        >
            {loading ? (
                <>
                    <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mr-2"/>
                    {t('common.loading')}
                </>
            ) : isSubscribed ? (
                <>
                    <UserPlus className="w-4 h-4 mr-2"/>
                    {t('common.subscribed')}
                </>
            ) : (
                t('common.subscribe')
            )}
        </Button>
    );
};

export default SubscribeButton;
