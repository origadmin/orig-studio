import {Spinner} from "@/components/ui/spinner"
import {createFileRoute} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/home/MediaEdit'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

/**
 * Media edit page.
 * Now protected by _authenticated layout route (P4 security fix).
 * Portal layout (Header + Sidebar) provided by _authenticated/_portal layout.
 */
export const Route = createFileRoute('/_authenticated/_portal/media/$shortToken/edit')({
    component: () => <Suspense fallback={<PageLoader/>}><Page/></Suspense>,
});
