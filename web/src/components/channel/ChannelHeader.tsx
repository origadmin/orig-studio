import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {
    Settings,
    Upload,
    MoreHorizontal,
    Edit3,
    BadgeCheck,
    Eye,
    Users,
    Film,
} from 'lucide-react';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
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
    isFromMeChannel = false,
    isSubscribed = false,
    subscriberCount = 0,
    subscribing = false,
    onSubscribe,
    onUnsubscribe,
    onNotificationSettingChange,
}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const [showUnsubscribeDialog, setShowUnsubscribeDialog] = useState(false);
    const [unsubscribing, setUnsubscribing] = useState(false);
    const [descriptionExpanded, setDescriptionExpanded] = useState(false);

    const handleUnsubscribeConfirm = async () => {
        try {
            setUnsubscribing(true);
            await onUnsubscribe?.();
            setShowUnsubscribeDialog(false);
        } finally {
            setUnsubscribing(false);
        }
    };

    const videoCount = channel.video_count || channel.media_count || 0;
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

                {isOwner && (
                    <button
                        onClick={() =>
                            navigate({
                                to: '/u/$id',
                                params: {id: channel?.friendly_token || channel?.slug || ''},
                                search: {tab: 'appearance'},
                            })
                        }
                        className="absolute top-4 right-4 z-20 px-3 py-1.5 bg-black/60 hover:bg-black/80 text-white text-sm rounded-lg backdrop-blur-sm transition-all opacity-0 group-hover:opacity-100 flex items-center gap-1.5"
                    >
                        <Edit3 className="w-4 h-4"/>
                        {t('channel.customizeChannel')}
                    </button>
                )}
            </div>

            {/* Channel Info Bar */}
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex flex-col sm:flex-row items-start sm:items-end gap-4 sm:gap-6 -mt-10 sm:-mt-14 relative z-10 pb-4">
                    {/* Avatar - 120px on desktop, smaller on mobile */}
                    <div className="w-20 h-20 sm:w-28 sm:h-28 md:w-[120px] md:h-[120px] rounded-full bg-background border-4 border-background overflow-hidden shadow-lg flex-shrink-0">
                        {channel.avatar ? (
                            <img
                                src={channel.avatar}
                                alt={channel.name}
                                className="w-full h-full object-cover"
                            />
                        ) : (
                            <div className="w-full h-full bg-muted flex items-center justify-center text-3xl sm:text-4xl md:text-5xl font-bold text-muted-foreground">
                                {channel.name?.charAt(0)?.toUpperCase()}
                            </div>
                        )}
                    </div>

                    {/* Channel Info */}
                    <div className="flex-1 pt-2 sm:pt-4 min-w-0">
                        <h1 className="text-xl sm:text-2xl lg:text-3xl font-bold text-foreground mb-1.5 truncate flex items-center gap-2">
                            {channel.name}
                            {channel.is_verified && (
                                <BadgeCheck className="w-5 h-5 sm:w-6 sm:h-6 text-info flex-shrink-0"/>
                            )}
                        </h1>

                        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground mb-2">
                            {channel.handle && (
                                <span className="font-medium text-foreground/70">
                                    @{channel.handle}
                                </span>
                            )}
                            <span className="flex items-center gap-1">
                                <Users className="w-3.5 h-3.5"/>
                                {formatCount(subCount)} {t('channel.subscribers')}
                            </span>
                            <span className="flex items-center gap-1">
                                <Film className="w-3.5 h-3.5"/>
                                {videoCount} {t('channel.videoCount')}
                            </span>
                            {viewCount > 0 && (
                                <span className="flex items-center gap-1">
                                    <Eye className="w-3.5 h-3.5"/>
                                    {formatCount(viewCount)} {t('channel.views')}
                                </span>
                            )}
                        </div>

                        {/* Description - expandable */}
                        {description && (
                            <div className="text-sm text-muted-foreground">
                                <p className={descriptionExpanded ? '' : 'line-clamp-2 max-w-2xl'}>
                                    {description}
                                </p>
                                {description.length > 120 && (
                                    <button
                                        onClick={() => setDescriptionExpanded(!descriptionExpanded)}
                                        className="text-primary hover:underline text-sm mt-0.5"
                                    >
                                        {descriptionExpanded
                                            ? (t('channel.showLess') || 'Show less')
                                            : (t('channel.showMore') || 'Show more')
                                        }
                                    </button>
                                )}
                                {isOwner && !descriptionExpanded && (
                                    <button
                                        onClick={() =>
                                            navigate({
                                                to: '/u/$id',
                                                params: {id: channel?.friendly_token || channel?.slug || ''},
                                                search: {tab: 'branding'},
                                            })
                                        }
                                        className="ml-2 text-primary hover:underline text-sm"
                                    >
                                        {t('channel.editDescription')}
                                    </button>
                                )}
                            </div>
                        )}
                    </div>

                    {/* Action Buttons */}
                    <div className="flex items-center gap-2 pt-2 sm:pt-4 flex-shrink-0 flex-wrap">
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

                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="rounded-full"
                                    title={t('channel.moreOptions')}
                                >
                                    <MoreHorizontal className="w-5 h-5"/>
                                </Button>
                            </>
                        ) : (
                            <>
                                <Button
                                    variant="outline"
                                    onClick={() => navigate({to: '/u/$id', params: {id: channel?.friendly_token || channel?.slug || ''}})}
                                >
                                    <Settings className="w-4 h-4 mr-1"/>
                                    {t('channel.manageChannel')}
                                </Button>
                                <Button
                                    variant="outline"
                                    onClick={() => navigate({to: '/me/upload'})}
                                >
                                    <Upload className="w-4 h-4 mr-1"/>
                                    {t('channel.uploadVideo')}
                                </Button>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="rounded-full"
                                    title={t('channel.moreOptions')}
                                >
                                    <MoreHorizontal className="w-5 h-5"/>
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
