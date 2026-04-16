/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import LatestPage from '../pages/home/Latest';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/latest',
  component: () => <LatestPage />,
});
