/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Upload Page - 支持分片上传与独立文件信息管理
 */

import React, {useState} from 'react';
import {useNavigate} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {UploadComponent} from '@/components/upload/UploadComponent';
import {useMyChannel, useChannelLimits} from '@/hooks/queries';
import {useAuth} from '@/hooks/useAuth';
import {Button} from '@/components/ui/button';
import {CreateChannelDialog} from '@/components/channel/CreateChannelDialog';
import {Tv, Plus} from 'lucide-react';

const UploadPage = () => {
    const {t} = useTranslation();
    const navigate = useNavigate();
    const {isAuthenticated} = useAuth();
    const {data: myChannel, isLoading: channelLoading} = useMyChannel(isAuthenticated);
    const [createDialogOpen, setCreateDialogOpen] = useState(false);

    const hasChannel = !!myChannel;

    if (channelLoading) {
        return (
            <div className="flex items-center justify-center min-h-[60vh]">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand"/>
            </div>
        );
    }

    if (!hasChannel) {
        return (
            <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <div className="flex flex-col items-center justify-center min-h-[400px] text-center gap-6">
                    <div className="w-20 h-20 rounded-full bg-muted flex items-center justify-center">
                        <Tv size={40} className="text-muted-foreground/50"/>
                    </div>
                    <div>
                        <h2 className="text-lg font-medium text-foreground">{t('channel.uploadNeedsChannelTitle')}</h2>
                        <p className="text-sm text-muted-foreground mt-2 max-w-sm">
                            {t('channel.uploadNeedsChannelDescription')}
                        </p>
                    </div>
                    <div className="flex gap-3">
                        <Button variant="outline" onClick={() => navigate({to: '/me/channels'})}>
                            <Tv size={16} className="mr-2"/>
                            {t('channel.myChannels')}
                        </Button>
                        <Button onClick={() => setCreateDialogOpen(true)}>
                            <Plus size={16} className="mr-2"/>
                            {t('channel.create.title')}
                        </Button>
                    </div>
                </div>
                <CreateChannelDialog
                    open={createDialogOpen}
                    onOpenChange={setCreateDialogOpen}
                    onSuccess={() => window.location.reload()}
                />
            </div>
        );
    }

    return (
        <div className="max-w-6xl mx-auto p-6 space-y-8">
            <div className="text-center">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">{t('upload.title')}</h1>
                <p className="text-gray-500 dark:text-muted-foreground">{t('upload.description')}</p>
            </div>

            <UploadComponent onCancel={() => navigate({to: '/'})}/>
        </div>
    );
};

export default UploadPage;
