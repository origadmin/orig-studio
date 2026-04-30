import {Spinner} from "@/components/ui/spinner"
import {createFileRoute, redirect} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/auth/SignIn/index'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

/**
 * Sign-in route.
 *
 * Redirects authenticated users to the home page.
 * Supports a redirect search param to return to the original target
 * after successful login.
 */
export const Route = createFileRoute('/auth/signin')({
    validateSearch: (search: Record<string, unknown>): { redirect?: string } => ({
        redirect: typeof search.redirect === 'string' ? search.redirect : undefined,
    }),
    beforeLoad: ({context}) => {
        if (context.auth.isAuthenticated) {
            throw redirect({to: '/'});
        }
    },
    component: () => <Suspense fallback={<PageLoader/>}><Page/></Suspense>,
});
