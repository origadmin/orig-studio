import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const HomePage = lazy(() => import('@/pages/home/index'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <div className="animate-spin w-8 h-8 border-3 border-emerald-600 border-t-transparent rounded-full"/>
    </div>
);

export const Route = createFileRoute('/_portal/')({
    component: () => (
        <Suspense fallback={<PageLoader />}>
            <HomePage />
        </Suspense>
    ),
});
