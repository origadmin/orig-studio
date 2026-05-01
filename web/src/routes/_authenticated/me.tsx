import {createFileRoute, Outlet} from '@tanstack/react-router';

/**
 * Personal center layout route.
 *
 * _authenticated already ensures the user is authenticated,
 * so no additional auth check is needed here.
 * Renders Outlet for child routes.
 *
 * Replaces the previous _portal/me/route.tsx.
 */
export const Route = createFileRoute('/_authenticated/me')({
    component: () => <Outlet/>,
});
