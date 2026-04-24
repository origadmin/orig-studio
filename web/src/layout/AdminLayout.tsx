/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React, {useState} from 'react';
import {Outlet, Link, useRouterState} from '@tanstack/react-router';
import {
    LayoutDashboard,
    Film,
    Users,
    Settings,
    LogOut,
    FolderTree,
    Radio,
    Tags,
    MessageSquare,
    PlayCircle,
    Cpu,
    Activity,
    PanelLeftClose,
    PanelLeftOpen,
    Home,
    FileText
} from 'lucide-react';
import {useTranslation} from 'react-i18next';
import {Separator} from '@/components/ui/separator';

const AdminLayout = () => {
    const {t} = useTranslation();
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
    const routerState = useRouterState();

    const menuItems = [
        {id: "dashboard", icon: LayoutDashboard, label: t('admin.dashboard'), path: "/admin"},
        {id: "media", icon: Film, label: t('admin.media'), path: "/admin/media"},
        {
            id: "transcoding-profiles",
            icon: Cpu,
            label: t('admin.transcodingProfiles') || "Transcoding Profiles",
            path: "/admin/transcoding/profiles"
        },
        {
            id: "transcoding-status",
            icon: Activity,
            label: t('admin.transcodingStatus') || "Transcoding Status",
            path: "/admin/transcoding/status"
        },
        {id: "users", icon: Users, label: t('admin.users'), path: "/admin/users"},
        {id: "categories", icon: FolderTree, label: t('admin.categories'), path: "/admin/categories"},
        {id: "channels", icon: Radio, label: t('admin.channels'), path: "/admin/channels"},
        {id: "tags", icon: Tags, label: t('admin.tags'), path: "/admin/tags"},
        {id: "comments", icon: MessageSquare, label: t('admin.comments'), path: "/admin/comments"},
        {id: "playlists", icon: PlayCircle, label: t('admin.playlists'), path: "/admin/playlists"},
        {id: "content", icon: FileText, label: t('admin.content'), path: "/admin/content"},
        {id: "settings", icon: Settings, label: t('admin.settings'), path: "/admin/settings"},
    ];

    // 生成面包屑路径
    const getBreadcrumbs = () => {
        const path = routerState.location.pathname;
        const breadcrumbs = [
            {label: "首页", path: "/admin", icon: Home}
        ];

        if (path.startsWith("/admin/media")) {
            breadcrumbs.push({label: "Media", path: "/admin/media"});
        } else if (path.startsWith("/admin/transcoding/profiles")) {
            breadcrumbs.push({label: "Media", path: "/admin/media"});
            breadcrumbs.push({label: "Encode profiles", path: "/admin/transcoding/profiles"});
        } else if (path.startsWith("/admin/transcoding/status")) {
            breadcrumbs.push({label: "Media", path: "/admin/media"});
            breadcrumbs.push({label: "Transcoding status", path: "/admin/transcoding/status"});
        } else if (path.startsWith("/admin/users")) {
            breadcrumbs.push({label: "Users", path: "/admin/users"});
        } else if (path.startsWith("/admin/categories")) {
            breadcrumbs.push({label: "Categories", path: "/admin/categories"});
        } else if (path.startsWith("/admin/channels")) {
            breadcrumbs.push({label: "Channels", path: "/admin/channels"});
        } else if (path.startsWith("/admin/tags")) {
            breadcrumbs.push({label: "Tags", path: "/admin/tags"});
        } else if (path.startsWith("/admin/comments")) {
            breadcrumbs.push({label: "Comments", path: "/admin/comments"});
        } else if (path.startsWith("/admin/playlists")) {
            breadcrumbs.push({label: "Playlists", path: "/admin/playlists"});
        } else if (path.startsWith("/admin/content")) {
            breadcrumbs.push({label: "Content", path: "/admin/content"});
        } else if (path.startsWith("/admin/settings")) {
            breadcrumbs.push({label: "Settings", path: "/admin/settings"});
        }

        return breadcrumbs;
    };

    const breadcrumbs = getBreadcrumbs();

    return (
        <div className="min-h-screen bg-background flex">
            {/* Sidebar */}
            <aside
                className={`${sidebarCollapsed ? 'w-20' : 'w-64'} bg-slate-900 text-white flex-shrink-0 flex flex-col transition-all duration-300 ease-in-out`}>
                <div className={`${sidebarCollapsed ? 'p-2' : 'p-6'} flex items-center justify-center`}>
                    <Link to="/admin" className={`flex items-center justify-center transition-all duration-300 ease-in-out ${sidebarCollapsed ? '' : 'w-full'}`}>
                        {sidebarCollapsed ? (
                            <div className="w-14 h-14 bg-blue-400 rounded-lg flex items-center justify-center text-white font-bold text-xl">OC</div>
                        ) : (
                            <span className="text-xl font-bold tracking-tight text-blue-400">OrigCMS Admin</span>
                        )}
                    </Link>
                </div>
                <nav className={`flex-grow space-y-2 ${sidebarCollapsed ? 'px-2' : 'px-4'}`}>
                    {menuItems.map((item) => (
                        <NavItem
                            key={item.path}
                            to={item.path}
                            icon={<item.icon size={24}/>}
                            label={item.label}
                            exact={item.path === "/admin"}
                            collapsed={sidebarCollapsed}
                        />
                    ))}
                </nav>
                <div className={`${sidebarCollapsed ? 'p-3' : 'p-6'} border-t border-slate-800`}>
                    <Link
                        to="/"
                        className={`flex items-center ${sidebarCollapsed ? 'justify-center' : 'space-x-3'} text-slate-400 hover:text-white transition-colors`}
                    >
                        <LogOut size={20}/>
                        {!sidebarCollapsed && <span>Exit Admin</span>}
                    </Link>
                </div>
            </aside>

            {/* Main Content */}
            <div className="flex-grow flex flex-col min-w-0 relative">
                {/* Floating toggle button on the divider line */}
                <button
                    onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
                    className="absolute top-12 z-50 w-8 h-8 bg-white dark:bg-gray-800 border border-slate-200 dark:border-slate-700 rounded-full shadow-md flex items-center justify-center hover:bg-slate-50 dark:hover:bg-slate-700 transition-all duration-200 hover:scale-110"
                    style={{ left: -16 }}
                    title={sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'}
                >
                    {sidebarCollapsed ? (
                        <PanelLeftOpen size={18} className="text-slate-600 dark:text-slate-300" />
                    ) : (
                        <PanelLeftClose size={18} className="text-slate-600 dark:text-slate-300" />
                    )}
                </button>

                <header className="h-16 bg-card border-b flex items-center justify-between px-8">
                    <div className="flex items-center gap-2">
                        {breadcrumbs.map((crumb, index) => (
                            <React.Fragment key={crumb.path}>
                                {index > 0 && (
                                    <span className="text-muted-foreground"> {'>'} </span>
                                )}
                                <Link
                                    to={crumb.path}
                                    className={`flex items-center gap-1 text-sm ${
                                        index === breadcrumbs.length - 1
                                            ? 'text-foreground font-medium'
                                            : 'text-muted-foreground hover:text-foreground'
                                    }`}
                                >
                                    {crumb.icon && <crumb.icon size={14}/>}
                                    <span>{crumb.label}</span>
                                </Link>
                            </React.Fragment>
                        ))}
                    </div>
                    <div className="flex items-center space-x-4">
                        <div
                            className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold">A
                        </div>
                        <span className="text-sm font-medium text-foreground">Administrator</span>
                    </div>
                </header>
                <main className="flex-grow p-8 overflow-auto">
                    <Outlet/>
                </main>
            </div>
        </div>
    );
};

const NavItem = ({to, icon, label, exact = false, collapsed = false}: {
    to: string;
    icon: React.ReactNode;
    label: string;
    exact?: boolean;
    collapsed?: boolean;
}) => {
    const state = useRouterState();
    const isActive = exact
        ? state.location.pathname === to
        : state.location.pathname.startsWith(to);
    return (
        <Link
            to={to}
            className={`flex items-center rounded-lg transition-all duration-300 ease-in-out ${
                collapsed
                    ? 'justify-center w-12 h-12 mx-auto'
                    : 'space-x-3 px-4 py-2'
            } ${
                isActive ? 'bg-blue-600 text-white' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
            }`}
            title={collapsed ? label : undefined}
        >
            {collapsed ? React.cloneElement(icon as React.ReactElement, { size: 28 }) : icon}
            {!collapsed && <span className="font-medium text-sm">{label}</span>}
        </Link>
    );
};

export default AdminLayout;
