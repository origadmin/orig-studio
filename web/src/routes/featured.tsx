/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

import { createRoute } from '@tanstack/react-router';
import { Route as LayoutRoute } from './_layout';
import FeaturedPage from '../pages/home/Featured';

export const Route = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/featured',
  component: () => <FeaturedPage />,
});
