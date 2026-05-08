import {Spinner} from "@/components/ui/spinner"
import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

const Page = lazy(() => import('@/pages/home/Search'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

export const Route = createFileRoute('/_portal/search')({
    validateSearch: (search: Record<string, unknown>) => ({
        q: typeof search.q === 'string' ? search.q : undefined,
        tag: typeof search.tag === 'string' ? search.tag : undefined,
        category_id: typeof search.category_id === 'string' ? search.category_id : undefined,
    } as { q?: string; tag?: string; category_id?: string }),
    component: () => <Suspense fallback={<PageLoader />}><Page /></Suspense>,
});
