import {Spinner} from "@/components/ui/spinner"
import {createFileRoute} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/admin/ArticleEdit'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

export const Route = createFileRoute('/_authenticated/admin/articles/new')({
    component: () => (
        <Suspense fallback={<PageLoader/>}>
            <Page mode="create"/>
        </Suspense>
    ),
});
