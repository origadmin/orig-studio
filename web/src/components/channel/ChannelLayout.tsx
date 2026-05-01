import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate, Link} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {Input} from '@/components/ui/input';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {Search, ExternalLink, Globe, Link2, Users, UserPlus, Loader2} from 'lucide-react';
import type {ChannelDetail, ChannelPlaylist} from '@/lib/api/channel';
import type {Media} from '@/lib/api/media';
import {
    useChannelVideos,
    useChannelPlaylists,
    useSubscribe,
    useUnsubscribe,
    useUpdateNotificationSetting,
} from '@/hooks/queries';
import {mediaApi} from '@/lib/api/media';
import {subscriptionApi, type SubscriptionListResponse} from '@/lib/api/subscription';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Spinner} from '@/components/ui/spinner';
import {formatDate} from '@/lib/format';
import {PAGINATION_CONFIG} from '@/config/pagination';
import ChannelHeader from './ChannelHeader';
import ChannelNav from './ChannelNav';
import VideoCard from './widgets/VideoCard';
import PlaylistCard from './widgets/PlaylistCard';
import RecommendedChannels from './widgets/RecommendedChannels';
import EmptyState from './widgets/EmptyState';

interface ChannelLayoutProps {
    channel: ChannelDetail;
    isOwner: boolean;
    isFromMeChannel?: boolean;
    isSubscribed?: boolean;
    subscriptionLoading?: boolean;
}

const ChannelLayout: React.FC<ChannelLayoutProps> = ({
    channel,
    isOwner,
    isFromMeChannel = false,
    isSubscribed = false,
    subscriptionLoading = false,
}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = useState('home');
    const [subscriberCount, setSubscriberCount] = useState(channel.subscriber_count || 0);
    const [subscribed, setSubscribed] = useState(isSubscribed);
    const [contentEmpty, setContentEmpty] = useState(false);

    const subscribeMutation = useSubscribe();
    const unsubscribeMutation = useUnsubscribe();
    const notificationMutation = useUpdateNotificationSetting();

    const channelToken = channel.friendly_token || channel.slug;

    const handleTabChange = (tab: string) => {
        setActiveTab(tab);
        setContentEmpty(false);
    };

    const handleContentEmptyChange = (empty: boolean) => {
        setContentEmpty(empty);
    };

    const handleSubscribe = () => {
        if (!channelToken) return;
        setSubscribed(true);
        setSubscriberCount((prev) => prev + 1);
        subscribeMutation.mutate(channelToken, {
            onError: () => {
                setSubscribed(false);
                setSubscriberCount((prev) => Math.max(0, prev - 1));
            },
        });
    };

    const handleUnsubscribe = async () => {
        if (!channelToken) return;
        setSubscribed(false);
        setSubscriberCount((prev) => Math.max(0, prev - 1));
        unsubscribeMutation.mutate(channelToken, {
            onError: () => {
                setSubscribed(true);
                setSubscriberCount((prev) => prev + 1);
            },
        });
    };

    const handleNotificationSettingChange = async (setting: string) => {
        if (!channelToken) return;
        await notificationMutation.mutateAsync({channelToken, setting});
    };

    return (
        <div className="channel-page min-h-screen bg-background">
            <div className="max-w-[1920px] mx-auto">
                <ChannelHeader
                    channel={channel}
                    isOwner={isOwner}
                    isFromMeChannel={isFromMeChannel}
                    isSubscribed={subscribed}
                    subscriberCount={subscriberCount}
                    subscribing={subscribeMutation.isPending || unsubscribeMutation.isPending}
                    onSubscribe={handleSubscribe}
                    onUnsubscribe={handleUnsubscribe}
                    onNotificationSettingChange={handleNotificationSettingChange}
                />

                <ChannelNav
                    activeTab={activeTab}
                    onTabChange={handleTabChange}
                    isOwner={isOwner}
                />

                <div className={`max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 ${contentEmpty ? '' : 'flex flex-col lg:flex-row gap-6'}`}>
                    <main className={contentEmpty ? 'w-full' : 'flex-1 min-w-0'}>
                        {activeTab === 'home' && (
                            <HomeTabContent
                                channelToken={channelToken}
                                channelId={channel.id}
                                isOwner={isOwner}
                                channelName={channel.name}
                                onTabChange={handleTabChange}
                                onEmptyChange={handleContentEmptyChange}
                            />
                        )}
                        {activeTab === 'videos' && (
                            <VideosTabContent
                                channelToken={channelToken}
                                channelId={channel.id}
                                isOwner={isOwner}
                                onEmptyChange={handleContentEmptyChange}
                            />
                        )}
                        {activeTab === 'playlists' && (
                            <PlaylistsTabContent
                                channelToken={channelToken}
                                channelId={channel.id}
                                isOwner={isOwner}
                                onEmptyChange={handleContentEmptyChange}
                            />
                        )}
                        {activeTab === 'community' && (
                            <CommunityTabContent
                                channelId={channel.id}
                                isOwner={isOwner}
                                channelName={channel.name}
                                onEmptyChange={handleContentEmptyChange}
                            />
                        )}
                        {activeTab === 'subscriptions' && isOwner && (
                            <SubscriptionsTabContent
                                onEmptyChange={handleContentEmptyChange}
                            />
                        )}
                        {activeTab === 'about' && (
                            <AboutTabContent
                                channel={channel}
                                isOwner={isOwner}
                                subscriberCount={subscriberCount}
                                onEmptyChange={handleContentEmptyChange}
                            />
                        )}
                    </main>

                    {!contentEmpty && (
                        <aside className="hidden lg:block w-80 flex-shrink-0">
                            <RecommendedChannels currentChannelId={channel.id}/>
                        </aside>
                    )}
                </div>

                <div className="lg:hidden max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-6">
                    <RecommendedChannels currentChannelId={channel.id}/>
                </div>
            </div>
        </div>
    );
};

// ================================
// Home Tab - Shows featured + latest videos from API
// ================================
const HomeTabContent: React.FC<{
    channelToken?: string;
    channelId?: string;
    isOwner: boolean;
    channelName?: string;
    onTabChange: (tab: string) => void;
    onEmptyChange?: (empty: boolean) => void;
}> = ({channelToken, channelId, isOwner, channelName, onTabChange, onEmptyChange}) => {
    const {t} = useTranslation();

    const {data: videosData, isLoading} = useChannelVideos(channelToken || null, {
        sort: 'newest',
        page_size: 8,
    });

    const videos = videosData?.items || [];

    React.useEffect(() => {
        if (!isLoading) {
            onEmptyChange?.(videos.length === 0);
        }
    }, [isLoading, videos.length, onEmptyChange]);

    if (isLoading) {
        return (
            <div className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
                        <div key={i} className="animate-pulse">
                            <div className="aspect-video bg-muted rounded-lg"/>
                            <div className="mt-2 space-y-2">
                                <div className="h-4 bg-muted rounded w-3/4"/>
                                <div className="h-3 bg-muted rounded w-1/2"/>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    if (videos.length === 0) {
        return <EmptyState type="home" isOwner={isOwner}/>;
    }

    return (
        <div className="space-y-8">
            <section>
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-lg font-semibold flex items-center gap-2">
                        <span>{t('home.latestVideos')}</span>
                    </h2>
                    <button
                        onClick={() => onTabChange('videos')}
                        className="text-sm text-primary hover:underline font-medium"
                    >
                        {t('home.viewAll')}
                    </button>
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {videos.map((video) => (
                        <VideoCard
                            key={video.id}
                            video={mapMediaToVideo(video)}
                            showChannelInfo={false}
                            isOwner={isOwner}
                            showProgress
                        />
                    ))}
                </div>
            </section>
        </div>
    );
};

// ================================
// Videos Tab - Full video list with sort and search
// ================================
const VideosTabContent: React.FC<{
    channelToken?: string;
    channelId?: string;
    isOwner: boolean;
    onEmptyChange?: (empty: boolean) => void;
}> = ({channelToken, channelId, isOwner, onEmptyChange}) => {
    const {t} = useTranslation();
    const [sortBy, setSortBy] = useState('newest');
    const [searchKeyword, setSearchKeyword] = useState('');

    const {data: videosData, isLoading} = useChannelVideos(
        channelToken || null,
        {
            sort: sortBy,
            keyword: searchKeyword || undefined,
            page_size: 20,
        }
    );

    const videos = videosData?.items || [];
    const total = videosData?.total || 0;

    React.useEffect(() => {
        if (!isLoading) {
            onEmptyChange?.(videos.length === 0 && !searchKeyword);
        }
    }, [isLoading, videos.length, searchKeyword, onEmptyChange]);

    const sortOptions = [
        {value: 'newest', label: t('channel.sortNewest') || 'Newest'},
        {value: 'popular', label: t('channel.sortPopular') || 'Most popular'},
        {value: 'oldest', label: t('channel.sortOldest') || 'Oldest'},
    ];

    if (isLoading) {
        return (
            <div className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {[1, 2, 3, 4, 5, 6, 7, 8].map((i) => (
                        <div key={i} className="animate-pulse">
                            <div className="aspect-video bg-muted rounded-lg"/>
                            <div className="mt-2 space-y-2">
                                <div className="h-4 bg-muted rounded w-3/4"/>
                                <div className="h-3 bg-muted rounded w-1/2"/>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    if (videos.length === 0 && !searchKeyword) {
        return <EmptyState type="videos" isOwner={isOwner}/>;
    }

    return (
        <div className="space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-4">
                <h2 className="text-lg font-semibold">
                    {t('channel.allVideos')} ({total})
                </h2>
                <div className="flex items-center gap-3">
                    <div className="relative">
                        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"/>
                        <Input
                            placeholder={t('channel.searchVideos') || 'Search videos...'}
                            value={searchKeyword}
                            onChange={(e) => setSearchKeyword(e.target.value)}
                            className="pl-9 h-9 w-48 sm:w-64"
                        />
                    </div>
                    <Select value={sortBy} onValueChange={setSortBy}>
                        <SelectTrigger className="w-[160px] h-9">
                            <SelectValue placeholder={t('channel.sortBy')}/>
                        </SelectTrigger>
                        <SelectContent>
                            {sortOptions.map((opt) => (
                                <SelectItem key={opt.value} value={opt.value}>
                                    {opt.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>
            </div>

            {videos.length === 0 && searchKeyword ? (
                <div className="text-center py-12 text-muted-foreground">
                    <Search className="w-12 h-12 mx-auto mb-3 opacity-50"/>
                    <p>{t('channel.noSearchResults') || 'No videos found'}</p>
                </div>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {videos.map((video) => (
                        <VideoCard
                            key={video.id}
                            video={mapMediaToVideo(video)}
                            showChannelInfo={false}
                            isOwner={isOwner}
                            showProgress
                            onEdit={(id) => console.log('Edit video:', id)}
                            onViewStats={(id) => console.log('View stats:', id)}
                        />
                    ))}
                </div>
            )}

            {videos.length < total && (
                <div className="flex justify-center pt-4">
                    <Button
                        variant="outline"
                        size="sm"
                    >
                        {t('channel.loadMore') || 'Load more'}
                    </Button>
                </div>
            )}
        </div>
    );
};

// ================================
// Playlists Tab - Channel playlists from API
// ================================
const PlaylistsTabContent: React.FC<{
    channelToken?: string;
    channelId?: string;
    isOwner: boolean;
    onEmptyChange?: (empty: boolean) => void;
}> = ({channelToken, channelId, isOwner, onEmptyChange}) => {
    const {t} = useTranslation();

    const {data: playlistsData, isLoading} = useChannelPlaylists(channelToken || null);

    const playlists: ChannelPlaylist[] = (playlistsData as any)?.items || [];
    const total = (playlistsData as any)?.total || 0;

    React.useEffect(() => {
        if (!isLoading) {
            onEmptyChange?.(playlists.length === 0);
        }
    }, [isLoading, playlists.length, onEmptyChange]);

    if (isLoading) {
        return (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                    <div key={i} className="animate-pulse p-4 border rounded-lg">
                        <div className="aspect-video bg-muted rounded-md mb-3"/>
                        <div className="h-4 bg-muted rounded w-3/4 mb-2"/>
                        <div className="h-3 bg-muted rounded w-1/2"/>
                    </div>
                ))}
            </div>
        );
    }

    if (playlists.length === 0 && !isOwner) {
        return <EmptyState type="playlists" isOwner={isOwner}/>;
    }

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold">
                    {t('channel.playlists')} ({total})
                </h2>
                {isOwner && (
                    <Button size="sm" onClick={() => console.log('Create playlist')}>
                        + {t('channel.createPlaylist')}
                    </Button>
                )}
            </div>

            {playlists.length === 0 ? (
                <EmptyState type="playlists" isOwner={isOwner}/>
            ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                    {playlists.map((playlist) => (
                        <PlaylistCard key={playlist.id} playlist={playlist} isOwner={isOwner}/>
                    ))}
                </div>
            )}
        </div>
    );
};

// ================================
// Community Tab - Placeholder for future implementation
// ================================
const CommunityTabContent: React.FC<{
    channelId?: string;
    isOwner: boolean;
    channelName?: string;
    onEmptyChange?: (empty: boolean) => void;
}> = ({channelId, isOwner, channelName, onEmptyChange}) => {
    const {t} = useTranslation();

    React.useEffect(() => {
        onEmptyChange?.(true);
    }, [onEmptyChange]);

    return (
        <div className="space-y-4">
            {isOwner && (
                <div className="mb-6 p-4 border border-dashed rounded-lg hover:border-primary/50 transition-colors cursor-pointer group">
                    <button className="w-full text-left text-muted-foreground group-hover:text-foreground transition-colors flex items-center gap-2">
                        <span className="text-lg">✏️</span>
                        <span>{t('channel.createNewPost')}...</span>
                    </button>
                </div>
            )}

            <EmptyState type="community" isOwner={isOwner}/>
        </div>
    );
};

// ================================
// About Tab - Channel description, stats, links, tags
// ================================
const AboutTabContent: React.FC<{
    channel: ChannelDetail;
    isOwner: boolean;
    subscriberCount?: number;
    onEmptyChange?: (empty: boolean) => void;
}> = ({channel, isOwner, subscriberCount = 0, onEmptyChange}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();

    const hasContent = !!(channel.description || channel.links?.length || channel.tags?.length);

    React.useEffect(() => {
        onEmptyChange?.(!hasContent);
    }, [hasContent, onEmptyChange]);

    const formatCount = (num: number): string => {
        if (!num) return '0';
        if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
        if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
        return num.toString();
    };

    const stats = [
        {label: t('channel.subscribers'), value: formatCount(subscriberCount), icon: '👥'},
        {label: t('channel.videoCount'), value: String(channel.video_count || channel.media_count || 0), icon: '🎬'},
        {label: t('channel.views'), value: formatCount(channel.total_views || 0), icon: '👁️'},
        {label: t('channel.joinDate'), value: channel.create_time ? new Date(channel.create_time).toLocaleDateString() : t('channel.notAvailable') || 'N/A', icon: '📅'},
    ];

    const links = channel.links || [];
    const tags = channel.tags || [];

    const getLinkIcon = (type: string, platform?: string) => {
        switch (platform?.toLowerCase() || type) {
            case 'website':
                return <Globe className="w-5 h-5"/>;
            default:
                return <Link2 className="w-5 h-5"/>;
        }
    };

    return (
        <div className="space-y-8">
            <section>
                <h2 className="text-lg font-semibold mb-4">{t('channel.description')}</h2>
                <div className="p-4 sm:p-6 bg-card rounded-lg border">
                    {channel.description ? (
                        <p className="text-foreground whitespace-pre-wrap leading-relaxed">
                            {channel.description}
                        </p>
                    ) : (
                        <p className="text-muted-foreground/60 italic">
                            {t('channel.noDescription') || 'This channel has no description yet.'}
                        </p>
                    )}
                    {isOwner && (
                        <button
                            onClick={() =>
                                navigate({
                                    to: '/u/$id',
                                    params: {id: channel?.friendly_token || channel?.slug || ''},
                                    search: {tab: 'branding'},
                                })
                            }
                            className="mt-3 text-sm text-primary hover:underline inline-flex items-center gap-1"
                        >
                            ✏️ {channel.description ? t('channel.editDescription') : (t('channel.addDescription') || 'Add description')}
                        </button>
                    )}
                </div>
            </section>

            <section>
                <h2 className="text-lg font-semibold mb-4">{t('channel.stats')}</h2>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    {stats.map((stat, idx) => (
                        <div
                            key={idx}
                            className="p-4 bg-card rounded-lg border text-center hover:border-primary/30 transition-colors"
                        >
                            <p className="text-2xl mb-1">{stat.icon}</p>
                            <p className="text-xl sm:text-2xl font-bold text-foreground">
                                {stat.value}
                            </p>
                            <p className="text-xs sm:text-sm text-muted-foreground mt-1">
                                {stat.label}
                            </p>
                        </div>
                    ))}
                </div>
            </section>

            {links.length > 0 && (
                <section>
                    <h2 className="text-lg font-semibold mb-4">{t('channel.links')}</h2>
                    <div className="space-y-2">
                        {links.map((link, idx) => (
                            <a
                                key={idx}
                                href={link.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="flex items-center gap-3 p-3 rounded-lg border hover:bg-accent hover:border-primary/30 transition-all group"
                            >
                                <span className="w-8 h-8 flex items-center justify-center text-muted-foreground group-hover:text-primary transition-colors">
                                    {getLinkIcon(link.type, link.platform)}
                                </span>
                                <div className="flex-1 min-w-0">
                                    <p className="font-medium truncate group-hover:text-primary transition-colors">
                                        {link.title}
                                    </p>
                                    <p className="text-xs text-muted-foreground truncate">
                                        {link.url}
                                    </p>
                                </div>
                                <ExternalLink className="w-4 h-4 text-muted-foreground opacity-50 group-hover:opacity-100 transition-opacity"/>
                            </a>
                        ))}
                    </div>
                </section>
            )}

            {tags.length > 0 && (
                <section>
                    <h2 className="text-lg font-semibold mb-4">{t('channel.tags')}</h2>
                    <div className="flex flex-wrap gap-2">
                        {tags.map((tag, idx) => (
                            <span
                                key={idx}
                                className="px-3 py-1.5 bg-secondary text-secondary-foreground rounded-full text-sm hover:bg-primary/10 hover:text-primary transition-colors cursor-default"
                            >
                                #{tag}
                            </span>
                        ))}
                    </div>
                </section>
            )}

            <section className="p-4 bg-card rounded-lg border border-dashed">
                <p className="text-sm text-muted-foreground text-center">
                    {t('channel.channelId')}: <code className="bg-background px-2 py-0.5 rounded text-xs">{channel.id}</code>
                </p>
            </section>
        </div>
    );
};

// ================================
// Subscriptions Tab - Shows subscribed channels (owner only)
// ================================
const SubscriptionsTabContent: React.FC<{
    onEmptyChange?: (empty: boolean) => void;
}> = ({onEmptyChange}) => {
    const {t} = useTranslation();
    const [activeSubTab, setActiveSubTab] = useState<'subscriptions' | 'followers'>('subscriptions');
    const [subscriptions, setSubscriptions] = useState<SubscriptionListResponse['items']>([]);
    const [followers, setFollowers] = useState<SubscriptionListResponse['items']>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);

    const fetchData = async (tab: 'subscriptions' | 'followers', pageNum: number) => {
        try {
            setLoading(true);
            setError(null);

            let response: SubscriptionListResponse;
            if (tab === 'subscriptions') {
                response = await subscriptionApi.getSubscriptions({page: pageNum, page_size: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE});
                setSubscriptions(prev => pageNum === 1 ? response.items : [...prev, ...response.items]);
                setHasMore(response.items.length === PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
            } else {
                response = await subscriptionApi.getFollowers({page: pageNum, page_size: PAGINATION_CONFIG.DEFAULT_PAGE_SIZE});
                setFollowers(prev => pageNum === 1 ? response.items : [...prev, ...response.items]);
                setHasMore(response.items.length === PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);
            }
        } catch (err) {
            setError('Failed to fetch data');
            console.error('Failed to fetch subscription data:', err);
        } finally {
            setLoading(false);
        }
    };

    React.useEffect(() => {
        setPage(1);
        setSubscriptions([]);
        setFollowers([]);
        fetchData(activeSubTab, 1);
    }, [activeSubTab]);

    React.useEffect(() => {
        if (!loading) {
            const list = activeSubTab === 'subscriptions' ? subscriptions : followers;
            onEmptyChange?.(list.length === 0);
        }
    }, [loading, activeSubTab, subscriptions, followers, onEmptyChange]);

    const handleLoadMore = () => {
        if (!loading && hasMore) {
            const nextPage = page + 1;
            setPage(nextPage);
            fetchData(activeSubTab, nextPage);
        }
    };

    const list = activeSubTab === 'subscriptions' ? subscriptions : followers;

    if (loading && page === 1) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    if (error) {
        return (
            <div className="text-center py-12 text-muted-foreground">
                <p>{error}</p>
                <Button variant="outline" className="mt-4" onClick={() => fetchData(activeSubTab, 1)}>
                    {t('common.retry')}
                </Button>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center gap-3">
                <Users size={24} className="text-emerald-600"/>
                <h2 className="text-lg font-semibold">{t('subscriptions.title')}</h2>
            </div>

            <div className="flex border-b dark:border-gray-700">
                <button
                    className={`px-4 py-3 font-medium text-sm ${
                        activeSubTab === 'subscriptions'
                            ? 'border-b-2 border-emerald-600 text-emerald-600'
                            : 'text-gray-500 dark:text-muted-foreground hover:text-gray-900 dark:hover:text-gray-100'
                    }`}
                    onClick={() => setActiveSubTab('subscriptions')}
                >
                    {t('subscriptions.subscriptions')}
                </button>
                <button
                    className={`px-4 py-3 font-medium text-sm ${
                        activeSubTab === 'followers'
                            ? 'border-b-2 border-emerald-600 text-emerald-600'
                            : 'text-gray-500 dark:text-muted-foreground hover:text-gray-900 dark:hover:text-gray-100'
                    }`}
                    onClick={() => setActiveSubTab('followers')}
                >
                    {t('subscriptions.followers')}
                </button>
            </div>

            {list.length === 0 ? (
                <div className="text-center py-16 text-muted-foreground">
                    <Users size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>{activeSubTab === 'subscriptions' ? t('subscriptions.noSubscriptions') : t('subscriptions.noFollowers')}</p>
                </div>
            ) : (
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
                                    <p className="text-xs text-gray-500 dark:text-muted-foreground">
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
                                    t('common.loading')
                                )}
                            </Button>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

// ================================
// Utility: Map Media API type to VideoCard-compatible type
// ================================
function mapMediaToVideo(media: any): {
    id: string;
    short_token?: string;
    title: string;
    thumbnail?: string;
    duration?: number;
    view_count?: number;
    published_at?: string;
    progress?: number;
} {
    return {
        id: media.id,
        short_token: media.short_token,
        title: media.title,
        thumbnail: media.thumbnail || media.poster,
        duration: media.duration,
        view_count: media.view_count,
        published_at: media.published_at || media.create_time,
        progress: 0,
    };
}

export default ChannelLayout;
