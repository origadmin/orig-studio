/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Sidebar: 浏览 / 我的(登录可见) / 管理(管理员可见)
 */

import React from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {
    Home,
    Star,
    Clock,
    Tag,
    LayoutGrid,
    Users,
    History,
    Heart,
    ListVideo,
    Upload,
    Shield,
    Settings,
    Info,
    Sun,
    Moon,
} from 'lucide-react';
import {useAuth} from '../../hooks/useAuth';

/* ── Props ───────────────────────────────────────────────────────────────── */

interface SidebarProps {
    darkMode: boolean;
    onToggleDarkMode: () => void;
    collapsed?: boolean;
    onToggleCollapse?: () => void;
}

interface NavItem {
    icon: React.ReactNode;
    label: string;
    to: string;
}

/* ── Component ───────────────────────────────────────────────────────────── */

const Sidebar: React.FC<SidebarProps> = ({
                                             darkMode,
                                             onToggleDarkMode,
                                             collapsed = false,
                                         }) => {
    const location = useLocation();
    const pathname = location.pathname;
    const {isAuthenticated, isAdmin} = useAuth();

    const isActive = (to: string) => pathname === to || pathname.startsWith(to + '/');

    // ── 导航分组 ──

    const browseItems: NavItem[] = [
        {icon: <Home size={18}/>, label: '主页', to: '/'},
        {icon: <Star size={18}/>, label: '精选', to: '/featured'},
        {icon: <Clock size={18}/>, label: '最新', to: '/latest'},
        {icon: <LayoutGrid size={18}/>, label: '分类', to: '/categories'},
        {icon: <Tag size={18}/>, label: '标签', to: '/tags'},
        {icon: <Users size={18}/>, label: '成员', to: '/members'},
    ];

    const myItems: NavItem[] = [
        {icon: <Upload size={18}/>, label: '我的上传', to: '/me/upload'},
        {icon: <ListVideo size={18}/>, label: '我的播放列表', to: '/me/playlists'},
        {icon: <History size={18}/>, label: '历史记录', to: '/me/history'},
        {icon: <Heart size={18}/>, label: '我的收藏', to: '/me/favorites'},
    ];

    const adminItems: NavItem[] = [
        {icon: <Shield size={18}/>, label: '管理媒体', to: '/admin/media'},
        {icon: <Users size={18}/>, label: '管理用户', to: '/admin/users'},
        {icon: <Settings size={18}/>, label: '系统设置', to: '/admin/settings'},
    ];

    const otherItems: NavItem[] = [
        {icon: <Info size={18}/>, label: '关于', to: '/about'},
    ];

    // ── 渲染区段 ──

    const NavSection: React.FC<{ title?: string; items: NavItem[] }> = ({title, items}) => (
        <div className="space-y-0.5">
            {title && !collapsed && (
                <p className="px-3 pt-4 pb-1.5 text-[11px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">
                    {title}
                </p>
            )}
            {collapsed && title && <div className="my-2 mx-3 border-t border-gray-100 dark:border-gray-800"/>}
            {items.map((item) => (
                <Link
                    key={item.to}
                    to={item.to}
                    className={`flex items-center gap-3 px-3 py-2 rounded-lg transition-all group ${
                        collapsed ? 'justify-center' : ''
                    } ${
                        isActive(item.to)
                            ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-medium'
                            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200'
                    }`}
                    title={collapsed ? item.label : undefined}
                >
                    <span
                        className={`shrink-0 ${isActive(item.to) ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300'}`}>
                        {item.icon}
                    </span>
                    {!collapsed && <span className="text-[13px]">{item.label}</span>}
                </Link>
            ))}
        </div>
    );

    return (
        <aside
            className={`fixed left-0 top-14 bottom-0 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 z-40 transition-all duration-300 overflow-hidden ${
                collapsed ? 'w-[60px]' : 'w-[220px]'
            }`}
        >
            <div className="h-full flex flex-col overflow-y-auto py-2 scrollbar-thin">
                {/* 浏览 */}
                <div className={collapsed ? 'px-2' : 'px-2'}>
                    <NavSection items={browseItems}/>
                </div>

                {/* 我的 - 仅登录可见 */}
                {isAuthenticated && (
                    <>
                        <div className={collapsed ? 'px-2' : 'px-2'}>
                            <NavSection title="我的" items={myItems}/>
                        </div>
                    </>
                )}

                {/* 管理 - 仅管理员可见 */}
                {isAdmin && (
                    <>
                        <div className={collapsed ? 'px-2' : 'px-2'}>
                            <NavSection title="管理" items={adminItems}/>
                        </div>
                    </>
                )}

                {/* 其他 */}
                <div className={collapsed ? 'px-2' : 'px-2'}>
                    <NavSection title="其他" items={otherItems}/>
                </div>

                {/* 底部弹性空间 */}
                <div className="flex-1"/>

                {/* 主题切换 */}
                <div className={`px-2 py-3 ${collapsed ? 'px-2' : ''}`}>
                    <button
                        onClick={onToggleDarkMode}
                        className={`flex items-center gap-3 w-full px-3 py-2 rounded-lg text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 transition-all ${
                            collapsed ? 'justify-center' : ''
                        }`}
                        title={collapsed ? (darkMode ? '亮色模式' : '暗色模式') : undefined}
                    >
                        {darkMode ? (
                            <Sun size={18} className="text-amber-500 shrink-0"/>
                        ) : (
                            <Moon size={18} className="text-gray-400 shrink-0"/>
                        )}
                        {!collapsed && (
                            <span className="text-[13px]">{darkMode ? '亮色模式' : '暗色模式'}</span>
                        )}
                    </button>
                </div>
            </div>
        </aside>
    );
};

export default Sidebar;
