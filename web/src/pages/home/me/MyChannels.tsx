import React, {useState} from 'react';
import {useTranslation} from 'react-i18next';
import {Link, useNavigate} from '@tanstack/react-router';
import {useMyChannels, useChannelLimits} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {useQueryClient} from '@tanstack/react-query';
import {channelApi, type Channel} from '@/lib/api/channel';
import {Button} from '@/components/ui/button';
import {Card, CardContent} from '@/components/ui/card';
import {Badge} from '@/components/ui/badge';
import {Avatar, AvatarImage, AvatarFallback} from '@/components/ui/avatar';
import {Input} from '@/components/ui/input';
import {Textarea} from '@/components/ui/textarea';
import {Label} from '@/components/ui/label';
import {CreateChannelDialog} from '@/components/channel/CreateChannelDialog';
import {getImageUrl} from '@/lib/imageUtils';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import {
    Tv,
    Plus,
    Users,
    Video,
    Eye,
    Settings,
    ExternalLink,
    FileText,
    Loader2,
} from 'lucide-react';

const MyChannels = () => {
    const {t} = useTranslation();
    const {isAuthenticated} = useAuth();
    const navigate = useNavigate();
    const queryClient = useQueryClient();
    const {data: channels, isLoading} = useMyChannels(isAuthenticated);
    const {data: limits} = useChannelLimits(isAuthenticated);
    const [createDialogOpen, setCreateDialogOpen] = useState(false);
    const [editChannel, setEditChannel] = useState<Channel | null>(null);
    const [editLoading, setEditLoading] = useState(false);
    const [editForm, setEditForm] = useState({name: '', description: ''});

    const channelList = channels || [];
    const canCreate = limits ? limits.can_create : true;
    const currentCount = limits ? limits.current_count : 0;
    const maxChannels = limits ? limits.max_channels : -1;

    const handleCreateSuccess = ({handle, short_token}: {id: string; handle: string; short_token?: string}) => {
        setCreateDialogOpen(false);
        if (short_token) {
            navigate({to: '/c/$id', params: {id: short_token}});
        } else {
            navigate({to: '/$handle', params: {handle: `@${handle}`}});
        }
    };

    const openEditDialog = (channel: Channel) => {
        setEditChannel(channel);
        setEditForm({name: channel.name, description: channel.description || ''});
    };

    const handleEditSave = async () => {
        if (!editChannel?.short_token) return;
        setEditLoading(true);
        try {
            await channelApi.update(editChannel.short_token, {
                name: editForm.name,
                description: editForm.description,
            });
            queryClient.invalidateQueries({queryKey: ['channels', 'me']});
            setEditChannel(null);
        } catch {
        } finally {
            setEditLoading(false);
        }
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand"/>
            </div>
        );
    }

    return (
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="flex items-center justify-between mb-8">
                <div>
                    <h1 className="text-2xl font-bold">{t('channel.myChannels')}</h1>
                    <p className="text-muted-foreground mt-1">
                        {t('channel.myChannelsDescription', {
                            current: currentCount,
                            max: maxChannels === -1 ? t('common.unlimited') : maxChannels,
                        })}
                    </p>
                </div>
                <Button
                    onClick={() => setCreateDialogOpen(true)}
                    disabled={!canCreate}
                >
                    <Plus size={16} className="mr-2"/>
                    {t('channel.create.title')}
                </Button>
            </div>

            {channelList.length === 0 ? (
                <div className="flex flex-col items-center justify-center min-h-[400px] text-center gap-6">
                    <div className="w-20 h-20 rounded-full bg-muted flex items-center justify-center">
                        <Tv size={40} className="text-muted-foreground/50"/>
                    </div>
                    <div>
                        <h2 className="text-lg font-medium text-foreground">{t('channel.noChannelsTitle')}</h2>
                        <p className="text-sm text-muted-foreground mt-2 max-w-sm">
                            {t('channel.noChannelsDescription')}
                        </p>
                    </div>
                    {canCreate && (
                        <Button size="lg" onClick={() => setCreateDialogOpen(true)}>
                            <Plus size={16} className="mr-2"/>
                            {t('channel.create.title')}
                        </Button>
                    )}
                </div>
            ) : (
                <div className="grid gap-4">
                    {channelList.map(channel => (
                        <Card key={channel.id} className="overflow-hidden hover:shadow-md transition-shadow">
                            <CardContent className="p-0">
                                <div className="flex flex-col sm:flex-row">
                                    {channel.banner ? (
                                        <div className="sm:w-48 h-24 sm:h-auto flex-shrink-0">
                                            <img
                                                src={getImageUrl(channel.banner, 'cover')}
                                                alt={channel.name}
                                                className="w-full h-full object-cover"
                                            />
                                        </div>
                                    ) : (
                                        <div className="sm:w-48 h-24 sm:h-auto flex-shrink-0 bg-gradient-to-r from-blue-600 via-purple-600 to-pink-500"/>
                                    )}
                                    <div className="flex-1 p-4 sm:p-5">
                                        <div className="flex items-start gap-4">
                                            <Avatar className="w-16 h-16 border-2 border-background shadow">
                                                <AvatarImage
                                                    src={getImageUrl(channel.avatar, 'avatar')}
                                                    alt={channel.name}
                                                />
                                                <AvatarFallback className="text-lg font-bold bg-muted text-muted-foreground">
                                                    {channel.name?.charAt(0)?.toUpperCase() || '?'}
                                                </AvatarFallback>
                                            </Avatar>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-center gap-2 flex-wrap">
                                                    <h3 className="text-lg font-semibold truncate">{channel.name}</h3>
                                                    {channel.is_verified && (
                                                        <Badge variant="secondary" className="bg-blue-100 text-blue-800 text-xs">
                                                            ✓ {t('channel.verified')}
                                                        </Badge>
                                                    )}
                                                    {channel.status && channel.status !== 'ACTIVE' && (
                                                        <Badge variant="outline" className="text-xs">
                                                            {channel.status}
                                                        </Badge>
                                                    )}
                                                </div>
                                                {channel.handle && (
                                                    <p className="text-sm text-muted-foreground">@{channel.handle}</p>
                                                )}
                                                {channel.description && (
                                                    <p className="text-sm text-muted-foreground mt-1 line-clamp-2">{channel.description}</p>
                                                )}
                                                <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                                                    <span className="flex items-center gap-1">
                                                        <Users size={14}/> {channel.subscriber_count || 0} {t('channel.subscribers')}
                                                    </span>
                                                    <span className="flex items-center gap-1">
                                                        <Video size={14}/> {channel.media_count || 0} {t('common.videos')}
                                                    </span>
                                                    {channel.article_count !== undefined && (
                                                        <span className="flex items-center gap-1">
                                                            <FileText size={14}/> {channel.article_count} {t('common.articles')}
                                                        </span>
                                                    )}
                                                    {channel.total_views !== undefined && (
                                                        <span className="flex items-center gap-1">
                                                            <Eye size={14}/> {channel.total_views} {t('common.views')}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-2 flex-shrink-0">
                                                <Link
                                                    to="/c/$id"
                                                    params={{id: channel.short_token || channel.id}}
                                                >
                                                    <Button variant="outline" size="sm">
                                                        <ExternalLink size={14} className="mr-1"/>
                                                        {t('channel.viewChannel')}
                                                    </Button>
                                                </Link>
                                                <Button variant="ghost" size="sm" onClick={() => openEditDialog(channel)}>
                                                    <Settings size={14} className="mr-1"/>
                                                    {t('channel.channelSettings')}
                                                </Button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            <CreateChannelDialog
                open={createDialogOpen}
                onOpenChange={setCreateDialogOpen}
                onSuccess={handleCreateSuccess}
            />

            <Dialog open={!!editChannel} onOpenChange={(open) => !open && setEditChannel(null)}>
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>{t('channel.channelSettings')}</DialogTitle>
                        <DialogDescription>
                            {editChannel?.name}
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <Label>{t('channel.create.name_label')}</Label>
                            <Input
                                value={editForm.name}
                                onChange={(e) => setEditForm(prev => ({...prev, name: e.target.value}))}
                                maxLength={150}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label>{t('channel.create.description_label')}</Label>
                            <Textarea
                                value={editForm.description}
                                onChange={(e) => setEditForm(prev => ({...prev, description: e.target.value}))}
                                rows={3}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setEditChannel(null)} disabled={editLoading}>
                            {t('common.cancel')}
                        </Button>
                        <Button onClick={handleEditSave} disabled={editLoading || !editForm.name.trim()}>
                            {editLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin"/>}
                            {t('common.save')}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    );
};

export default MyChannels;
