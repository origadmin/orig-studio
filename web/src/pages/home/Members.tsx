import {Spinner} from "@/components/ui/spinner"
/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * 成员页 - 展示平台用户
 */

import React, {useState, useEffect} from 'react';
import {Link} from '@tanstack/react-router';
import {Users, Search, Shield, Edit, UserIcon} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {userApi, getUserStatusLabel} from '@/lib/api/user';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {Avatar, AvatarImage, AvatarFallback} from '@/components/ui/avatar';
import ErrorPage from '@/components/common/ErrorPage';
import {PAGINATION_CONFIG} from '@/config/pagination';

const formatNumber = (n: number, t: (key: string) => string) => n >= 10000 ? `${(n / 10000).toFixed(1)}${t('common.wan')}` : n >= 1000 ? `${(n / 1000).toFixed(1)}K` : String(n);

// Role label mapping for display
const getRoleLabel = (role: string, t: (key: string) => string): {label: string, icon: React.ReactNode} => {
    const roles: Record<string, {label: string, icon: React.ReactNode}> = {
        admin: {label: t('admin.admin') || 'Admin', icon: <Shield className="w-3 h-3"/>},
        editor: {label: t('admin.editor') || 'Editor', icon: <Edit className="w-3 h-3"/>},
        user: {label: t('admin.user') || 'User', icon: <UserIcon className="w-3 h-3"/>},
    };
    return roles[role] || {label: role, icon: <UserIcon className="w-3 h-3"/>};
};

// Status label mapping for display
const getStatusLabel = (status: number | string, t: (key: string) => string): string => {
    const statusLabel = getUserStatusLabel(status);
    const statusMap: Record<string, string> = {
        active: t('admin.active') || 'Active',
        inactive: t('admin.inactive') || 'Inactive',
        pending: t('admin.pending') || 'Pending',
        suspended: t('admin.suspended') || 'Suspended',
        rejected: t('admin.rejected') || 'Rejected',
        unspecified: t('admin.unspecified') || 'Unspecified',
    };
    return statusMap[statusLabel] || statusLabel;
};

const MembersPage = () => {
    const {t} = useTranslation();
    const [filter, setFilter] = useState('');
    const [members, setMembers] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    // 获取用户列表
    useEffect(() => {
        const fetchMembers = async () => {
            try {
                setLoading(true);
                const response = await userApi.list({page_size: PAGINATION_CONFIG.MAX_PAGE_SIZE});
                setMembers(response.items || []);
            } catch (err) {
                setError(t('common.error'));
                console.error('Failed to fetch members:', err);
            } finally {
                setLoading(false);
            }
        };

        fetchMembers();
    }, [t]);

    const filtered = members.filter(m =>
        (m.username || '').toLowerCase().includes(filter.toLowerCase())
    );


    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Spinner />
            </div>
        );
    }

    if (error) {
        return <ErrorPage message={error}/>;
    }

    if (members.length === 0) {
        return (
            <div className="text-center py-16 text-muted-foreground">
                <Users size={48} className="mx-auto mb-3 opacity-30"/>
                <p>{t('members.noMatch')}</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 标题 */}
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                    <Users size={24} className="text-emerald-600"/>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{t('members.title')}</h1>
                </div>
                <span className="text-sm text-gray-500">{t('members.creatorCount', {count: members.length})}</span>
            </div>

            {/* 搜索 */}
            <div className="relative max-w-md">
                <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"/>
                <input
                    type="text"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder={t('members.searchPlaceholder')}
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
                            <Avatar className="w-12 h-12 ring-2 ring-gray-100 dark:ring-gray-700 group-hover:ring-emerald-200 dark:group-hover:ring-emerald-800 transition-all">
                                <AvatarImage
                                    src={getImageUrl(member.avatar, 'avatar')}
                                    alt={member.username}
                                    onError={(e) => handleImageError(e, 'avatar')}
                                />
                                <AvatarFallback className="text-sm font-semibold bg-muted text-muted-foreground">
                                    {(member.nickname || member.username || '?').charAt(0).toUpperCase()}
                                </AvatarFallback>
                            </Avatar>
                            <div className="min-w-0">
                                <h3 className="text-sm font-semibold text-gray-900 dark:text-white truncate group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                                    {member.username}
                                </h3>
                                <p className="text-xs text-muted-foreground">@{member.username}</p>
                            </div>
                        </div>
                        <p className="text-xs text-gray-500 dark:text-muted-foreground mt-3 line-clamp-2">{member.email}</p>
                        <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                            <span className="flex items-center gap-1">{getRoleLabel(member.role, t).icon} {getRoleLabel(member.role, t).label}</span>
                            <span>{getStatusLabel(member.status, t)}</span>
                        </div>
                    </Link>
                ))}
            </div>

            {filtered.length === 0 && (
                <div className="text-center py-16 text-muted-foreground">
                    <Users size={48} className="mx-auto mb-3 opacity-30"/>
                    <p>{t('members.noMatch')}</p>
                </div>
            )}
        </div>
    );
};

export default MembersPage;
