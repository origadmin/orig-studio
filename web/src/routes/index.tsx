/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import HomePage from '../pages/home/index';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/',
  component: () => <HomePage />,
});
