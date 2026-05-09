import React, {useState, useEffect, useCallback, useMemo} from 'react';
import {useParams, Link} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {Badge} from '@/components/ui/badge';
import {Button} from '@/components/ui/button';
import {formatDuration, formatViews, formatDate} from '@/lib/format';
import {userApi, type User} from '@/lib/api/user';
import {subscriptionApi, type SubscriptionListResponse} from '@/lib/api/subscription';
import {channelApi, type Channel, type ChannelList} from '@/lib/api/channel';
import {mediaApi, normalizeMediaList} from '@/lib/api/media';
import {playlistApi, type Playlist} from '@/lib/api/playlist';
import {articleApi, type Article} from '@/lib/api/article';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import ErrorPage from '@/components/common/ErrorPage';
import SubscribeButton from '@/components/common/SubscribeButton';
import {useAuth} from '@/hooks/useAuth';
import {Play, Info, Tv, ListVideo, Shield, Eye, Calendar, Edit3, BadgeCheck, Users, Film, Home, FileText, Settings} from 'lucide-react';

function formatCount(num: number): string {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
}

interface ProfilePageProps {
    userId?: string;
}

const ProfilePage: React.FC<ProfilePageProps> = ({userId: propUserId}) => {
    const params = useParams({strict: false}) as {id?: string; username?: string; handle?: string};
    const {t} = useTranslation();
    const {user: currentUser, isAuthenticated} = useAuth();
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState('home');
    const [descriptionExpanded, setDescriptionExpanded] = useState(false);

    const [followers, setFollowers] = useState<SubscriptionListResponse['items']>([]);
    const [followersLoading, setFollowersLoading] = useState(false);
    const [followersTotal, setFollowersTotal] = useState(0);
    const [followersPage, setFollowersPage] = useState(1);

    const [channels, setChannels] = useState<Channel[]>([]);
    const [channelsLoading, setChannelsLoading] = useState(false);

    const [videos, setVideos] = useState<any[]>([]);
    const [videosLoading, setVideosLoading] = useState(false);

    const [playlists, setPlaylists] = useState<Playlist[]>([]);
    const [playlistsLoading, setPlaylistsLoading] = useState(false);

    const [articles, setArticles] = useState<Article[]>([]);
    const [articlesLoading, setArticlesLoading] = useState(false);

    const isMe = useMemo(() => {
        if (!user) return false;
        if (user.is_me !== undefined) return user.is_me;
        if (!isAuthenticated || !currentUser) return false;
        return String(currentUser.id) === String(user.id) || currentUser.username === user.username;
    }, [user, isAuthenticated, currentUser]);

    const fetchFollowers = useCallback(async (page: number = 1) => {
        if (!isMe) return;
        try {
            setFollowersLoading(true);
            const res = await subscriptionApi.getFollowers({page, page_size: 20});
            if (page === 1) {
                setFollowers(res.items || []);
            } else {
                setFollowers(prev => [...prev, ...(res.items || [])]);
            }
            setFollowersTotal(res.total || 0);
            setFollowersPage(page);
        } catch (err) {
            console.error('Failed to fetch followers:', err);
        } finally {
            setFollowersLoading(false);
        }
    }, [isMe]);

    const fetchChannels = useCallback(async () => {
        if (!user) return;
        try {
            setChannelsLoading(true);
            let res: ChannelList;
            if (isMe) {
                res = await channelApi.getMyChannels({page_size: 50});
            } else {
                res = await channelApi.get({user_id: String(user.id)}) as ChannelList;
            }
            setChannels(res.items || []);
        } catch (err) {
            console.error('Failed to fetch channels:', err);
        } finally {
            setChannelsLoading(false);
        }
    }, [user]);

    const fetchVideos = useCallback(async () => {
        if (!user) return;
        try {
            setVideosLoading(true);
            const res = await mediaApi.list({user_id: user.id || undefined, page_size: 20});
            const items = (res as any)?.data?.items || (res as any)?.items || [];
            setVideos(normalizeMediaList(items));
        } catch (err) {
            console.error('Failed to fetch videos:', err);
        } finally {
            setVideosLoading(false);
        }
    }, [user]);

    const fetchPlaylists = useCallback(async () => {
        if (!user) return;
        try {
            setPlaylistsLoading(true);
            const res = await playlistApi.getMyPlaylists({page_size: 20});
            const data = res as any;
            setPlaylists(data?.items || []);
        } catch (err) {
            console.error('Failed to fetch playlists:', err);
        } finally {
            setPlaylistsLoading(false);
        }
    }, [user]);

    const fetchArticles = useCallback(async () => {
        if (!user) return;
        try {
            setArticlesLoading(true);
            const res = await articleApi.list({user_id: String(user.id), page_size: 20});
            const data = res as any;
            setArticles(data?.data?.items || data?.items || []);
        } catch (err) {
            console.error('Failed to fetch articles:', err);
        } finally {
            setArticlesLoading(false);
        }
    }, [user]);

    useEffect(() => {
        if (!user) return;
        if ((activeTab === 'home' || activeTab === 'channels') && channels.length === 0) {
            fetchChannels();
        }
        if ((activeTab === 'home' || activeTab === 'videos') && videos.length === 0) {
            fetchVideos();
        }
        if ((activeTab === 'home' || activeTab === 'articles') && articles.length === 0) {
            fetchArticles();
        }
        if (activeTab === 'followers' && isMe && followers.length === 0) {
            fetchFollowers(1);
        }
        if (activeTab === 'playlists' && playlists.length === 0) {
            fetchPlaylists();
        }
    }, [activeTab, user?.id, isMe, followers.length, channels.length, videos.length, articles.length, playlists.length]);

    useEffect(() => {
        if (user && channels.length === 0 && !channelsLoading) {
            fetchChannels();
        }
    }, [user?.id, channels.length, channelsLoading]);

    useEffect(() => {
        const fetchUser = async () => {
            try {
                setLoading(true);
                setError(null);
                let userResponse;
                const resolvedUsername = params.handle
                    ? params.handle.startsWith('@') ? params.handle.slice(1) : params.handle
                    : params.username;
                if (propUserId) {
                    userResponse = await userApi.get(propUserId);
                } else if (resolvedUsername) {
                    userResponse = await userApi.getByUsername(resolvedUsername);
                } else if (params.id) {
                    userResponse = await userApi.get(params.id);
                } else {
                    throw new Error('No user identifier provided');
                }
                const raw = userResponse as any;
                const userData = raw?.user ?? raw?.data?.user ?? raw?.data ?? raw;
                setUser(userData);
            } catch (err: any) {
                if (err.response && err.response.status === 404) {
                    setUser(null);
                } else {
                    setError('Failed to fetch user data');
                }
                console.error('Failed to fetch user data:', err);
            } finally {
                setLoading(false);
            }
        };
        fetchUser();
    }, [propUserId, params.id, params.username, params.handle]);

    const getPrivacyBadgeVariant = (privacy?: string) => {
        switch (privacy?.toUpperCase()) {
            case 'PUBLIC': return 'default';
            case 'PRIVATE': return 'secondary';
            case 'UNLISTED': return 'outline';
            default: return 'default';
        }
    };

    const getPrivacyLabel = (privacy?: string) => {
        switch (privacy?.toUpperCase()) {
            case 'PUBLIC': return t('common.public');
            case 'PRIVATE': return t('common.private');
            case 'UNLISTED': return t('common.unlisted');
            default: return t('common.public');
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-background">
                <div className="max-w-[1920px] mx-auto">
                    <div className="relative">
                        <div className="w-full h-[150px] sm:h-[200px] md:h-[250px] bg-muted animate-pulse"/>
                    </div>
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <div className="flex flex-col sm:flex-row items-start sm:items-end gap-4 sm:gap-6 -mt-10 sm:-mt-14 relative z-10 pb-4">
                            <div className="w-20 h-20 sm:w-28 sm:h-28 md:w-[120px] md:h-[120px] rounded-full bg-muted animate-pulse border-4 border-background flex-shrink-0"/>
                            <div className="flex-1 pt-2 sm:pt-4 min-w-0 space-y-3">
                                <div className="h-7 bg-muted rounded w-1/3 animate-pulse"/>
                                <div className="h-4 bg-muted rounded w-1/2 animate-pulse"/>
                                <div className="h-4 bg-muted rounded w-2/3 animate-pulse"/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    if (!user) {
        return (
            <ErrorPage
                statusCode={404}
                title={t('profile.userNotFound')}
                message={t('error.404Message')}
            />
        );
    }

    const description = user.bio || '';
    const subCount = user.subscriber_count || 0;
    const viewCount = user.total_views || 0;
    const videoCount = channels.reduce((sum, ch) => sum + (ch.media_count || 0), 0);

    const tabs = [
        {id: 'home', label: t('channel.tabHome') || 'Home', icon: Home},
        {id: 'channels', label: t('profile.tabChannels'), icon: Tv},
        {id: 'videos', label: t('channel.tabVideos') || 'Videos', icon: Film},
        {id: 'articles', label: t('profile.tabArticles') || 'Articles', icon: FileText},
        {id: 'playlists', label: t('channel.tabPlaylists') || 'Playlists', icon: ListVideo},
        ...(isMe ? [{id: 'followers', label: t('profile.tabFollowers'), icon: Users}] : []),
        {id: 'about', label: t('channel.tabAbout') || 'About', icon: Info},
    ];

    return (
        <div className="min-h-screen bg-background">
            <div className="max-w-[1920px] mx-auto">
                <div className="relative">
                    <div className="relative group">
                        {user.cover ? (
                            <div
                                className="w-full h-[150px] sm:h-[200px] md:h-[250px] bg-cover bg-center"
                                style={{backgroundImage: `url(${getImageUrl(user.cover, 'cover')})`}}
                            />
                        ) : (
                            <div className="w-full h-[150px] sm:h-[200px] md:h-[250px] bg-gradient-to-r from-blue-600 via-purple-600 to-pink-500"/>
                        )}
                        <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent"/>
                        {isMe && (
                            <Link
                                to="/me/channels"
                                className="absolute top-4 right-4 z-20 px-3 py-1.5 bg-black/60 hover:bg-black/80 text-white text-sm rounded-lg backdrop-blur-sm transition-all opacity-0 group-hover:opacity-100 flex items-center gap-1.5"
                            >
                                <Edit3 className="w-4 h-4"/>
                                {t('profile.customizeProfile')}
                            </Link>
                        )}
                    </div>

                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <div className="flex flex-col sm:flex-row items-start sm:items-end gap-4 sm:gap-6 -mt-10 sm:-mt-14 relative z-10 pb-4">
                            <Avatar className="w-20 h-20 sm:w-28 sm:h-28 md:w-[120px] md:h-[120px] border-4 border-background shadow-lg flex-shrink-0">
                                {user.avatar ? (
                                    <AvatarImage src={getImageUrl(user.avatar, 'avatar')} loading="lazy"
                                                 onError={(e) => handleImageError(e, 'avatar')}/>
                                ) : null}
                                <AvatarFallback className="text-3xl sm:text-4xl md:text-5xl font-bold bg-muted text-muted-foreground">
                                    {user.username ? user.username.charAt(0).toUpperCase() : 'U'}
                                </AvatarFallback>
                            </Avatar>

                            <div className="flex-1 pt-2 sm:pt-4 min-w-0">
                                <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold text-foreground mb-1.5 truncate flex items-center gap-2">
                                    {user.username}
                                    {user.is_verified && (
                                        <BadgeCheck className="w-5 h-5 sm:w-6 sm:h-6 text-info flex-shrink-0"/>
                                    )}
                                </h1>

                                <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground mb-2">
                                    <span className="flex items-center gap-1">
                                        <Users className="w-3.5 h-3.5"/>
                                        {formatCount(subCount)} {t('common.followers')}
                                    </span>
                                    <span className="flex items-center gap-1">
                                        <Film className="w-3.5 h-3.5"/>
                                        {videoCount} {t('profile.videosCount')}
                                    </span>
                                    {viewCount > 0 && (
                                        <span className="flex items-center gap-1">
                                            <Eye className="w-3.5 h-3.5"/>
                                            {formatCount(viewCount)} {t('common.views')}
                                        </span>
                                    )}
                                </div>

                                {description ? (
                                    <div className="text-sm text-muted-foreground">
                                        <p className={descriptionExpanded ? '' : 'line-clamp-2 max-w-2xl'}>
                                            {description}
                                        </p>
                                        {description.length > 120 && (
                                            <button
                                                onClick={() => setDescriptionExpanded(!descriptionExpanded)}
                                                className="text-primary hover:underline text-sm mt-0.5"
                                            >
                                                {descriptionExpanded ? (t('channel.showLess') || 'Show less') : (t('channel.showMore') || 'Show more')}
                                            </button>
                                        )}
                                    </div>
                                ) : null}
                            </div>

                            <div className="flex items-center gap-2 pt-2 sm:pt-4 flex-shrink-0 flex-wrap">
                                {isMe ? (
                                    <Link to="/me/channels">
                                        <Button variant="outline">
                                            <Tv className="w-4 h-4 mr-1"/>
                                            {t('nav.channelManagement')}
                                        </Button>
                                    </Link>
                                ) : (
                                    <SubscribeButton
                                        channelId={user.channel_id || user.id || ''}
                                        initialSubscriberCount={user.subscriber_count || 0}
                                    />
                                )}
                            </div>
                        </div>
                    </div>
                </div>

                <div className="border-b border-border sticky top-16 z-30 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
                    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                        <nav className="flex gap-1 sm:gap-2 overflow-x-auto scrollbar-hide -mb-px" role="tablist">
                            {tabs.map((tab) => {
                                const Icon = tab.icon;
                                const isActive = activeTab === tab.id;
                                return (
                                    <button
                                        key={tab.id}
                                        onClick={() => setActiveTab(tab.id)}
                                        role="tab"
                                        aria-selected={isActive}
                                        className={`relative py-3 sm:py-4 px-3 sm:px-4 font-medium text-sm whitespace-nowrap transition-all duration-200 group flex items-center gap-1.5 ${
                                            isActive
                                                ? 'text-primary'
                                                : 'text-muted-foreground hover:text-foreground'
                                        }`}
                                    >
                                        <Icon className={`w-4 h-4 transition-colors ${
                                            isActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'
                                        }`}/>
                                        <span className="relative z-10">{tab.label}</span>
                                        {isActive && (
                                            <span
                                                className="absolute bottom-0 left-0 h-0.5 bg-primary rounded-full transition-all duration-300"
                                                style={{
                                                    width: 'calc(100% + 8px)',
                                                    left: '-4px',
                                                }}
                                            />
                                        )}
                                        {!isActive && (
                                            <span
                                                className="absolute bottom-0 left-0 h-0.5 bg-transparent group-hover:bg-muted-foreground/30 rounded-full transition-all duration-200"
                                                style={{
                                                    width: 'calc(100% + 8px)',
                                                    left: '-4px',
                                                }}
                                            />
                                        )}
                                    </button>
                                );
                            })}
                        </nav>
                    </div>
                </div>

                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                    {activeTab === 'home' && (
                        (channelsLoading || videosLoading) && channels.length === 0 && videos.length === 0 ? (
                            <div className="space-y-8">
                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                    {[1, 2, 3].map((i) => (
                                        <div key={i} className="animate-pulse">
                                            <div className="h-24 bg-muted rounded-lg"/>
                                            <div className="mt-2 space-y-2">
                                                <div className="h-4 bg-muted rounded w-3/4"/>
                                                <div className="h-3 bg-muted rounded w-1/2"/>
                                            </div>
                                        </div>
                                    ))}
                                </div>
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
                        ) : channels.length === 0 && videos.length === 0 ? (
                            <div className="text-center py-16 text-muted-foreground">
                                <Home className="w-16 h-16 mx-auto mb-4 opacity-30"/>
                                <p className="text-lg font-medium mb-2">{isMe ? t('profile.noChannelsSelf') : t('profile.noChannels')}</p>
                                {isMe && (
                                    <Link to="/me/channels" className="text-primary hover:underline text-sm">
                                        {t('profile.createChannel') || 'Create a channel'}
                                    </Link>
                                )}
                            </div>
                        ) : (
                            <div className="space-y-8">
                                {channels.length > 0 && (
                                    <section>
                                        <div className="flex items-center justify-between mb-4">
                                            <h2 className="text-lg font-semibold">{t('profile.tabChannels')}</h2>
                                            <button
                                                onClick={() => setActiveTab('channels')}
                                                className="text-sm text-primary hover:underline font-medium"
                                            >
                                                {t('home.viewAll') || 'View all'}
                                            </button>
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                            {channels.slice(0, 3).map(ch => (
                                                <Link key={ch.id} to="/c/$id" params={{id: ch.short_token || ch.id}} className="group">
                                                    <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1 border border-gray-100 dark:border-gray-700">
                                                        {ch.banner && (
                                                            <div className="h-20 bg-cover bg-center relative"
                                                                 style={{backgroundImage: `url(${getImageUrl(ch.banner, 'cover')})`}}>
                                                                <div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/20"/>
                                                            </div>
                                                        )}
                                                        <div className="p-3">
                                                            <div className="flex items-center gap-3">
                                                                <Avatar className="w-10 h-10 border-2 border-white dark:border-gray-700 shadow-sm flex-shrink-0">
                                                                    <AvatarImage src={getImageUrl(ch.avatar, 'avatar')} loading="lazy"
                                                                                 onError={(e) => handleImageError(e, 'avatar')}/>
                                                                    <AvatarFallback className="text-sm">{ch.name ? ch.name.charAt(0).toUpperCase() : 'C'}</AvatarFallback>
                                                                </Avatar>
                                                                <div className="flex-1 min-w-0">
                                                                    <h3 className="font-semibold text-sm text-slate-900 dark:text-white group-hover:text-emerald-600 transition-colors truncate">{ch.name}</h3>
                                                                    <p className="text-xs text-slate-500 dark:text-muted-foreground">{ch.media_count || 0} {t('profile.videosCount')}</p>
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </Link>
                                            ))}
                                        </div>
                                    </section>
                                )}
                                {videos.length > 0 && (
                                    <section>
                                        <div className="flex items-center justify-between mb-4">
                                            <h2 className="text-lg font-semibold">{t('home.latestVideos') || 'Latest Videos'}</h2>
                                            <button
                                                onClick={() => setActiveTab('videos')}
                                                className="text-sm text-primary hover:underline font-medium"
                                            >
                                                {t('home.viewAll') || 'View all'}
                                            </button>
                                        </div>
                                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                                            {videos.slice(0, 8).map(video => (
                                                <Link key={video.id} to="/watch" search={{v: video.short_token}} className="group">
                                                    <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                                        <div className="relative aspect-video">
                                                            {video.thumbnail ? (
                                                                <img src={video.thumbnail} alt={video.title}
                                                                     className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                                            ) : (
                                                                <div className="w-full h-full bg-muted dark:bg-gray-700 flex items-center justify-center">
                                                                    <Play className="w-10 h-10 text-muted-foreground"/>
                                                                </div>
                                                            )}
                                                            {video.duration != null ? (
                                                                <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                                            ) : null}
                                                        </div>
                                                        <div className="p-3">
                                                            <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                                            <p className="text-xs text-slate-500 dark:text-muted-foreground mt-2">{formatViews(video.view_count)} {t('common.views')}</p>
                                                        </div>
                                                    </div>
                                                </Link>
                                            ))}
                                        </div>
                                    </section>
                                )}
                            </div>
                        )
                    )}

                    {activeTab === 'channels' && (
                        <>
                            {isMe && (
                                <div className="flex justify-end mb-4">
                                    <Link to="/me/channels">
                                        <Button variant="outline" size="sm">
                                            <Settings className="w-4 h-4 mr-1"/>
                                            {t('nav.channelManagement')}
                                        </Button>
                                    </Link>
                                </div>
                            )}
                            {channelsLoading && channels.length === 0 ? (
                                <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
                            ) : channels.length > 0 ? (
                                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                                {channels.map(ch => (
                                    <Link key={ch.id} to="/c/$id" params={{id: ch.short_token || ch.id}} className="group">
                                        <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1 border border-gray-100 dark:border-gray-700">
                                            {ch.banner && (
                                                <div className="h-24 bg-cover bg-center relative"
                                                     style={{backgroundImage: `url(${getImageUrl(ch.banner, 'cover')})`}}>
                                                    <div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/20"/>
                                                </div>
                                            )}
                                            <div className="p-4">
                                                <div className="flex items-start gap-4">
                                                    <Avatar className="w-16 h-16 border-2 border-white dark:border-gray-700 shadow-sm -mt-8 relative z-10">
                                                        <AvatarImage src={getImageUrl(ch.avatar, 'avatar')} loading="lazy"
                                                                     onError={(e) => handleImageError(e, 'avatar')}/>
                                                        <AvatarFallback className="text-lg">{ch.name ? ch.name.charAt(0).toUpperCase() : 'C'}</AvatarFallback>
                                                    </Avatar>
                                                    <div className="flex-1 min-w-0">
                                                        <div className="flex items-center gap-2">
                                                            <h3 className="font-semibold text-slate-900 dark:text-white group-hover:text-emerald-600 transition-colors truncate">{ch.name}</h3>
                                                            {ch.is_verified && (
                                                                <Badge variant="default" className="bg-emerald-500 text-xs px-1.5 py-0">{t('common.verified')}</Badge>
                                                            )}
                                                        </div>
                                                    </div>
                                                </div>
                                                {ch.description && (
                                                    <p className="text-sm text-slate-500 dark:text-muted-foreground mt-3 line-clamp-2">{ch.description}</p>
                                                )}
                                                <div className="flex items-center gap-3 mt-3 text-xs text-slate-500 dark:text-muted-foreground flex-nowrap">
                                                    <span className="whitespace-nowrap">{ch.media_count || 0} {t('profile.videosCount')}</span>
                                                    {ch.article_count !== undefined && (
                                                        <span className="whitespace-nowrap">{ch.article_count} {t('profile.articlesCount')}</span>
                                                    )}
                                                    <span className="whitespace-nowrap">{formatViews(ch.subscriber_count || 0)} {t('profile.subscribersCount')}</span>
                                                </div>
                                                <div className="flex items-center justify-between mt-3">
                                                    <Badge variant={getPrivacyBadgeVariant(ch.privacy)} className="text-xs">
                                                        <Shield className="w-3 h-3 mr-1"/>
                                                        {getPrivacyLabel(ch.privacy)}
                                                    </Badge>
                                                    {isMe && ch.short_token && (
                                                        <Link to="/c/$id" params={{id: ch.short_token}}
                                                              onClick={(e) => e.stopPropagation()}
                                                              className="text-xs text-emerald-600 hover:text-emerald-700 dark:text-emerald-400 dark:hover:text-emerald-300 font-medium flex items-center gap-1">
                                                            {t('profile.manageChannel')}
                                                        </Link>
                                                    )}
                                                </div>
                                            </div>
                                        </div>
                                    </Link>
                                ))}
                            </div>
                        ) : (
                            <div className="text-center py-12 text-muted-foreground">
                                {isMe ? t('profile.noChannelsSelf') : t('profile.noChannels')}
                            </div>
                        )}
                        </>
                    )}

                    {activeTab === 'videos' && (
                        <>
                            {isMe && (
                                <div className="flex justify-end mb-4">
                                    <Link to="/me/videos">
                                        <Button variant="outline" size="sm">
                                            <Settings className="w-4 h-4 mr-1"/>
                                            {t('profile.manageVideos')}
                                        </Button>
                                    </Link>
                                </div>
                            )}
                            {videosLoading && videos.length === 0 ? (
                                <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
                            ) : videos.length > 0 ? (
                                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
                                    {videos.map(video => (
                                        <Link key={video.id} to="/watch" search={{v: video.short_token}} className="group">
                                            <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1">
                                                <div className="relative aspect-video">
                                                    {video.thumbnail ? (
                                                        <img src={video.thumbnail} alt={video.title}
                                                             className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                                    ) : (
                                                        <div className="w-full h-full bg-muted dark:bg-gray-700 flex items-center justify-center">
                                                            <Play className="w-10 h-10 text-muted-foreground"/>
                                                        </div>
                                                    )}
                                                    {video.duration != null ? (
                                                        <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">{formatDuration(video.duration)}</div>
                                                    ) : null}
                                                </div>
                                                <div className="p-3">
                                                    <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 text-sm group-hover:text-emerald-600 transition-colors">{video.title}</h3>
                                                    <p className="text-xs text-slate-500 dark:text-muted-foreground mt-2">{formatViews(video.view_count)} {t('common.views')}</p>
                                                </div>
                                            </div>
                                        </Link>
                                    ))}
                                </div>
                            ) : (
                                <div className="text-center py-12 text-muted-foreground">{t('profile.noVideos')}</div>
                            )}
                        </>
                    )}

                    {activeTab === 'articles' && (
                        <>
                            {isMe && (
                                <div className="flex justify-end mb-4">
                                    <Link to="/me/articles">
                                        <Button variant="outline" size="sm">
                                            <Settings className="w-4 h-4 mr-1"/>
                                            {t('profile.manageArticles')}
                                        </Button>
                                    </Link>
                                </div>
                            )}
                            {articlesLoading && articles.length === 0 ? (
                                <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
                            ) : articles.length > 0 ? (
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                                    {articles.map(article => (
                                        <Link key={article.id} to="/articles/$slug" params={{slug: article.slug || article.id}} className="group">
                                            <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1 border border-gray-100 dark:border-gray-700">
                                                {article.thumbnail ? (
                                                    <div className="relative aspect-video">
                                                        <img src={article.thumbnail} alt={article.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"/>
                                                    </div>
                                                ) : (
                                                    <div className="relative aspect-video bg-muted dark:bg-gray-700 flex items-center justify-center">
                                                        <FileText className="w-10 h-10 text-muted-foreground"/>
                                                    </div>
                                                )}
                                                <div className="p-4">
                                                    <h3 className="font-semibold text-foreground line-clamp-2 group-hover:text-primary transition-colors">{article.title}</h3>
                                                    {article.summary && (
                                                        <p className="text-sm text-muted-foreground line-clamp-2 mt-1">{article.summary}</p>
                                                    )}
                                                    <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground">
                                                        <span>{formatViews(article.view_count)} {t('common.views')}</span>
                                                        <span>{formatDate(article.published_at || article.create_time)}</span>
                                                    </div>
                                                </div>
                                            </div>
                                        </Link>
                                    ))}
                                </div>
                            ) : (
                                <div className="text-center py-12 text-muted-foreground">
                                    <FileText className="w-12 h-12 mx-auto mb-3 opacity-30"/>
                                    <p>{isMe ? t('profile.noArticlesSelf') : t('profile.noArticles')}</p>
                                    {isMe && (
                                        <Link to="/me/articles/new" className="text-primary hover:underline text-sm mt-2 inline-block">
                                            {t('profile.writeArticle')}
                                        </Link>
                                    )}
                                </div>
                            )}
                        </>
                    )}

                    {activeTab === 'playlists' && (
                        <>
                            {isMe && (
                                <div className="flex justify-end mb-4">
                                    <Link to="/me/playlists">
                                        <Button variant="outline" size="sm">
                                            <Settings className="w-4 h-4 mr-1"/>
                                            {t('profile.managePlaylists')}
                                        </Button>
                                    </Link>
                                </div>
                            )}
                            {playlistsLoading && playlists.length === 0 ? (
                            <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
                        ) : playlists.length > 0 ? (
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                                {playlists.map(pl => (
                                    <Link key={pl.id} to="/playlist/$token" params={{token: pl.short_token || pl.id}} className="group">
                                        <div className="bg-white dark:bg-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all hover:-translate-y-1 border border-gray-100 dark:border-gray-700">
                                            <div className="relative aspect-video bg-muted dark:bg-gray-700">
                                                <div className="w-full h-full flex items-center justify-center">
                                                    <ListVideo className="w-10 h-10 text-muted-foreground"/>
                                                </div>
                                                <div className="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-2 py-1 rounded">
                                                    {pl.media_items?.length || 0} {t('common.videos')}
                                                </div>
                                            </div>
                                            <div className="p-4">
                                                <h3 className="font-semibold text-slate-900 dark:text-white line-clamp-2 group-hover:text-emerald-600 transition-colors">{pl.title}</h3>
                                                <p className="text-xs text-slate-500 dark:text-muted-foreground mt-1">
                                                    {t('playlists.updated', {date: formatDate(pl.update_time || pl.create_time)})}
                                                </p>
                                            </div>
                                        </div>
                                    </Link>
                                ))}
                            </div>
                        ) : (
                            <div className="text-center py-12 text-muted-foreground">{t('profile.noPlaylists')}</div>
                        )}
                        </>
                    )}

                    {isMe && activeTab === 'followers' && (
                        followersLoading && followers.length === 0 ? (
                            <div className="text-center py-12 text-muted-foreground">{t('common.loading')}</div>
                        ) : followers.length > 0 ? (
                            <div className="space-y-3">
                                <p className="text-sm text-muted-foreground">
                                    {t('profile.followersCount', {count: followersTotal})}
                                </p>
                                {followers.map(follower => (
                                    <Link key={follower.id} to="/u/$id" params={{id: follower.id}}
                                          className="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
                                        <Avatar className="w-10 h-10">
                                            <AvatarImage src={getImageUrl(follower.avatar, 'avatar')} loading="lazy"
                                                         onError={(e) => handleImageError(e, 'avatar')}/>
                                            <AvatarFallback>{follower.username ? follower.username.charAt(0).toUpperCase() : 'U'}</AvatarFallback>
                                        </Avatar>
                                        <div className="flex-1 min-w-0">
                                            <p className="font-medium text-foreground truncate">{follower.username}</p>
                                            <p className="text-xs text-muted-foreground">
                                                {t('profile.followedAt', {date: formatDate(follower.subscribed_at)})}
                                            </p>
                                        </div>
                                    </Link>
                                ))}
                                {followers.length < followersTotal && (
                                    <div className="flex justify-center pt-4">
                                        <Button variant="outline" size="sm"
                                                onClick={() => fetchFollowers(followersPage + 1)}
                                                disabled={followersLoading}>
                                            {followersLoading ? t('common.loading') : t('channel.loadMore')}
                                        </Button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <div className="text-center py-12 text-muted-foreground">{t('profile.noFollowers')}</div>
                        )
                    )}

                    {activeTab === 'about' && (
                        <div className="max-w-2xl space-y-6">
                            <div>
                                <h3 className="font-semibold text-foreground mb-4">{t('profile.aboutBio')}</h3>
                                <p className="text-muted-foreground">{user.bio || t('profile.noBio')}</p>
                            </div>
                            <div className="border-t border-border pt-4">
                                <h3 className="font-semibold text-foreground mb-3">{t('profile.aboutDetails')}</h3>
                                <div className="space-y-3">
                                    <div className="flex items-center gap-3 text-sm">
                                        <Calendar className="w-4 h-4 text-muted-foreground"/>
                                        <span className="text-muted-foreground">
                                            {t('common.joinedAt', {date: formatDate(user.create_time || new Date().toISOString())})}
                                        </span>
                                    </div>
                                    <div className="flex items-center gap-3 text-sm">
                                        <Eye className="w-4 h-4 text-muted-foreground"/>
                                        <span className="text-muted-foreground">
                                            {formatViews(user.total_views || 0)} {t('common.views')}
                                        </span>
                                    </div>
                                    <div className="flex items-center gap-3 text-sm">
                                        <Users className="w-4 h-4 text-muted-foreground"/>
                                        <span className="text-muted-foreground">
                                            {formatViews(user.subscriber_count || 0)} {t('common.followers')}
                                        </span>
                                    </div>
                                </div>
                            </div>
                            {user.links && user.links.length > 0 && (
                                <div className="border-t border-border pt-4">
                                    <h3 className="font-semibold text-foreground mb-3">{t('channel.links')}</h3>
                                    <div className="space-y-2">
                                        {user.links.map((link, index) => (
                                            <a key={index} href={link.url} target="_blank" rel="noopener noreferrer"
                                               className="flex items-center gap-2 text-sm text-primary hover:underline">
                                                {link.title || link.url}
                                            </a>
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ProfilePage;
