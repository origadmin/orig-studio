/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Sidebar: 可展开/折叠模式 (YouTube风格)
 * - 桌面端: 默认展开240px，可折叠到72px
 * - 移动端: 抽屉模式240px
 */

import React, {useState, useEffect} from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {
    Home,
    Star,
    Clock,
    Tag,
    LayoutGrid,
    Users,
    Video,
    History,
    Heart,
    ListVideo,
    Upload,
    Shield,
    Settings,
    Info,
    Sun,
    Moon,
    ChevronLeft,
    Globe,
} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {useAuth} from '../../hooks/useAuth';

/* ── Props ───────────────────────────────────────────────────────────────── */

interface SidebarProps {
    darkMode: boolean;
    onToggleDarkMode: () => void;
    collapsed?: boolean;
    onToggleCollapse?: () => void;
    mobileOpen?: boolean;
    onCloseMobile?: () => void;
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
                                             onToggleCollapse,
                                             mobileOpen,
                                             onCloseMobile,
                                         }) => {
    const {t, i18n} = useTranslation();
    const location = useLocation();
    const pathname = location.pathname;
    const {isAuthenticated, isAdmin} = useAuth();

    // 响应式状态
    const [isMobile, setIsMobile] = useState(false);

    useEffect(() => {
        const checkScreen = () => setIsMobile(window.innerWidth < 768);
        checkScreen();
        window.addEventListener('resize', checkScreen);
        return () => window.removeEventListener('resize', checkScreen);
    }, []);
    
    const isActive = (to: string) => pathname === to || pathname.startsWith(to + '/');

    // ── 导航分组 ──
    const browseItems: NavItem[] = [
        {icon: <Home size={20}/>, label: '首页', to: '/'},
        {icon: <Star size={20}/>, label: '精选', to: '/featured'},
        {icon: <Clock size={20}/>, label: '最新', to: '/latest'},
        {icon: <LayoutGrid size={20}/>, label: '分类', to: '/categories'},
        {icon: <Tag size={20}/>, label: '标签', to: '/tags'},
        {icon: <Users size={20}/>, label: '会员', to: '/members'},
    ];

    const myItems: NavItem[] = [
        {icon: <Video size={20}/>, label: '我的上传', to: '/me/videos'},
        {icon: <Upload size={20}/>, label: '上传视频', to: '/me/upload'},
        {icon: <ListVideo size={20}/>, label: '播放列表', to: '/me/playlists'},
        {icon: <History size={20}/>, label: '历史记录', to: '/me/history'},
        {icon: <Heart size={20}/>, label: '我的收藏', to: '/me/favorites'},
    ];

    const adminItems: NavItem[] = [
        {icon: <Shield size={20}/>, label: '媒体管理', to: '/admin/media'},
        {icon: <Users size={20}/>, label: '用户管理', to: '/admin/users'},
        {icon: <Settings size={20}/>, label: '系统设置', to: '/admin/settings'},
    ];

    const otherItems: NavItem[] = [
        {icon: <Info size={20}/>, label: '关于', to: '/about'},
    ];

    // 图标链接组件 - 折叠模式只显示图标
    const IconNavLink = ({item}: { item: NavItem }) => {
        const [showTooltip, setShowTooltip] = useState(false);
        const active = isActive(item.to);

        return (
            <div className="relative">
                <Link
                    to={item.to}
                    onClick={() => isMobile && onCloseMobile?.()}
                    onMouseEnter={() => setShowTooltip(true)}
                    onMouseLeave={() => setShowTooltip(false)}
                    className={`flex flex-col items-center justify-center py-2 px-1 min-h-[60px] rounded-lg transition-colors ${
                        active
                            ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                            : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'
                    }`}
                >
                    <span
                        className={`flex-shrink-0 ${active ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-gray-500'}`}>
                        {item.icon}
                    </span>
                </Link>

                {/* 悬停提示 */}
                {showTooltip && (
                    <div
                        className="absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-xs rounded whitespace-nowrap z-50 pointer-events-none">
                        {item.label}
                        <div
                            className="absolute left-0 top-1/2 -translate-x-1 -translate-y-1/2 border-4 border-transparent border-r-gray-900"/>
                    </div>
                )}
            </div>
        );
    };

    // 完整导航链接 - 展开模式显示图标+文字
    const FullNavLink = ({item}: { item: NavItem }) => {
        const active = isActive(item.to);
        return (
            <Link
                to={item.to}
                className={`flex items-center gap-3 py-2 px-3 rounded-lg transition-colors ${
                    active
                        ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-medium'
                        : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
            >
                <span
                    className={`flex-shrink-0 ${active ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-gray-500'}`}>
                    {item.icon}
                </span>
                <span className="text-[13px]">{item.label}</span>
            </Link>
        );
    };

    // 图标导航分组
    const IconNavSection = ({items}: { items: NavItem[] }) => (
        <div className="py-0.5">
            <div className="space-y-0">
                {items.map((item) => (
                    <IconNavLink key={item.to} item={item}/>
                ))}
            </div>
        </div>
    );

    // 完整导航分组
    const FullNavSection = ({items, title}: { items: NavItem[]; title?: string }) => (
        <div className="py-2">
            {title && <h3 className="px-3 py-1 text-xs font-medium text-gray-400 uppercase tracking-wider">{title}</h3>}
            <div className="space-y-0.5">
                {items.map((item) => (
                    <FullNavLink key={item.to} item={item}/>
                ))}
            </div>
        </div>
    );

    // 分割线
    const IconDivider = () => (
        <div className="border-t border-gray-200 dark:border-gray-800 my-0.5 mx-2"/>
    );

    const FullDivider = () => (
        <div className="border-t border-gray-200 dark:border-gray-800 my-2"/>
    );

    // 桌面端 Sidebar
    const DesktopSidebar = () => {
        const width = collapsed ? 72 : 240;
        return (
            <aside
                style={{width}}
                className="fixed left-0 top-14 bottom-0 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 z-50 hidden md:flex flex-col transition-all duration-300"
            >
                {/* 导航内容 - 顶部有 Header 菜单按钮 */}
                <div className={`flex-1 overflow-y-auto ${collapsed ? 'px-1 py-2' : 'px-3 py-2'}`}>
                    {collapsed ? (
                        <>
                            <IconNavSection items={browseItems}/>
                            {isAuthenticated && <><IconDivider/><IconNavSection items={myItems}/></>}
                            {isAdmin && <><IconDivider/><IconNavSection items={adminItems}/></>}
                            <IconDivider/>
                            <IconNavSection items={otherItems}/>
                        </>
                    ) : (
                        <>
                            <FullNavSection items={browseItems} title='浏览'/>
                            {isAuthenticated && <><FullDivider/><FullNavSection items={myItems} title='我的内容'/></>}
                            {isAdmin && <><FullDivider/><FullNavSection items={adminItems} title='管理'/></>}
                            <FullDivider/>
                            <FullNavSection items={otherItems}/>
                        </>
                    )}
                </div>

                {/* 底部主题切换 + 语言切换 */}
                <div
                    className={`border-t border-gray-200 dark:border-gray-800 ${collapsed ? 'px-1 py-2 space-y-1' : 'px-3 py-2 space-y-2'}`}>
                    {collapsed ? (
                        <>
                            <button
                                onClick={onToggleDarkMode}
                                className="flex items-center justify-center w-full py-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                                title={darkMode ? '切换亮色' : '切换暗色'}
                            >
                                {darkMode ? <Sun size={18} className="text-amber-500"/> : <Moon size={18}/>}
                            </button>
                            <button
                                onClick={() => i18n.changeLanguage(i18n.language === 'zh' ? 'en' : 'zh')}
                                className="flex items-center justify-center w-full py-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                                title={i18n.language === 'zh' ? '切换到英文' : '切换到中文'}
                            >
                                <Globe size={18} className="text-emerald-500"/>
                            </button>
                        </>
                    ) : (
                        <>
                            <button
                                onClick={onToggleDarkMode}
                                className="flex items-center gap-2 w-full py-2 px-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                            >
                                {darkMode ? <Sun size={18} className="text-amber-500"/> : <Moon size={18}/>}
                                <span className="text-[13px]">{darkMode ? '切换亮色' : '切换暗色'}</span>
                            </button>
                            <button
                                onClick={() => i18n.changeLanguage(i18n.language === 'zh' ? 'en' : 'zh')}
                                className="flex items-center gap-2 w-full py-2 px-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                            >
                                <Globe size={18} className="text-emerald-500"/>
                                <span className="text-[13px]">{i18n.language === 'zh' ? 'English' : '中文'}</span>
                            </button>
                        </>
                    )}
                </div>
            </aside>
        );
    };

    // 标准导航链接 - 用于移动端抽屉
    const NavLink = ({item}: { item: NavItem }) => {
        const active = isActive(item.to);
        return (
            <Link
                to={item.to}
                onClick={() => onCloseMobile?.()}
                className={`flex items-center gap-3 py-2 px-3 rounded-lg transition-colors ${
                    active
                        ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-medium'
                        : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'
                }`}
            >
                <span
                    className={`flex-shrink-0 ${active ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400 dark:text-gray-500'}`}>
                    {item.icon}
                </span>
                <span className="text-[13px]">{item.label}</span>
            </Link>
        );
    };

    const NavSection = ({items}: { items: NavItem[] }) => (
        <div className="py-2">
            <div className="space-y-0.5">
                {items.map((item) => (
                    <NavLink key={item.to} item={item}/>
                ))}
            </div>
        </div>
    );

    const Divider = () => (
        <div className="border-t border-gray-200 dark:border-gray-800 my-2"/>
    );

    // ── 移动端抽屉 Sidebar ──
    const MobileSidebar = () => (
        <>
            {/* 遮罩层 */}
            {mobileOpen && (
                <div
                    className="fixed inset-0 bg-black/50 z-40 md:hidden"
                    onClick={onCloseMobile}
                />
            )}

            {/* 抽屉 */}
            <aside
                className={`fixed left-0 top-0 bottom-0 w-[240px] bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 z-50 transition-transform duration-300 ease-in-out md:hidden flex flex-col ${
                    mobileOpen ? 'translate-x-0' : '-translate-x-full'
                }`}
            >
                {/* 头部 */}
                <div
                    className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
                    <span className="text-lg font-semibold">Menu</span>
                    <button
                        onClick={onCloseMobile}
                        className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full"
                    >
                        <ChevronLeft size={20}/>
                    </button>
                </div>

                {/* 导航内容 - 可滚动 */}
                <div className="flex-1 overflow-y-auto scrollbar-thin px-2 py-2">
                    <NavSection items={browseItems}/>
                    {isAuthenticated && (
                        <>
                            <Divider/>
                            <NavSection items={myItems}/>
                        </>
                    )}
                    {isAdmin && (
                        <>
                            <Divider/>
                            <NavSection items={adminItems}/>
                        </>
                    )}
                    <Divider/>
                    <NavSection items={otherItems}/>
                </div>

                {/* 底部操作区 */}
                <div className="shrink-0 px-2 py-3 border-t border-gray-200 dark:border-gray-800 space-y-2">
                    <button
                        onClick={onToggleDarkMode}
                        className="flex items-center gap-3 w-full py-2 px-3 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                    >
                        {darkMode ? <Sun size={20} className="text-amber-500"/> : <Moon size={20}/>}
                        <span className="text-[13px]">{darkMode ? '切换亮色' : '切换暗色'}</span>
                    </button>
                    <button
                        onClick={() => i18n.changeLanguage(i18n.language === 'zh' ? 'en' : 'zh')}
                        className="flex items-center justify-center gap-2 w-full py-2 px-3 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                    >
                        <span
                            className="text-emerald-500 font-medium">{i18n.language === 'zh' ? 'English' : '中文'}</span>
                    </button>
                </div>
            </aside>
        </>
    );

    return (
        <>
            <DesktopSidebar/>
            <MobileSidebar/>
        </>
    );
};

export default Sidebar;
