import {Spinner} from "@/components/ui/spinner"
import {createFileRoute} from '@tanstack/react-router';
import {lazy, Suspense} from 'react';

const Page = lazy(() => import('@/pages/portal/ArticleList'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[80vh] bg-background text-foreground">
        <Spinner/>
    </div>
);

export const Route = createFileRoute('/_authenticated/_portal/articles/')({
    component: () => <Suspense fallback={<PageLoader/>}><Page/></Suspense>,
});
