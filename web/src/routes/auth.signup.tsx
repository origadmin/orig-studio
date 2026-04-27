import {Spinner} from "@/components/ui/spinner"
import { createFileRoute, redirect } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';
import { getStoredToken } from '@/hooks/useAuth';

const Page = lazy(() => import('@/pages/auth/SignUp/index'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

function redirectIfAuth() {
    const token = getStoredToken();
    if (token) throw redirect({ to: '/' });
}

export const Route = createFileRoute('/auth/signup')({
    beforeLoad: redirectIfAuth,
    component: () => <Suspense fallback={<PageLoader />}><Page /></Suspense>,
});
