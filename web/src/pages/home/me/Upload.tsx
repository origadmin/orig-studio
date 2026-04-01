/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Upload Page - 支持分片上传与独立文件信息管理
 */

import React from 'react';
import {useNavigate} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {UploadComponent} from '@/components/upload/UploadComponent';

const UploadPage = () => {
    const {t} = useTranslation();
    const navigate = useNavigate();

    return (
        <div className="max-w-6xl mx-auto p-6 space-y-8">
            <div className="text-center">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">{t('upload.title')}</h1>
                <p className="text-gray-500 dark:text-gray-400">{t('upload.description')}</p>
            </div>

            <UploadComponent onCancel={() => navigate({to: '/'})}/>
        </div>
    );
};

export default UploadPage;
