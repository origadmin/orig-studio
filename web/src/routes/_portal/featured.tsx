import {Spinner} from "@/components/ui/spinner"
import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const FeaturedPage = lazy(() => import('@/pages/home/Featured'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

export const Route = createFileRoute('/_portal/featured')({
    component: () => (
        <Suspense fallback={<PageLoader />}>
            <FeaturedPage />
        </Suspense>
    ),
});
