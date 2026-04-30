import React, {useState, useEffect} from 'react';
import {Outlet} from '@tanstack/react-router';
import Header from '@/components/portal/Header';
import Sidebar from '@/components/portal/Sidebar';
import {useTheme} from '@/themes';

const PortalLayout = () => {
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
    const [isDesktop, setIsDesktop] = useState(false);
    const {isDark, toggleDark} = useTheme();

    useEffect(() => {
        const saved = localStorage.getItem('sidebarCollapsed') === 'true';
        setSidebarCollapsed(saved);
    }, []);

    useEffect(() => {
        const checkScreen = () => setIsDesktop(window.innerWidth >= 768);
        checkScreen();
        window.addEventListener('resize', checkScreen);
        return () => window.removeEventListener('resize', checkScreen);
    }, []);

    const toggleSidebar = () => {
        const next = !sidebarCollapsed;
        setSidebarCollapsed(next);
        localStorage.setItem('sidebarCollapsed', String(next));
    };

    return (
        <div className="min-h-screen bg-background transition-colors">
            <Header
                onToggleSidebar={toggleSidebar}
                sidebarCollapsed={sidebarCollapsed}
                darkMode={isDark}
                onToggleDarkMode={toggleDark}
            />

            <Sidebar
                collapsed={sidebarCollapsed}
                onToggleCollapse={toggleSidebar}
            />

            <main
                className="pt-14 min-h-screen transition-all duration-300 bg-background relative z-10"
                style={{
                    marginLeft: isDesktop ? (sidebarCollapsed ? 72 : 240) : 0
                }}
            >
                <div className="p-4 md:p-6 lg:p-8 max-w-7xl mx-auto">
                    <Outlet/>
                </div>
            </main>
        </div>
    );
};

export default PortalLayout;