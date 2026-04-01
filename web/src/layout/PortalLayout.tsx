/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * PortalLayout: Header + Sidebar + Content
 */

import React, {useState, useEffect} from 'react';
import {Outlet} from '@tanstack/react-router';
import Header from '../components/portal/Header';
import Sidebar from '../components/portal/Sidebar';

const PortalLayout = () => {
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
    const [sidebarOpen, setSidebarOpen] = useState(false); // 移动端 overlay
    const [darkMode, setDarkMode] = useState(false);

    // 恢复主题
    useEffect(() => {
        const saved = localStorage.getItem('darkMode') === 'true';
        setDarkMode(saved);
        document.documentElement.classList.toggle('dark', saved);
    }, []);

    // 恢复侧边栏状态
    useEffect(() => {
        const saved = localStorage.getItem('sidebarCollapsed') === 'true';
        setSidebarCollapsed(saved);
    }, []);

    const toggleDarkMode = () => {
        const next = !darkMode;
        setDarkMode(next);
        localStorage.setItem('darkMode', String(next));
        document.documentElement.classList.toggle('dark', next);
    };

    const toggleSidebar = () => {
        if (window.innerWidth < 768) {
            setSidebarOpen(!sidebarOpen);
        } else {
            const next = !sidebarCollapsed;
            setSidebarCollapsed(next);
            localStorage.setItem('sidebarCollapsed', String(next));
        }
    };

    // 关闭移动端 sidebar
    const closeMobileSidebar = () => setSidebarOpen(false);

    return (
        <div className="min-h-screen bg-gray-50 dark:bg-gray-950 transition-colors">
            {/* Header */}
            <Header
                onToggleSidebar={toggleSidebar}
                sidebarCollapsed={sidebarCollapsed}
            />

            {/* 移动端遮罩 */}
            {sidebarOpen && (
                <div
                    className="fixed inset-0 bg-black/30 z-40 md:hidden"
                    onClick={closeMobileSidebar}
                />
            )}

            {/* Sidebar */}
            <div className="hidden md:block">
                <Sidebar
                    darkMode={darkMode}
                    onToggleDarkMode={toggleDarkMode}
                    collapsed={sidebarCollapsed}
                    onToggleCollapse={toggleSidebar}
                />
            </div>

            {/* 移动端 Sidebar (overlay) */}
            <div className="md:hidden">
                <div
                    className={`fixed left-0 top-14 bottom-0 z-50 transition-transform duration-300 ${
                        sidebarOpen ? 'translate-x-0' : '-translate-x-full'
                    }`}
                >
                    <Sidebar
                        darkMode={darkMode}
                        onToggleDarkMode={toggleDarkMode}
                        collapsed={false}
                    />
                </div>
            </div>

            {/* Main Content */}
            <main
                className={`pt-14 transition-all duration-300 ${
                    sidebarCollapsed ? 'md:ml-[60px]' : 'md:ml-[220px]'
                }`}
            >
                <div className="p-4 md:p-6 lg:p-8 max-w-[1600px]">
                    <Outlet/>
                </div>
            </main>
        </div>
    );
};

export default PortalLayout;
