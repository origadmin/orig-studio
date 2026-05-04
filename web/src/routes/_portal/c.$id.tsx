import {Spinner} from "@/components/ui/spinner"
import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const ChannelPage = lazy(() => import('@/pages/home/UnifiedChannelPage'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

/**
 * /c/{token} route: Channel page (RESTful, recommended)
 *
 * Uses short_token as the path parameter.
 * This is the primary entry point for viewing channels.
 */
export const Route = createFileRoute('/_portal/c/$id')({
    component: () => <Suspense fallback={<PageLoader />}><ChannelPage /></Suspense>,
});
