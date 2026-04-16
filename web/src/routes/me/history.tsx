/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as MeLayoutRoute } from './_layout';
import HistoryPage from '../../pages/home/me/History';

export const Route = createRoute({
  getParentRoute: () => MeLayoutRoute,
  path: '/me/history',
  component: () => <HistoryPage />,
});
