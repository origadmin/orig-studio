import {Spinner} from "@/components/ui/spinner"
import {createFileRoute, redirect} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/auth/SignUp/index'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

/**
 * Sign-up route.
 *
 * Redirects authenticated users to the home page.
 */
export const Route = createFileRoute('/auth/signup')({
    beforeLoad: ({context}) => {
        if (context.auth.isAuthenticated) {
            throw redirect({to: '/'});
        }
    },
    component: () => <Suspense fallback={<PageLoader/>}><Page/></Suspense>,
});
