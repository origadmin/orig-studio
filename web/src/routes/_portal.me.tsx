import { createFileRoute, redirect, Outlet } from '@tanstack/react-router';
import { getStoredToken } from '@/hooks/useAuth';

function requireAuth() {
    const token = getStoredToken();
    if (!token) throw redirect({ to: '/auth/signin' });
}

export const Route = createFileRoute('/_portal/me')({
    beforeLoad: requireAuth,
    component: () => <Outlet />,
});
