/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 成员页 - 展示平台用户
 */

import React, {useState} from 'react';
import {Link} from '@tanstack/react-router';
import {Users, Search} from 'lucide-react';

interface MemberInfo {
    id: number;
    username: string;
    displayName: string;
    avatar: string;
    bio: string;
    mediaCount: number;
    followerCount: number;
}

const mockMembers: MemberInfo[] = [
    {
        id: 1,
        username: 'gopher_expert',
        displayName: 'Gopher Expert',
        avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&q=80&w=200',
        bio: 'Go 语言爱好者，微服务架构师',
        mediaCount: 24,
        followerCount: 12500
    },
    {
        id: 2,
        username: 'react_master',
        displayName: 'React Master',
        avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&q=80&w=200',
        bio: '前端工程师，React 生态贡献者',
        mediaCount: 18,
        followerCount: 8900
    },
    {
        id: 3,
        username: 'devops_pro',
        displayName: 'DevOps Pro',
        avatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?auto=format&fit=crop&q=80&w=200',
        bio: 'DevOps 工程师，K8s 认证专家',
        mediaCount: 12,
        followerCount: 6700
    },
    {
        id: 4,
        username: 'ts_guru',
        displayName: 'TypeScript Guru',
        avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&q=80&w=200',
        bio: '全栈开发者，TypeScript 布道师',
        mediaCount: 31,
        followerCount: 15200
    },
    {
        id: 5,
        username: 'data_scientist',
        displayName: 'Data Scientist',
        avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?auto=format&fit=crop&q=80&w=200',
        bio: '数据科学家，机器学习研究员',
        mediaCount: 9,
        followerCount: 4300
    },
    {
        id: 6,
        username: 'cloud_expert',
        displayName: 'Cloud Expert',
        avatar: 'https://images.unsplash.com/photo-1560250097-0b93528c311a?auto=format&fit=crop&q=80&w=200',
        bio: 'AWS 解决方案架构师',
        mediaCount: 15,
        followerCount: 9800
    },
    {
        id: 7,
        username: 'vue_master',
        displayName: 'Vue Master',
        avatar: 'https://images.unsplash.com/photo-1580489944761-15a19d654956?auto=format&fit=crop&q=80&w=200',
        bio: 'Vue.js 核心团队成员',
        mediaCount: 22,
        followerCount: 18900
    },
    {
        id: 8,
        username: 'fullstack_dev',
        displayName: 'Full Stack Dev',
        avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&q=80&w=200',
        bio: '全栈工程师，开源贡献者',
        mediaCount: 27,
        followerCount: 11300
    },
];

const formatNumber = (n: number) => n >= 10000 ? `${(n / 10000).toFixed(1)}万` : n >= 1000 ? `${(n / 1000).toFixed(1)}K` : String(n);

const MembersPage = () => {
    const [filter, setFilter] = useState('');
    const filtered = mockMembers.filter(m =>
        m.displayName.toLowerCase().includes(filter.toLowerCase()) ||
        m.username.toLowerCase().includes(filter.toLowerCase())
    );

    return (
        <div className="space-y-6">
            {/* 标题 */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <Users size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">成员</h1>
                </div>
                <span className="text-sm text-gray-500">{mockMembers.length} 位创作者</span>
            </div>

            {/* 搜索 */}
            <div className="relative max-w-md">
                <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"/>
                <input
                    type="text"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder="搜索成员..."
                    className="w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg pl-9 pr-4 py-2 text-sm focus:ring-2 focus:ring-emerald-500 focus:border-transparent outline-none"
                />
            </div>

            {/* 成员卡片网格 */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {filtered.map((member) => (
                    <Link
                        key={member.id}
                        to="/u/$id"
                        params={{id: String(member.id)}}
                        className="group p-4 bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 rounded-xl hover:shadow-lg hover:border-emerald-200 dark:hover:border-emerald-800 transition-all"
                    >
                        <div className="flex items-center gap-3">
                            <img
                                src={member.avatar}
                                alt={member.displayName}
                                className="w-12 h-12 rounded-full object-cover ring-2 ring-gray-100 dark:ring-gray-700 group-hover:ring-emerald-200 dark:group-hover:ring-emerald-800 transition-all"
                            />
                            <div className="min-w-0">
                                <h3 className="text-sm font-semibold text-gray-900 dark:text-white truncate group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                    {member.displayName}
                                </h3>
                                <p className="text-xs text-gray-400">@{member.username}</p>
                            </div>
                        </div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-3 line-clamp-2">{member.bio}</p>
                        <div className="flex items-center gap-4 mt-3 text-xs text-gray-400">
                            <span>{member.mediaCount} 个视频</span>
                            <span>{formatNumber(member.followerCount)} 粉丝</span>
                        </div>
                    </Link>
                ))}
            </div>

            {filtered.length === 0 && (
                <div className="text-center py-16 text-gray-400">
                    <Users size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>没有找到匹配的成员</p>
                </div>
            )}
        </div>
    );
};

export default MembersPage;
