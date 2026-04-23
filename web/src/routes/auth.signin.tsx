import { createFileRoute, redirect } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';
import { getStoredToken } from '@/hooks/useAuth';

const Page = lazy(() => import('@/pages/auth/SignIn/index'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <div className="animate-spin w-8 h-8 border-3 border-emerald-600 border-t-transparent rounded-full"/>
    </div>
);

function redirectIfAuth() {
    const token = getStoredToken();
    if (token) throw redirect({ to: '/' });
}

export const Route = createFileRoute('/auth/signin')({
    beforeLoad: redirectIfAuth,
    component: () => <Suspense fallback={<PageLoader />}><Page /></Suspense>,
});
