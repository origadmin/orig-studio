import {createFileRoute, redirect, Outlet} from '@tanstack/react-router';

/**
 * Authenticated layout route.
 *
 * Unified entry point for all routes requiring authentication.
 * - Unauthenticated users are redirected to /auth/signin with the
 *   original target URL preserved in the redirect search param.
 * - Authenticated users proceed to child routes normally.
 *
 * This replaces the scattered requireAuth() function definitions
 * that previously existed in multiple route files (P3 fix).
 */
export const Route = createFileRoute('/_authenticated')({
    beforeLoad: ({context, location}) => {
        if (!context.auth.isAuthenticated) {
            throw redirect({
                to: '/auth/signin',
                search: {redirect: location.href},
            });
        }
    },
    component: () => <Outlet/>,
});
