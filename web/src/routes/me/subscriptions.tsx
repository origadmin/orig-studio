/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import SubscriptionsPage from '../../pages/home/Subscriptions';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/subscriptions',
  component: () => <SubscriptionsPage />,
});
