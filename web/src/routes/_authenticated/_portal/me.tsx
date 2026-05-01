import {createFileRoute, Outlet} from '@tanstack/react-router';

/**
 * Personal center layout route.
 *
 * _authenticated already ensures the user is authenticated,
 * so no additional auth check is needed here.
 * Renders Outlet for child routes.
 *
 * Portal layout (Header + Sidebar) provided by _authenticated/_portal layout.
 */
export const Route = createFileRoute('/_authenticated/_portal/me')({
    component: () => <Outlet/>,
});
