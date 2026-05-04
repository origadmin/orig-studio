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
    Info,
    Bell,
    Radio,
    TrendingUp,
    UserCircle,
    FileText,
    Tv,
} from 'lucide-react';
import type {NavSection} from '@/types/nav';

export const NAV_CONFIG: NavSection[] = [
    {
        id: 'browse',
        title: 'nav.browse',
        items: [
            {id: 'home', label: 'nav.home', to: '/', icon: Home},
            {id: 'featured', label: 'nav.featured', to: '/featured', icon: Star, module: 'videos'},
            {id: 'latest', label: 'nav.latest', to: '/latest', icon: Clock, module: 'videos'},
            {id: 'categories', label: 'nav.categories', to: '/categories', icon: LayoutGrid},
            {id: 'tags', label: 'nav.tags', to: '/tags', icon: Tag},
            {id: 'members', label: 'nav.members', to: '/members', icon: UsersIcon},
            {id: 'articles', label: 'nav.articles', to: '/articles', icon: FileText, module: 'articles'},
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
                module: 'videos',
            },
        ],
    },
    {
        id: 'you',
        title: 'nav.you',
        requiresAuth: true,
        items: [
            {id: 'my-profile', label: 'nav.myProfile', to: '/@__dynamic__', icon: UserCircle, isDynamic: true},
            {id: 'my-channels', label: 'nav.myChannels', to: '/me/channels', icon: Tv, module: 'videos'},
            {id: 'my-videos', label: 'nav.myVideos', to: '/me/videos', icon: Video, module: 'videos'},
            {id: 'my-articles', label: 'nav.myArticles', to: '/me/articles', icon: FileText, module: 'articles'},
        ],
    },
    {
        id: 'library',
        title: 'nav.library',
        requiresAuth: true,
        items: [
            {id: 'history', label: 'nav.history', to: '/me/history', icon: History},
            {id: 'favorites', label: 'nav.favorites', to: '/me/favorites', icon: Heart},
            {id: 'playlists', label: 'nav.playlists', to: '/me/playlists', icon: ListVideo, module: 'videos'},
        ],
    },
    {
        id: 'notifications',
        title: 'nav.notifications',
        requiresAuth: true,
        items: [
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
        id: 'other',
        title: 'nav.other',
        items: [
            {id: 'about', label: 'nav.about', to: '/about', icon: Info},
        ],
    },
];
