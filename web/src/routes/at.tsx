/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import ProfilePage from '../pages/home/Profile';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/@$username',
  validateParams: (params: Record<string, unknown>) => {
    return {
      username: String(params.username),
    };
  },
  component: () => <ProfilePage />,
});