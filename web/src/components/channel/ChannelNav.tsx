import React from 'react';
import {useTranslation} from 'react-i18next';
import {Film, ListVideo, Info, Home, MessageSquare} from 'lucide-react';

interface ChannelNavProps {
    activeTab: string;
    onTabChange: (tab: string) => void;
    isOwner?: boolean;
}

const ChannelNav: React.FC<ChannelNavProps> = ({
    activeTab,
    onTabChange,
    isOwner = false,
}) => {
    const {t} = useTranslation();

    const tabs = [
        {id: 'home', label: t('channel.tabHome') || 'Home', icon: Home},
        {id: 'videos', label: t('channel.tabVideos') || 'Videos', icon: Film},
        {id: 'playlists', label: t('channel.tabPlaylists') || 'Playlists', icon: ListVideo},
        {id: 'community', label: t('channel.tabCommunity') || 'Community', icon: MessageSquare},
        {id: 'about', label: t('channel.tabAbout') || 'About', icon: Info},
    ];

    return (
        <div className="border-b border-border sticky top-16 z-30 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <nav className="flex gap-1 sm:gap-2 overflow-x-auto scrollbar-hide -mb-px" role="tablist">
                    {tabs.map((tab) => {
                        const Icon = tab.icon;
                        const isActive = activeTab === tab.id;
                        return (
                            <button
                                key={tab.id}
                                onClick={() => onTabChange(tab.id)}
                                role="tab"
                                aria-selected={isActive}
                                className={`relative py-3 sm:py-4 px-3 sm:px-4 font-medium text-sm whitespace-nowrap transition-all duration-200 group flex items-center gap-1.5 ${
                                    isActive
                                        ? 'text-primary'
                                        : 'text-muted-foreground hover:text-foreground'
                                }`}
                            >
                                <Icon className={`w-4 h-4 transition-colors ${
                                    isActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'
                                }`}/>
                                <span className="relative z-10">{tab.label}</span>
                                {isActive && (
                                    <span
                                        className="absolute bottom-0 left-0 h-0.5 bg-primary rounded-full transition-all duration-300"
                                        style={{
                                            width: 'calc(100% + 8px)',
                                            left: '-4px',
                                        }}
                                    />
                                )}
                                {!isActive && (
                                    <span
                                        className="absolute bottom-0 left-0 h-0.5 bg-transparent group-hover:bg-muted-foreground/30 rounded-full transition-all duration-200"
                                        style={{
                                            width: 'calc(100% + 8px)',
                                            left: '-4px',
                                        }}
                                    />
                                )}
                            </button>
                        );
                    })}
                </nav>
            </div>
        </div>
    );
};

export default ChannelNav;
