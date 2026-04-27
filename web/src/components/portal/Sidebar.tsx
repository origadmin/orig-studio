import React, {useState, useEffect, useMemo, useRef} from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {useAuth} from '../../hooks/useAuth';
import {useSubscribedChannels} from '../../hooks/useSubscriptions';
import {NAV_CONFIG} from '../../config/navigation';
import type {NavSection, NavItem} from '../../types/nav';

interface SidebarProps {
    collapsed?: boolean;
    onToggleCollapse?: () => void;
}

interface RenderNavItem {
    id: string;
    icon: React.ReactNode;
    label: string;
    to: string;
}

function toRenderItems(items: NavItem[]): RenderNavItem[] {
    return items.map((item) => ({
        id: item.id,
        icon: item.icon ? <item.icon size={22}/> : null,
        label: item.label,
        to: item.to,
    }));
}

const Sidebar: React.FC<SidebarProps> = ({collapsed = false}) => {
    const {t} = useTranslation();
    const location = useLocation();
    const pathname = location.pathname;
    const {isAuthenticated, isAdmin} = useAuth();
    const {channels: subChannels} = useSubscribedChannels();
    const [hoveredSection, setHoveredSection] = useState<NavSection | null>(null);
    const [hoveredItems, setHoveredItems] = useState<RenderNavItem[]>([]);
    const [hoverPos, setHoverPos] = useState({top: 0});
    const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        return () => {
            if (closeTimerRef.current) clearTimeout(closeTimerRef.current);
        };
    }, []);

    const isActive = (to: string) => pathname === to || pathname.startsWith(to + '/');

    const visibleSections = useMemo((): { section: NavSection; items: RenderNavItem[] }[] => {
        return NAV_CONFIG.filter((section) => {
            if (section.requiresAuth && !isAuthenticated) return false;
            if (section.requiresAdmin && !isAdmin) return false;
            return true;
        }).map((section) => {
            const baseItems = toRenderItems(section.items);
            if (section.id === 'subscriptions' && subChannels.length > 0) {
                return {section, items: [...baseItems, ...subChannels as RenderNavItem[]]};
            }
            return {section, items: baseItems};
        });
    }, [isAuthenticated, isAdmin, subChannels]);

    const handleSectionEnter = (e: React.MouseEvent, section: NavSection, items: RenderNavItem[]) => {
        if (closeTimerRef.current) clearTimeout(closeTimerRef.current);
        const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
        setHoverPos({top: rect.top});
        setHoveredSection(section);
        setHoveredItems(items);
    };

    const handleSectionLeave = () => {
        closeTimerRef.current = setTimeout(() => {
            setHoveredSection(null);
            setHoveredItems([]);
        }, 150);
    };

    const handlePopupEnter = () => {
        if (closeTimerRef.current) clearTimeout(closeTimerRef.current);
    };

    const handlePopupLeave = () => {
        closeTimerRef.current = setTimeout(() => {
            setHoveredSection(null);
            setHoveredItems([]);
        }, 100);
    };

    const FullNavLink = ({item}: { item: RenderNavItem }) => {
        const active = isActive(item.to);
        return (
            <Link
                to={item.to}
                className={`flex items-center gap-3 py-2.5 px-3 rounded-lg transition-colors ${
                    active
                        ? 'bg-gray-100 dark:bg-gray-800 font-medium'
                        : 'hover:bg-gray-100 dark:hover:bg-gray-800'
                }`}
            >
                <span className={`flex-shrink-0 ${active ? 'text-black dark:text-white' : 'text-gray-700 dark:text-gray-300'}`}>
                    {item.icon}
                </span>
                <span className="text-[14px]">{t(item.label)}</span>
            </Link>
        );
    };

    const CollapsedSectionButton = ({section, items}: { section: NavSection; items: RenderNavItem[] }) => (
        <button
            onMouseEnter={(e) => handleSectionEnter(e, section, items)}
            onMouseLeave={handleSectionLeave}
            className="w-full flex flex-col items-center gap-1 py-2.5 px-1 rounded-lg transition-colors hover:bg-gray-100 dark:hover:bg-gray-800"
        >
            {items[0]?.icon && (
                <span className="text-gray-700 dark:text-gray-300 scale-110">{items[0].icon}</span>
            )}
            <span className="text-[12px] font-medium text-gray-900 dark:text-gray-100 leading-tight">
                {t(section.title)}
            </span>
        </button>
    );

    const PopupNavLink = ({item}: { item: RenderNavItem }) => {
        const active = isActive(item.to);
        return (
            <Link
                to={item.to}
                className={`flex items-center gap-3 px-4 py-2.5 rounded-lg transition-colors ${
                    active
                        ? 'bg-gray-100 dark:bg-gray-800 font-medium'
                        : 'hover:bg-gray-100 dark:hover:bg-gray-800'
                }`}
                onMouseEnter={handlePopupEnter}
            >
                <span className={`flex-shrink-0 ${active ? 'text-black dark:text-white' : 'text-gray-700 dark:text-gray-300'}`}>
                    {item.icon}
                </span>
                <span className="text-sm">{t(item.label)}</span>
            </Link>
        );
    };

    const FullNavSection = ({items, title}: { items: RenderNavItem[]; title?: string }) => (
        <div className="py-0.5">
            {title && (
                <h3 className="px-3 py-1.5 text-xs font-medium text-gray-500 dark:text-muted-foreground uppercase tracking-wider">
                    {title}
                </h3>
            )}
            <div className="space-y-0.5 px-2">
                {items.map((item) => (
                    <FullNavLink key={item.id} item={item}/>
                ))}
            </div>
        </div>
    );

    const FullDivider = () => (
        <div className="border-t border-gray-200/60 dark:border-gray-700/60 my-1.5 mx-3"/>
    );

    const CollapsedDivider = () => (
        <div className="border-t border-gray-200/60 dark:border-gray-700/60 my-1 mx-2"/>
    );

    const width = collapsed ? 72 : 240;

    const renderFullContent = () => {
        const sections: React.ReactNode[] = [];
        visibleSections.forEach(({section, items}, idx) => {
            if (idx > 0) sections.push(<FullDivider key={`d-${idx}`}/>);
            sections.push(<FullNavSection key={section.id} items={items} title={t(section.title)}/>);
        });
        return sections;
    };

    const renderCollapsedContent = () => {
        const sections: React.ReactNode[] = [];
        visibleSections.forEach(({section, items}, idx) => {
            if (idx > 0) sections.push(<CollapsedDivider key={`d-${idx}`}/>);
            sections.push(
                <CollapsedSectionButton key={section.id} section={section} items={items}/>
            );
        });
        return sections;
    };

    return (
        <>
            <aside
                style={{width}}
                className="fixed left-0 top-14 bottom-0 bg-white dark:bg-gray-900 z-40 hidden md:flex flex-col transition-all duration-200"
            >
                {collapsed ? (
                    <nav className="flex-1 overflow-hidden py-2 relative">
                        {renderCollapsedContent()}

                        {hoveredSection && hoveredItems.length > 0 && (
                            <div
                                className="fixed bg-white dark:bg-gray-900 rounded-xl shadow-xl border border-gray-200/60 dark:border-gray-700/60 py-1.5 z-50 animate-in fade-in slide-in-from-left-1 duration-150"
                                style={{
                                    left: width + 6,
                                    top: hoverPos.top,
                                    minWidth: 200,
                                    maxHeight: Math.min(480, window.innerHeight - hoverPos.top - 16),
                                    overflowY: 'auto',
                                }}
                                onMouseEnter={handlePopupEnter}
                                onMouseLeave={handlePopupLeave}
                            >
                                <div className="px-3 py-1.5 border-b border-gray-100 dark:border-gray-800">
                                    <span className="text-xs font-semibold text-gray-500 dark:text-muted-foreground uppercase tracking-wider">
                                        {t(hoveredSection.title)}
                                    </span>
                                </div>
                                <div className="py-0.5">
                                    {hoveredItems.map((item) => (
                                        <PopupNavLink key={item.id} item={item}/>
                                    ))}
                                </div>
                            </div>
                        )}
                    </nav>
                ) : (
                    <nav className="flex-1 overflow-y-auto yt-scrollbar py-2">
                        {renderFullContent()}
                    </nav>
                )}
            </aside>
        </>
    );
};

export default Sidebar;