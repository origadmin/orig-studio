import React, {useMemo} from 'react';
import {useLocation, useNavigate} from '@tanstack/react-router';
import {useAuth} from '@/hooks/useAuth';
import {
    useChannelByToken,
    useChannelByHandle,
    useMyChannel,
    useSubscriptionStatus,
} from '@/hooks/queries';
import type {ChannelDetail} from '@/lib/api/channel';
import ChannelLayout from '@/components/channel/ChannelLayout';
import ChannelSkeleton from '@/components/channel/ChannelSkeleton';
import ChannelNotFound from '@/components/channel/ChannelNotFound';

/**
 * UnifiedChannelPage: Handles three routing modes
 *   /@{handle}      - @username query mode (two-step)
 *   /c/{token}      - short_token path mode (RESTful, recommended)
 *   /channel/{token} - short_token path mode (alternative)
 *   /me/channel      - my channel (requires auth)
 */
const UnifiedChannelPage: React.FC = () => {
    const location = useLocation();
    const pathname = location.pathname;
    const navigate = useNavigate();
    const {user, isAuthenticated} = useAuth();

    // Parse routing mode from pathname
    let handle: string | undefined;
    let token: string | undefined;

    if (pathname.startsWith('/@')) {
        handle = pathname.slice(2);
    } else if (pathname.startsWith('/c/')) {
        token = pathname.slice(3);
    } else if (pathname.startsWith('/channel/')) {
        token = pathname.slice(9);
    }

    const isFromMeChannel = useMemo(() => {
        return !handle && !token;
    }, [handle, token]);

    // Fetch channel data using TanStack Query based on routing mode
    const tokenQuery = useChannelByToken(token || null);
    const handleQuery = useChannelByHandle(handle || null);
    const myChannelQuery = useMyChannel(isFromMeChannel && isAuthenticated);

    // Determine which query result to use
    const activeQuery = token
        ? tokenQuery
        : handle
            ? handleQuery
            : myChannelQuery;

    const channel = activeQuery.data as ChannelDetail | undefined;
    const loading = activeQuery.isLoading;
    const error = activeQuery.error;

    // Fetch subscription status when channel is loaded and user is not the owner
    const channelToken = channel?.short_token || null;
    const isOwner = useMemo(() => {
        if (!user || !channel) return false;
        if (isFromMeChannel) return true;
        return String(user.id) === String(channel.owner_id) || String(user.id) === String(channel.user_id);
    }, [user, channel, isFromMeChannel]);

    const subscriptionQuery = useSubscriptionStatus(
        channelToken && !isOwner && isAuthenticated ? channelToken : null
    );

    if (loading) {
        return <ChannelSkeleton/>;
    }

    if (error || !channel) {
        const errorMessage = error
            ? (error as any)?.response?.data?.message || (error as Error).message || 'Failed to load channel'
            : isFromMeChannel && !isAuthenticated
                ? 'Please log in first'
                : 'Channel not found';
        return (
            <ChannelNotFound
                message={errorMessage}
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

export default UnifiedChannelPage;
