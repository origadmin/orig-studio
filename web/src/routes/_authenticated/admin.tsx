import {createFileRoute, redirect} from '@tanstack/react-router';
import AdminLayout from '@/layout/AdminLayout';

/**
 * Admin layout route.
 *
 * - _authenticated already ensures the user is authenticated.
 * - This route additionally checks for the admin role.
 * - Non-admin users are redirected to the home page.
 *
 * Replaces the previous requireAdmin() function in admin/route.tsx.
 */
export const Route = createFileRoute('/_authenticated/admin')({
    beforeLoad: ({context}) => {
        if (!context.auth.isAdmin) {
            throw redirect({to: '/'});
        }
    },
    component: AdminLayout,
});
