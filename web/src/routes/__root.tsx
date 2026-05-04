import {createRootRouteWithContext, Outlet} from '@tanstack/react-router';
import type {AuthContextValue} from '@/contexts/auth/types';
import NotFoundPage from '@/pages/NotFoundPage';

export interface RouterContext {
    auth: AuthContextValue;
}

export const Route = createRootRouteWithContext<RouterContext>()({
    component: () => <Outlet/>,
    notFoundComponent: NotFoundPage,
});
