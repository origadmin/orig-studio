import {
    Home,
    Star,
    Clock,
    Tag,
    LayoutGrid,
    Users as UsersIcon,
    Video,
    History,
    Heart,
    ListVideo,
    Upload,
    Shield,
    Settings,
    Info,
    Bell,
    Radio,
    TrendingUp,
} from 'lucide-react';
import type {NavSection} from '@/types/nav';

export const NAV_CONFIG: NavSection[] = [
    {
        id: 'browse',
        title: 'nav.browse',
        items: [
            {id: 'home', label: 'nav.home', to: '/', icon: Home},
            {id: 'featured', label: 'nav.featured', to: '/featured', icon: Star},
            {id: 'latest', label: 'nav.latest', to: '/latest', icon: Clock},
            {id: 'categories', label: 'nav.categories', to: '/categories', icon: LayoutGrid},
            {id: 'tags', label: 'nav.tags', to: '/tags', icon: Tag},
            {id: 'members', label: 'nav.members', to: '/members', icon: UsersIcon},
        ],
    },
    {
        id: 'subscriptions',
        title: 'nav.subscriptions',
        requiresAuth: true,
        items: [
            {
                id: 'subs-feed',
                label: 'nav.subsFeed',
                to: '/subscriptions',
                icon: Radio,
            },
        ],
    },
    {
        id: 'you',
        title: 'nav.you',
        requiresAuth: true,
        items: [
            {id: 'history', label: 'nav.history', to: '/me/history', icon: History},
            {id: 'my-videos', label: 'nav.myVideos', to: '/me/videos', icon: Video},
            {id: 'upload', label: 'nav.upload', to: '/me/upload', icon: Upload},
            {id: 'favorites', label: 'nav.favorites', to: '/me/favorites', icon: Heart},
            {id: 'playlists', label: 'nav.playlists', to: '/me/playlists', icon: ListVideo},
            {id: 'notifications', label: 'nav.notifications', to: '/me/notifications', icon: Bell},
        ],
    },
    {
        id: 'explore',
        title: 'nav.explore',
        items: [
            {
                id: 'trending',
                label: 'nav.trending',
                to: '/explore',
                icon: TrendingUp,
            },
        ],
    },
    {
        id: 'admin',
        title: 'nav.admin',
        requiresAdmin: true,
        items: [
            {id: 'media-admin', label: 'nav.mediaAdmin', to: '/admin/media', icon: Shield},
            {id: 'users-admin', label: 'nav.usersAdmin', to: '/admin/users', icon: UsersIcon},
            {id: 'settings-admin', label: 'nav.settingsAdmin', to: '/admin/settings', icon: Settings},
        ],
    },
    {
        id: 'other',
        title: 'nav.other',
        items: [
            {id: 'about', label: 'nav.about', to: '/about', icon: Info},
        ],
    },
];
