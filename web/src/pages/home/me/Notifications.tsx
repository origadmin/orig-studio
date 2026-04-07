/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Notifications Page
 */

import React from 'react';
import {Bell} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import NotificationCenter from '@/components/common/NotificationCenter';

const NotificationsPage = () => {
    const {t} = useTranslation();

    return (
        <div className="max-w-4xl mx-auto space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
                    <Bell className="w-6 h-6"/>
                    {t('notifications.title')}
                </h1>
            </div>

            {/* Notification Center */}
            <NotificationCenter/>
        </div>
    );
};

export default NotificationsPage;
