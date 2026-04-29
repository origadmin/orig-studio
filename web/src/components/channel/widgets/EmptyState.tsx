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
            icon: <FileVideo className="w-16 h-16"/>,
            title: t('channel.noVideos'),
            desc: isOwner
                ? '上传您的第一个视频，与观众分享精彩内容'
                : '创作者可能正在准备精彩内容',
            action: isOwner ? (
                <Button onClick={() => navigate({to: '/me/upload'})}>
                    <Upload className="w-4 h-4 mr-1"/>
                    上传视频
                </Button>
            ) : null,
        },
        playlists: {
            icon: <ListVideo className="w-16 h-16"/>,
            title: t('channel.noPlaylists'),
            desc: isOwner
                ? '创建您的第一个播放列表来整理视频'
                : '该频道暂无公开播放列表',
            action: isOwner ? (
                <Button variant="outline" onClick={() => console.log('Create playlist')}>
                    <ListVideo className="w-4 h-4 mr-1"/>
                    创建播放列表
                </Button>
            ) : null,
        },
        community: {
            icon: <MessageSquare className="w-16 h-16"/>,
            title: t('channel.noCommunity'),
            desc: isOwner
                ? '发布动态，与订阅者互动'
                : '暂无社区帖子',
            action: isOwner ? (
                <Button variant="outline" onClick={() => console.log('Create post')}>
                    <MessageSquare className="w-4 h-4 mr-1"/>
                    发布帖子
                </Button>
            ) : null,
        },
        home: {
            icon: <Info className="w-16 h-16"/>,
            title: '频道首页为空',
            desc: isOwner ? '上传内容让您的频道焕发生机' : '该频道暂无内容',
            action: isOwner ? (
                <Button onClick={() => navigate({to: '/me/upload'})}>
                    <Upload className="w-4 h-4 mr-1"/>
                    开始创作
                </Button>
            ) : null,
        },
    };

    const config = configs[type];

    return (
        <div className="text-center py-16 px-4">
            <div className="text-muted-foreground/50 mb-4 flex justify-center">
                {config.icon}
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">
                {config.title}
            </h3>
            <p className="text-sm text-muted-foreground mb-6 max-w-sm mx-auto">
                {config.desc}
            </p>
            {config.action && <div>{config.action}</div>}
        </div>
    );
};

export default EmptyState;
