import { createFileRoute, redirect } from '@tanstack/react-router';
import AdminLayout from '@/layout/AdminLayout';
import { getStoredToken, getStoredUser } from '@/hooks/useAuth';

function requireAdmin() {
    const token = getStoredToken();
    if (!token) throw redirect({ to: '/auth/signin' });
    const user = getStoredUser();
    if (!user?.roles?.includes('admin')) throw redirect({ to: '/' });
}

export const Route = createFileRoute('/admin')({
    beforeLoad: requireAdmin,
    component: AdminLayout,
});
