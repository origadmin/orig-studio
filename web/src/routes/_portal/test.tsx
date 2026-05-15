import {Spinner} from "@/components/ui/spinner"
import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const Page = lazy(() => import('@/pages/test/index'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

export const Route = createFileRoute('/_portal/test')({
    component: () => <Suspense fallback={<PageLoader />}><Page /></Suspense>,
});
