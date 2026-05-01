import {createFileRoute} from '@tanstack/react-router';
import PortalLayout from '@/layout/PortalLayout';

/**
 * Authenticated portal layout route.
 *
 * Provides PortalLayout (Header + Sidebar) for authenticated pages
 * that are part of the portal experience (subscriptions, me, media edit).
 *
 * This sits between _authenticated (auth guard) and the leaf routes,
 * ensuring portal-style pages get the full navigation chrome while
 * admin pages retain their own AdminLayout.
 */
export const Route = createFileRoute('/_authenticated/_portal')({
    component: PortalLayout,
});
