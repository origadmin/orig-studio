import React from 'react';
import {useTranslation} from 'react-i18next';
import {Button} from '@/components/ui/button';
import {CheckCircle, UserPlus, Loader2} from 'lucide-react';

interface SubscribeButtonProps {
    isSubscribed: boolean;
    isOwner: boolean;
    subscriberCount?: number;
    subscribing?: boolean;
    onSubscribe?: () => void;
    onUnsubscribeClick?: () => void;
}

const SubscribeButton: React.FC<SubscribeButtonProps> = ({
    isSubscribed,
    isOwner,
    subscriberCount: _subscriberCount = 0,
    subscribing = false,
    onSubscribe,
    onUnsubscribeClick,
}) => {
    const {t} = useTranslation();

    if (isOwner) return null;

    return (
        <Button
            variant={isSubscribed ? 'outline' : 'default'}
            size={isSubscribed ? 'sm' : 'default'}
            className={`${
                isSubscribed
                    ? 'text-primary border-primary hover:bg-primary/10'
                    : ''
            } ${subscribing ? 'pointer-events-none' : ''}`}
            onClick={isSubscribed ? onUnsubscribeClick : onSubscribe}
            disabled={subscribing}
        >
            {subscribing ? (
                <>
                    <Loader2 className="w-4 h-4 mr-1 animate-spin"/>
                    {t('channel.unsubscribing')}
                </>
            ) : isSubscribed ? (
                <>
                    <CheckCircle className="w-4 h-4 mr-1"/>
                    {t('channel.subscribed')}
                </>
            ) : (
                <>
                    <UserPlus className="w-4 h-4 mr-1"/>
                    {t('channel.subscribe')}
                </>
            )}
        </Button>
    );
};

export default SubscribeButton;
