import React, {useState, useCallback} from 'react';
import {useTranslation} from 'react-i18next';
import {Link} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {
    Settings,
    Upload,
    BadgeCheck,
    Eye,
    Users,
    Film,
    Share2,
    Flag,
    ChevronDown,
    Check,
    Link2,
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
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import SubscribeButton from './SubscribeButton';
import NotificationBell from './NotificationBell';
import type {ChannelDetail} from '@/lib/api/channel';

interface ChannelHeaderProps {
    channel: ChannelDetail;
    isOwner: boolean;
    isFromMeChannel?: boolean;
    isSubscribed?: boolean;
    subscriberCount?: number;
    subscribing?: boolean;
    onSubscribe?: () => void;
    onUnsubscribe?: () => Promise<void>;
    onNotificationSettingChange?: (setting: string) => Promise<void> | void;
}

const ChannelHeader: React.FC<ChannelHeaderProps> = ({
    channel,
    isOwner,
    isFromMeChannel: _isFromMeChannel = false,
    isSubscribed = false,
    subscriberCount = 0,
    subscribing = false,
    onSubscribe,
    onUnsubscribe,
    onNotificationSettingChange,
}) => {
    const {t} = useTranslation();
    const [showUnsubscribeDialog, setShowUnsubscribeDialog] = useState(false);
    const [unsubscribing, setUnsubscribing] = useState(false);
    const [descriptionExpanded, setDescriptionExpanded] = useState(false);

    // Share state
    const [showShareDialog, setShowShareDialog] = useState(false);
    const [shareCopied, setShareCopied] = useState(false);
    const [shareError, setShareError] = useState<string | null>(null);

    // Build the canonical channel share URL using /c/{short_token}
    const channelShareUrl = channel.short_token
        ? `${window.location.origin}/c/${channel.short_token}`
        : `${window.location.origin}/channel/${channel.id}`;

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
                    title: channel.name,
                    url: channelShareUrl,
                });
            } catch (err) {
                // User cancelled or share failed - not an error
            }
        }
    }, [channel.name, channelShareUrl]);

    const handleUnsubscribeConfirm = async () => {
        try {
            setUnsubscribing(true);
            await onUnsubscribe?.();
            setShowUnsubscribeDialog(false);
        } finally {
            setUnsubscribing(false);
        }
    };

    const videoCount = channel.media_count || 0;
    const subCount = subscriberCount || channel.subscriber_count || 0;
    const viewCount = channel.total_views || 0;
    const description = channel.description || '';

    return (
        <div className="relative">
            {/* Banner Section - 250px height, full-width background */}
            <div className="relative group">
                {channel.banner ? (
                    <div
                        className="w-full h-[150px] sm:h-[200px] md:h-[250px] bg-cover bg-center"
                        style={{backgroundImage: `url(${channel.banner})`}}
                    />
                ) : (
                    <div className="w-full h-[150px] sm:h-[200px] md:h-[250px] bg-gradient-to-r from-blue-600 via-purple-600 to-pink-500"/>
                )}

                {/* Gradient overlay for better text readability */}
                <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent"/>
            </div>

            {/* Channel Info Bar */}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:gap-6 -mt-8 sm:-mt-14 relative z-10 pb-4">
                    {/* Avatar */}
                    <Avatar className="w-16 h-16 sm:w-28 sm:h-28 md:w-[120px] md:h-[120px] border-4 border-background shadow-lg flex-shrink-0">
                        <AvatarImage
                            src={getImageUrl(channel.avatar, 'avatar')}
                            alt={channel.name}
                            onError={(e) => handleImageError(e, 'avatar')}
                        />
                        <AvatarFallback className="text-2xl sm:text-4xl md:text-5xl font-bold bg-muted text-muted-foreground">
                            {channel.name?.charAt(0)?.toUpperCase() || '?'}
                        </AvatarFallback>
                    </Avatar>

                    {/* Channel Info */}
                    <div className="flex-1 min-w-0">
                        <h1 className="text-lg sm:text-2xl lg:text-3xl font-bold text-foreground mb-0.5 sm:mb-1 flex items-center gap-2 min-w-0">
                            <span className="truncate">{channel.name}</span>
                            {channel.is_verified && (
                                <BadgeCheck className="w-4 h-4 sm:w-6 sm:h-6 text-info flex-shrink-0"/>
                            )}
                        </h1>

                        <div className="flex items-center gap-x-2 sm:gap-x-3 text-xs sm:text-sm text-muted-foreground">
                            <span className="flex items-center gap-1 whitespace-nowrap">
                                <Users className="w-3 h-3 sm:w-3.5 sm:h-3.5"/>
                                {formatCount(subCount)} {t('channel.subscribers')}
                            </span>
                            <span className="flex items-center gap-1 whitespace-nowrap">
                                <Film className="w-3 h-3 sm:w-3.5 sm:h-3.5"/>
                                {videoCount} {t('channel.videoCount')}
                            </span>
                            {viewCount > 0 && (
                                <span className="flex items-center gap-1 whitespace-nowrap">
                                    <Eye className="w-3 h-3 sm:w-3.5 sm:h-3.5"/>
                                    {formatCount(viewCount)} {t('channel.views')}
                                </span>
                            )}
                        </div>

                        {/* Description - expandable */}
                        {description ? (
                            <div className="text-xs sm:text-sm text-muted-foreground">
                                <p className={descriptionExpanded ? '' : 'line-clamp-2 max-w-2xl'}>
                                    {description}
                                </p>
                                {description.length > 120 && (
                                    <button
                                        onClick={() => setDescriptionExpanded(!descriptionExpanded)}
                                        className="text-primary hover:underline text-xs sm:text-sm mt-0.5"
                                    >
                                        {descriptionExpanded
                                            ? (t('channel.showLess') || 'Show less')
                                            : (t('channel.showMore') || 'Show more')
                                        }
                                    </button>
                                )}
                            </div>
                        ) : isOwner ? (
                            <Link
                                to="/me/channels"
                                className="text-xs sm:text-sm text-muted-foreground/60 hover:text-primary transition-colors italic"
                            >
                                {t('channel.addDescription') || 'Add a description...'}
                            </Link>
                        ) : null}
                    </div>

                    {/* Action Buttons */}
                    <div className="flex items-center gap-2 flex-shrink-0 flex-wrap">
                        {!isOwner ? (
                            <>
                                <SubscribeButton
                                    isSubscribed={isSubscribed}
                                    isOwner={isOwner}
                                    subscriberCount={subscriberCount}
                                    subscribing={subscribing}
                                    onSubscribe={onSubscribe}
                                    onUnsubscribeClick={() =>
                                        setShowUnsubscribeDialog(true)
                                    }
                                />

                                <NotificationBell
                                    isSubscribed={isSubscribed}
                                    onSettingChange={
                                        onNotificationSettingChange
                                    }
                                />

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
                                        <DropdownMenuSeparator/>
                                        <DropdownMenuItem>
                                            <Flag className="w-4 h-4 mr-2"/>
                                            {t('channel.reportChannel')}
                                        </DropdownMenuItem>
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </>
                        ) : (
                            <>
                                <Button asChild size="sm">
                                    <Link to="/me/upload">
                                        <Upload className="w-4 h-4"/>
                                    </Link>
                                </Button>
                                <Button asChild variant="outline" size="sm">
                                    <Link to="/me/channels">
                                        <Settings className="w-4 h-4"/>
                                    </Link>
                                </Button>
                                <Button variant="ghost" size="icon" className="rounded-full" onClick={handleShareClick}>
                                    <Share2 className="w-4 h-4"/>
                                </Button>
                            </>
                        )}
                    </div>
                </div>
            </div>

            {/* Unsubscribe Confirmation Dialog */}
            <Dialog
                open={showUnsubscribeDialog}
                onOpenChange={setShowUnsubscribeDialog}
            >
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('channel.confirmUnsubscribeTitle')}</DialogTitle>
                        <DialogDescription>
                            {t('channel.confirmUnsubscribeDesc', {
                                channel: channel.name,
                            })}
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => setShowUnsubscribeDialog(false)}
                            disabled={unsubscribing}
                        >
                            {t('common.cancel')}
                        </Button>
                        <Button
                            variant="destructive"
                            onClick={handleUnsubscribeConfirm}
                            disabled={unsubscribing}
                        >
                            {unsubscribing
                                ? t('channel.unsubscribing')
                                : t('channel.unsubscribe')}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Share Channel Dialog */}
            <Dialog open={showShareDialog} onOpenChange={setShowShareDialog}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t('channel.shareChannel') || 'Share Channel'}</DialogTitle>
                        <DialogDescription>
                            {t('channel.shareDescription', {channel: channel.name}) || `Share ${channel.name} with your friends`}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4">
                        {/* Share Link */}
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

                        {/* Social Share Buttons */}
                        <div className="grid grid-cols-4 gap-2">
                            <a
                                href={`https://twitter.com/intent/tweet?url=${encodeURIComponent(channelShareUrl)}&text=${encodeURIComponent(channel.name)}`}
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
                                href={`https://t.me/share/url?url=${encodeURIComponent(channelShareUrl)}&text=${encodeURIComponent(channel.name)}`}
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

function formatCount(num: number): string {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
}

export default ChannelHeader;
