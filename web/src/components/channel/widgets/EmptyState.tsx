import React from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from '@tanstack/react-router';
import {Button} from '@/components/ui/button';
import {Upload, FileVideo, ListVideo, MessageSquare, Info} from 'lucide-react';

interface EmptyStateProps {
    type: 'videos' | 'playlists' | 'community' | 'home';
    isOwner?: boolean;
    channelId?: string;
}

const EmptyState: React.FC<EmptyStateProps> = ({type, isOwner = false, channelId: _channelId}) => {
    const {t} = useTranslation();
    const navigate = useNavigate();

    const configs = {
        videos: {
            icon: <FileVideo className="w-20 h-20"/>,
            title: t('channel.noVideos'),
            desc: isOwner
                ? t('channel.emptyVideosDescOwner')
                : t('channel.emptyVideosDescVisitor'),
            action: isOwner ? (
                <Button onClick={() => navigate({to: '/me/upload'})}>
                    <Upload className="w-4 h-4 mr-1"/>
                    {t('channel.uploadVideo')}
                </Button>
            ) : null,
        },
        playlists: {
            icon: <ListVideo className="w-20 h-20"/>,
            title: t('channel.noPlaylists'),
            desc: isOwner
                ? t('channel.emptyPlaylistsDescOwner')
                : t('channel.emptyPlaylistsDescVisitor'),
            action: isOwner ? (
                <Button variant="outline" onClick={() => console.log('Create playlist')}>
                    <ListVideo className="w-4 h-4 mr-1"/>
                    {t('channel.createPlaylist')}
                </Button>
            ) : null,
        },
        community: {
            icon: <MessageSquare className="w-20 h-20"/>,
            title: t('channel.noCommunity'),
            desc: isOwner
                ? t('channel.emptyCommunityDescOwner')
                : t('channel.emptyCommunityDescVisitor'),
            action: isOwner ? (
                <Button variant="outline" onClick={() => console.log('Create post')}>
                    <MessageSquare className="w-4 h-4 mr-1"/>
                    {t('channel.createPost')}
                </Button>
            ) : null,
        },
        home: {
            icon: <Info className="w-20 h-20"/>,
            title: t('channel.emptyHomeTitle'),
            desc: isOwner ? t('channel.emptyHomeDescOwner') : t('channel.emptyHomeDescVisitor'),
            action: isOwner ? (
                <Button onClick={() => navigate({to: '/me/upload'})}>
                    <Upload className="w-4 h-4 mr-1"/>
                    {t('channel.startCreating')}
                </Button>
            ) : null,
        },
    };

    const config = configs[type];

    return (
        <div className="flex flex-col items-center justify-center min-h-[400px] text-center px-4">
            <div className="text-muted-foreground/30 mb-6 flex justify-center">
                {config.icon}
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">
                {config.title}
            </h3>
            <p className="text-sm text-muted-foreground mb-8 max-w-sm mx-auto">
                {config.desc}
            </p>
            {config.action && <div className="mt-2">{config.action}</div>}
        </div>
    );
};

export default EmptyState;
