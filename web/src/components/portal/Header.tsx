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
    Settings,
    Shield,
    ChevronDown,
    Plus,
} from 'lucide-react';
import {useAuth} from '../../hooks/useAuth';

/* ── QuickLink 定义 ──────────────────────────────────────────────────────── */

interface QuickLink {
    label: string;
    to: string;
    icon?: React.ReactNode;
}

const defaultQuickLinks: QuickLink[] = [
    {label: '精选', to: '/featured'},
    {label: '最新', to: '/latest'},
    {label: '分类', to: '/categories'},
    {label: '标签', to: '/tags'},
    {label: '成员', to: '/members'},
];

/** 顶部最多显示 N 个，超过收进"更多"下拉 */
const VISIBLE_QUICK_LINKS = 4;

/* ── Props ───────────────────────────────────────────────────────────────── */

interface HeaderProps {
    onToggleSidebar?: () => void;
    sidebarCollapsed?: boolean;
}

/* ── Component ───────────────────────────────────────────────────────────── */

const Header: React.FC<HeaderProps> = ({onToggleSidebar, sidebarCollapsed}) => {
    const [search, setSearch] = useState('');
    const [userMenuOpen, setUserMenuOpen] = useState(false);
    const [moreMenuOpen, setMoreMenuOpen] = useState(false);
    const userMenuRef = useRef<HTMLDivElement>(null);
    const moreMenuRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();
    const location = useLocation();
    const {isAuthenticated, user, logout, isAdmin} = useAuth();

    // 从 localStorage 读取自定义 QuickLinks
    const [customLinks, setCustomLinks] = useState<QuickLink[]>(() => {
        try {
            const raw = localStorage.getItem('origcms_quicklinks');
            return raw ? JSON.parse(raw) : [];
        } catch {
            return [];
        }
    });

    const allLinks = [...defaultQuickLinks, ...customLinks];
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
            className="fixed top-0 left-0 right-0 h-14 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 z-50">
            <div className="h-full flex items-center px-4 gap-3">
                {/* 左侧: 汉堡菜单 + Logo */}
                <button
                    onClick={onToggleSidebar}
                    className="p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors shrink-0"
                    title={sidebarCollapsed ? '展开菜单' : '收起菜单'}
                >
                    <Menu size={20}/>
                </button>

                <Link to="/" className="flex items-center gap-2 shrink-0">
                    <div className="w-8 h-8 bg-emerald-600 rounded-lg flex items-center justify-center">
                        <span className="text-white font-bold text-sm">O</span>
                    </div>
                    <span className="text-lg font-bold text-gray-900 dark:text-white hidden sm:inline">
                        OrigCMS
                    </span>
                </Link>

                {/* 中间: QuickLinks */}
                <nav className="hidden md:flex items-center gap-1 ml-4">
                    {visibleLinks.map((link) => (
                        <Link
                            key={link.to}
                            to={link.to}
                            className={`px-3 py-1.5 text-sm rounded-full transition-colors whitespace-nowrap ${
                                isActive(link.to)
                                    ? 'bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-400 font-medium'
                                    : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
                            }`}
                        >
                            {link.label}
                        </Link>
                    ))}

                    {/* 更多下拉 */}
                    {moreLinks.length > 0 && (
                        <div className="relative" ref={moreMenuRef}>
                            <button
                                onClick={() => setMoreMenuOpen(!moreMenuOpen)}
                                className="flex items-center gap-1 px-3 py-1.5 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full transition-colors"
                            >
                                更多
                                <ChevronDown size={14}
                                             className={`transition-transform ${moreMenuOpen ? 'rotate-180' : ''}`}/>
                            </button>
                            {moreMenuOpen && (
                                <div
                                    className="absolute top-full left-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg py-1 min-w-[140px]">
                                    {moreLinks.map((link) => (
                                        <Link
                                            key={link.to}
                                            to={link.to}
                                            onClick={() => setMoreMenuOpen(false)}
                                            className={`block px-4 py-2 text-sm transition-colors ${
                                                isActive(link.to)
                                                    ? 'bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-medium'
                                                    : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
                                            }`}
                                        >
                                            {link.label}
                                        </Link>
                                    ))}
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
                            className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
                        />
                        <input
                            type="search"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            placeholder="搜索视频、标签、用户..."
                            className="w-full bg-gray-100 dark:bg-gray-800 border-0 rounded-full pl-9 pr-4 py-1.5 text-sm focus:ring-2 focus:ring-emerald-500 focus:bg-white dark:focus:bg-gray-700 transition-all outline-none"
                        />
                    </div>
                </form>

                {/* 右侧: 用户操作 */}
                <div className="flex items-center gap-2 shrink-0">
                    {isAuthenticated && user ? (
                        <>
                            {/* 上传按钮 */}
                            <Link
                                to="/me/upload"
                                className="hidden sm:flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full transition-colors"
                            >
                                <Plus size={16}/>
                                <span className="hidden lg:inline">上传</span>
                            </Link>

                            {/* 用户头像 + 下拉菜单 */}
                            <div className="relative" ref={userMenuRef}>
                                <button
                                    onClick={() => setUserMenuOpen(!userMenuOpen)}
                                    className="flex items-center gap-2 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                                >
                                    {user.avatarUrl ? (
                                        <img
                                            src={user.avatarUrl}
                                            alt={user.displayName}
                                            className="w-7 h-7 rounded-full object-cover"
                                        />
                                    ) : (
                                        <div
                                            className="w-7 h-7 bg-emerald-100 dark:bg-emerald-900 rounded-full flex items-center justify-center">
                                            <User size={14} className="text-emerald-600 dark:text-emerald-400"/>
                                        </div>
                                    )}
                                    <span
                                        className="text-sm font-medium text-gray-700 dark:text-gray-300 hidden sm:inline max-w-[100px] truncate">
                                        {user.displayName || user.username}
                                    </span>
                                </button>

                                {userMenuOpen && (
                                    <div
                                        className="absolute right-0 top-full mt-1 w-56 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg py-1">
                                        {/* 用户信息 */}
                                        <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-700">
                                            <p className="text-sm font-medium text-gray-900 dark:text-white">
                                                {user.displayName || user.username}
                                            </p>
                                            <p className="text-xs text-gray-500 dark:text-gray-400">
                                                @{user.username}
                                            </p>
                                        </div>

                                        <Link
                                            to={`/u/${user.id}`}
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                                        >
                                            <User size={16}/> 我的频道
                                        </Link>
                                        <Link
                                            to="/me/upload"
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                                        >
                                            <Upload size={16}/> 上传视频
                                        </Link>
                                        <Link
                                            to="/me/favorites"
                                            onClick={() => setUserMenuOpen(false)}
                                            className="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
                                        >
                                            <Settings size={16}/> 我的收藏
                                        </Link>

                                        {isAdmin && (
                                            <Link
                                                to="/admin"
                                                onClick={() => setUserMenuOpen(false)}
                                                className="flex items-center gap-3 px-4 py-2.5 text-sm text-emerald-600 dark:text-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/20"
                                            >
                                                <Shield size={16}/> 管理后台
                                            </Link>
                                        )}

                                        <div className="my-1 border-t border-gray-100 dark:border-gray-700"/>

                                        <button
                                            onClick={() => {
                                                setUserMenuOpen(false);
                                                logout();
                                                navigate({to: '/'});
                                            }}
                                            className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
                                        >
                                            <LogOut size={16}/> 退出登录
                                        </button>
                                    </div>
                                )}
                            </div>
                        </>
                    ) : (
                        <Link
                            to="/auth/signin"
                            className="flex items-center gap-2 px-4 py-1.5 bg-emerald-600 text-white text-sm font-medium rounded-full hover:bg-emerald-700 transition-colors"
                        >
                            <LogIn size={16}/>
                            登录
                        </Link>
                    )}
                </div>
            </div>
        </header>
    );
};

export default Header;
