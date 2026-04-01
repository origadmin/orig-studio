/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 关于页面
 */

import React from 'react';
import {Info, Heart, Code, Users} from 'lucide-react';
import {useTranslation} from 'react-i18next';

const AboutPage = () => {
    const {t} = useTranslation();
    return (
        <div className="space-y-8 max-w-3xl">
            <div className="flex items-center gap-3">
                <Info size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('about.title')}</h1>
            </div>

            <div
                className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-8 space-y-6">
                <div>
                    <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-3">{t('about.whatIs')}</h2>
                    <p className="text-gray-600 dark:text-gray-300 leading-relaxed">
                        {t('about.description')}
                    </p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="text-center p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                        <Code size={28} className="mx-auto text-emerald-500 mb-2"/>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">{t('about.goReact')}</p>
                        <p className="text-xs text-gray-500 mt-1">{t('about.modernStack')}</p>
                    </div>
                    <div className="text-center p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                        <Heart size={28} className="mx-auto text-red-500 mb-2"/>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">{t('about.openSource')}</p>
                        <p className="text-xs text-gray-500 mt-1">MIT License</p>
                    </div>
                    <div className="text-center p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                        <Users size={28} className="mx-auto text-blue-500 mb-2"/>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">{t('about.communityDriven')}</p>
                        <p className="text-xs text-gray-500 mt-1">{t('about.welcomeContrib')}</p>
                    </div>
                </div>

                <div>
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">{t('about.features')}</h2>
                    <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-300">
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            {t('about.feature1')}
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            {t('about.feature2')}
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            {t('about.feature3')}
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            {t('about.feature4')}
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            {t('about.feature5')}
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            {t('about.feature6')}
                        </li>
                    </ul>
                </div>

                <div className="text-center pt-4 border-t border-gray-100 dark:border-gray-700">
                    <p className="text-sm text-gray-400">
                        Powered by <span
                        className="font-medium text-gray-600 dark:text-gray-300">OrigAdmin</span> &middot; MIT
                        License &middot; 2024
                    </p>
                </div>
            </div>
        </div>
    );
};

export default AboutPage;
