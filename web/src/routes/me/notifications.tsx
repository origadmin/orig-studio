/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import NotificationsPage from '../../pages/home/me/Notifications';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/notifications',
  component: () => <NotificationsPage />,
});
