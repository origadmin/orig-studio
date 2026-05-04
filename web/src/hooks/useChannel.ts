import {useState, useEffect, useMemo, useCallback} from 'react';
import {useLocation} from '@tanstack/react-router';
import {useAuth} from '@/hooks/useAuth';
import {channelApi, type ChannelDetail, type ChannelQueryParams} from '@/lib/api/channel';

interface UseChannelOptions {
    enabled?: boolean;
}

interface UseChannelReturn {
    channel: ChannelDetail | null;
    loading: boolean;
    error: string | null;
    isOwner: boolean;
    isFromMeChannel: boolean;
    refetch: () => void;
}

/**
 * useChannel Hook v4.0 (F019)
 *
 * 支持多种路由来源:
 * - /c/{token} 或 /channel/{id} → 路径参数方式 (getByToken)
 * - /@{handle}              → Handle resolution (resolveHandle)
 * - /me/channel             → 当前用户频道 (getMyChannels)
 */
const useChannel = (options: UseChannelOptions = {}): UseChannelReturn => {
    const {enabled = true} = options;

    const location = useLocation();
    const pathname = location.pathname;

    let handle: string | undefined;
    let token: string | undefined;
    let id: string | undefined;

    if (pathname.startsWith('/@')) {
        handle = pathname.slice(2);
    } else if (pathname.startsWith('/c/')) {
        token = pathname.slice(3);
    } else if (pathname.startsWith('/channel/')) {
        token = pathname.slice(9);
    }

    const {user, isAuthenticated} = useAuth();

    const [channel, setChannel] = useState<ChannelDetail | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const isFromMeChannel = useMemo(() => {
        return !handle && !token && !id;
    }, [handle, token, id]);

    const fetchChannel = useCallback(async () => {
        if (!enabled) return;

        let isMounted = true;

        try {
            setLoading(true);
            setError(null);

            let response: any;

            if (token) {
                // Mode 1: 路径参数方式 (RESTful, 推荐)
                response = await channelApi.getByToken(token);
            } else if (handle) {
                // Mode 2: @handle resolution (F019: uses resolveHandle API)
                response = await channelApi.resolveHandle(handle);
            } else if (isAuthenticated && user?.id) {
                // Mode 3: 我的频道
                response = await channelApi.getMyChannels();
            } else {
                if (isMounted) {
                    setError('请先登录');
                    return;
                }
            }

            if (isMounted && response?.code === 0) {
                const data = response.data;
                // Handle resolution returns { type, channel, user }
                if (data?.type === 'channel' && data?.channel) {
                    setChannel(data.channel as ChannelDetail);
                } else if (data?.items && Array.isArray(data.items)) {
                    // getMyChannels returns list
                    setChannel(data.items.length > 0 ? data.items[0] as ChannelDetail : null);
                } else if (data?.id) {
                    setChannel(data as ChannelDetail);
                } else {
                    setChannel(null);
                }
            }
        } catch (err: any) {
            console.error('Failed to fetch channel:', err);
            if (isMounted) {
                setError(err.response?.data?.message || err.message || '加载频道失败');
            }
        } finally {
            if (isMounted) {
                setLoading(false);
            }
        }

        return () => {
            isMounted = false;
        };
    }, [token, handle, user?.id, isAuthenticated, enabled]);

    useEffect(() => {
        fetchChannel();
    }, [fetchChannel]);

    const isOwner = useMemo(() => {
        if (!user || !channel) return false;

        if (isFromMeChannel) return true;

        return String(user.id) === String(channel.user_id);
    }, [user, channel, isFromMeChannel]);

    return {
        channel,
        loading,
        error,
        isOwner,
        isFromMeChannel,
        refetch: fetchChannel,
    };
};

export default useChannel;
