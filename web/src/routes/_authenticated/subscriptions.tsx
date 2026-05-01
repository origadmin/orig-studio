import {Spinner} from "@/components/ui/spinner"
import {createFileRoute} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/home/Subscriptions'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

/**
 * Subscriptions page.
 * Auth is handled by _authenticated layout route, no beforeLoad needed.
 */
export const Route = createFileRoute('/_authenticated/subscriptions')({
    component: () => <Suspense fallback={<PageLoader/>}><Page/></Suspense>,
});
