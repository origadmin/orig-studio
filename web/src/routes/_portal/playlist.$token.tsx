import {Spinner} from "@/components/ui/spinner"
import { createFileRoute } from '@tanstack/react-router';
import { lazy, Suspense } from 'react';

// Portal playlist detail route uses short_token (not database id) per A005 design principle.
// Consistent with /c/$token (ChannelHandler) and /watch?v={short_token} (MediaHandler).
// This displays a single playlist with its videos, NOT the user's playlist list.
const Page = lazy(() => import('@/pages/home/PlaylistDetail'));

const PageLoader = () => (
    <div className="flex items-center justify-center min-h-[60vh] bg-background text-foreground">
        <Spinner />
    </div>
);

export const Route = createFileRoute('/_portal/playlist/$token')({
    component: () => <Suspense fallback={<PageLoader />}><Page /></Suspense>,
});
