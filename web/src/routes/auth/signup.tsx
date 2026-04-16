/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute, redirect } from '@tanstack/react-router';
import { Route as RootRoute } from '../__root';
import SignUpPage from '../../pages/auth/SignUp/index';
import { getStoredToken } from '../../hooks/useAuth';

function redirectIfAuth() {
  const token = getStoredToken();
  if (token) throw redirect({ to: '/' });
}

export const Route = createRoute({
  getParentRoute: () => RootRoute,
  path: '/auth/signup',
  beforeLoad: redirectIfAuth,
  component: () => <SignUpPage />,
});
