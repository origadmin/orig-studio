/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute, redirect } from '@tanstack/react-router';
import { Outlet } from '@tanstack/react-router';
import { Route as RootRoute } from '../__root';
import AdminLayout from '../../layout/AdminLayout';
import { getStoredToken, getStoredUser } from '../../hooks/useAuth';

function requireAdmin() {
  const token = getStoredToken();
  if (!token) throw redirect({ to: '/auth/signin' });
  const user = getStoredUser();
  if (!user?.roles?.includes('admin')) throw redirect({ to: '/' });
}

export const Route = createRoute({
  getParentRoute: () => RootRoute,
  id: 'admin-layout',
  beforeLoad: requireAdmin,
  component: () => <AdminLayout />,
});
