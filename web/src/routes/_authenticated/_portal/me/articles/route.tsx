import {createFileRoute, Outlet} from '@tanstack/react-router';

export const Route = createFileRoute('/_authenticated/_portal/me/articles')({
    component: () => <Outlet/>,
});
