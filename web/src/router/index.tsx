/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import React, {lazy, Suspense} from 'react';
import {
    createRouter,
    createRootRoute,
    createRoute,
    RouterProvider,
    Outlet,
    redirect,
} from '@tanstack/react-router';

import PortalLayout from '../layout/PortalLayout';
import AdminLayout from '../layout/AdminLayout';
import {getStoredToken, getStoredUser} from '../hooks/useAuth';

// ── Lazy pages ────────────────────────────────────────────────────────────────

// Portal pages
const HomePage = lazy(() => import('../pages/home/index'));
const WatchPage = lazy(() => import('../pages/home/Watch'));
const SearchPage = lazy(() => import('../pages/home/Search'));
const ChannelPage = lazy(() => import('../pages/home/Channel'));
const ProfilePage = lazy(() => import('../pages/home/Profile'));
const FeaturedPage = lazy(() => import('../pages/home/Featured'));
const LatestPage = lazy(() => import('../pages/home/Latest'));
const TagsPage = lazy(() => import('../pages/home/Tags'));
const MembersPage = lazy(() => import('../pages/home/Members'));
const AboutPage = lazy(() => import('../pages/home/About'));

// Auth pages
const SignInPage = lazy(() => import('../pages/auth/SignIn/index'));
const SignUpPage = lazy(() => import('../pages/auth/SignUp/index'));

// User center (protected)
const UploadPage = lazy(() => import('../pages/home/me/Upload'));
const FavoritesPage = lazy(() => import('../pages/home/me/Favorites'));
const NotificationsPage = lazy(() => import('../pages/home/me/Notifications'));
const HistoryPage = lazy(() => import('../pages/home/me/History'));
const PlaylistsPage = lazy(() => import('../pages/home/me/Playlists'));

// Admin pages
const AdminDashboard = lazy(() => import('../pages/admin/Dashboard'));
const AdminMedia = lazy(() => import('../pages/admin/Media'));
const AdminUsers = lazy(() => import('../pages/admin/Users'));
const AdminContent = lazy(() => import('../pages/admin/Content'));
const AdminCategories = lazy(() => import('../pages/admin/Categories'));
const AdminChannels = lazy(() => import('../pages/admin/Channels'));
const AdminTags = lazy(() => import('../pages/admin/Tags'));
const AdminComments = lazy(() => import('../pages/admin/Comments'));
const AdminPlaylists = lazy(() => import('../pages/admin/Playlists'));
const AdminSettings = lazy(() => import('../pages/admin/Settings'));

// ── Loading fallback ──────────────────────────────────────────────────────────

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh]">
        <div className="animate-spin w-8 h-8 border-3 border-emerald-600 border-t-transparent rounded-full"/>
    </div>
);

const Lazy = ({children}: { children: React.ReactNode }) => (
    <Suspense fallback={<PageLoader/>}>{children}</Suspense>
);

// ── Auth helpers ──────────────────────────────────────────────────────────────

function requireAuth() {
    const token = getStoredToken();
    if (!token) throw redirect({to: '/auth/signin'});
}

function requireAdmin() {
    const token = getStoredToken();
    if (!token) throw redirect({to: '/auth/signin'});
    const user = getStoredUser();
    if (!user?.roles?.includes('admin')) throw redirect({to: '/'});
}

function redirectIfAuth() {
    const token = getStoredToken();
    if (token) throw redirect({to: '/'});
}

// ── Route tree ────────────────────────────────────────────────────────────────

const rootRoute = createRootRoute({
    component: () => <Outlet/>,
});

// Portal layout wrapper
const portalLayoutRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'portal',
    component: () => <PortalLayout/>,
});

// ── Portal routes ──

const homeRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/',
    component: () => <Lazy><HomePage/></Lazy>,
});

const featuredRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/featured',
    component: () => <Lazy><FeaturedPage/></Lazy>,
});

const latestRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/latest',
    component: () => <Lazy><LatestPage/></Lazy>,
});

const categoriesRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/categories',
    component: () => <Lazy><SearchPage/></Lazy>,
});

const tagsRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/tags',
    component: () => <Lazy><TagsPage/></Lazy>,
});

const membersRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/members',
    component: () => <Lazy><MembersPage/></Lazy>,
});

const aboutRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/about',
    component: () => <Lazy><AboutPage/></Lazy>,
});

const watchRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/v/$id',
    component: () => <Lazy><WatchPage/></Lazy>,
});

const searchRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/search',
    validateSearch: (search: Record<string, unknown>) => search as { q?: string },
    component: () => <Lazy><SearchPage/></Lazy>,
});

const channelRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/c/$id',
    component: () => <Lazy><ChannelPage/></Lazy>,
});

const profileRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    path: '/u/$id',
    component: () => <Lazy><ProfilePage/></Lazy>,
});

// ── Auth routes ──

const signInRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/signin',
    beforeLoad: redirectIfAuth,
    component: () => <Lazy><SignInPage/></Lazy>,
});

const signUpRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/auth/signup',
    beforeLoad: redirectIfAuth,
    component: () => <Lazy><SignUpPage/></Lazy>,
});

// ── Protected user routes (/me/*) ──

const meLayoutRoute = createRoute({
    getParentRoute: () => portalLayoutRoute,
    id: 'me',
    beforeLoad: requireAuth,
});

const uploadRoute = createRoute({
    getParentRoute: () => meLayoutRoute,
    path: '/me/upload',
    component: () => <Lazy><UploadPage/></Lazy>,
});

const favoritesRoute = createRoute({
    getParentRoute: () => meLayoutRoute,
    path: '/me/favorites',
    component: () => <Lazy><FavoritesPage/></Lazy>,
});

const notificationsRoute = createRoute({
    getParentRoute: () => meLayoutRoute,
    path: '/me/notifications',
    component: () => <Lazy><NotificationsPage/></Lazy>,
});

const historyRoute = createRoute({
    getParentRoute: () => meLayoutRoute,
    path: '/me/history',
    component: () => <Lazy><HistoryPage/></Lazy>,
});

const playlistsRoute = createRoute({
    getParentRoute: () => meLayoutRoute,
    path: '/me/playlists',
    component: () => <Lazy><PlaylistsPage/></Lazy>,
});

// ── Admin routes ──

const adminLayoutRoute = createRoute({
    getParentRoute: () => rootRoute,
    id: 'admin-layout',
    beforeLoad: requireAdmin,
    component: () => <AdminLayout/>,
});

const adminIndexRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin',
    component: () => <Lazy><AdminDashboard/></Lazy>,
});

const adminMediaRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/media',
    component: () => <Lazy><AdminMedia/></Lazy>,
});

const adminUsersRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/users',
    component: () => <Lazy><AdminUsers/></Lazy>,
});

const adminContentRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/content',
    component: () => <Lazy><AdminContent/></Lazy>,
});

const adminCategoriesRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/categories',
    component: () => <Lazy><AdminCategories/></Lazy>,
});

const adminChannelsRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/channels',
    component: () => <Lazy><AdminChannels/></Lazy>,
});

const adminTagsRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/tags',
    component: () => <Lazy><AdminTags/></Lazy>,
});

const adminCommentsRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/comments',
    component: () => <Lazy><AdminComments/></Lazy>,
});

const adminPlaylistsRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/playlists',
    component: () => <Lazy><AdminPlaylists/></Lazy>,
});

const adminSettingsRoute = createRoute({
    getParentRoute: () => adminLayoutRoute,
    path: '/admin/settings',
    component: () => <Lazy><AdminSettings/></Lazy>,
});

// ── Router ────────────────────────────────────────────────────────────────────

const routeTree = rootRoute.addChildren([
    portalLayoutRoute.addChildren([
        homeRoute,
        featuredRoute,
        latestRoute,
        categoriesRoute,
        tagsRoute,
        membersRoute,
        aboutRoute,
        watchRoute,
        searchRoute,
        channelRoute,
        profileRoute,
        meLayoutRoute.addChildren([
            uploadRoute,
            favoritesRoute,
            notificationsRoute,
            historyRoute,
            playlistsRoute,
        ]),
    ]),
    signInRoute,
    signUpRoute,
    adminLayoutRoute.addChildren([
        adminIndexRoute,
        adminMediaRoute,
        adminUsersRoute,
        adminContentRoute,
        adminCategoriesRoute,
        adminChannelsRoute,
        adminTagsRoute,
        adminCommentsRoute,
        adminPlaylistsRoute,
        adminSettingsRoute,
    ]),
]);

export const router = createRouter({routeTree});

declare module '@tanstack/react-router' {
    interface Register {
        router: typeof router;
    }
}

const AppRouter: React.FC = () => <RouterProvider router={router}/>;

export default AppRouter;
