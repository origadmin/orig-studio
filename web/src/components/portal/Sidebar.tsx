import React, {useState, useEffect, useMemo, useRef} from 'react';
import {Link, useLocation} from '@tanstack/react-router';
import {useTranslation} from 'react-i18next';
import {useAuth} from '@/hooks/useAuth';
import {useSubscribedChannels, type ChannelSummary} from '@/hooks/useSubscriptions';
import {useModuleState} from '@/contexts/ModuleConfigContext';
import {NAV_CONFIG} from '@/config/navigation';
import type {NavSection, NavItem} from '@/types/nav';
import type {User as AuthUser} from '@/contexts/auth/types';
import {Avatar, AvatarFallback, AvatarImage} from '@/components/ui/avatar';
import {getImageUrl, handleImageError} from '@/lib/imageUtils';
import {ChevronDown, ChevronUp} from 'lucide-react';

interface SidebarProps {
    collapsed?: boolean;
    onToggleCollapse?: () => void;
}

interface RenderNavItem {
    id: string;
    icon: React.ReactNode;
    label: string;
    to: string;
    params?: Record<string, string>;
}

const SUBS_DEFAULT_SHOW = 5;

function toRenderItems(items: NavItem[], currentUser?: AuthUser | null): RenderNavItem[] {
    return items.map((item) => {
        let to = item.to;
        let params: Record<string, string> | undefined;
        // Resolve dynamic paths: replace __dynamic__ placeholder with current user's username
        if (item.isDynamic && item.to.includes('__dynamic__')) {
            const username = currentUser?.username;
            if (username) {
                to = '/$handle';
                params = {handle: `@${username}`};
            } else {
                to = '/auth/signin';
            }
        }
        return {
            id: item.id,
            icon: item.icon ? <item.icon size={22}/> : null,
            label: item.label,
            to,
            params,
        };
    });
}

const Sidebar: React.FC<SidebarProps> = ({collapsed = false}) => {
    const {t} = useTranslation();
    const location = useLocation();
    const pathname = location.pathname;
    const {isAuthenticated, user} = useAuth();
    const {modules} = useModuleState();
    const {channels: subChannels, channelDetails} = useSubscribedChannels();
    const [hoveredSection, setHoveredSection] = useState<NavSection | null>(null);
    const [hoveredItems, setHoveredItems] = useState<RenderNavItem[]>([]);
    const [hoverPos, setHoverPos] = useState({top: 0});
    const closeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
    const [subsExpanded, setSubsExpanded] = useState(false);

    useEffect(() => {
        return () => {
            if (closeTimerRef.current) clearTimeout(closeTimerRef.current);
        };
    }, []);

    const isActive = (to: string) => pathname === to || pathname.startsWith(to + '/');

    const visibleSections = useMemo((): { section: NavSection; items: RenderNavItem[] }[] => {
        return NAV_CONFIG.filter((section) => {
            if (section.requiresAuth && !isAuthenticated) return false;
            return true;
        }).map((section) => {
            const filteredItems = section.items.filter((item) => {
                if (!item.module) return true;
                return modules[item.module] === true;
            });
            const baseItems = toRenderItems(filteredItems, user);
            return {section, items: baseItems};
        }).filter(({items}) => items.length > 0);
    }, [isAuthenticated, user, modules]);

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
                params={item.params}
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

    // YouTube-style subscription channel link with avatar
    const SubsChannelLink = ({channel}: { channel: ChannelSummary }) => {
        const linkTo = channel.short_token ? '/c/$id' : channel.username ? '/$handle' : '/u/$id';
        const linkParams = channel.short_token ? {id: channel.short_token} : channel.username ? {handle: `@${channel.username}`} : {id: channel.id};
        const active = channel.short_token ? isActive(`/c/${channel.short_token}`) : channel.username ? isActive(`/@${channel.username}`) : isActive(`/u/${channel.id}`);
        const displayName = channel.name || channel.username;
        return (
            <Link
                to={linkTo}
                params={linkParams}
                className={`flex items-center gap-3 py-1.5 px-3 rounded-lg transition-colors ${
                    active
                        ? 'bg-gray-100 dark:bg-gray-800 font-medium'
                        : 'hover:bg-gray-100 dark:hover:bg-gray-800'
                }`}
            >
                <Avatar className="w-6 h-6 flex-shrink-0">
                    <AvatarImage
                        src={getImageUrl(channel.avatar, 'avatar')}
                        loading="lazy"
                        onError={(e) => handleImageError(e, 'avatar')}
                    />
                    <AvatarFallback className="text-[10px]">
                        {displayName ? displayName.charAt(0).toUpperCase() : 'U'}
                    </AvatarFallback>
                </Avatar>
                <span className={`text-[14px] truncate ${active ? 'text-black dark:text-white' : 'text-gray-700 dark:text-gray-300'}`}>
                    {displayName}
                </span>
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
                params={item.params}
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

    // Popup version of subscription channel link with avatar
    const PopupSubsChannelLink = ({channel}: { channel: ChannelSummary }) => {
        const linkTo = channel.short_token ? '/c/$id' : channel.username ? '/$handle' : '/u/$id';
        const linkParams = channel.short_token ? {id: channel.short_token} : channel.username ? {handle: `@${channel.username}`} : {id: channel.id};
        const active = channel.short_token ? isActive(`/c/${channel.short_token}`) : channel.username ? isActive(`/@${channel.username}`) : isActive(`/u/${channel.id}`);
        const displayName = channel.name || channel.username;
        return (
            <Link
                to={linkTo}
                params={linkParams}
                className={`flex items-center gap-3 px-4 py-2 rounded-lg transition-colors ${
                    active
                        ? 'bg-gray-100 dark:bg-gray-800 font-medium'
                        : 'hover:bg-gray-100 dark:hover:bg-gray-800'
                }`}
                onMouseEnter={handlePopupEnter}
            >
                <Avatar className="w-6 h-6 flex-shrink-0">
                    <AvatarImage
                        src={getImageUrl(channel.avatar, 'avatar')}
                        loading="lazy"
                        onError={(e) => handleImageError(e, 'avatar')}
                    />
                    <AvatarFallback className="text-[10px]">
                        {displayName ? displayName.charAt(0).toUpperCase() : 'U'}
                    </AvatarFallback>
                </Avatar>
                <span className={`text-sm truncate ${active ? 'text-black dark:text-white' : 'text-gray-700 dark:text-gray-300'}`}>
                    {displayName}
                </span>
            </Link>
        );
    };

    const FullNavSection = ({items, title, sectionId}: { items: RenderNavItem[]; title?: string; sectionId?: string }) => {
        const isSubscriptions = sectionId === 'subscriptions';
        const hasChannels = isSubscriptions && channelDetails.length > 0;
        const displayedChannels = hasChannels
            ? (subsExpanded ? channelDetails : channelDetails.slice(0, SUBS_DEFAULT_SHOW))
            : [];
        const canExpand = hasChannels && channelDetails.length > SUBS_DEFAULT_SHOW;

        return (
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
                    {/* YouTube-style subscription channels with avatars */}
                    {displayedChannels.map((channel) => (
                        <SubsChannelLink key={`ch-${channel.id}`} channel={channel}/>
                    ))}
                    {canExpand && (
                        <button
                            onClick={() => setSubsExpanded(!subsExpanded)}
                            className="flex items-center gap-3 py-2 px-3 rounded-lg transition-colors hover:bg-gray-100 dark:hover:bg-gray-800 w-full text-left"
                        >
                            {subsExpanded ? (
                                <ChevronUp className="w-5 h-5 text-gray-700 dark:text-gray-300 flex-shrink-0"/>
                            ) : (
                                <ChevronDown className="w-5 h-5 text-gray-700 dark:text-gray-300 flex-shrink-0"/>
                            )}
                            <span className="text-[14px] text-gray-700 dark:text-gray-300">
                                {subsExpanded ? t('nav.showLess') : t('nav.showMore')}
                            </span>
                        </button>
                    )}
                </div>
            </div>
        );
    };

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
            sections.push(<FullNavSection key={section.id} items={items} title={t(section.title)} sectionId={section.id}/>);
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

    // Build popup items for subscriptions section (includes channel links)
    const subsPopupItems = useMemo((): RenderNavItem[] => {
        return subChannels as RenderNavItem[];
    }, [subChannels]);

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
                                    {hoveredSection.id === 'subscriptions' && channelDetails.length > 0 ? (
                                        <>
                                            {hoveredItems.map((item) => (
                                                <PopupNavLink key={item.id} item={item}/>
                                            ))}
                                            {channelDetails.map((channel) => (
                                                <PopupSubsChannelLink key={`ch-${channel.id}`} channel={channel}/>
                                            ))}
                                        </>
                                    ) : (
                                        hoveredItems.map((item) => (
                                            <PopupNavLink key={item.id} item={item}/>
                                        ))
                                    )}
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
