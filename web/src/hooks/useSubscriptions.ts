import {useState, useEffect} from 'react';
import {useAuth} from './useAuth';
import {subscriptionApi, type SubscriptionListResponse} from '../lib/api/subscription';
import type {NavItem} from '../types/nav';
import {User} from 'lucide-react';

interface ChannelSummary {
    id: string;
    username: string;
    avatar?: string;
}

export interface UseSubscribedChannelsReturn {
    channels: NavItem[];
    loading: boolean;
    error: Error | null;
    refetch: () => void;
}

const MAX_CHANNELS = 8;

function toNavItems(channels: ChannelSummary[]): NavItem[] {
    return channels.slice(0, MAX_CHANNELS).map((ch) => ({
        id: `ch-${ch.id}`,
        label: ch.username,
        to: `/u/${ch.id}`,
        icon: User,
    }));
}

export function useSubscribedChannels(): UseSubscribedChannelsReturn {
    const {isAuthenticated} = useAuth();
    const [channels, setChannels] = useState<NavItem[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);
    const [version, setVersion] = useState(0);

    useEffect(() => {
        if (!isAuthenticated) {
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
                    username: item.username,
                    avatar: item.avatar,
                }));
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

    return {channels, loading, error, refetch};
}
