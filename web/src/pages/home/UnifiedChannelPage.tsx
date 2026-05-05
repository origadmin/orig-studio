import React, {useMemo, useState} from 'react';
import {useLocation, useNavigate, useParams} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {
    useChannelByToken,
    useChannelByHandle,
    useMyChannel,
    useSubscriptionStatus,
} from '@/hooks/queries';
import {type ChannelDetail} from '@/lib/api/channel';
import ChannelLayout from '@/components/channel/ChannelLayout';
import ChannelSkeleton from '@/components/channel/ChannelSkeleton';
import ChannelNotFound from '@/components/channel/ChannelNotFound';
import {CreateChannelDialog} from '@/components/channel/CreateChannelDialog';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';

const ChannelPage: React.FC = () => {
    const {t} = useTranslation();
    const location = useLocation();
    const pathname = location.pathname;
    const params = useParams({strict: false});
    const navigate = useNavigate();
    const {user, isAuthenticated} = useAuth();
    const [createDialogOpen, setCreateDialogOpen] = useState(false);

    let handle: string | undefined;
    let token: string | undefined;

    if (pathname.startsWith('/@')) {
        handle = pathname.slice(2);
    } else if (pathname.startsWith('/c/')) {
        token = pathname.slice(3);
    } else if (pathname.startsWith('/channel/')) {
        token = pathname.slice(9);
    }

    const isFromMeChannel = !token && !handle;

    const handleQuery = useChannelByHandle(handle || null);
    const tokenQuery = useChannelByToken(token || null);
    const myChannelQuery = useMyChannel(isFromMeChannel && isAuthenticated);

    const activeQuery = handle
        ? handleQuery
        : token
            ? tokenQuery
            : myChannelQuery;

    const channel = activeQuery.data as ChannelDetail | undefined;
    const loading = activeQuery.isLoading;
    const error = activeQuery.error;

    const isOwner = useMemo(() => {
        if (!user || !channel) return false;
        if (isFromMeChannel) return true;
        return String(user.id) === String(channel.user_id);
    }, [user, channel, isFromMeChannel]);

    const channelToken = channel?.short_token || null;
    const subscriptionQuery = useSubscriptionStatus(
        channelToken && !isOwner && isAuthenticated ? channelToken : null
    );

    if (loading) {
        return <ChannelSkeleton/>;
    }

    if (error || !channel) {
        if (isFromMeChannel && !isAuthenticated) {
            return <ChannelNotFound message={t('channel.login_required')} onBack={() => navigate({to: '/'})}/>;
        }
        if (isFromMeChannel) {
            return (
                <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6 p-8">
                    <Avatar className="h-24 w-24">
                        <AvatarImage src={user?.avatarUrl} alt={user?.username}/>
                        <AvatarFallback>{user?.username?.charAt(0)?.toUpperCase() || '?'}</AvatarFallback>
                    </Avatar>
                    <div className="text-center">
                        <h1 className="text-2xl font-bold">{user?.username}</h1>
                    </div>
                    <p className="text-muted-foreground text-center max-w-md">
                        {t('channel.profile.no_channel_self')}
                    </p>
                    <Button onClick={() => setCreateDialogOpen(true)} size="lg">
                        {t('channel.create.title')}
                    </Button>
                    <CreateChannelDialog
                        open={createDialogOpen}
                        onOpenChange={setCreateDialogOpen}
                        onSuccess={({handle: _newHandle, short_token: newToken}) => {
                            if (newToken) {
                                navigate({to: '/c/$id', params: {id: newToken}});
                            } else {
                                // short_token should always be returned from backend; reload as last resort
                                window.location.reload();
                            }
                        }}
                    />
                </div>
            );
        }
        if (handle) {
            return (
                <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4 p-8">
                    <Avatar className="h-24 w-24">
                        <AvatarFallback>{handle.charAt(0).toUpperCase()}</AvatarFallback>
                    </Avatar>
                    <div className="text-center">
                        <h1 className="text-2xl font-bold">{handle}</h1>
                        <p className="text-sm text-muted-foreground mt-1">@{handle}</p>
                    </div>
                    <p className="text-muted-foreground text-center max-w-md">
                        {t('channel.profile.no_channel_other', {name: handle})}
                    </p>
                </div>
            );
        }
        return (
            <ChannelNotFound
                message={error ? (error as any)?.response?.data?.message || (error as Error).message || 'Failed to load channel' : t('channel.not_found')}
                onBack={() => navigate({to: '/'})}
            />
        );
    }

    return (
        <ChannelLayout
            channel={channel}
            isOwner={isOwner}
            isFromMeChannel={isFromMeChannel}
            isSubscribed={subscriptionQuery.data?.is_subscribed ?? false}
            subscriptionLoading={subscriptionQuery.isLoading}
        />
    );
};

export default ChannelPage;
