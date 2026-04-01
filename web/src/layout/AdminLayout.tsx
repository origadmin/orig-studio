/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React from 'react';
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
    PlayCircle
} from 'lucide-react';
import {useTranslation} from 'react-i18next';

const AdminLayout = () => {
    const {t} = useTranslation();

    const menuItems = [
        {id: "dashboard", icon: LayoutDashboard, label: t('admin.dashboard'), path: "/admin"},
        {id: "media", icon: Film, label: t('admin.media'), path: "/admin/media"},
        {id: "users", icon: Users, label: t('admin.users'), path: "/admin/users"},
        {id: "categories", icon: FolderTree, label: t('admin.categories'), path: "/admin/categories"},
        {id: "channels", icon: Radio, label: t('admin.channels'), path: "/admin/channels"},
        {id: "tags", icon: Tags, label: t('admin.tags'), path: "/admin/tags"},
        {id: "comments", icon: MessageSquare, label: t('admin.comments'), path: "/admin/comments"},
        {id: "playlists", icon: PlayCircle, label: t('admin.playlists'), path: "/admin/playlists"},
        {id: "content", icon: Settings, label: t('admin.content'), path: "/admin/content"},
        {id: "settings", icon: Settings, label: t('admin.settings'), path: "/admin/settings"},
    ];

    return (
        <div className="min-h-screen bg-gray-50 flex">
            {/* Sidebar */}
            <aside className="w-64 bg-slate-900 text-white flex-shrink-0 flex flex-col">
                <div className="p-6">
                    <Link to="/admin" className="text-xl font-bold tracking-tight text-blue-400">
                        OrigCMS Admin
                    </Link>
                </div>
                <nav className="flex-grow px-4 space-y-2">
                    {menuItems.map((item) => (
                        <NavItem
                            key={item.path}
                            to={item.path}
                            icon={<item.icon size={20}/>}
                            label={item.label}
                            exact={item.path === "/admin"}
                        />
                    ))}
                </nav>
                <div className="p-6 border-t border-slate-800">
                    <Link
                        to="/"
                        className="flex items-center space-x-3 text-slate-400 hover:text-white transition-colors"
                    >
                        <LogOut size={20}/>
                        <span>Exit Admin</span>
                    </Link>
                </div>
            </aside>

            {/* Main Content */}
            <div className="flex-grow flex flex-col min-w-0">
                <header className="h-16 bg-white border-b flex items-center justify-between px-8">
                    <h2 className="text-lg font-semibold text-gray-800">Backstage Control</h2>
                    <div className="flex items-center space-x-4">
                        <div
                            className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold">A
                        </div>
                        <span className="text-sm font-medium">Administrator</span>
                    </div>
                </header>
                <main className="flex-grow p-8 overflow-auto">
                    <Outlet/>
                </main>
            </div>
        </div>
    );
};

const NavItem = ({to, icon, label, exact = false}: {
    to: string;
    icon: React.ReactNode;
    label: string;
    exact?: boolean
}) => {
    const state = useRouterState();
    const isActive = exact
        ? state.location.pathname === to
        : state.location.pathname.startsWith(to);
    return (
        <Link
            to={to}
            className={`flex items-center space-x-3 px-4 py-3 rounded-lg transition-all ${
                isActive ? 'bg-blue-600 text-white' : 'text-slate-400 hover:bg-slate-800 hover:text-white'
            }`}
        >
            {icon}
            <span className="font-medium text-sm">{label}</span>
        </Link>
    );
};

export default AdminLayout;
