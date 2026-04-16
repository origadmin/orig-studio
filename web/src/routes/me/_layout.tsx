/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute, redirect } from '@tanstack/react-router';
import { Outlet } from '@tanstack/react-router';
import { Route as LayoutRoute } from '../_layout';
import { getStoredToken } from '../../hooks/useAuth';

function requireAuth() {
  const token = getStoredToken();
  if (!token) throw redirect({ to: '/auth/signin' });
}

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  id: 'me',
  beforeLoad: requireAuth,
  component: () => <Outlet />,
});
