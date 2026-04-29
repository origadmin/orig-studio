import { createFileRoute } from '@tanstack/react-router';
import PortalLayout from '@/layout/PortalLayout';

export const Route = createFileRoute('/_portal')({
    component: PortalLayout,
});
