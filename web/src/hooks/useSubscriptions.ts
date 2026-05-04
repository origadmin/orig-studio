import {useState, useEffect} from 'react';
import {useAuth} from './useAuth';
import {subscriptionApi, type SubscriptionListResponse} from '../lib/api/subscription';
import type {NavItem} from '../types/nav';
import {User} from 'lucide-react';

export interface ChannelSummary {
    id: string;
    name: string;
    username: string;
    avatar?: string;
    short_token?: string;
    channel_id?: string;
}

export interface UseSubscribedChannelsReturn {
    channels: NavItem[];
    channelDetails: ChannelSummary[];
    loading: boolean;
    error: Error | null;
    refetch: () => void;
}

const MAX_CHANNELS = 8;

function toNavItems(channels: ChannelSummary[]): NavItem[] {
    return channels.slice(0, MAX_CHANNELS).map((ch) => {
        const item: NavItem = {
            id: `ch-${ch.id}`,
            label: ch.name || ch.username,
            to: ch.short_token ? '/c/$id' : ch.username ? '/$handle' : '/u/$id',
            icon: User,
        };
        if (ch.short_token) {
            item.params = {id: ch.short_token};
        } else if (ch.username) {
            item.params = {handle: `@${ch.username}`};
        } else {
            item.params = {id: ch.id};
        }
        return item;
    });
}

export function useSubscribedChannels(): UseSubscribedChannelsReturn {
    const {isAuthenticated} = useAuth();
    const [channelDetails, setChannelDetails] = useState<ChannelSummary[]>([]);
    const [channels, setChannels] = useState<NavItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);
    const [version, setVersion] = useState(0);

    useEffect(() => {
        if (!isAuthenticated) {
            setChannelDetails([]);
            setChannels([]);
            return;
        }

        let cancelled = false;
        setLoading(true);
        setError(null);

        subscriptionApi
            .getSubscriptions({page_size: MAX_CHANNELS})
            .then((res: SubscriptionListResponse) => {
                if (cancelled) return;
                const items: ChannelSummary[] = (res?.items || []).map((item) => ({
                    id: item.id,
                    name: item.name || item.username,
                    username: item.username,
                    avatar: item.avatar,
                    short_token: item.short_token,
                    channel_id: item.channel_id,
                }));
                setChannelDetails(items);
                setChannels(toNavItems(items));
            })
            .catch((err) => {
                if (cancelled) return;
                setError(err instanceof Error ? err : new Error(String(err)));
            })
            .finally(() => {
                if (!cancelled) setLoading(false);
            });

        return () => {
            cancelled = true;
        };
    }, [isAuthenticated, version]);

    const refetch = () => setVersion((v) => v + 1);

    return {channels, channelDetails, loading, error, refetch};
}
