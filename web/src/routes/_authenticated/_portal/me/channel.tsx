import {Spinner} from "@/components/ui/spinner"
import {createFileRoute} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/home/UnifiedChannelPage'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

/**
 * User channel page.
 * Auth is handled by _authenticated layout route, no beforeLoad needed.
 * Portal layout (Header + Sidebar) provided by _authenticated/_portal layout.
 */
export const Route = createFileRoute('/_authenticated/_portal/me/channel')({
    component: () => <Suspense fallback={<PageLoader/>}><Page/></Suspense>,
});
