import React, {useState, useCallback} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from '@tanstack/react-router';
import {
    usePublicProfile,
    useMediaList,
    useMyChannels,
    useFavoriteList,
    useHistoryList,
    useSubscriptionStatus,
    useSubscribe,
    useUnsubscribe,
} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {getImageUrl} from '@/lib/imageUtils';
import {Avatar, AvatarImage, AvatarFallback} from '@/components/ui/avatar';
import {Button} from '@/components/ui/button';
import {Badge} from '@/components/ui/badge';
import {Spinner} from '@/components/ui/spinner';
import EmptyState from '@/components/channel/widgets/EmptyState';
import VideoCard from '@/components/channel/widgets/VideoCard';
import {
    Pencil,
    Film,
    ListVideo,
    Heart,
    Info,
    Users,
    Calendar,
    MapPin,
    Link as LinkIcon,
    BadgeCheck,
    Bell,
    Upload,
    Share2,
    ChevronDown,
    Tv,
    Video,
    FileText,
    History,
    UserCheck,
    Check,
    Link2,
    ArrowRight,
    Settings,
} from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';

interface ProfileHomePageProps {
    username: string;
}

type OwnerTab = 'videos' | 'channels' | 'articles' | 'followers' | 'favorites' | 'playlists' | 'history' | 'about';
type VisitorTab = 'videos' | 'playlists' | 'about';

const OWNER_TABS: {key: OwnerTab; icon: React.ElementType; labelKey: string; manageTo: string}[] = [
    {key: 'videos', icon: Video, labelKey: 'nav.myVideos', manageTo: '/me/videos'},
    {key: 'channels', icon: Tv, labelKey: 'nav.myChannels', manageTo: '/me/channels'},
    {key: 'articles', icon: FileText, labelKey: 'nav.myArticles', manageTo: '/me/articles'},
    {key: 'followers', icon: UserCheck, labelKey: 'profile.myFollowers', manageTo: '/u/$id'},
    {key: 'favorites', icon: Heart, labelKey: 'nav.myFavorites', manageTo: '/me/favorites'},
    {key: 'playlists', icon: ListVideo, labelKey: 'nav.myPlaylists', manageTo: '/me/playlists'},
    {key: 'history', icon: History, labelKey: 'nav.history', manageTo: '/me/history'},
    {key: 'about', icon: Info, labelKey: 'profile.tabAbout', manageTo: ''},
];

const ProfileHomePage: React.FC<ProfileHomePageProps> = ({username}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const {user: currentUser, isAuthenticated} = useAuth();
    const [ownerTab, setOwnerTab] = useState<OwnerTab>('videos');
    const [visitorTab, setVisitorTab] = useState<VisitorTab>('videos');

    const [showShareDialog, setShowShareDialog] = useState(false);
    const [shareCopied, setShareCopied] = useState(false);
    const [shareError, setShareError] = useState<string | null>(null);

    const {data: profile, isLoading, error} = usePublicProfile(username);
    // Use profile.is_owner (computed by usePublicProfile from auth state) as primary source.
    // Fallback to client-side computation for defensive robustness.
    const isOwner = profile?.is_owner === true
        || (isAuthenticated && !!currentUser && !!profile && currentUser.username === profile.username);

    const {data: videosData, isLoading: videosLoading} = useMediaList({
        user_id: profile?.id,
        page_size: 6,
        order_by: 'create_time',
        descending: true,
    });
    const videos = (videosData as any)?.items || (videosData as any)?.medias || [];

    const {data: channelsData, isLoading: channelsLoading} = useMyChannels(isOwner);
    const channels = Array.isArray(channelsData) ? channelsData : (channelsData as any)?.items || [];

    const {data: favoritesData, isLoading: favoritesLoading} = useFavoriteList(
        {page_size: 6},
        isOwner && profile?.id ? profile.id : undefined
    );
    const favorites = (favoritesData as any)?.items || (favoritesData as any)?.favorites || [];

    const {data: historyData, isLoading: historyLoading} = useHistoryList({
        page_size: 6,
        isAuthenticated,
        userId: profile?.id,
    });
    const historyItems = (historyData as any)?.items || (historyData as any)?.histories || [];

    const channelToken = profile?.default_channel_token || null;
    const subscriptionQuery = useSubscriptionStatus(
        channelToken && !isOwner && isAuthenticated ? channelToken : null
    );
    const subscribeMutation = useSubscribe();
    const unsubscribeMutation = useUnsubscribe();

    const handleSubscribe = () => {
        if (!channelToken) return;
        subscribeMutation.mutate(channelToken);
    };

    const handleUnsubscribe = () => {
        if (!channelToken) return;
        unsubscribeMutation.mutate(channelToken);
    };

    // Use /c/{short_token} for channel share URL, fallback to /@username only when no token
    const channelShareUrl = channelToken
        ? `${window.location.origin}/c/${channelToken}`
        : `${window.location.origin}/@${username}`;

    const handleShareClick = useCallback(() => {
        setShowShareDialog(true);
        setShareCopied(false);
        setShareError(null);
    }, []);

    const handleCopyShareLink = useCallback(async () => {
        try {
            await navigator.clipboard.writeText(channelShareUrl);
            setShareCopied(true);
            setShareError(null);
            setTimeout(() => setShareCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy share link:', err);
            setShareError(t('channel.shareCopyFailed') || 'Failed to copy link');
        }
    }, [channelShareUrl, t]);

    const handleNativeShare = useCallback(async () => {
        if (navigator.share) {
            try {
                await navigator.share({
                    title: profile?.nickname || profile?.username || username,
                    url: channelShareUrl,
                });
            } catch (err) {}
        }
    }, [profile, username, channelShareUrl]);

    const handleManageClick = (tab: typeof OWNER_TABS[number]) => {
        if (tab.key === 'followers' && profile) {
            navigate({to: tab.manageTo, params: {id: profile.id}, search: {tab: 'followers'}});
        } else if (tab.manageTo) {
            navigate({to: tab.manageTo});
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <Spinner/>
            </div>
        );
    }

    if (error || !profile) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4 p-8">
                <Avatar className="h-24 w-24">
                    <AvatarFallback>{username.charAt(0).toUpperCase()}</AvatarFallback>
                </Avatar>
                <div className="text-center">
                    <h1 className="text-2xl font-bold">{username}</h1>
                    <p className="text-sm text-muted-foreground mt-1">@{username}</p>
                </div>
                <p className="text-muted-foreground text-center max-w-md">
                    {t('profile.notFound', {name: username})}
                </p>
            </div>
        );
    }

    const visitorTabs: {id: VisitorTab; label: string; icon: React.ElementType}[] = [
        {id: 'videos', label: t('profile.tabVideos') || 'Videos', icon: Film},
        {id: 'playlists', label: t('profile.tabPlaylists') || 'Playlists', icon: ListVideo},
        {id: 'about', label: t('profile.tabAbout') || 'About', icon: Info},
    ];

    const renderOwnerTabContent = () => {
        switch (ownerTab) {
            case 'videos':
                return (
                    <ContentSection
                        loading={videosLoading}
                        items={videos}
                        tab={OWNER_TABS[0]}
                        onManage={() => handleManageClick(OWNER_TABS[0])}
                        renderItem={(item) => <VideoCard key={item.id} video={item} showChannelInfo={true}/>}
                        emptyType="videos"
                    />
                );
            case 'channels':
                return (
                    <ContentSection
                        loading={channelsLoading}
                        items={channels}
                        tab={OWNER_TABS[1]}
                        onManage={() => handleManageClick(OWNER_TABS[1])}
                        renderItem={(ch) => (
                            <div key={ch.id} className="flex items-center gap-3 p-3 rounded-lg border hover:bg-accent cursor-pointer transition-colors"
                                 onClick={() => navigate({to: '/c/$token', params: {token: ch.token || ch.short_token}})}>
                                <div className="w-12 h-12 rounded-lg bg-muted flex items-center justify-center flex-shrink-0">
                                    {ch.logo ? <img src={getImageUrl(ch.logo, 'avatar')} alt="" className="w-12 h-12 rounded-lg object-cover"/> : <Tv className="w-6 h-6 text-muted-foreground"/>}
                                </div>
                                <div className="min-w-0 flex-1">
                                    <p className="font-medium text-sm line-clamp-2">{ch.name}</p>
                                    <p className="text-xs text-muted-foreground line-clamp-2">{ch.description || t('profile.noDescription') || 'No description'}</p>
                                </div>
                                <Settings className="w-4 h-4 text-muted-foreground flex-shrink-0"/>
                            </div>
                        )}
                        emptyType="channels"
                    />
                );
            case 'articles':
                return (
                    <ContentSection
                        loading={false}
                        items={[]}
                        tab={OWNER_TABS[2]}
                        onManage={() => handleManageClick(OWNER_TABS[2])}
                        emptyType="articles"
                    />
                );
            case 'followers':
                return (
                    <ContentSection
                        loading={false}
                        items={[]}
                        tab={OWNER_TABS[3]}
                        onManage={() => handleManageClick(OWNER_TABS[3])}
                        emptyType="followers"
                    />
                );
            case 'favorites':
                return (
                    <ContentSection
                        loading={favoritesLoading}
                        items={favorites}
                        tab={OWNER_TABS[4]}
                        onManage={() => handleManageClick(OWNER_TABS[4])}
                        renderItem={(item) => <VideoCard key={item.id || item.media_id} video={item.media || item}/>}
                        emptyType="favorites"
                    />
                );
            case 'playlists':
                return (
                    <ContentSection
                        loading={false}
                        items={[]}
                        tab={OWNER_TABS[5]}
                        onManage={() => handleManageClick(OWNER_TABS[5])}
                        emptyType="playlists"
                    />
                );
            case 'history':
                return (
                    <ContentSection
                        loading={historyLoading}
                        items={historyItems}
                        tab={OWNER_TABS[6]}
                        onManage={() => handleManageClick(OWNER_TABS[6])}
                        renderItem={(item) => <VideoCard key={item.id || item.media_id} video={item.media || item}/>}
                        emptyType="history"
                    />
                );
            case 'about':
                return <ProfileAboutTab profile={profile}/>;
        }
    };

    return (
        <div className="max-w-6xl mx-auto">
            {/* Banner: pure gradient background */}
            <div className="h-32 sm:h-40 md:h-48 bg-gradient-to-r from-blue-600 via-purple-600 to-pink-500 relative"/>

            {/* Profile info section: entirely below the banner */}
            <div className="px-4 sm:px-6 lg:px-8 pt-4 sm:pt-5">
                <div className="flex flex-col sm:flex-row sm:items-center gap-3 sm:gap-5">
                    {/* Avatar: below banner, no overlap */}
                    <Avatar className="w-20 h-20 sm:w-24 sm:h-24 border-4 border-background shadow-lg flex-shrink-0">
                        <AvatarImage src={getImageUrl(profile.avatar, 'avatar')} alt={profile.username}/>
                        <AvatarFallback className="text-2xl font-bold bg-muted text-muted-foreground">
                            {profile.username?.charAt(0)?.toUpperCase() || '?'}
                        </AvatarFallback>
                    </Avatar>

                    {/* User info: name, @username, stats */}
                    <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                            <h1 className="text-xl sm:text-2xl font-bold truncate">{profile.nickname || profile.username}</h1>
                            {profile.is_featured && (
                                <Badge variant="secondary" className="bg-blue-100 text-blue-800 text-xs flex-shrink-0">
                                    <BadgeCheck size={12} className="mr-1"/>
                                    {t('profile.featured')}
                                </Badge>
                            )}
                        </div>
                        <p className="text-sm text-muted-foreground mt-0.5">@{profile.username}</p>
                        {profile.title && (
                            <p className="text-sm text-muted-foreground mt-0.5">{profile.title}</p>
                        )}
                        <div className="flex items-center gap-4 mt-1.5 text-sm text-muted-foreground">
                            {isOwner ? (
                                <>
                                    <span className="flex items-center gap-1">
                                        <Film size={14}/> {videos.length} {t('profile.videos')}
                                    </span>
                                    <span className="flex items-center gap-1">
                                        <Users size={14}/> {profile.subscriber_count || 0} {t('common.followers')}
                                    </span>
                                </>
                            ) : (
                                <>
                                    <span className="flex items-center gap-1">
                                        <Film size={14}/> {profile.media_count || 0} {t('profile.videos')}
                                    </span>
                                    <span className="flex items-center gap-1">
                                        <Users size={14}/> {profile.subscriber_count || 0} {t('common.followers')}
                                    </span>
                                </>
                            )}
                        </div>
                    </div>

                    {/* Action buttons: right side */}
                    <div className="flex items-center gap-2 flex-shrink-0">
                        {isOwner ? (
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button>
                                        <Settings className="w-4 h-4 mr-1"/>
                                        {t('profile.manage') || 'Manage'}
                                        <ChevronDown className="w-4 h-4 ml-1"/>
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end" className="w-48">
                                    <DropdownMenuItem onClick={() => navigate({to: '/me/upload'})}>
                                        <Upload className="w-4 h-4 mr-2"/>
                                        {t('profile.uploadContent')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => navigate({to: '/me/channels'})}>
                                        <Tv className="w-4 h-4 mr-2"/>
                                        {t('profile.createChannel')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => navigate({to: '/me/articles'})}>
                                        <FileText className="w-4 h-4 mr-2"/>
                                        {t('profile.createArticle')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => navigate({to: '/me/playlists'})}>
                                        <ListVideo className="w-4 h-4 mr-2"/>
                                        {t('profile.createPlaylist')}
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator/>
                                    <DropdownMenuItem onClick={() => navigate({to: '/u/$id', params: {id: profile.id}, search: {tab: 'profile'}})}>
                                        <Pencil className="w-4 h-4 mr-2"/>
                                        {t('profile.editProfile')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={handleShareClick}>
                                        <Share2 className="w-4 h-4 mr-2"/>
                                        {t('profile.shareProfile')}
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        ) : (
                            <>
                                <Button
                                    variant={subscriptionQuery.data?.is_subscribed ? 'outline' : 'default'}
                                    onClick={subscriptionQuery.data?.is_subscribed ? handleUnsubscribe : handleSubscribe}
                                    disabled={subscribeMutation.isPending || unsubscribeMutation.isPending}
                                >
                                    {subscriptionQuery.data?.is_subscribed ? t('common.subscribed') : t('common.subscribe')}
                                </Button>
                                {subscriptionQuery.data?.is_subscribed && (
                                    <Button variant="ghost" size="icon" className="rounded-full">
                                        <Bell size={18}/>
                                    </Button>
                                )}
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="ghost" size="icon" className="rounded-full">
                                            <ChevronDown className="w-5 h-5"/>
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <DropdownMenuItem onClick={handleShareClick}>
                                            <Share2 className="w-4 h-4 mr-2"/>
                                            {t('channel.shareChannel')}
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </>
                        )}
                    </div>
                </div>

                {profile.bio && (
                    <p className="text-sm text-muted-foreground mt-3 max-w-2xl">{profile.bio}</p>
                )}
            </div>

            {isOwner ? (
                <div className="px-4 sm:px-6 lg:px-8 mt-6">
                    <div className="flex border-b dark:border-gray-700 overflow-x-auto">
                        {OWNER_TABS.map(tab => (
                            <button
                                key={tab.key}
                                onClick={() => setOwnerTab(tab.key)}
                                className={`flex items-center gap-1.5 px-3 py-2.5 font-medium text-sm border-b-2 transition-colors whitespace-nowrap flex-shrink-0 ${
                                    ownerTab === tab.key
                                        ? 'border-brand text-brand'
                                        : 'border-transparent text-muted-foreground hover:text-foreground'
                                }`}
                            >
                                <tab.icon size={15}/>
                                {t(tab.labelKey)}
                            </button>
                        ))}
                    </div>

                    <div className="py-6">
                        {renderOwnerTabContent()}
                    </div>
                </div>
            ) : (
                <div className="px-4 sm:px-6 lg:px-8 mt-6">
                    <div className="flex border-b dark:border-gray-700">
                        {visitorTabs.map(tab => (
                            <button
                                key={tab.id}
                                onClick={() => setVisitorTab(tab.id)}
                                className={`flex items-center gap-2 px-4 py-3 font-medium text-sm border-b-2 transition-colors ${
                                    visitorTab === tab.id
                                        ? 'border-brand text-brand'
                                        : 'border-transparent text-muted-foreground hover:text-foreground'
                                }`}
                            >
                                <tab.icon size={16}/>
                                {tab.label}
                            </button>
                        ))}
                    </div>

                    <div className="py-6">
                        {visitorTab === 'videos' && (
                            <ProfileVideosTab videos={videos} loading={videosLoading} isOwner={false}/>
                        )}
                        {visitorTab === 'playlists' && (
                            <EmptyState type="playlists" isOwner={false}/>
                        )}
                        {visitorTab === 'about' && (
                            <ProfileAboutTab profile={profile}/>
                        )}
                    </div>
                </div>
            )}

            <Dialog open={showShareDialog} onOpenChange={setShowShareDialog}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('channel.shareChannel') || 'Share Channel'}</DialogTitle>
                        <DialogDescription>
                            {t('channel.shareDescription', {channel: profile?.nickname || profile?.username || username}) || `Share this channel with your friends`}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        <div className="flex items-center gap-2">
                            <div className="flex-1 flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-800 rounded-lg">
                                <Link2 className="w-4 h-4 text-gray-500 flex-shrink-0"/>
                                <input
                                    type="text"
                                    value={channelShareUrl}
                                    readOnly
                                    className="flex-1 bg-transparent text-sm text-gray-700 dark:text-gray-300 outline-none min-w-0"
                                />
                            </div>
                            <Button
                                size="sm"
                                onClick={handleCopyShareLink}
                                className={shareCopied ? 'bg-green-600 hover:bg-green-700' : 'bg-emerald-600 hover:bg-emerald-700'}
                            >
                                {shareCopied ? <Check className="w-4 h-4"/> : (t('watch.copyLink') || 'Copy')}
                            </Button>
                        </div>

                        {shareError && (
                            <p className="text-sm text-destructive">{shareError}</p>
                        )}

                        <div className="grid grid-cols-4 gap-2">
                            <a
                                href={`https://twitter.com/intent/tweet?url=${encodeURIComponent(channelShareUrl)}&text=${encodeURIComponent(profile?.nickname || username)}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                            >
                                <div className="w-10 h-10 bg-black rounded-full flex items-center justify-center">
                                    <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                        <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                                    </svg>
                                </div>
                                <span className="text-xs text-gray-600 dark:text-muted-foreground">X</span>
                            </a>
                            <a
                                href={`https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(channelShareUrl)}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                            >
                                <div className="w-10 h-10 bg-blue-600 rounded-full flex items-center justify-center">
                                    <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                        <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
                                    </svg>
                                </div>
                                <span className="text-xs text-gray-600 dark:text-muted-foreground">Facebook</span>
                            </a>
                            <a
                                href={`https://t.me/share/url?url=${encodeURIComponent(channelShareUrl)}&text=${encodeURIComponent(profile?.nickname || username)}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                            >
                                <div className="w-10 h-10 bg-info rounded-full flex items-center justify-center">
                                    <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 24 24">
                                        <path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
                                    </svg>
                                </div>
                                <span className="text-xs text-gray-600 dark:text-muted-foreground">Telegram</span>
                            </a>
                            {'share' in navigator && (
                                <button
                                    onClick={handleNativeShare}
                                    className="flex flex-col items-center gap-1 p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    <div className="w-10 h-10 bg-muted dark:bg-gray-700 rounded-full flex items-center justify-center">
                                        <Share2 className="w-5 h-5 text-gray-700 dark:text-gray-300"/>
                                    </div>
                                    <span className="text-xs text-gray-600 dark:text-muted-foreground">More</span>
                                </button>
                            )}
                        </div>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
};

const ContentSection: React.FC<{
    loading: boolean;
    items: any[];
    tab: typeof OWNER_TABS[number];
    onManage: () => void;
    onAction?: () => void;
    renderItem?: (item: any) => React.ReactNode;
    emptyType: string;
}> = ({loading, items, tab, onManage, renderItem, emptyType}) => {
    const {t} = useTranslation();

    if (loading) {
        return (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[1, 2, 3].map(i => (
                    <div key={i} className="animate-pulse">
                        <div className="aspect-video bg-muted rounded-lg mb-2"/>
                        <div className="h-4 bg-muted rounded w-3/4 mb-1"/>
                        <div className="h-3 bg-muted rounded w-1/2"/>
                    </div>
                ))}
            </div>
        );
    }

    const hasManage = tab.manageTo && onManage;

    return (
        <div>
            {items.length > 0 && renderItem ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                    {items.map(item => renderItem(item))}
                </div>
            ) : (
                <EmptyState type={emptyType as any} isOwner={true}/>
            )}

            {hasManage && (
                <div className="mt-4 flex justify-end">
                    <Button variant="ghost" size="sm" onClick={onManage} className="text-muted-foreground hover:text-foreground">
                        {t('profile.viewAll') || 'View All'}
                        <ArrowRight className="w-4 h-4 ml-1"/>
                    </Button>
                </div>
            )}
        </div>
    );
};

const ProfileVideosTab: React.FC<{videos: any[]; loading: boolean; isOwner: boolean}> = ({videos, loading, isOwner}) => {
    if (loading) {
        return (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[1, 2, 3, 4, 5, 6].map(i => (
                    <div key={i} className="animate-pulse">
                        <div className="aspect-video bg-muted rounded-lg mb-2"/>
                        <div className="h-4 bg-muted rounded w-3/4 mb-1"/>
                        <div className="h-3 bg-muted rounded w-1/2"/>
                    </div>
                ))}
            </div>
        );
    }

    if (videos.length === 0) {
        return <EmptyState type="videos" isOwner={isOwner}/>;
    }

    return (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {videos.map(video => (
                <VideoCard key={video.id} video={video}/>
            ))}
        </div>
    );
};

const ProfileAboutTab: React.FC<{profile: any}> = ({profile}) => {
    const {t} = useTranslation();

    return (
        <div className="space-y-6 max-w-2xl">
            {profile.bio && (
                <div>
                    <h3 className="text-sm font-medium text-muted-foreground mb-1">{t('profile.bio') || 'Bio'}</h3>
                    <p className="text-foreground whitespace-pre-wrap">{profile.bio}</p>
                </div>
            )}
            {profile.location && (
                <div className="flex items-center gap-2 text-sm">
                    <MapPin size={16} className="text-muted-foreground"/>
                    <span>{profile.location}</span>
                </div>
            )}
            {profile.website && (
                <div className="flex items-center gap-2 text-sm">
                    <LinkIcon size={16} className="text-muted-foreground"/>
                    <a href={profile.website} target="_blank" rel="noopener noreferrer" className="text-brand hover:underline">
                        {profile.website}
                    </a>
                </div>
            )}
            {profile.created_at && (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Calendar size={16}/>
                    <span>{t('profile.joined') || 'Joined'} {new Date(profile.created_at).toLocaleDateString()}</span>
                </div>
            )}
            {!profile.bio && !profile.location && !profile.website && (
                <p className="text-muted-foreground/60 italic">{t('profile.noInfo') || 'No additional information'}</p>
            )}
        </div>
    );
};

export default ProfileHomePage;
