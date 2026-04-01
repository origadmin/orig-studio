/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 关于页面
 */

import React from 'react';
import {Info, Heart, Code, Users} from 'lucide-react';

const AboutPage = () => {
    return (
        <div className="space-y-8 max-w-3xl">
            <div className="flex items-center gap-3">
                <Info size={24} className="text-emerald-600"/>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">关于 OrigCMS</h1>
            </div>

            <div
                className="bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700 p-8 space-y-6">
                <div>
                    <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-3">什么是 OrigCMS？</h2>
                    <p className="text-gray-600 dark:text-gray-300 leading-relaxed">
                        OrigCMS 是一个开源的视频内容管理系统，基于 Go + React 构建。
                        它提供了完整的视频上传、转码、管理和播放功能，适用于个人、团队和企业使用。
                    </p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="text-center p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                        <Code size={28} className="mx-auto text-emerald-500 mb-2"/>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">Go + React</p>
                        <p className="text-xs text-gray-500 mt-1">现代技术栈</p>
                    </div>
                    <div className="text-center p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                        <Heart size={28} className="mx-auto text-red-500 mb-2"/>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">开源免费</p>
                        <p className="text-xs text-gray-500 mt-1">MIT License</p>
                    </div>
                    <div className="text-center p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                        <Users size={28} className="mx-auto text-blue-500 mb-2"/>
                        <p className="text-sm font-medium text-gray-900 dark:text-white">社区驱动</p>
                        <p className="text-xs text-gray-500 mt-1">欢迎参与贡献</p>
                    </div>
                </div>

                <div>
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">主要功能</h2>
                    <ul className="space-y-2 text-sm text-gray-600 dark:text-gray-300">
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            视频上传、转码和管理
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            用户系统（注册、登录、个人中心）
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            分类、标签、搜索
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            播放列表、收藏、历史记录
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            管理后台（媒体、用户、评论、设置）
                        </li>
                        <li className="flex items-start gap-2">
                            <span className="text-emerald-500 mt-1">&#10003;</span>
                            亮色/暗色主题切换
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
