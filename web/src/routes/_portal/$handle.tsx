import {Spinner} from "@/components/ui/spinner"
import { createFileRoute, notFound } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const UnifiedChannelPage = lazy(() => import('@/pages/home/UnifiedChannelPage'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

export const Route = createFileRoute('/_portal/$handle')({
    beforeLoad: ({ params }) => {
        if (!params.handle.startsWith('@')) {
            throw notFound();
        }
    },
    component: () => (
        <Suspense fallback={<PageLoader />}>
            <UnifiedChannelPage />
        </Suspense>
    ),
});
