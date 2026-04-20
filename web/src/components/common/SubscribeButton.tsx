import React, {useState, useEffect} from 'react';
import {Button} from '@/components/ui/button';
import {UserPlus, UserCheck, Loader2} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {subscriptionApi} from '@/lib/api/subscription';
import {useAuth} from '@/hooks/useAuth';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

interface SubscribeButtonProps {
    channelId: string;
    initialSubscriberCount?: number;
    className?: string;
    size?: 'sm' | 'default' | 'lg';
    variant?: 'default' | 'outline';
}

const SubscribeButton: React.FC<SubscribeButtonProps> = ({
                                                             channelId,
                                                             initialSubscriberCount = 0,
                                                             className = '',
                                                             size = 'default',
                                                             variant = 'default'
                                                         }) => {
    const {t} = useTranslation();
    const {isAuthenticated, user} = useAuth();
    const [isSubscribed, setIsSubscribed] = useState(false);
    const [subscriberCount, setSubscriberCount] = useState(initialSubscriberCount);
    const [loading, setLoading] = useState(false);
    const [initialLoading, setInitialLoading] = useState(true);
    const [showLoginDialog, setShowLoginDialog] = useState(false);

    useEffect(() => {
        const fetchStatus = async () => {
            if (!isAuthenticated || !channelId) {
                setInitialLoading(false);
                return;
            }
            try {
                const response = await subscriptionApi.getStatus(channelId);
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
    }, [channelId, isAuthenticated]);

    const handleSubscribe = async () => {
        if (!isAuthenticated) {
            setShowLoginDialog(true);
            return;
        }

        if (!channelId) {
            return;
        }

        try {
            setLoading(true);
            if (isSubscribed) {
                await subscriptionApi.unsubscribe(channelId);
                setIsSubscribed(false);
                setSubscriberCount(prev => Math.max(0, prev - 1));
            } else {
                await subscriptionApi.subscribe(channelId);
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
            <Button
                disabled
                size={size}
                variant="outline"
                className={className}
            >
                <Loader2 className="w-4 h-4 animate-spin mr-2"/>
                {t('common.loading')}
            </Button>
        );
    }

    const buttonVariant = isSubscribed ? 'outline' : 'default';
    const buttonClass = isSubscribed
        ? 'border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
        : 'bg-red-600 hover:bg-red-700 text-white';

    return (
        <>
            <Button
                onClick={handleSubscribe}
                disabled={loading}
                size={size}
                variant={buttonVariant}
                className={`${className} ${buttonClass}`}
            >
                {loading ? (
                    <>
                        <Loader2 className="w-4 h-4 animate-spin mr-2"/>
                        {t('common.loading')}
                    </>
                ) : isSubscribed ? (
                    <>
                        <UserCheck className="w-4 h-4 mr-2"/>
                        {t('common.subscribed')}
                    </>
                ) : (
                    <>
                        <UserPlus className="w-4 h-4 mr-2"/>
                        {t('common.subscribe')}
                    </>
                )}
            </Button>

            {/* Login Dialog */}
            <Dialog open={showLoginDialog} onOpenChange={setShowLoginDialog}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('auth.loginRequired') || 'Login Required'}</DialogTitle>
                        <DialogDescription>
                            {t('auth.loginToSubscribe') || 'Please login to subscribe to this channel.'}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex justify-end gap-3 mt-4">
                        <Button variant="outline" onClick={() => setShowLoginDialog(false)}>
                            {t('common.cancel')}
                        </Button>
                        <Button
                            className="bg-emerald-600 hover:bg-emerald-700"
                            onClick={() => window.location.href = '/auth/signin'}
                        >
                            {t('auth.signin')}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>
        </>
    );
};

export default SubscribeButton;
