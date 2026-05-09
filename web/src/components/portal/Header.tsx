/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * Header: Logo | QuickLinks + 更多下拉 | 搜索框 | 用户菜单
 */

import React, {useState, useRef, useEffect} from 'react';
import {Link, useNavigate, useLocation} from '@tanstack/react-router';
import {
    Search,
    Menu,
    LogIn,
    User,
    LogOut,
    Upload,
    Heart,
    Shield,
    ChevronDown,
    Plus,
    Sun,
    Moon,
    FileText,
    Tv,
    UserCircle,
    ExternalLink,
} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useModuleConfig} from '@/hooks/useModuleConfig';
import {useModuleState} from '@/contexts/ModuleConfigContext';
import {usePortalConfig} from '@/hooks/queries';
import LanguageSwitcher from '@/components/common/LanguageSwitcher';

/* ── QuickLink 定义 ──────────────────────────────────────────────────────── */

interface QuickLink {
    label: string;
    to: string;
    icon?: React.ReactNode;
}

/** 顶部最多显示 N 个，超过收进"更多"下拉 */
const VISIBLE_QUICK_LINKS = 4;

/* ── Props ───────────────────────────────────────────────────────────────── */

interface HeaderProps {
    onToggleSidebar?: () => void;
    onOpenMobileSidebar?: () => void;
    sidebarCollapsed?: boolean;
    darkMode?: boolean;
    onToggleDarkMode?: () => void;
}

/* ── Component ───────────────────────────────────────────────────────────── */

const Header: React.FC<HeaderProps> = ({onToggleSidebar, onOpenMobileSidebar, sidebarCollapsed, darkMode, onToggleDarkMode}) => {
    const {t, i18n} = useTranslation();
    const [search, setSearch] = useState('');
    const [userMenuOpen, setUserMenuOpen] = useState(false);
    const [moreMenuOpen, setMoreMenuOpen] = useState(false);
    const [isMobile, setIsMobile] = useState(false);
    const userMenuRef = useRef<HTMLDivElement>(null);
    const moreMenuRef = useRef<HTMLDivElement>(null);

    // 检测移动端
    useEffect(() => {
        const checkMobile = () => setIsMobile(window.innerWidth < 768);
        checkMobile();
        window.addEventListener('resize', checkMobile);
        return () => window.removeEventListener('resize', checkMobile);
    }, []);

    // 菜单按钮点击处理
    const handleMenuClick = () => {
        if (isMobile && onOpenMobileSidebar) {
            onOpenMobileSidebar();
        } else if (onToggleSidebar) {
            onToggleSidebar();
        }
    };
    const navigate = useNavigate();
    const location = useLocation();
    const {isAuthenticated, user, logout, isAdmin} = useAuth();
    const {data: moduleConfig} = useModuleConfig();
    const {data: portalConfig} = usePortalConfig();
    const {site} = useModuleState();
    const articlesEnabled = moduleConfig?.modules?.articles !== false;

    const dynamicNavItems = (portalConfig?.navigation?.items || [])
        .filter(item => item.is_visible !== false)
        .sort((a, b) => a.sequence - b.sequence);

    const quickLinks: QuickLink[] = dynamicNavItems.map(item => ({
        label: item.label_i18n?.[i18n.language] || item.label,
        to: item.url,
    }));

    const allLinks = [...quickLinks];
    const visibleLinks = allLinks.slice(0, VISIBLE_QUICK_LINKS);
    const moreLinks = allLinks.slice(VISIBLE_QUICK_LINKS);

    // 点击外部关闭菜单
    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
                setUserMenuOpen(false);
            }
            if (moreMenuRef.current && !moreMenuRef.current.contains(e.target as Node)) {
                setMoreMenuOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        if (search.trim()) navigate({to: '/search', search: {q: search}});
    };

    const isActive = (to: string) => location.pathname === to;

    return (
        <header
            className="fixed top-0 left-0 right-0 h-14 bg-background border-b border-border z-50">
            <div className="h-full flex items-center px-4 gap-3">
                {/* 左侧: 汉堡菜单 + Logo */}
                {/* 移动端菜单按钮 */}
                <button
                    onClick={handleMenuClick}
                    className="p-2 text-muted-foreground hover:bg-accent rounded-lg transition-colors shrink-0 md:hidden"
                    title={t('nav.menu')}
                >
                    <Menu size={20}/>
                </button>

                {/* 桌面端Sidebar折叠按钮 */}
                <button
                    onClick={onToggleSidebar}
                    className="hidden md:flex w-8 h-8 text-muted-foreground hover:bg-accent rounded-lg transition-colors shrink-0 flex-col items-center justify-center gap-1"
                    title={sidebarCollapsed ? t('nav.expand') : t('nav.collapse')}
                >
                    <div className="w-4 h-0.5 bg-muted-foreground rounded"></div>
                    <div className="w-4 h-0.5 bg-muted-foreground rounded"></div>
                    <div className="w-4 h-0.5 bg-muted-foreground rounded"></div>
                </button>

                <Link to="/" className="flex items-center gap-2 shrink-0">
                    <img src="/logo.svg" alt={site?.site_name || 'OrigCMS'} className="h-8 w-8" />
                    <span className="text-lg font-bold text-foreground hidden sm:inline">
                        {site?.site_name || 'OrigCMS'}
                    </span>
                </Link>

                {/* 中间: QuickLinks */}
                <nav className="hidden md:flex items-center gap-1 ml-4">
                    {visibleLinks.map((link, idx) => {
                        const navItem = dynamicNavItems[idx];
                        const isExternal = navItem?.type === 'external_link' || link.to.startsWith('http');
                        if (isExternal) {
                            return (
                                <a
                                    key={link.to + idx}
                                    href={link.to}
                                    target={navItem?.open_new_tab ? '_blank' : undefined}
                                    rel={navItem?.open_new_tab ? 'noopener noreferrer' : undefined}
                                    className="px-3 py-1.5 text-sm rounded-full transition-colors whitespace-nowrap text-muted-foreground hover:bg-accent flex items-center gap-1"
                                >
                                    {link.label}
                                    <ExternalLink className="h-3 w-3"/>
                                </a>
                            );
                        }
                        return (
                            <Link
                                key={link.to}
                                to={link.to}
                                className={`px-3 py-1.5 text-sm rounded-full transition-colors whitespace-nowrap ${
                                    isActive(link.to)
                                        ? 'bg-brand/10 dark:bg-brand/20 text-brand dark:text-brand font-medium'
                                        : 'text-muted-foreground hover:bg-accent'
                                }`}
                            >
                                {link.label}
                            </Link>
                        );
                    })}

                    {moreLinks.length > 0 && (
                        <div className="relative" ref={moreMenuRef}>
                            <button
                                onClick={() => setMoreMenuOpen(!moreMenuOpen)}
                                className="flex items-center gap-1 px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent rounded-full transition-colors"
                            >
                                {t('nav.more')}
                                <ChevronDown size={14}
                                             className={`transition-transform ${moreMenuOpen ? 'rotate-180' : ''}`}/>
                            </button>
                            {moreMenuOpen && (
                                <div
                                    className="absolute top-full left-0 mt-1 bg-popover border border-border rounded-xl shadow-lg py-1 min-w-[140px]">
                                    {moreLinks.map((link, idx) => {
                                        const navItem = dynamicNavItems[VISIBLE_QUICK_LINKS + idx];
                                        const isExternal = navItem?.type === 'external_link' || link.to.startsWith('http');
                                        if (isExternal) {
                                            return (
                                                <a
                                                    key={link.to + idx}
                                                    href={link.to}
                                                    target={navItem?.open_new_tab ? '_blank' : undefined}
                                                    rel={navItem?.open_new_tab ? 'noopener noreferrer' : undefined}
                                                    onClick={() => setMoreMenuOpen(false)}
                                                    className="block px-4 py-2 text-sm transition-colors text-muted-foreground hover:bg-accent flex items-center gap-1"
                                                >
                                                    {link.label}
                                                    <ExternalLink className="h-3 w-3"/>
                                                </a>
                                            );
                                        }
                                        return (
                                            <Link
                                                key={link.to}
                                                to={link.to}
                                                onClick={() => setMoreMenuOpen(false)}
                                                className={`block px-4 py-2 text-sm transition-colors ${
                                                    isActive(link.to)
                                                        ? 'bg-brand/10 dark:bg-brand/20 text-brand dark:text-brand font-medium'
                                                        : 'text-muted-foreground hover:bg-accent'
                                                }`}
                                            >
                                                {link.label}
                                            </Link>
                                        );
                                    })}
                                </div>
                            )}
                        </div>
                    )}
                </nav>

                {/* 搜索框 */}
                <form onSubmit={handleSearch} className="flex-1 max-w-xl mx-auto">
                    <div className="relative">
                        <Search
                            size={16}
                            className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                        />
                        <input
                            type="search"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            placeholder={t('header.searchPlaceholder')}
                            className="w-full bg-muted border-0 rounded-full pl-9 pr-4 py-1.5 text-sm focus:ring-2 focus:ring-brand focus:bg-background transition-all outline-none"
                        />
                    </div>
                </form>

                {/* 右侧: 用户操作 */}
                <div className="flex items-center gap-1 shrink-0">
                    {onToggleDarkMode && (
                        <button
                            onClick={onToggleDarkMode}
                            className="w-10 h-10 flex items-center justify-center text-muted-foreground hover:bg-accent rounded-lg transition-colors"
                            title={darkMode ? t('nav.toggleLight') : t('nav.toggleDark')}
                        >
                            {darkMode ? <Sun size={18} className="text-warning"/> : <Moon size={18}/>}
                        </button>
                    )}
                    <LanguageSwitcher
                        buttonClassName="text-muted-foreground hover:bg-accent"
                    />

                    {isAuthenticated && user ? (
                        <>
                            {/* 通知徽章 - 暂时禁用 */}
                            {/* <NotificationBadge/> */}

                            {/* Write Article button */}
                            {articlesEnabled && (
                            <Link
                                to="/me/articles/new"
                                className="hidden sm:flex items-center gap-1.5 h-10 px-3.5 text-sm font-medium text-muted-foreground hover:bg-accent rounded-full transition-colors"
                            >
                                <FileText size={16}/>
                                <span className="hidden lg:inline">{t('nav.write')}</span>
                            </Link>
                            )}

                            {/* 上传按钮 */}
                            <Link
                                to="/me/upload"
                                className="hidden sm:flex items-center gap-1.5 h-10 px-3.5 text-sm font-medium text-muted-foreground hover:bg-accent rounded-full transition-colors"
                            >
                                <Plus size={16}/>
                                <span className="hidden lg:inline">{t('nav.upload')}</span>
                            </Link>

                            {/* 用户头像 + 下拉菜单 */}
                            <div className="relative" ref={userMenuRef}>
                                <button
                                    onClick={() => setUserMenuOpen(!userMenuOpen)}
                                    className="w-10 h-10 flex items-center justify-center rounded-full hover:bg-accent transition-colors"
                                >
                                    {user.avatarUrl ? (
                                        <img
                                            src={user.avatarUrl}
                                            alt={user.displayName}
                                            loading="lazy"
                                            className="w-8 h-8 rounded-full object-cover"
                                        />
                                    ) : (
                                        <div
                                            className="w-8 h-8 bg-brand/10 dark:bg-brand/20 rounded-full flex items-center justify-center">
                                            <User size={16} className="text-brand dark:text-brand"/>
                                        </div>
                                    )}
                                </button>

                                {userMenuOpen && (
                                    <div
                                        className="absolute right-0 top-full mt-1 w-56 bg-popover border border-border rounded-xl shadow-lg py-1">
                                        {/* 用户信息 */}
                                        <div className="px-4 py-3 border-b border-border">
                                            <p className="text-sm font-medium text-foreground">
                                                {user.displayName || user.username}
                                            </p>
                                            <p className="text-xs text-muted-foreground">
                                                @{user.username}
                                            </p>
                                        </div>

                                        <Link
                                            to="/$handle"
                                            params={{handle: `@${user.username}`}}
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground hover:bg-accent"
                                        >
                                            <UserCircle size={16}/> {t('nav.myProfile')}
                                        </Link>
                                        <Link
                                            to="/me/channels"
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground hover:bg-accent"
                                        >
                                            <Tv size={16}/> {t('nav.channelManagement')}
                                        </Link>
                                        <Link
                                            to="/me/upload"
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground hover:bg-accent"
                                        >
                                            <Upload size={16}/> {t('nav.uploadVideo')}
                                        </Link>
                                        <Link
                                            to="/me/favorites"
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-muted-foreground hover:bg-accent"
                                        >
                                            <Heart size={16}/> {t('nav.myFavorites')}
                                        </Link>

                                        {isAdmin && (
                                            <Link
                                                to="/admin"
                                                onClick={() => setUserMenuOpen(false)}
                                                className="flex items-center gap-3 px-4 py-2.5 text-sm text-brand dark:text-brand hover:bg-brand/10 dark:hover:bg-brand/20"
                                            >
                                                <Shield size={16}/> {t('nav.admin')}
                                            </Link>
                                        )}

                                        <div className="my-1 border-t border-border"/>

                                        <button
                                            onClick={() => {
                                                setUserMenuOpen(false);
                                                logout();
                                                navigate({to: '/'});
                                            }}
                                            className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-destructive hover:bg-destructive/10"
                                        >
                                            <LogOut size={16}/> {t('nav.logout')}
                                        </button>
                                    </div>
                                )}
                            </div>
                        </>
                    ) : (
                        <Link
                            to="/auth/signin"
                            className="flex items-center gap-2 h-10 px-4 bg-brand text-primary-foreground text-sm font-medium rounded-full hover:bg-brand/90 transition-colors"
                        >
                            <LogIn size={16}/>
                            {t('nav.login')}
                        </Link>
                    )}
                </div>
            </div>
        </header>
    );
};

export default Header;
