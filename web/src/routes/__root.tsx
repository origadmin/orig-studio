import {createRootRouteWithContext, Outlet} from '@tanstack/react-router';
import type {AuthContextValue} from '@/contexts/auth/types';
import NotFoundPage from '@/pages/NotFoundPage';

/** Router context type definition - injects auth state into route guards */
export interface RouterContext {
    auth: AuthContextValue;
}

export const Route = createRootRouteWithContext<RouterContext>()({
    component: () => <Outlet/>,
    notFoundComponent: NotFoundPage,
});
