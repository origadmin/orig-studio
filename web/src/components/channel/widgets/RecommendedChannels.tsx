import React, {useState, useEffect} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {CheckCircle, Users, ChevronRight, Loader2} from 'lucide-react';
import {channelApi} from '@/lib/api/channel';

interface RecommendedChannelsProps {
    currentChannelId: string;
}

interface RecommendedChannel {
    id: string;
    name: string;
    handle?: string;
    avatar?: string;
    subscriber_count?: number;
    is_verified?: boolean;
    is_subscribed?: boolean;
}

const RecommendedChannels: React.FC<RecommendedChannelsProps> = ({
    currentChannelId,
}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const [channels, setChannels] = useState<RecommendedChannel[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchRecommended = async () => {
            try {
                setLoading(true);
                const response = await channelApi.listAll();
                if (response) {
                    const data = response as any;
                    if (Array.isArray(data?.items)) {
                        setChannels(data.items);
                    } else if (Array.isArray(data)) {
                        setChannels(data);
                    }
                }
            } catch (error) {
                console.error('Failed to fetch recommended channels:', error);
                setChannels([]);
            } finally {
                setLoading(false);
            }
        };

        fetchRecommended();
    }, [currentChannelId]);

    const handleSubscribe = async (
        channelId: string,
        e: React.MouseEvent
    ) => {
        e.preventDefault();
        e.stopPropagation();

        setChannels((prev) =>
            prev.map((ch) =>
                ch.id === channelId
                    ? {...ch, is_subscribed: true, subscriber_count: (ch.subscriber_count || 0) + 1}
                    : ch
            )
        );
    };

    const formatCount = (num: number): string => {
        if (!num) return '0';
        if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
        if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
        return num.toString();
    };

    if (loading) {
        return (
            <div className="space-y-3 animate-pulse">
                {[...Array(3)].map((_, i) => (
                    <div key={i} className="flex items-center gap-3 p-3">
                        <div className="w-10 h-10 rounded-full bg-muted dark:bg-gray-700"/>
                        <div className="flex-1 space-y-2">
                            <div className="h-4 w-32 bg-muted dark:bg-gray-700 rounded"/>
                            <div className="h-3 w-20 bg-muted dark:bg-gray-700 rounded"/>
                        </div>
                    </div>
                ))}
            </div>
        );
    }

    if (channels.length === 0) {
        return null;
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h3 className="font-semibold text-sm text-foreground flex items-center gap-2">
                    <Users className="w-4 h-4"/>
                    {t('channel.recommendedChannels')}
                </h3>
                <button
                    onClick={() => navigate({to: '/members'})}
                    className="text-xs text-primary hover:underline flex items-center gap-1"
                >
                    查看全部
                    <ChevronRight className="w-3 h-3"/>
                </button>
            </div>

            <div className="space-y-2">
                {channels.map((channel) => (
                    <div
                        key={channel.id}
                        className="flex items-center gap-3 p-2 rounded-lg hover:bg-accent/50 transition-colors cursor-pointer group"
                        onClick={() =>
                            navigate({
                                to:
                                    channel.handle &&
                                    `/@${channel.handle}` ||
                                      `/c/${channel.id}`,
                                params: channel.handle ? {handle: `@${channel.handle}`} : undefined,
                            })
                        }
                    >
                        <div className="w-10 h-10 rounded-full overflow-hidden bg-muted dark:bg-gray-700 flex-shrink-0">
                            {channel.avatar ? (
                                <img
                                    src={channel.avatar}
                                    alt={channel.name}
                                    className="w-full h-full object-cover"
                                />
                            ) : (
                                <div className="w-full h-full flex items-center justify-center text-sm font-bold text-muted-foreground">
                                    {channel.name?.charAt(0)?.toUpperCase()}
                                </div>
                            )}
                        </div>

                        <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium truncate group-hover:text-primary transition-colors flex items-center gap-1">
                                {channel.name}
                                {channel.is_verified && (
                                    <CheckCircle className="w-3.5 h-3.5 text-info flex-shrink-0"/>
                                )}
                            </p>
                            <p className="text-xs text-muted-foreground">
                                {formatCount(channel.subscriber_count || 0)}{' '}
                                {t('channel.subscribers')}
                            </p>
                        </div>

                        <button
                            onClick={(e) => handleSubscribe(channel.id, e)}
                            className={`px-3 py-1 text-xs font-medium rounded-full transition-colors flex-shrink-0 ${
                                channel.is_subscribed
                                    ? 'bg-secondary text-secondary-foreground'
                                    : 'bg-primary text-primary-foreground hover:bg-primary/90'
                            }`}
                        >
                            {channel.is_subscribed ? (
                                <span className="flex items-center gap-1">
                                    <CheckCircle className="w-3 h-3"/>
                                    {t('channel.subscribed')}
                                </span>
                            ) : (
                                t('channel.subscribe')
                            )}
                        </button>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default RecommendedChannels;
