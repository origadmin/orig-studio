import { createFileRoute, redirect } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';
import { getStoredToken } from '@/hooks/useAuth';

const Page = lazy(() => import('@/pages/home/Subscriptions'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <div className="animate-spin w-8 h-8 border-3 border-emerald-600 border-t-transparent rounded-full"/>
    </div>
);

function requireAuth() {
    const token = getStoredToken();
    if (!token) throw redirect({ to: '/auth/signin' });
}

export const Route = createFileRoute('/_portal/subscriptions')({
    beforeLoad: requireAuth,
    component: () => <Suspense fallback={<PageLoader />}><Page /></Suspense>,
});
